//go:build ignore

// PumpSwap Event Parsing with Detailed Performance Metrics
//
// Demonstrates how to:
// - Subscribe to PumpSwap protocol events through SubscribeDexEvents
// - Measure gRPC recv time, queue recv time, and end-to-end latency per event
// - Display periodic 10s summaries
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

	solparser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
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

	cfg := solparser.DefaultClientConfig()
	cfg.EnableMetrics = true
	cfg.OrderMode = solparser.OrderModeUnordered
	client := solparser.NewYellowstoneGrpc(endpoint, cfg)
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

	go reportStats("PumpSwap Performance Stats (10s window)", &eventCount, &totalLatency, &minLatency, &maxLatency, &lastCount)

	protocols := []solparser.Protocol{solparser.ProtocolPumpSwap}
	txFilter := solparser.TransactionFilterForProtocols(protocols)
	accountFilter := solparser.AccountFilterForProtocols(protocols)
	voteF := false
	failedF := false
	txFilter.Vote = &voteF
	txFilter.Failed = &failedF
	eventFilter := solparser.EventTypeFilterIncludeOnly([]solparser.EventType{
		solparser.EventTypePumpSwapBuy,
		solparser.EventTypePumpSwapSell,
		solparser.EventTypePumpSwapCreatePool,
		solparser.EventTypePumpSwapLiquidityAdded,
		solparser.EventTypePumpSwapLiquidityRemoved,
	})

	sub, err := client.SubscribeDexEvents(
		[]solparser.TransactionFilter{txFilter},
		[]solparser.AccountFilter{accountFilter},
		eventFilter,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", err)
		os.Exit(1)
	}
	defer sub.Cancel()

	fmt.Printf("✅ gRPC client created successfully\n")
	fmt.Printf("📋 Event Filter: Buy, Sell, CreatePool, LiquidityAdded, LiquidityRemoved\n")
	fmt.Printf("✅ Subscribed (id=%s)\n", sub.ID)
	fmt.Println("🛑 Press Ctrl+C to stop...")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev, ok := <-sub.Events:
			if !ok {
				return
			}
			queueRecvUs := solparser.NowUs()
			meta := ev.GetMetadata()
			latencyUs := queueRecvUs - meta.GrpcRecvUs
			if latencyUs < 0 {
				latencyUs = 0
			}

			atomic.AddInt64(&eventCount, 1)
			atomic.AddInt64(&totalLatency, latencyUs)
			updateMin(&minLatency, latencyUs)
			updateMax(&maxLatency, latencyUs)

			fmt.Printf("\n================================================\n")
			fmt.Printf("gRPC recv time : %d μs\n", meta.GrpcRecvUs)
			fmt.Printf("Queue recv time: %d μs\n", queueRecvUs)
			fmt.Printf("Latency        : %d μs\n", latencyUs)
			fmt.Printf("================================================\n")
			fmt.Printf("Event: %s\n", ev.Type)
			printPumpSwapFields(ev)
			fmt.Println()
		case err, ok := <-sub.Errors:
			if ok {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		case <-interrupt:
			fmt.Println("\n👋 Shutting down gracefully...")
			return
		}
	}
}

func printPumpSwapFields(ev solparser.DexEvent) {
	switch d := ev.Data.(type) {
	case *solparser.PumpSwapBuyEvent:
		fmt.Printf("  pool : %s\n", d.Pool)
		fmt.Printf("  user : %s\n", d.User)
		fmt.Printf("  base_mint : %s\n", d.BaseMint)
		fmt.Printf("  quote_mint: %s\n", d.QuoteMint)
	case *solparser.PumpSwapSellEvent:
		fmt.Printf("  pool : %s\n", d.Pool)
		fmt.Printf("  user : %s\n", d.User)
		fmt.Printf("  base_mint : %s\n", d.BaseMint)
		fmt.Printf("  quote_mint: %s\n", d.QuoteMint)
	case *solparser.PumpSwapCreatePoolEvent:
		fmt.Printf("  pool : %s\n", d.Pool)
		fmt.Printf("  creator : %s\n", d.Creator)
		fmt.Printf("  base_mint : %s\n", d.BaseMint)
		fmt.Printf("  quote_mint: %s\n", d.QuoteMint)
	}
}

func reportStats(title string, eventCount, totalLatency, minLatency, maxLatency, lastCount *int64) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		count := atomic.LoadInt64(eventCount)
		total := atomic.LoadInt64(totalLatency)
		minL := atomic.LoadInt64(minLatency)
		maxL := atomic.LoadInt64(maxLatency)
		last := atomic.LoadInt64(lastCount)
		if count == 0 {
			continue
		}
		avg := total / count
		rate := float64(count-last) / 10.0
		if minL == 1<<62-1 {
			minL = 0
		}
		fmt.Println("\n╔════════════════════════════════════════════════════╗")
		fmt.Printf("║  %-48s║\n", title)
		fmt.Println("╠════════════════════════════════════════════════════╣")
		fmt.Printf("║  Total Events : %10d                              ║\n", count)
		fmt.Printf("║  Events/sec   : %10.1f                              ║\n", rate)
		fmt.Printf("║  Avg Latency  : %10d μs                           ║\n", avg)
		fmt.Printf("║  Min Latency  : %10d μs                           ║\n", minL)
		fmt.Printf("║  Max Latency  : %10d μs                           ║\n", maxL)
		fmt.Println("╚════════════════════════════════════════════════════╝\n")
		atomic.StoreInt64(lastCount, count)
	}
}

func updateMin(target *int64, value int64) {
	for {
		cur := atomic.LoadInt64(target)
		if value >= cur || atomic.CompareAndSwapInt64(target, cur, value) {
			return
		}
	}
}

func updateMax(target *int64, value int64) {
	for {
		cur := atomic.LoadInt64(target)
		if value <= cur || atomic.CompareAndSwapInt64(target, cur, value) {
			return
		}
	}
}
