//go:build ignore

// PumpFun Quick Test
//
// Quick connection test - subscribes to ALL PumpFun DexEvents,
// prints the first 10, then exits.
//
// Run: GRPC_URL=host:443 GRPC_TOKEN=your_token go run examples/pumpfun_quick_test.go

package main

import (
	"fmt"
	"os"
	"time"

	solparser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
)

func main() {
	endpoint := os.Getenv("GRPC_URL")
	if endpoint == "" {
		endpoint = "solana-yellowstone-grpc.publicnode.com:443"
	}
	token := os.Getenv("GRPC_TOKEN")

	fmt.Println("🚀 Quick Test - Subscribing to ALL PumpFun events...")
	fmt.Printf("📡 Endpoint: %s\n\n", endpoint)

	client := solparser.NewYellowstoneGrpc(endpoint)
	if token != "" {
		client.SetXToken(token)
	}
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connect failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	protocols := []solparser.Protocol{solparser.ProtocolPumpFun}
	txFilter := solparser.TransactionFilterForProtocols(protocols)
	voteF := false
	failedF := false
	txFilter.Vote = &voteF
	txFilter.Failed = &failedF

	sub, err := client.SubscribeDexEvents([]solparser.TransactionFilter{txFilter}, nil, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", err)
		os.Exit(1)
	}
	defer sub.Cancel()

	fmt.Println("✅ Subscribing... (no event filter - will show ALL PumpFun events)")
	fmt.Println("🎧 Listening for events... (waiting up to 60 seconds)\n")

	eventCount := 0
	timeout := time.After(60 * time.Second)

	for {
		select {
		case ev, ok := <-sub.Events:
			if !ok {
				return
			}
			if !ev.IsPumpFun() {
				continue
			}
			eventCount++
			meta := ev.GetMetadata()
			fmt.Printf("✅ Event #%d: %s (slot=%d)\n", eventCount, ev.Type, meta.Slot)
			if eventCount >= 10 {
				fmt.Printf("\n🎉 Received %d events! Test successful!\n", eventCount)
				return
			}
		case err, ok := <-sub.Errors:
			if ok {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		case <-timeout:
			if eventCount == 0 {
				fmt.Println("⏰ Timeout: No events received in 60 seconds.")
				fmt.Println("   This might indicate:")
				fmt.Println("   - Network connectivity issues")
				fmt.Println("   - gRPC endpoint is down")
				fmt.Println("   - Missing or invalid API token")
			} else {
				fmt.Printf("\n✅ Received %d events in 60 seconds\n", eventCount)
			}
			return
		}
	}
}
