package shredstream

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "sol-parser-sdk-golang/shredstream/pb"
)

// Client Jito ShredStream gRPC 客户端（与 `0xfnzero/solana-streamer` 中 prost 生成的
// `Shredstream` + `ShredstreamProxy` 服务对齐）。
type Client struct {
	endpoint string
	config   ShredStreamConfig
	conn     *grpc.ClientConn
	proxy    pb.ShredstreamProxyClient
	shred    pb.ShredstreamClient // SendHeartbeat
}

// Dial 建立连接（测试连接与 Rust `new_with_config` 类似）。
func Dial(ctx context.Context, endpoint string, cfg ShredStreamConfig) (*Client, error) {
	if cfg.MaxDecodingMessageSize <= 0 {
		cfg = DefaultShredStreamConfig()
	}
	dialCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ConnectionTimeoutMs)*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.MaxDecodingMessageSize),
			grpc.MaxCallSendMsgSize(cfg.MaxDecodingMessageSize),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("shredstream dial: %w", err)
	}
	return &Client{
		endpoint: endpoint,
		config:   cfg,
		conn:     conn,
		proxy:    pb.NewShredstreamProxyClient(conn),
		shred:    pb.NewShredstreamClient(conn),
	}, nil
}

// Close 关闭底层连接。
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendHeartbeat 调用 `/shredstream.Shredstream/SendHeartbeat`（部分部署用于维持订阅）。
func (c *Client) SendHeartbeat(ctx context.Context, req *pb.Heartbeat) (*pb.HeartbeatResponse, error) {
	if c.shred == nil {
		return nil, fmt.Errorf("shredstream: client not connected")
	}
	return c.shred.SendHeartbeat(ctx, req)
}

// SubscribeEntries 返回 gRPC 流。对 `Entry.entries` 的二进制解码请使用
// DecodeEntriesBincode 或 DecodeGRPCEntry（布局与 solana-streamer / shredstream-sdk-go 一致）。
func (c *Client) SubscribeEntries(ctx context.Context) (grpc.ServerStreamingClient[pb.Entry], error) {
	if c.proxy == nil {
		return nil, fmt.Errorf("shredstream: client not connected")
	}
	return c.proxy.SubscribeEntries(ctx, &pb.SubscribeEntriesRequest{})
}
