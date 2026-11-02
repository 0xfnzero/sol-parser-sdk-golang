package shredstream

import (
	"context"
	pb "github.com/0xfnzero/sol-parser-sdk-golang/shredstream/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
	"testing"
)

type tokenProxy struct{ tokens []string }

func (p *tokenProxy) SubscribeEntries(ctx context.Context, _ *pb.SubscribeEntriesRequest, _ ...grpc.CallOption) (grpc.ServerStreamingClient[pb.Entry], error) {
	md, _ := metadata.FromOutgoingContext(ctx)
	p.tokens = md.Get("x-token")
	return nil, nil
}
func TestSubscribeToken(t *testing.T) {
	for _, token := range []string{"", "  ", "  test-token  ", "secret\nbad"} {
		proxy := &tokenProxy{}
		client := &Client{proxy: proxy, config: ShredStreamConfig{AuthToken: &token}}
		_, err := client.SubscribeEntries(context.Background())
		if strings.Contains(token, "\n") {
			if err == nil || err.Error() != "invalid x-token metadata" {
				t.Fatal(err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(token) == "" {
			if len(proxy.tokens) != 0 {
				t.Fatal(proxy.tokens)
			}
		} else if len(proxy.tokens) != 1 || proxy.tokens[0] != "test-token" {
			t.Fatal(proxy.tokens)
		}
	}
}
