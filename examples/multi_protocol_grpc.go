//go:build ignore

// Multi-Protocol gRPC Example
//
// Subscribe to multiple DEX protocols through SubscribeDexEvents.
//
// Run: go run examples/multi_protocol_grpc.go  (from github.com/0xfnzero/sol-parser-sdk-golang/)

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
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

	fmt.Println("🚀 Multi-Protocol gRPC Example")
	fmt.Println("================================\n")

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

	protocols := []solparser.Protocol{
		solparser.ProtocolPumpFun,
		solparser.ProtocolPumpSwap,
		solparser.ProtocolBonk,
		solparser.ProtocolRaydiumCpmm,
		solparser.ProtocolRaydiumClmm,
		solparser.ProtocolRaydiumAmmV4,
		solparser.ProtocolOrcaWhirlpool,
		solparser.ProtocolMeteoraPools,
		solparser.ProtocolMeteoraDammV2,
		solparser.ProtocolMeteoraDlmm,
	}
	txFilter := solparser.TransactionFilterForProtocols(protocols)
	accountFilter := solparser.AccountFilterForProtocols(protocols)
	voteF := false
	failedF := false
	txFilter.Vote = &voteF
	txFilter.Failed = &failedF

	sub, err := client.SubscribeDexEvents(
		[]solparser.TransactionFilter{txFilter},
		[]solparser.AccountFilter{accountFilter},
		nil,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe failed: %v\n", err)
		os.Exit(1)
	}
	defer sub.Cancel()

	stats := make(map[string]int)
	statsMu := sync.RWMutex{}

	go func() {
		for range time.Tick(30 * time.Second) {
			statsMu.RLock()
			if len(stats) == 0 {
				statsMu.RUnlock()
				continue
			}
			snapshot := make(map[string]int, len(stats))
			for k, v := range stats {
				snapshot[k] = v
			}
			statsMu.RUnlock()

			fmt.Println("\n📊 Event Statistics:")
			keys := make([]string, 0, len(snapshot))
			for k := range snapshot {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				return snapshot[keys[i]] > snapshot[keys[j]]
			})
			for _, k := range keys {
				fmt.Printf("  %-35s: %d\n", k, snapshot[k])
			}
			fmt.Println()
		}
	}()

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
			key := string(ev.Type)
			statsMu.Lock()
			stats[key]++
			statsMu.Unlock()

			meta := ev.GetMetadata()
			data, _ := json.Marshal(ev)
			s := string(data)
			if len(s) > 240 {
				s = s[:240] + "..."
			}
			fmt.Printf("[%s] %s | slot=%d sig=%s\n", key, s, meta.Slot, shortSig(meta.Signature))
			fmt.Println()
		case err, ok := <-sub.Errors:
			if ok {
				fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
			}
		case <-interrupt:
			printFinalStats(stats, &statsMu)
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

func printFinalStats(stats map[string]int, statsMu *sync.RWMutex) {
	statsMu.RLock()
	defer statsMu.RUnlock()
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
