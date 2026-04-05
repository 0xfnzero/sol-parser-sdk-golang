//go:build ignore

// PumpFun Trade Event Filter Example
//
// Demonstrates how to:
// - Subscribe to PumpFun protocol events
// - Filter specific trade types: Buy, Sell, BuyExactSolIn, Create
// - Display trade details with latency metrics
//
// Run: go run examples/pumpfun_trade_filter.go  (from sol-parser-sdk-golang/)
// Or:  GEYSER_ENDPOINT=xxx GEYSER_API_TOKEN=yyy go run examples/pumpfun_trade_filter.go

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	base58 "github.com/mr-tron/base58"
	solparser "sol-parser-sdk-golang/solparser"
)

const defaultEndpoint = "solana-yellowstone-grpc.publicnode.com:443"
const defaultToken = ""

var pumpFunProgramIDs = []string{
	"6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P", // PumpFun
}

func nowUs() int64 {
	return time.Now().UnixMicro()
}

func main() {
	endpoint := os.Getenv("GEYSER_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	token := os.Getenv("GEYSER_API_TOKEN")
	if token == "" {
		token = defaultToken
	}

	fmt.Println("🚀 PumpFun Trade Event Filter Example")
	fmt.Println("======================================\n")
	fmt.Printf("📡 Endpoint: %s\n", endpoint)
	fmt.Printf("🎯 Program: %s\n\n", pumpFunProgramIDs[0])

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
		eventCount    int
		buyCount      int
		sellCount     int
		buyExactCount int
		createCount   int
	)

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude:  pumpFunProgramIDs,
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
			if len(sigStr) > 16 {
				sigStr = sigStr[:16]
			}
			slot := update.Transaction.Slot
			queueRecvUs := nowUs()

			events := solparser.ParseLogsOnly(logs, sigStr+"...", slot, nil)

			for _, ev := range events {
				for key := range ev {
					data := ev[key]
					dataMap, _ := data.(map[string]any)
					eventCount++

					var grpcRecvUs int64
					if md, ok := dataMap["metadata"].(map[string]any); ok {
						if v, ok := md["grpc_recv_us"].(float64); ok {
							grpcRecvUs = int64(v)
						}
					}
					latencyUs := queueRecvUs - grpcRecvUs

					switch key {
					case "PumpFunBuy":
						buyCount++
						fmt.Println("┌─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 🟢 PumpFun BUY #%d\n", eventCount)
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ Slot       : %d\n", slot)
						printField(dataMap, "mint", "Mint")
						printField(dataMap, "sol_amount", "SOL Amount")
						printField(dataMap, "token_amount", "Token Amt")
						printField(dataMap, "user", "User")
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 📊 Latency : %d μs\n", latencyUs)
						fmt.Printf("│ 📊 Stats   : Buy=%d Sell=%d BuyExact=%d\n", buyCount, sellCount, buyExactCount)
						fmt.Println("└─────────────────────────────────────────────────────────────\n")

					case "PumpFunBuyExactSolIn":
						buyExactCount++
						fmt.Println("┌─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 🟡 PumpFun BUY_EXACT_SOL_IN #%d\n", eventCount)
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ Slot       : %d\n", slot)
						printField(dataMap, "mint", "Mint")
						printField(dataMap, "sol_amount", "SOL Amount")
						printField(dataMap, "user", "User")
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 📊 Latency : %d μs\n", latencyUs)
						fmt.Printf("│ 📊 Stats   : Buy=%d Sell=%d BuyExact=%d\n", buyCount, sellCount, buyExactCount)
						fmt.Println("└─────────────────────────────────────────────────────────────\n")

					case "PumpFunSell":
						sellCount++
						fmt.Println("┌─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 🔴 PumpFun SELL #%d\n", eventCount)
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ Slot       : %d\n", slot)
						printField(dataMap, "mint", "Mint")
						printField(dataMap, "sol_amount", "SOL Amount")
						printField(dataMap, "user", "User")
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 📊 Latency : %d μs\n", latencyUs)
						fmt.Printf("│ 📊 Stats   : Buy=%d Sell=%d BuyExact=%d\n", buyCount, sellCount, buyExactCount)
						fmt.Println("└─────────────────────────────────────────────────────────────\n")

					case "PumpFunCreate":
						createCount++
						fmt.Println("┌─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 🆕 PumpFun CREATE #%d\n", eventCount)
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ Slot       : %d\n", slot)
						printField(dataMap, "name", "Name")
						printField(dataMap, "symbol", "Symbol")
						printField(dataMap, "mint", "Mint")
						printField(dataMap, "creator", "Creator")
						fmt.Println("├─────────────────────────────────────────────────────────────")
						fmt.Printf("│ 📊 Latency : %d μs\n", latencyUs)
						fmt.Printf("│ 📊 Creates : %d\n", createCount)
						fmt.Println("└─────────────────────────────────────────────────────────────\n")

					default:
						b, _ := json.Marshal(ev)
						fmt.Printf("[%s] %s\n\n", key, string(b)[:min(len(string(b)), 300)])
					}
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
	fmt.Printf("\n👋 Total events: %d (Buy=%d Sell=%d BuyExact=%d Create=%d)\n",
		eventCount, buyCount, sellCount, buyExactCount, createCount)
}

func printField(m map[string]any, key, label string) {
	if v, ok := m[key]; ok {
		fmt.Printf("│ %-11s: %v\n", label, v)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
