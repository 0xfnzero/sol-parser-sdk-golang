package shredstream

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/0xfnzero/sol-parser-sdk-golang/shredstream/pb"
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
	target, transport, err := endpointTransport(endpoint)
	if err != nil {
		return nil, err
	}
	if cfg.MaxDecodingMessageSize <= 0 {
		token := cfg.AuthToken
		cfg = DefaultShredStreamConfig()
		cfg.AuthToken = token
	}
	dialCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.ConnectionTimeoutMs)*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, target,
		grpc.WithTransportCredentials(transport),
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
	if c.config.AuthToken != nil {
		token := strings.TrimSpace(*c.config.AuthToken)
		for _, b := range []byte(token) {
			if b < 32 || b > 126 {
				return nil, fmt.Errorf("invalid x-token metadata")
			}
		}
		if token != "" {
			md, _ := metadata.FromOutgoingContext(ctx)
			md = md.Copy()
			md.Set("x-token", token)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
	}
	return c.proxy.SubscribeEntries(ctx, &pb.SubscribeEntriesRequest{})
}

// Keep bare host:port endpoints compatible, but never discard an explicit TLS scheme.
func endpointTransport(endpoint string) (string, credentials.TransportCredentials, error) {
	endpoint = strings.TrimSpace(endpoint)
	if !strings.Contains(endpoint, "://") {
		if endpoint == "" {
			return "", nil, fmt.Errorf("shredstream: endpoint is empty")
		}
		return endpoint, insecure.NewCredentials(), nil
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return "", nil, fmt.Errorf("shredstream: invalid endpoint")
	}
	port := u.Port()
	switch u.Scheme {
	case "https":
		if port == "" {
			port = "443"
		}
		return net.JoinHostPort(u.Hostname(), port), credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12}), nil
	case "http":
		if port == "" {
			port = "80"
		}
		return net.JoinHostPort(u.Hostname(), port), insecure.NewCredentials(), nil
	default:
		return "", nil, fmt.Errorf("shredstream: endpoint must use http or https")
	}
}
