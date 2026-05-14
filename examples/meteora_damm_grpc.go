//go:build ignore

// Meteora DAMM V2 gRPC Example
//
// Demonstrates subscribing to Meteora DAMM V2 events through SubscribeDexEvents:
// Swap, AddLiquidity, RemoveLiquidity, CreatePosition, ClosePosition
//
// Run: go run examples/meteora_damm_grpc.go  (from github.com/0xfnzero/sol-parser-sdk-golang/)

package main

import (
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

	fmt.Println("🚀 Meteora DAMM V2 gRPC Example")
	fmt.Println("=================================\n")

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

	protocols := []solparser.Protocol{solparser.ProtocolMeteoraDammV2}
	txFilter := solparser.TransactionFilterForProtocols(protocols)
	voteF := false
	failedF := false
	txFilter.Vote = &voteF
	txFilter.Failed = &failedF
	eventFilter := solparser.EventTypeFilterIncludeOnly([]solparser.EventType{
		solparser.EventTypeMeteoraDammV2Swap,
		solparser.EventTypeMeteoraDammV2AddLiquidity,
		solparser.EventTypeMeteoraDammV2RemoveLiquidity,
		solparser.EventTypeMeteoraDammV2CreatePosition,
		solparser.EventTypeMeteoraDammV2ClosePosition,
	})

	sub, err := client.SubscribeDexEvents([]solparser.TransactionFilter{txFilter}, nil, eventFilter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", err)
		os.Exit(1)
	}
	defer sub.Cancel()

	var swapCount, addLiqCount, removeLiqCount, createPosCount, closePosCount int

	fmt.Printf("✅ Subscribed (id=%s)\n", sub.ID)
	fmt.Printf("📊 Protocols: %v\n", protocols)
	fmt.Println("🛑 Press Ctrl+C to stop...\n")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case ev, ok := <-sub.Events:
			if !ok {
				return
			}
			meta := ev.GetMetadata()
			switch ev.Type {
			case solparser.EventTypeMeteoraDammV2Swap:
				swapCount++
				fmt.Printf("🔄 SWAP #%d | sig=%s slot=%d\n", swapCount, shortSig(meta.Signature), meta.Slot)
				if d := ev.AsMeteoraDammV2Swap(); d != nil {
					fmt.Printf("   amount_in=%d amount_out=%d pool=%s\n", d.AmountIn, d.OutputAmount, d.Pool)
				}
			case solparser.EventTypeMeteoraDammV2AddLiquidity:
				addLiqCount++
				fmt.Printf("💧 ADD_LIQUIDITY #%d | sig=%s slot=%d\n", addLiqCount, shortSig(meta.Signature), meta.Slot)
			case solparser.EventTypeMeteoraDammV2RemoveLiquidity:
				removeLiqCount++
				fmt.Printf("🔥 REMOVE_LIQUIDITY #%d | sig=%s slot=%d\n", removeLiqCount, shortSig(meta.Signature), meta.Slot)
			case solparser.EventTypeMeteoraDammV2CreatePosition:
				createPosCount++
				fmt.Printf("📌 CREATE_POSITION #%d | sig=%s slot=%d\n", createPosCount, shortSig(meta.Signature), meta.Slot)
			case solparser.EventTypeMeteoraDammV2ClosePosition:
				closePosCount++
				fmt.Printf("❌ CLOSE_POSITION #%d | sig=%s slot=%d\n", closePosCount, shortSig(meta.Signature), meta.Slot)
			}
		case err, ok := <-sub.Errors:
			if ok {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		case <-interrupt:
			fmt.Printf("\n📊 Stats: Swap=%d AddLiq=%d RemoveLiq=%d CreatePos=%d ClosePos=%d\n",
				swapCount, addLiqCount, removeLiqCount, createPosCount, closePosCount)
			return
		}
	}
}

func shortSig(sig string) string {
	if len(sig) <= 16 {
		return sig
	}
	return sig[:16] + "..."
}
