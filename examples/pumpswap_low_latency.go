//go:build ignore

// PumpSwap Low-Latency Example
//
// Demonstrates:
// - Subscribe to PumpSwap protocol events through SubscribeDexEvents
// - Use the latest protocol/event filters instead of hand-written Program IDs
// - Measure gRPC recv time, queue recv time, and per-event latency
//
// Run: go run examples/pumpswap_low_latency.go  (from github.com/0xfnzero/sol-parser-sdk-golang/)

package main

import (
	"fmt"
	"math"
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

	fmt.Println("🚀 PumpSwap Low-Latency Test")
	fmt.Println("============================\n")

	cfg := solparser.DefaultClientConfig()
	cfg.EnableMetrics = true
	cfg.OrderMode = solparser.OrderModeUnordered
	cfg.BufferSize = 16_384

	client := solparser.NewYellowstoneGrpc(endpoint, cfg)
	if token != "" {
		client.SetXToken(token)
	}
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connect failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

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

	var (
		eventCount   int64
		totalLatency int64
		minLatency   int64 = math.MaxInt64
		maxLatency   int64
	)

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

	fmt.Printf("✅ Subscribed (id=%s)\n", sub.ID)
	fmt.Printf("📊 Protocols: %v | OrderMode=%s\n", protocols, cfg.OrderMode)
	fmt.Println("🛑 Press Ctrl+C to stop...\n")

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
			fmt.Printf("  sig  : %s\n", meta.Signature)
			fmt.Printf("  slot : %d\n", meta.Slot)
			printPumpSwapFields(ev)
			fmt.Println()
		case err, ok := <-sub.Errors:
			if ok {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		case <-interrupt:
			return
		}
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

func printPumpSwapFields(ev solparser.DexEvent) {
	switch d := ev.Data.(type) {
	case *solparser.PumpSwapBuyEvent:
		fmt.Printf("  pool : %s\n", d.Pool)
		fmt.Printf("  user : %s\n", d.User)
	case *solparser.PumpSwapSellEvent:
		fmt.Printf("  pool : %s\n", d.Pool)
		fmt.Printf("  user : %s\n", d.User)
	case *solparser.PumpSwapCreatePoolEvent:
		fmt.Printf("  pool : %s\n", d.Pool)
		fmt.Printf("  user : %s\n", d.Creator)
	}
}
