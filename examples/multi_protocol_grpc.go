//go:build ignore

// Multi-Protocol gRPC Example
//
// Subscribe to multiple DEX protocols simultaneously:
// PumpFun, PumpSwap, Raydium, Orca, Meteora, Bonk
//
// Run: go run examples/multi_protocol_grpc.go  (from sol-parser-sdk-golang/)

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	base58 "github.com/mr-tron/base58"
	solparser "sol-parser-sdk-golang/solparser"
)

var allProgramIDs = []string{
	solparser.PUMPFUN_PROGRAM_ID,
	solparser.PUMPSWAP_PROGRAM_ID,
	solparser.GrpcPumpSwapFeesProgramID,
	solparser.RAYDIUM_AMM_V4_PROGRAM_ID,
	solparser.GrpcRaydiumClmmProgramID,
	solparser.RAYDIUM_CPMM_PROGRAM_ID,
	solparser.ORCA_WHIRLPOOL_PROGRAM_ID,
	solparser.GrpcMeteoraDammV2ProgramID,
	solparser.METEORA_DLMM_PROGRAM_ID,
	solparser.METEORA_POOLS_PROGRAM_ID,
	solparser.GrpcBonkProgramID,
}

func main() {
	endpoint := os.Getenv("GEYSER_ENDPOINT")
	if endpoint == "" {
		endpoint = "solana-yellowstone-grpc.publicnode.com:443"
	}
	token := os.Getenv("GEYSER_API_TOKEN")

	fmt.Println("🚀 Multi-Protocol gRPC Example")
	fmt.Println("================================\n")

	client := solparser.NewYellowstoneGrpc(endpoint)
	if token != "" {
		client.SetXToken(token)
	}
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connect failed: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	stats := make(map[string]int)

	// Print stats every 30s
	go func() {
		for range time.Tick(30 * time.Second) {
			if len(stats) == 0 {
				continue
			}
			fmt.Println("\n📊 Event Statistics:")
			keys := make([]string, 0, len(stats))
			for k := range stats {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				return stats[keys[i]] > stats[keys[j]]
			})
			for _, k := range keys {
				fmt.Printf("  %-35s: %d\n", k, stats[k])
			}
			fmt.Println()
		}
	}()

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude:  allProgramIDs,
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
					stats[key]++
					data, _ := json.Marshal(map[string]any{key: val})
					s := string(data)
					if len(s) > 200 {
						s = s[:200] + "..."
					}
					fmt.Printf("[%s] %s | slot=%d sig=%s...\n", key, s, slot, shortSig)
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
	fmt.Println("\n📊 Final Event Statistics:")
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return stats[keys[i]] > stats[keys[j]]
	})
	for _, k := range keys {
		fmt.Printf("  %-35s: %d\n", k, stats[k])
	}
}
