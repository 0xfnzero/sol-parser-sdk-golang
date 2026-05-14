//go:build ignore

// PumpFun Trade Event Filter Example
//
// Demonstrates how to:
// - Subscribe to PumpFun protocol events through SubscribeDexEvents
// - Filter specific trade types: Buy, Sell, BuyExactSolIn, Create
// - Display trade details with latency metrics
//
// Run: go run examples/pumpfun_trade_filter.go  (from github.com/0xfnzero/sol-parser-sdk-golang/)

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	solparser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
)

func main() {
	endpoint := os.Getenv("GRPC_URL")
	if endpoint == "" {
		endpoint = "solana-yellowstone-grpc.publicnode.com:443"
	}
	token := os.Getenv("GRPC_TOKEN")

	fmt.Println("🚀 PumpFun Trade Event Filter Example")
	fmt.Println("======================================\n")
	fmt.Printf("📡 Endpoint: %s\n", endpoint)
	fmt.Println("🎯 Protocol: PumpFun\n")

	cfg := solparser.DefaultClientConfig()
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

	protocols := []solparser.Protocol{solparser.ProtocolPumpFun}
	txFilter := solparser.TransactionFilterForProtocols(protocols)
	voteF := false
	failedF := false
	txFilter.Vote = &voteF
	txFilter.Failed = &failedF

	eventFilter := solparser.EventTypeFilterIncludeOnly([]solparser.EventType{
		solparser.EventTypePumpFunBuy,
		solparser.EventTypePumpFunSell,
		solparser.EventTypePumpFunBuyExactSolIn,
		solparser.EventTypePumpFunCreate,
	})

	sub, err := client.SubscribeDexEvents(
		[]solparser.TransactionFilter{txFilter},
		nil,
		eventFilter,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", err)
		os.Exit(1)
	}
	defer sub.Cancel()

	var (
		eventCount    int
		buyCount      int
		sellCount     int
		buyExactCount int
		createCount   int
	)

	fmt.Printf("✅ Subscribed (id=%s)\n", sub.ID)
	fmt.Println("🛑 Press Ctrl+C to stop...\n")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev, ok := <-sub.Events:
			if !ok {
				return
			}
			eventCount++
			meta := ev.GetMetadata()
			latencyUs := solparser.NowUs() - meta.GrpcRecvUs
			if latencyUs < 0 {
				latencyUs = 0
			}

			switch ev.Type {
			case solparser.EventTypePumpFunBuy:
				buyCount++
				printTrade("🟢 PumpFun BUY", eventCount, meta, latencyUs, ev)
				fmt.Printf("│ 📊 Stats   : Buy=%d Sell=%d BuyExact=%d\n", buyCount, sellCount, buyExactCount)
				fmt.Println("└─────────────────────────────────────────────────────────────\n")
			case solparser.EventTypePumpFunBuyExactSolIn:
				buyExactCount++
				printTrade("🟡 PumpFun BUY_EXACT_SOL_IN", eventCount, meta, latencyUs, ev)
				fmt.Printf("│ 📊 Stats   : Buy=%d Sell=%d BuyExact=%d\n", buyCount, sellCount, buyExactCount)
				fmt.Println("└─────────────────────────────────────────────────────────────\n")
			case solparser.EventTypePumpFunSell:
				sellCount++
				printTrade("🔴 PumpFun SELL", eventCount, meta, latencyUs, ev)
				fmt.Printf("│ 📊 Stats   : Buy=%d Sell=%d BuyExact=%d\n", buyCount, sellCount, buyExactCount)
				fmt.Println("└─────────────────────────────────────────────────────────────\n")
			case solparser.EventTypePumpFunCreate:
				createCount++
				printCreate(eventCount, meta, latencyUs, ev)
				fmt.Printf("│ 📊 Creates : %d\n", createCount)
				fmt.Println("└─────────────────────────────────────────────────────────────\n")
			default:
				b, _ := json.Marshal(ev)
				fmt.Printf("[%s] %s\n\n", ev.Type, truncate(string(b), 300))
			}
		case err, ok := <-sub.Errors:
			if ok {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		case <-interrupt:
			fmt.Printf("\n👋 Total events: %d (Buy=%d Sell=%d BuyExact=%d Create=%d)\n",
				eventCount, buyCount, sellCount, buyExactCount, createCount)
			return
		}
	}
}

func printTrade(title string, count int, meta solparser.EventMetadata, latencyUs int64, ev solparser.DexEvent) {
	trade := ev.AsPumpFunTrade()
	fmt.Println("┌─────────────────────────────────────────────────────────────")
	fmt.Printf("│ %s #%d\n", title, count)
	fmt.Println("├─────────────────────────────────────────────────────────────")
	fmt.Printf("│ Signature  : %s\n", meta.Signature)
	fmt.Printf("│ Slot       : %d\n", meta.Slot)
	fmt.Println("├─────────────────────────────────────────────────────────────")
	if trade != nil {
		fmt.Printf("│ Mint       : %s\n", trade.Mint)
		fmt.Printf("│ SOL Amount : %d\n", trade.SolAmount)
		fmt.Printf("│ Token Amt  : %d\n", trade.TokenAmount)
		fmt.Printf("│ User       : %s\n", trade.User)
	}
	fmt.Println("├─────────────────────────────────────────────────────────────")
	fmt.Printf("│ 📊 Latency : %d μs\n", latencyUs)
}

func printCreate(count int, meta solparser.EventMetadata, latencyUs int64, ev solparser.DexEvent) {
	create := ev.AsPumpFunCreate()
	fmt.Println("┌─────────────────────────────────────────────────────────────")
	fmt.Printf("│ 🆕 PumpFun CREATE #%d\n", count)
	fmt.Println("├─────────────────────────────────────────────────────────────")
	fmt.Printf("│ Signature  : %s\n", meta.Signature)
	fmt.Printf("│ Slot       : %d\n", meta.Slot)
	fmt.Println("├─────────────────────────────────────────────────────────────")
	if create != nil {
		fmt.Printf("│ Name       : %s\n", create.Name)
		fmt.Printf("│ Symbol     : %s\n", create.Symbol)
		fmt.Printf("│ Mint       : %s\n", create.Mint)
		fmt.Printf("│ Creator    : %s\n", create.Creator)
	}
	fmt.Println("├─────────────────────────────────────────────────────────────")
	fmt.Printf("│ 📊 Latency : %d μs\n", latencyUs)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
