//go:build ignore

// Meteora DAMM V2 gRPC Example
//
// Demonstrates subscribing to Meteora DAMM V2 events:
// Swap, AddLiquidity, RemoveLiquidity, CreatePosition, ClosePosition
//
// Run: go run examples/meteora_damm_grpc.go  (from sol-parser-sdk-golang/)

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	base58 "github.com/mr-tron/base58"
	solparser "sol-parser-sdk-golang/solparser"
)

var meteoraProgramIDs = []string{
	"Eo7WjKq67rjJQDd1d4dSYkT7LeHVAaFL1K7dajEgrpwz", // Meteora DAMM V2
}

func main() {
	endpoint := os.Getenv("GEYSER_ENDPOINT")
	if endpoint == "" {
		endpoint = "solana-yellowstone-grpc.publicnode.com:443"
	}
	token := os.Getenv("GEYSER_API_TOKEN")

	fmt.Println("🚀 Meteora DAMM V2 gRPC Example")
	fmt.Println("=================================\n")

	client := solparser.NewYellowstoneGrpc(endpoint)
	if token != "" {
		client.SetXToken(token)
	}
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connect failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	var swapCount, addLiqCount, removeLiqCount, createPosCount, closePosCount int

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude:  meteoraProgramIDs,
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
			shortSig := sigStr
			if len(shortSig) > 16 {
				shortSig = shortSig[:16]
			}
			slot := update.Transaction.Slot

			events := solparser.ParseLogsOnly(logs, sigStr, slot, nil)
			for _, ev := range events {
				for key, val := range ev {
					data, _ := val.(map[string]any)
					switch key {
					case "MeteoraDammV2Swap":
						swapCount++
						fmt.Printf("🔄 SWAP #%d | sig=%s... slot=%d\n", swapCount, shortSig, slot)
						if v, ok := data["amount_in"]; ok {
							fmt.Printf("   amount_in=%v amount_out=%v\n", v, data["amount_out"])
						}
					case "MeteoraDammV2AddLiquidity":
						addLiqCount++
						fmt.Printf("💧 ADD_LIQUIDITY #%d | sig=%s... slot=%d\n", addLiqCount, shortSig, slot)
					case "MeteoraDammV2RemoveLiquidity":
						removeLiqCount++
						fmt.Printf("🔥 REMOVE_LIQUIDITY #%d | sig=%s... slot=%d\n", removeLiqCount, shortSig, slot)
					case "MeteoraDammV2CreatePosition":
						createPosCount++
						fmt.Printf("📌 CREATE_POSITION #%d | sig=%s... slot=%d\n", createPosCount, shortSig, slot)
					case "MeteoraDammV2ClosePosition":
						closePosCount++
						fmt.Printf("❌ CLOSE_POSITION #%d | sig=%s... slot=%d\n", closePosCount, shortSig, slot)
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
	fmt.Printf("\n📊 Stats: Swap=%d AddLiq=%d RemoveLiq=%d CreatePos=%d ClosePos=%d\n",
		swapCount, addLiqCount, removeLiqCount, createPosCount, closePosCount)
}
