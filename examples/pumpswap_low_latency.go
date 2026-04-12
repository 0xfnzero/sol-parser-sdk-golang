//go:build ignore

// PumpSwap Low-Latency Example
//
// Demonstrates:
// - Subscribe to PumpSwap protocol events
// - Measure end-to-end latency
// - Per-event and periodic performance statistics
//
// Run: go run examples/pumpswap_low_latency.go  (from sol-parser-sdk-golang/)

package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	base58 "github.com/mr-tron/base58"
	solparser "sol-parser-sdk-golang/solparser"
)

var pumpSwapProgramIDs = []string{
	solparser.PUMPSWAP_PROGRAM_ID,
	solparser.GrpcPumpSwapFeesProgramID,
}

func nowUsPumpSwap() int64 {
	return time.Now().UnixMicro()
}

func main() {
	endpoint := os.Getenv("GRPC_URL")
	if endpoint == "" {
		endpoint = "solana-yellowstone-grpc.publicnode.com:443"
	}
	token := os.Getenv("GRPC_TOKEN")

	fmt.Println("🚀 PumpSwap Low-Latency Test")
	fmt.Println("============================\n")

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
		minLatency   int64 = math.MaxInt64
		maxLatency   int64
	)

	// Stats reporter every 10s
	go func() {
		lastCount := int64(0)
		for range time.Tick(10 * time.Second) {
			count := atomic.LoadInt64(&eventCount)
			total := atomic.LoadInt64(&totalLatency)
			minL := atomic.LoadInt64(&minLatency)
			maxL := atomic.LoadInt64(&maxLatency)

			if count == 0 {
				continue
			}
			avg := total / count
			rate := float64(count-lastCount) / 10.0
			if minL == math.MaxInt64 {
				minL = 0
			}

			fmt.Println("\n╔════════════════════════════════════════════════════╗")
			fmt.Println("║          Performance Stats (10s window)            ║")
			fmt.Println("╠════════════════════════════════════════════════════╣")
			fmt.Printf("║  Total Events : %10d                              ║\n", count)
			fmt.Printf("║  Events/sec   : %10.1f                              ║\n", rate)
			fmt.Printf("║  Avg Latency  : %10d μs                           ║\n", avg)
			fmt.Printf("║  Min Latency  : %10d μs                           ║\n", minL)
			fmt.Printf("║  Max Latency  : %10d μs                           ║\n", maxL)
			fmt.Println("╚════════════════════════════════════════════════════╝\n")
			lastCount = count
		}
	}()

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude:  pumpSwapProgramIDs,
		AccountExclude:  []string{},
		AccountRequired: []string{},
		Vote:            &voteF,
		Failed:          &failedF,
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
			queueRecvUs := nowUsPumpSwap()

			events := solparser.ParseLogsOnly(logs, sigStr, slot, nil)
			for _, ev := range events {
				for key := range ev {
					if len(key) < 8 || key[:8] != "PumpSwap" {
						break
					}
					data, _ := ev[key].(map[string]any)
					var grpcRecvUs int64
					if md, ok := data["metadata"].(map[string]any); ok {
						if v, ok := md["grpc_recv_us"].(float64); ok {
							grpcRecvUs = int64(v)
						}
					}
					latencyUs := queueRecvUs - grpcRecvUs

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
					if v, ok := data["pool"]; ok {
						fmt.Printf("  pool : %v\n", v)
					}
					if v, ok := data["user"]; ok {
						fmt.Printf("  user : %v\n", v)
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
	fmt.Printf("✅ Subscribed (id=%s)\n", sub.ID)
	fmt.Println("🛑 Press Ctrl+C to stop...\n")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-done:
	case <-interrupt:
	}
	client.Unsubscribe(sub.ID)
}
