//go:build ignore

// PumpSwap Event Parsing with Detailed Performance Metrics
//
// Demonstrates how to:
// - Subscribe to PumpSwap protocol events
// - Measure end-to-end latency per event
// - Display periodic 10s summaries (total count, rate, avg/min/max latency)
//
// Run: GRPC_URL=host:443 GRPC_TOKEN=your_token go run examples/pumpswap_with_metrics.go

package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
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

	fmt.Println("PumpSwap event parsing with detailed performance metrics")
	fmt.Println("🚀 Subscribing to Yellowstone gRPC (PumpSwap)...")
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

	var (
		eventCount   int64
		totalLatency int64
		minLatency   int64 = 1<<62 - 1
		maxLatency   int64
		lastCount    int64
	)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			count := atomic.LoadInt64(&eventCount)
			total := atomic.LoadInt64(&totalLatency)
			minL := atomic.LoadInt64(&minLatency)
			maxL := atomic.LoadInt64(&maxLatency)
			last := atomic.LoadInt64(&lastCount)
			if count == 0 {
				continue
			}
			avg := total / count
			rate := float64(count-last) / 10.0
			if minL == 1<<62-1 {
				minL = 0
			}
			fmt.Println("\n╔════════════════════════════════════════════════════╗")
			fmt.Println("║      PumpSwap Performance Stats (10s window)       ║")
			fmt.Println("╠════════════════════════════════════════════════════╣")
			fmt.Printf("║  Total Events : %10d                              ║\n", count)
			fmt.Printf("║  Events/sec   : %10.1f                              ║\n", rate)
			fmt.Printf("║  Avg Latency  : %10d μs                           ║\n", avg)
			fmt.Printf("║  Min Latency  : %10d μs                           ║\n", minL)
			fmt.Printf("║  Max Latency  : %10d μs                           ║\n", maxL)
			fmt.Println("╚════════════════════════════════════════════════════╝\n")
			atomic.StoreInt64(&lastCount, count)
		}
	}()

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude: []string{solparser.PUMPSWAP_PROGRAM_ID, solparser.GrpcPumpSwapFeesProgramID},
		Vote:           &voteF,
		Failed:         &failedF,
	}

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
			queueRecvUs := time.Now().UnixMicro()

			events := solparser.ParseLogsOnly(logs, sigStr, slot, nil)
			for _, ev := range events {
				for key := range ev {
					if len(key) < 8 || key[:8] != "PumpSwap" {
						break
					}
					dataMap, _ := ev[key].(map[string]any)
					var grpcRecvUs int64
					if md, ok := dataMap["metadata"].(map[string]any); ok {
						if v, ok := md["grpc_recv_us"].(float64); ok {
							grpcRecvUs = int64(v)
						}
					}
					var latencyUs int64
					if grpcRecvUs > 0 {
						latencyUs = queueRecvUs - grpcRecvUs
					}

					atomic.AddInt64(&eventCount, 1)
					atomic.AddInt64(&totalLatency, latencyUs)
					for {
						cur := atomic.LoadInt64(&minLatency)
						if latencyUs >= cur || atomic.CompareAndSwapInt64(&minLatency, cur, latencyUs) {
							break
						}
					}
					for {
						cur := atomic.LoadInt64(&maxLatency)
						if latencyUs <= cur || atomic.CompareAndSwapInt64(&maxLatency, cur, latencyUs) {
							break
						}
					}

					fmt.Printf("\n================================================\n")
					fmt.Printf("gRPC recv time : %d μs\n", grpcRecvUs)
					fmt.Printf("Queue recv time: %d μs\n", queueRecvUs)
					fmt.Printf("Latency        : %d μs\n", latencyUs)
					fmt.Printf("================================================\n")
					fmt.Printf("Event: %s\n", key)
					if v, ok := dataMap["pool"]; ok {
						fmt.Printf("  pool : %v\n", v)
					}
					if v, ok := dataMap["user"]; ok {
						fmt.Printf("  user : %v\n", v)
					}
					if v, ok := dataMap["base_mint"]; ok {
						fmt.Printf("  base_mint : %v\n", v)
					}
					if v, ok := dataMap["quote_mint"]; ok {
						fmt.Printf("  quote_mint: %v\n", v)
					}
					fmt.Println()
					break
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
	fmt.Printf("✅ gRPC client created successfully\n")
	fmt.Printf("📋 Event Filter: Buy, Sell, CreatePool, LiquidityAdded, LiquidityRemoved\n")
	fmt.Printf("✅ Subscribed (id=%s)\n", sub.ID)
	fmt.Println("🛑 Press Ctrl+C to stop...")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-done:
	case <-interrupt:
	}
	client.Unsubscribe(sub.ID)
	fmt.Println("\n👋 Shutting down gracefully...")
}
