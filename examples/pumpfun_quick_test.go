//go:build ignore

// PumpFun Quick Test
//
// Quick connection test - subscribes to ALL PumpFun events,
// prints the first 10, then exits.
//
// Run: GRPC_URL=host:443 GRPC_TOKEN=your_token go run examples/pumpfun_quick_test.go

package main

import (
	"fmt"
	"os"
	"time"

	base58 "github.com/mr-tron/base58"
	solparser "sol-parser-sdk-golang/solparser"
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

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude: []string{"6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"},
		Vote:           &voteF,
		Failed:         &failedF,
	}

	fmt.Println("✅ Subscribing... (no event filter - will show ALL events)")
	fmt.Println("🎧 Listening for events... (waiting up to 60 seconds)\n")

	eventCount := 0
	start := time.Now()
	done := make(chan struct{})

	callbacks := solparser.SubscribeCallbacks{
		OnUpdate: func(update *solparser.SubscribeUpdate) {
			if update.Transaction == nil || update.Transaction.Transaction == nil {
				return
			}
			txInfo := update.Transaction.Transaction
			if txInfo.Meta == nil || len(txInfo.Meta.LogMessages) == 0 {
				return
			}
			logs := txInfo.Meta.LogMessages
			sigStr := base58.Encode(txInfo.Signature)
			slot := update.Transaction.Slot
			events := solparser.ParseLogsOnly(logs, sigStr, slot, nil)

			for _, ev := range events {
				for key := range ev {
					eventCount++
					fmt.Printf("✅ Event #%d: %s (slot=%d)\n", eventCount, key, slot)
					if eventCount >= 10 {
						fmt.Printf("\n🎉 Received %d events! Test successful!\n", eventCount)
						select {
						case <-done:
						default:
							close(done)
						}
						return
					}
					break
				}
			}

			if time.Since(start) > 60*time.Second {
				if eventCount == 0 {
					fmt.Println("⏰ Timeout: No events received in 60 seconds.")
					fmt.Println("   This might indicate:")
					fmt.Println("   - Network connectivity issues")
					fmt.Println("   - gRPC endpoint is down")
					fmt.Println("   - Missing or invalid API token")
				} else {
					fmt.Printf("\n✅ Received %d events in 60 seconds\n", eventCount)
				}
				select {
				case <-done:
				default:
					close(done)
				}
			}
		},
		OnError: func(err error) {
			fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
		},
		OnEnd: func() {
			fmt.Println("Stream ended")
			select {
			case <-done:
			default:
				close(done)
			}
		},
	}

	sub, err := client.SubscribeTransactions(filter, callbacks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Connected. Waiting for PumpFun events...\n\n")

	<-done
	client.Unsubscribe(sub.ID)
}
