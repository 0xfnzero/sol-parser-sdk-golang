package solparser

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mr-tron/base58"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
)

// tlsConfigForGRPCEndpoint 为 gRPC over TLS 设置 SNI（ServerName）。空 tls.Config 在部分环境下会导致握手阶段 EOF。
func tlsConfigForGRPCEndpoint(endpoint string) *tls.Config {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		cfg.ServerName = endpoint
		return cfg
	}
	cfg.ServerName = host
	return cfg
}

// normalizeGRPCEndpoint accepts the same http(s) GRPC_URL form as other languages.
// Never include the original URL in errors: it can contain credentials.
func normalizeGRPCEndpoint(endpoint string, configuredTLS bool) (string, bool, error) {
	if !strings.Contains(endpoint, "://") {
		return endpoint, configuredTLS, nil
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", false, fmt.Errorf("invalid gRPC endpoint; use host:port or http(s)://host:port")
	}
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return net.JoinHostPort(parsed.Hostname(), port), parsed.Scheme == "https", nil
}

// SubscribeCallbacks 订阅回调函数
type SubscribeCallbacks struct {
	OnUpdate func(update *SubscribeUpdate)
	OnError  func(err error)
	OnEnd    func()
}

// Subscription 订阅句柄
type Subscription struct {
	ID        string
	Filter    *TransactionFilter
	Cancel    context.CancelFunc
	callbacks SubscribeCallbacks
}

// GrpcStreamStatus reports lifecycle and sticky loss of continuity.
type GrpcStreamStatus struct {
	State            string
	ContinuityBroken bool
	Reconnects       uint64
	Dropped          uint64
}

// DexEventSubscription 直接产出解析后的 DexEvent。
// Events/Errors 使用有界缓冲；缓冲满时丢弃新消息，避免阻塞 gRPC 读循环。
type DexEventSubscription struct {
	ID     string
	Events <-chan DexEvent
	Errors <-chan error
	Cancel func()
	States <-chan GrpcStreamStatus
	Status func() GrpcStreamStatus
	Join   func()
}

// YellowstoneGrpc Yellowstone gRPC 客户端
type YellowstoneGrpc struct {
	endpoint     string
	config       ClientConfig
	xToken       string
	ctx          context.Context
	cancel       context.CancelFunc
	conn         *grpc.ClientConn
	client       pb.GeyserClient
	stream       pb.Geyser_SubscribeClient
	dexLifecycle sync.Mutex
	dexCancel    context.CancelFunc
	dexDone      chan struct{}
	dexControl   chan *pb.SubscribeRequest
	dexFilter    EventTypeFilter
	mu           sync.RWMutex
	connected    bool
	subscribers  map[string]*Subscription
}

// NewYellowstoneGrpc 创建新的 Yellowstone gRPC 客户端
func NewYellowstoneGrpc(endpoint string, config ...ClientConfig) *YellowstoneGrpc {
	cfg := DefaultClientConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &YellowstoneGrpc{
		endpoint:    endpoint,
		config:      cfg,
		ctx:         ctx,
		cancel:      cancel,
		subscribers: make(map[string]*Subscription),
	}
}

// SetXToken 设置 X-Token 认证
func (c *YellowstoneGrpc) SetXToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.xToken = token
}

// xTokenAuth 返回 x-token 认证拦截器
func (c *YellowstoneGrpc) xTokenAuth() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if c.xToken != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-token", c.xToken)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}

// streamXTokenAuth 返回流式 x-token 认证拦截器
func (c *YellowstoneGrpc) streamXTokenAuth() grpc.DialOption {
	return grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if c.xToken != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-token", c.xToken)
		}
		return streamer(ctx, desc, cc, method, opts...)
	})
}

// Connect 连接到 gRPC 服务器
//
// 参考实现:
// - https://github.com/rpcpool/yellowstone-grpc/examples/golang
// - https://github.com/ChainBuff/yellowstone-grpc
func (c *YellowstoneGrpc) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}
	if c.ctx == nil || c.ctx.Err() != nil {
		c.ctx, c.cancel = context.WithCancel(context.Background())
	}

	endpoint, enableTLS, err := normalizeGRPCEndpoint(c.endpoint, c.config.EnableTLS)
	if err != nil {
		return err
	}

	// 配置 keepalive 参数
	// 参考: yellowstone-grpc-golang 示例
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // 每 10 秒发送一次 ping
		Timeout:             time.Second,      // ping 超时时间为 1 秒
		PermitWithoutStream: true,             // 即使没有活动的流也发送 ping
	}

	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(kacp),
		c.xTokenAuth(),
		c.streamXTokenAuth(),
	}

	// 配置 TLS（显式 SNI，避免 publicnode 等域名出现 handshake EOF）
	if enableTLS {
		creds := credentials.NewTLS(tlsConfigForGRPCEndpoint(endpoint))
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Duration(c.config.ConnectionTimeoutMs)*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.client = pb.NewGeyserClient(conn)
	c.connected = true

	return nil
}

// Disconnect 断开连接
func (c *YellowstoneGrpc) Disconnect() error {
	c.dexLifecycle.Lock()
	defer c.dexLifecycle.Unlock()
	c.stopDexUnlocked()
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	// 取消所有订阅
	for _, sub := range c.subscribers {
		if sub.Cancel != nil {
			sub.Cancel()
		}
	}
	c.subscribers = make(map[string]*Subscription)

	// 取消主上下文
	if c.cancel != nil {
		c.cancel()
	}

	// 关闭连接
	if c.conn != nil {
		c.conn.Close()
	}

	c.connected = false
	c.client = nil
	c.conn = nil
	return nil
}

// SubscribeTransactions 订阅交易
func (c *YellowstoneGrpc) SubscribeTransactions(filter TransactionFilter, callbacks SubscribeCallbacks) (*Subscription, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected, call Connect() first")
	}
	client := c.client
	c.mu.RUnlock()

	// 创建订阅上下文
	subCtx, subCancel := context.WithCancel(c.ctx)

	sub := &Subscription{
		ID:        generateSubID(),
		Filter:    &filter,
		Cancel:    subCancel,
		callbacks: callbacks,
	}

	c.mu.Lock()
	c.subscribers[sub.ID] = sub
	c.mu.Unlock()

	// 打开双向流
	stream, err := client.Subscribe(subCtx)
	if err != nil {
		subCancel()
		c.mu.Lock()
		delete(c.subscribers, sub.ID)
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	// 构建订阅请求
	req := c.buildSubscribeRequest(filter)

	// 发送初始请求
	if err := stream.Send(req); err != nil {
		subCancel()
		c.mu.Lock()
		delete(c.subscribers, sub.ID)
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to send subscribe request: %w", err)
	}

	// 启动处理 goroutine
	go c.handleStream(subCtx, stream, sub)

	return sub, nil
}

// SubscribeDexEvents 订阅交易/账户更新并直接产出 DexEvent。
// 解析路径与 Rust gRPC 订阅对齐：交易走指令 + 日志 + 字段填充，账户更新走 ParseAccountUnified。
func (c *YellowstoneGrpc) SubscribeDexEvents(transactionFilters []TransactionFilter, accountFilters []AccountFilter, filter EventTypeFilter) (*DexEventSubscription, error) {
	c.dexLifecycle.Lock()
	defer c.dexLifecycle.Unlock()
	c.stopDexUnlocked()
	if err := c.Connect(); err != nil {
		return nil, err
	}

	bufferSize := c.config.BufferSize
	if bufferSize <= 0 {
		bufferSize = 8192
	}

	eventCh := make(chan DexEvent, bufferSize)
	errCh := make(chan error, bufferSize)
	stateCh := make(chan GrpcStreamStatus, bufferSize)
	statusMu := sync.RWMutex{}
	status := GrpcStreamStatus{State: "connecting"}
	stateCh <- status
	statusClosed := false
	transition := func(state string, broken bool, drop bool) {
		statusMu.Lock()
		defer statusMu.Unlock()
		if statusClosed {
			return
		}
		previous := status
		if state != "" {
			status.State = state
		}
		status.ContinuityBroken = status.ContinuityBroken || broken
		if drop {
			status.Dropped++
		}
		if state == "reconnecting" && previous.State != "reconnecting" {
			status.Reconnects++
		}
		if status == previous {
			return
		}
		select {
		case stateCh <- status:
		default:
		}
	}
	getStatus := func() GrpcStreamStatus { statusMu.RLock(); defer statusMu.RUnlock(); return status }
	closeStatus := func() {
		transition("stopped", true, false)
		statusMu.Lock()
		defer statusMu.Unlock()
		if !statusClosed {
			statusClosed = true
			close(stateCh)
		}
	}
	ctx, cancelCtx := context.WithCancel(c.ctx)
	done := make(chan struct{})
	c.dexCancel = cancelCtx
	c.dexDone = done
	dexID := generateSubID()
	controlCh := make(chan *pb.SubscribeRequest, 100)
	reqMu := sync.RWMutex{}
	currentReq := c.buildSubscribeRequestMulti(transactionFilters, accountFilters)
	if filter != nil && filter.ShouldInclude(EventTypeBlockMeta) {
		currentReq.BlocksMeta = map[string]*pb.SubscribeRequestFilterBlocksMeta{"block_meta": {}}
	}

	var closeOnce sync.Once
	var sendMu sync.RWMutex
	closed := false
	closeChannels := func() {
		closeOnce.Do(func() {
			sendMu.Lock()
			closed = true
			close(eventCh)
			close(errCh)
			closeStatus()
			sendMu.Unlock()
		})
	}
	sendError := func(err error) {
		if err == nil {
			return
		}
		sendMu.RLock()
		defer sendMu.RUnlock()
		if closed {
			return
		}
		select {
		case errCh <- err:
		default:
		}
	}
	sendEvent := func(event DexEvent) {
		if event.Type == "" {
			return
		}
		sendMu.RLock()
		defer sendMu.RUnlock()
		if closed {
			return
		}
		select {
		case eventCh <- event:
		default:
			transition("", true, true)
		}
	}

	var cancelOnce sync.Once
	cancel := func() {
		cancelOnce.Do(func() {
			cancelCtx()
			c.mu.Lock()
			delete(c.subscribers, dexID)
			c.mu.Unlock()
			closeChannels()
		})
	}

	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		cancel()
		close(done)
		return nil, fmt.Errorf("client not connected")
	}

	orderDispatcher := newDexOrderDispatcher(c.config)
	orderDispatcher.start(ctx, sendEvent)

	c.mu.Lock()
	c.subscribers[dexID] = &Subscription{ID: dexID, Cancel: cancelCtx}
	c.dexControl = controlCh
	c.dexFilter = filter
	c.mu.Unlock()

	go func() {
		defer close(done)
		defer func() {
			orderDispatcher.stop()
			orderDispatcher.flushAll(sendEvent)
			c.mu.Lock()
			delete(c.subscribers, dexID)
			if c.dexControl == controlCh {
				c.dexControl = nil
			}
			c.mu.Unlock()
			closeChannels()
		}()

		backoff := time.Duration(c.config.RetryDelayMs) * time.Millisecond
		if backoff <= 0 {
			backoff = time.Second
		}
		const maxBackoff = 60 * time.Second

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			for {
				select {
				case nextReq := <-controlCh:
					reqMu.Lock()
					currentReq = nextReq
					reqMu.Unlock()
				default:
					goto drainedControl
				}
			}
		drainedControl:

			streamCtx, cancelStream := context.WithCancel(ctx)
			stream, err := client.Subscribe(streamCtx)
			if err != nil {
				cancelStream()
				transition("reconnecting", true, false)
				if !sleepBeforeReconnect(ctx, backoff, sendError, err) {
					return
				}
				backoff = minDuration(backoff*2, maxBackoff)
				continue
			}

			sendMu := sync.Mutex{}
			reqMu.RLock()
			req := currentReq
			reqMu.RUnlock()
			sendMu.Lock()
			err = stream.Send(req)
			sendMu.Unlock()
			if err != nil {
				cancelStream()
				transition("reconnecting", true, false)
				if !sleepBeforeReconnect(ctx, backoff, sendError, err) {
					return
				}
				backoff = minDuration(backoff*2, maxBackoff)
				continue
			}

			backoff = time.Duration(c.config.RetryDelayMs) * time.Millisecond
			if backoff <= 0 {
				backoff = time.Second
			}
			transition("connected", false, false)
			streamDone := make(chan struct{})
			writerDone := make(chan struct{})
			var streamDoneOnce sync.Once
			closeStreamDone := func() {
				streamDoneOnce.Do(func() { close(streamDone); cancelStream() })
				<-writerDone
			}
			go func() {
				defer close(writerDone)
				for {
					select {
					case <-ctx.Done():
						return
					case <-streamDone:
						return
					case nextReq := <-controlCh:
						reqMu.Lock()
						currentReq = nextReq
						reqMu.Unlock()
						sendMu.Lock()
						err := stream.Send(nextReq)
						sendMu.Unlock()
						if err != nil {
							cancelStream()
							sendError(err)
							return
						}
					}
				}
			}()

			for {
				resp, err := stream.Recv()
				if err != nil {
					if ctx.Err() == nil {
						transition("reconnecting", true, false)
					}
					closeStreamDone()
					if !sleepBeforeReconnect(ctx, backoff, sendError, err) {
						return
					}
					backoff = minDuration(backoff*2, maxBackoff)
					break
				}

				if resp.GetPing() != nil {
					sendMu.Lock()
					err := stream.Send(&pb.SubscribeRequest{
						Ping: &pb.SubscribeRequestPing{Id: 1},
					})
					sendMu.Unlock()
					if err != nil {
						sendError(err)
						transition("reconnecting", true, false)
						closeStreamDone()
						break
					}
					continue
				}

				grpcRecvUs := NowUs()
				update := c.convertSubscribeUpdate(resp)
				blockTimeUs := update.CreatedAt
				if update.Transaction != nil && update.Transaction.Transaction != nil {
					events, perr := ParseSubscribeTransactionWithBlockTime(
						update.Transaction.Slot,
						update.Transaction.Transaction,
						filter,
						grpcRecvUs,
						blockTimeUs,
					)
					if perr != nil {
						sendError(perr)
					}
					orderDispatcher.pushTransactionEvents(events, update.Transaction.Slot, update.Transaction.Transaction.Index, sendEvent)
				}
				if update.BlockMeta != nil && (filter == nil || filter.ShouldInclude(EventTypeBlockMeta)) {
					sendEvent(ParseBlockMetaUpdate(update.BlockMeta, grpcRecvUs, blockTimeUs))
				}
				if update.Account != nil {
					if inc, ok := filter.(*IncludeOnlyFilter); ok && eventTypeSliceContains(inc.IncludeOnly, EventTypeAccountRawSnapshot) {
						sendEvent(parseAccountDexEvent(update.Account, filter, grpcRecvUs, blockTimeUs, true))
					}

					sendEvent(parseAccountDexEvent(update.Account, filter, grpcRecvUs, blockTimeUs))
				}

				select {
				case <-ctx.Done():
					closeStreamDone()
					return
				default:
				}
			}
			closeStreamDone()
		}
	}()

	return &DexEventSubscription{
		ID:     dexID,
		Events: eventCh,
		Errors: errCh,
		Cancel: cancel,
		States: stateCh, Status: getStatus, Join: func() { <-done },
	}, nil
}

// UpdateSubscription 动态更新当前 DEX 订阅过滤器（与 Rust update_subscription 对齐）。
func (c *YellowstoneGrpc) UpdateSubscription(transactionFilters []TransactionFilter, accountFilters []AccountFilter) error {
	c.dexLifecycle.Lock()
	defer c.dexLifecycle.Unlock()
	req := c.buildSubscribeRequestMulti(transactionFilters, accountFilters)
	c.mu.RLock()
	controlCh := c.dexControl
	filter := c.dexFilter
	c.mu.RUnlock()
	if filter != nil && filter.ShouldInclude(EventTypeBlockMeta) {
		req.BlocksMeta = map[string]*pb.SubscribeRequestFilterBlocksMeta{"block_meta": {}}
	}
	if controlCh == nil {
		return fmt.Errorf("no active DEX subscription")
	}
	select {
	case controlCh <- req:
		return nil
	default:
		return fmt.Errorf("subscription control channel is full")
	}
}

func sleepBeforeReconnect(ctx context.Context, backoff time.Duration, sendError func(error), err error) bool {
	if ctx.Err() != nil {
		return false
	}
	sendError(err)
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func parseAccountDexEvent(update *SubscribeUpdateAccount, filter EventTypeFilter, grpcRecvUs int64, blockTimeUs *int64, rawSnapshot ...bool) DexEvent {
	if update == nil || update.Account == nil {
		return DexEvent{}
	}
	acc := update.Account
	signature := ""
	if len(acc.TxnSignature) > 0 {
		signature = base58.Encode(acc.TxnSignature)
	}
	account := &AccountData{
		Pubkey:     base58.Encode(acc.Pubkey),
		Executable: acc.Executable,
		Lamports:   acc.Lamports,
		Owner:      base58.Encode(acc.Owner),
		RentEpoch:  acc.RentEpoch,
		Data:       acc.Data,
	}
	meta := makeMetadata(signature, update.Slot, 0, blockTimeUs, grpcRecvUs, "")

	if len(rawSnapshot) > 0 && rawSnapshot[0] {
		account.Data = append([]byte{}, account.Data...)
		return DexEvent{Type: EventTypeAccountRawSnapshot, Data: &RawAccountSnapshotEvent{meta, *account, acc.WriteVersion, update.IsStartup}}
	}
	return ParseAccountUnified(account, meta, filter)
}

// buildSubscribeRequest 构建订阅请求
func (c *YellowstoneGrpc) buildSubscribeRequest(filter TransactionFilter) *pb.SubscribeRequest {
	req := c.buildSubscribeRequestMulti(nil, nil)
	req.Transactions["client"] = transactionFilterToProto(filter)
	return req
}

func (c *YellowstoneGrpc) buildSubscribeRequestMulti(transactionFilters []TransactionFilter, accountFilters []AccountFilter) *pb.SubscribeRequest {
	commitment := pb.CommitmentLevel_PROCESSED
	req := &pb.SubscribeRequest{
		Accounts:           map[string]*pb.SubscribeRequestFilterAccounts{},
		Slots:              map[string]*pb.SubscribeRequestFilterSlots{},
		Transactions:       map[string]*pb.SubscribeRequestFilterTransactions{},
		TransactionsStatus: map[string]*pb.SubscribeRequestFilterTransactions{},
		Blocks:             map[string]*pb.SubscribeRequestFilterBlocks{},
		BlocksMeta:         map[string]*pb.SubscribeRequestFilterBlocksMeta{},
		Entry:              map[string]*pb.SubscribeRequestFilterEntry{},
		Commitment:         &commitment,
		AccountsDataSlice:  []*pb.SubscribeRequestAccountsDataSlice{},
	}
	for i, filter := range transactionFilters {
		req.Transactions[fmt.Sprintf("tx_%d", i)] = transactionFilterToProto(filter)
	}
	for i, filter := range accountFilters {
		req.Accounts[fmt.Sprintf("acc_%d", i)] = accountFilterToProto(filter)
	}
	return req
}

func transactionFilterToProto(filter TransactionFilter) *pb.SubscribeRequestFilterTransactions {
	out := &pb.SubscribeRequestFilterTransactions{
		AccountInclude:  filter.AccountInclude,
		AccountExclude:  filter.AccountExclude,
		AccountRequired: filter.AccountRequired,
	}
	if filter.Vote != nil {
		out.Vote = filter.Vote
	}
	if filter.Failed != nil {
		out.Failed = filter.Failed
	}
	if filter.Signature != "" {
		sig := filter.Signature
		out.Signature = &sig
	}
	return out
}

func accountFilterToProto(filter AccountFilter) *pb.SubscribeRequestFilterAccounts {
	out := &pb.SubscribeRequestFilterAccounts{
		Account: filter.Account,
		Owner:   filter.Owner,
		Filters: make([]*pb.SubscribeRequestFilterAccountsFilter, 0, len(filter.Filters)),
	}
	for _, item := range filter.Filters {
		if item == nil {
			continue
		}
		out.Filters = append(out.Filters, accountFilterItemToProto(item))
	}
	return out
}

func accountFilterItemToProto(filter *SubscribeRequestFilterAccountsFilter) *pb.SubscribeRequestFilterAccountsFilter {
	switch {
	case filter.Memcmp != nil:
		memcmp := &pb.SubscribeRequestFilterAccountsFilterMemcmp{
			Offset: filter.Memcmp.Offset,
		}
		switch {
		case len(filter.Memcmp.Bytes) > 0:
			memcmp.Data = &pb.SubscribeRequestFilterAccountsFilterMemcmp_Bytes{Bytes: filter.Memcmp.Bytes}
		case filter.Memcmp.Base58 != "":
			memcmp.Data = &pb.SubscribeRequestFilterAccountsFilterMemcmp_Base58{Base58: filter.Memcmp.Base58}
		case filter.Memcmp.Base64 != "":
			memcmp.Data = &pb.SubscribeRequestFilterAccountsFilterMemcmp_Base64{Base64: filter.Memcmp.Base64}
		}
		return &pb.SubscribeRequestFilterAccountsFilter{
			Filter: &pb.SubscribeRequestFilterAccountsFilter_Memcmp{Memcmp: memcmp},
		}
	case filter.Datasize != nil:
		return &pb.SubscribeRequestFilterAccountsFilter{
			Filter: &pb.SubscribeRequestFilterAccountsFilter_Datasize{Datasize: *filter.Datasize},
		}
	case filter.TokenAccountState != nil:
		return &pb.SubscribeRequestFilterAccountsFilter{
			Filter: &pb.SubscribeRequestFilterAccountsFilter_TokenAccountState{TokenAccountState: *filter.TokenAccountState},
		}
	case filter.Lamports != nil:
		return &pb.SubscribeRequestFilterAccountsFilter{
			Filter: &pb.SubscribeRequestFilterAccountsFilter_Lamports{Lamports: lamportsFilterToProto(filter.Lamports)},
		}
	default:
		return &pb.SubscribeRequestFilterAccountsFilter{}
	}
}

func lamportsFilterToProto(filter *SubscribeRequestFilterAccountsFilterLamports) *pb.SubscribeRequestFilterAccountsFilterLamports {
	out := &pb.SubscribeRequestFilterAccountsFilterLamports{}
	switch {
	case filter.Eq != nil:
		out.Cmp = &pb.SubscribeRequestFilterAccountsFilterLamports_Eq{Eq: *filter.Eq}
	case filter.Ne != nil:
		out.Cmp = &pb.SubscribeRequestFilterAccountsFilterLamports_Ne{Ne: *filter.Ne}
	case filter.Lt != nil:
		out.Cmp = &pb.SubscribeRequestFilterAccountsFilterLamports_Lt{Lt: *filter.Lt}
	case filter.Gt != nil:
		out.Cmp = &pb.SubscribeRequestFilterAccountsFilterLamports_Gt{Gt: *filter.Gt}
	}
	return out
}

// handleStream 处理流式响应
func (c *YellowstoneGrpc) handleStream(ctx context.Context, stream pb.Geyser_SubscribeClient, sub *Subscription) {
	defer func() {
		c.mu.Lock()
		delete(c.subscribers, sub.ID)
		c.mu.Unlock()
		if sub.callbacks.OnEnd != nil {
			sub.callbacks.OnEnd()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		resp, err := stream.Recv()
		if err != nil {
			if sub.callbacks.OnError != nil {
				sub.callbacks.OnError(err)
			}
			return
		}

		// Geyser 周期性下发 SubscribeUpdate.ping；必须在同一 Subscribe 双向流上回写 SubscribeRequest.ping，
		// 与 Rust / TypeScript 客户端一致，否则公共节点或 LB 可能 RST_STREAM。
		if resp.GetPing() != nil {
			if err := stream.Send(&pb.SubscribeRequest{
				Ping: &pb.SubscribeRequestPing{Id: 1},
			}); err != nil {
				if sub.callbacks.OnError != nil {
					sub.callbacks.OnError(err)
				}
				return
			}
			continue
		}

		if sub.callbacks.OnUpdate != nil {
			update := c.convertSubscribeUpdate(resp)
			sub.callbacks.OnUpdate(update)
		}
	}
}

// convertSubscribeUpdate 转换 protobuf 更新到本地类型
func (c *YellowstoneGrpc) convertSubscribeUpdate(pbUpdate *pb.SubscribeUpdate) *SubscribeUpdate {
	update := &SubscribeUpdate{
		Filters: pbUpdate.Filters,
	}
	if ts := pbUpdate.GetCreatedAt(); ts != nil {
		us := ts.Seconds*1_000_000 + int64(ts.Nanos)/1_000
		update.CreatedAt = &us
	}

	// 转换账户更新
	if pbUpdate.GetAccount() != nil {
		acc := pbUpdate.GetAccount()
		update.Account = &SubscribeUpdateAccount{
			Slot:      acc.Slot,
			IsStartup: acc.IsStartup,
		}
		if acc.Account != nil {
			update.Account.Account = &SubscribeUpdateAccountInfo{
				Pubkey:       acc.Account.Pubkey,
				Lamports:     acc.Account.Lamports,
				Owner:        acc.Account.Owner,
				Executable:   acc.Account.Executable,
				RentEpoch:    acc.Account.RentEpoch,
				Data:         acc.Account.Data,
				WriteVersion: acc.Account.WriteVersion,
				TxnSignature: acc.Account.TxnSignature,
			}
		}
	}

	// 转换 slot 更新
	if pbUpdate.GetSlot() != nil {
		slot := pbUpdate.GetSlot()
		update.Slot = &SubscribeUpdateSlot{
			Slot:   slot.Slot,
			Status: SlotStatus(slot.Status),
		}
		if slot.Parent != nil {
			update.Slot.Parent = slot.Parent
		}
		if slot.DeadError != nil {
			update.Slot.DeadError = slot.DeadError
		}
	}

	// 转换交易更新
	if pbUpdate.GetTransaction() != nil {
		tx := pbUpdate.GetTransaction()
		update.Transaction = &SubscribeUpdateTransaction{
			Slot: tx.Slot,
		}
		if tx.Transaction != nil {
			update.Transaction.Transaction = &SubscribeUpdateTransactionInfo{
				Signature:   tx.Transaction.Signature,
				IsVote:      tx.Transaction.IsVote,
				Transaction: tx.Transaction.Transaction,
				Meta:        tx.Transaction.Meta,
				Index:       tx.Transaction.Index,
			}
		}
	}

	// 转换区块更新
	if pbUpdate.GetBlock() != nil {
		block := pbUpdate.GetBlock()
		update.Block = &SubscribeUpdateBlock{
			Slot:                     block.Slot,
			Blockhash:                block.Blockhash,
			ParentSlot:               block.ParentSlot,
			ParentBlockhash:          block.ParentBlockhash,
			ExecutedTransactionCount: block.ExecutedTransactionCount,
		}
	}

	// 转换区块元数据更新
	if pbUpdate.GetBlockMeta() != nil {
		meta := pbUpdate.GetBlockMeta()
		update.BlockMeta = &SubscribeUpdateBlockMeta{
			Slot:                     meta.Slot,
			Blockhash:                meta.Blockhash,
			ParentSlot:               meta.ParentSlot,
			ParentBlockhash:          meta.ParentBlockhash,
			ExecutedTransactionCount: meta.ExecutedTransactionCount,
		}
	}

	if meta := pbUpdate.GetBlockMeta(); meta != nil && meta.BlockTime != nil {
		t := meta.BlockTime.Timestamp
		update.BlockMeta.BlockTime = &t
	}

	// 转换 Ping
	if pbUpdate.GetPing() != nil {
		update.Ping = &SubscribeUpdatePing{}
	}

	// 转换 Pong
	if pbUpdate.GetPong() != nil {
		pong := pbUpdate.GetPong()
		update.Pong = &SubscribeUpdatePong{
			ID: pong.Id,
		}
	}

	return update
}

// Unsubscribe 取消订阅
func (c *YellowstoneGrpc) Unsubscribe(subID string) error {
	c.mu.Lock()
	sub, exists := c.subscribers[subID]
	if !exists {
		c.mu.Unlock()
		return fmt.Errorf("subscription %s not found", subID)
	}
	delete(c.subscribers, subID)
	c.mu.Unlock()

	if sub.Cancel != nil {
		sub.Cancel()
	}

	return nil
}

// IsConnected 检查是否已连接
func (c *YellowstoneGrpc) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetConfig 获取客户端配置
func (c *YellowstoneGrpc) GetConfig() ClientConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// generateSubID 生成订阅 ID
var subIDCounter uint64
var subIDMu sync.Mutex

func generateSubID() string {
	subIDMu.Lock()
	defer subIDMu.Unlock()
	subIDCounter++
	return fmt.Sprintf("sub_%d_%d", time.Now().Unix(), subIDCounter)
}

// GetLatestBlockhash 获取最新区块哈希
func (c *YellowstoneGrpc) GetLatestBlockhash(commitment *CommitmentLevel) (*GetLatestBlockhashResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.GetLatestBlockhashRequest{}
	if commitment != nil {
		req.Commitment = (*pb.CommitmentLevel)(commitment)
	}

	resp, err := client.GetLatestBlockhash(c.ctx, req)
	if err != nil {
		return nil, err
	}

	return &GetLatestBlockhashResponse{
		Slot:                 resp.Slot,
		Blockhash:            resp.Blockhash,
		LastValidBlockHeight: resp.LastValidBlockHeight,
	}, nil
}

// GetBlockHeight 获取区块高度
func (c *YellowstoneGrpc) GetBlockHeight(commitment *CommitmentLevel) (*GetBlockHeightResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.GetBlockHeightRequest{}
	if commitment != nil {
		req.Commitment = (*pb.CommitmentLevel)(commitment)
	}

	resp, err := client.GetBlockHeight(c.ctx, req)
	if err != nil {
		return nil, err
	}

	return &GetBlockHeightResponse{
		BlockHeight: resp.BlockHeight,
	}, nil
}

// GetSlot 获取当前 Slot
func (c *YellowstoneGrpc) GetSlot(commitment *CommitmentLevel) (*GetSlotResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.GetSlotRequest{}
	if commitment != nil {
		req.Commitment = (*pb.CommitmentLevel)(commitment)
	}

	resp, err := client.GetSlot(c.ctx, req)
	if err != nil {
		return nil, err
	}

	return &GetSlotResponse{
		Slot: resp.Slot,
	}, nil
}

// GetVersion 获取服务器版本
func (c *YellowstoneGrpc) GetVersion() (*GetVersionResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.GetVersionRequest{}

	resp, err := client.GetVersion(c.ctx, req)
	if err != nil {
		return nil, err
	}

	return &GetVersionResponse{
		Version: resp.Version,
	}, nil
}

// IsBlockhashValid 验证区块哈希是否有效
func (c *YellowstoneGrpc) IsBlockhashValid(blockhash string, commitment *CommitmentLevel) (*IsBlockhashValidResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.IsBlockhashValidRequest{
		Blockhash: blockhash,
	}
	if commitment != nil {
		req.Commitment = (*pb.CommitmentLevel)(commitment)
	}

	resp, err := client.IsBlockhashValid(c.ctx, req)
	if err != nil {
		return nil, err
	}

	return &IsBlockhashValidResponse{
		Slot:  resp.Slot,
		Valid: resp.Valid,
	}, nil
}

// Ping 发送 Ping 请求
func (c *YellowstoneGrpc) Ping(count int32) (*PongResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.PingRequest{
		Count: count,
	}

	resp, err := client.Ping(c.ctx, req)
	if err != nil {
		return nil, err
	}

	return &PongResponse{
		Count: resp.Count,
	}, nil
}

// SubscribeReplayInfo 订阅重放信息
func (c *YellowstoneGrpc) SubscribeReplayInfo() (*SubscribeReplayInfoResponse, error) {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client not connected")
	}
	client := c.client
	c.mu.RUnlock()

	req := &pb.SubscribeReplayInfoRequest{}

	resp, err := client.SubscribeReplayInfo(c.ctx, req)
	if err != nil {
		return nil, err
	}

	result := &SubscribeReplayInfoResponse{}
	if resp.FirstAvailable != nil {
		result.FirstAvailable = resp.FirstAvailable
	}

	return result, nil
}

// NewTLSConfig 创建 TLS 配置
func NewTLSConfig(skipVerify bool) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: skipVerify,
	}
}

// ParseCommitmentLevel 解析承诺级别字符串
func ParseCommitmentLevel(s string) CommitmentLevel {
	switch s {
	case "confirmed":
		return CommitmentLevelConfirmed
	case "finalized":
		return CommitmentLevelFinalized
	default:
		return CommitmentLevelProcessed
	}
}

// Stop waits for the active DEX stream and all associated workers to exit.
func (c *YellowstoneGrpc) Stop() {
	c.dexLifecycle.Lock()
	defer c.dexLifecycle.Unlock()
	c.stopDexUnlocked()
}
func (c *YellowstoneGrpc) stopDexUnlocked() {
	if c.dexCancel != nil {
		c.dexCancel()
	}
	if c.dexDone != nil {
		<-c.dexDone
	}
	c.dexCancel = nil
	c.dexDone = nil
}
