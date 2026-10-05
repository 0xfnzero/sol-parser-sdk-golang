package solparser

import (
	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
	"google.golang.org/grpc"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type nativeLifecycleServer struct {
	pb.UnimplementedGeyserServer
	active  atomic.Int32
	started chan struct{}
}

func (s *nativeLifecycleServer) Subscribe(stream grpc.BidiStreamingServer[pb.SubscribeRequest, pb.SubscribeUpdate]) error {
	s.active.Add(1)
	defer s.active.Add(-1)
	if _, err := stream.Recv(); err != nil {
		return err
	}
	s.started <- struct{}{}
	for {
		if _, err := stream.Recv(); err != nil {
			return err
		}
	}
}
func TestNativeSingleDexSubscriptionLifecycle(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	implementation := &nativeLifecycleServer{started: make(chan struct{}, 4)}
	pb.RegisterGeyserServer(server, implementation)
	go server.Serve(listener)
	defer server.Stop()
	cfg := DefaultClientConfig()
	cfg.EnableTLS = false
	client := NewYellowstoneGrpc("http://"+listener.Addr().String(), cfg)
	defer client.Disconnect()
	wait := func() {
		t.Helper()
		select {
		case <-implementation.started:
		case <-time.After(3 * time.Second):
			t.Fatal("gRPC stream did not start")
		}
	}
	first, err := client.SubscribeDexEvents(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	wait()
	second, err := client.SubscribeDexEvents(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	wait()
	select {
	case _, ok := <-first.Events:
		if ok {
			t.Fatal("old events channel still active")
		}
	default:
		t.Fatal("old stream was not joined")
	}
	if err := client.UpdateSubscription([]TransactionFilter{{AccountInclude: []string{"test"}}}, nil); err != nil {
		t.Fatal(err)
	}
	client.Stop()
	select {
	case _, ok := <-second.Events:
		if ok {
			t.Fatal("second events channel active")
		}
	default:
		t.Fatal("stop did not join")
	}
	// Cancellation does not destroy the client's connection context.
	third, err := client.SubscribeDexEvents(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	wait()
	third.Cancel()
	client.Stop()
	third.Join()
	if third.Status().State != "stopped" {
		t.Fatal(third.Status())
	}
	client.Disconnect()
	fourth, err := client.SubscribeDexEvents(nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	wait()
	fourth.Cancel()
	fourth.Join()
}
func TestNativeLaunchLabAliases(t *testing.T) {
	var event StonkFunTradeEvent
	var original *RaydiumLaunchlabTradeEvent = &event
	var alias *LaunchLabTradeEvent = original
	if alias != original {
		t.Fatal("event aliases must retain identity")
	}
}
