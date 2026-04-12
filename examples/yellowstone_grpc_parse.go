//go:build ignore

// Yellowstone gRPC（Geyser Subscribe）：与 Rust `pumpswap_with_metrics` / `pumpfun_with_metrics` 打印风格一致：
// 分隔线、gRPC/事件时间与延迟、完整缩进 JSON（不截断）、末尾整段 signature。
// 解析使用 ParseSubscribeTransaction（指令账户 + 日志 Program data，并合并 PumpSwap 重复事件以补全 mint/池子字段）。
//
// 环境变量（与常见部署一致）：
//   GRPC_URL       如 https://solana-yellowstone-grpc.publicnode.com:443（可写 host:port）
//   GRPC_TOKEN     x-token（可与 GEYSER_API_TOKEN 二选一）
//   GEYSER_ENDPOINT / GEYSER_API_TOKEN  若未设置 GRPC_* 则回退
//
// 运行：
//
//	export GRPC_URL="https://solana-yellowstone-grpc.publicnode.com:443"
//	export GRPC_TOKEN="your_token"
//	go run examples/yellowstone_grpc_parse.go
//
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	solparser "sol-parser-sdk-golang/solparser"
)

// eventEnvelope 与 Rust `println!("{:?}", event)` 对应：完整 JSON，便于人工阅读（不截断）。
type eventEnvelope struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// grpcRecvUsFromEventData 从各事件 Data 内嵌的 metadata 提取 grpc_recv_us（与 Rust 延迟统计一致）。
func grpcRecvUsFromEventData(data any) int64 {
	b, err := json.Marshal(data)
	if err != nil {
		return 0
	}
	var aux struct {
		Metadata struct {
			GrpcRecvUs int64 `json:"grpc_recv_us"`
		} `json:"metadata"`
	}
	if json.Unmarshal(b, &aux) != nil {
		return 0
	}
	return aux.Metadata.GrpcRecvUs
}

func printEventLikeRust(ev solparser.DexEvent, slot uint64, signatureFull string) {
	if ev.Type == "" {
		return
	}
	key := string(ev.Type)
	queueRecvUs := time.Now().UnixMicro()
	grpcUs := grpcRecvUsFromEventData(ev.Data)

	fmt.Println()
	fmt.Println("================================================")
	if grpcUs > 0 {
		latency := queueRecvUs - grpcUs
		fmt.Printf("gRPC接收时间: %d μs\n", grpcUs)
		fmt.Printf("事件接收时间: %d μs\n", queueRecvUs)
		fmt.Printf("延迟时间: %d μs\n", latency)
	} else {
		fmt.Println("gRPC接收时间: (n/a，解析路径未写入 metadata.grpc_recv_us)")
		fmt.Printf("事件接收时间: %d μs\n", queueRecvUs)
	}
	fmt.Println("队列长度: (n/a)")
	fmt.Println("================================================")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(eventEnvelope{Type: key, Data: ev.Data}); err != nil {
		fmt.Printf("%s: %+v\n", key, ev.Data)
	} else {
		// Encode 末尾自带换行，去掉多余空行
		out := bytes.TrimSpace(buf.Bytes())
		fmt.Println(string(out))
	}
	fmt.Printf("\nslot=%d\nsignature=%s\n", slot, signatureFull)
}

func main() {
	endpoint := grpcEndpointFromEnv()
	token := firstNonEmpty(os.Getenv("GRPC_TOKEN"), os.Getenv("GEYSER_API_TOKEN"))

	fmt.Println("Yellowstone gRPC → ParseSubscribeTransaction")
	fmt.Println("=================================")
	fmt.Printf("endpoint: %s\n", endpoint)
	if token != "" {
		fmt.Println("token: (set)")
	} else {
		fmt.Println("warning: no GRPC_TOKEN / GEYSER_API_TOKEN — 部分节点会拒绝连接")
	}
	fmt.Println()

	client := solparser.NewYellowstoneGrpc(endpoint)
	if token != "" {
		client.SetXToken(token)
	}
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Disconnect()

	stats := make(map[string]int)

	go func() {
		for range time.Tick(30 * time.Second) {
			if len(stats) == 0 {
				continue
			}
			fmt.Println("\n📊 Event stats (30s):")
			keys := make([]string, 0, len(stats))
			for k := range stats {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool { return stats[keys[i]] > stats[keys[j]] })
			for _, k := range keys {
				fmt.Printf("  %-35s: %d\n", k, stats[k])
			}
			fmt.Println()
		}
	}()

	voteF := false
	failedF := false
	filter := solparser.TransactionFilter{
		AccountInclude: []string{
			solparser.PUMPFUN_PROGRAM_ID,
			solparser.PUMPSWAP_PROGRAM_ID,
			solparser.GrpcPumpSwapFeesProgramID,
		},
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
			info := update.Transaction.Transaction
			if info.IsVote {
				return
			}
			if info.Transaction == nil || info.Transaction.Message == nil {
				return
			}
			slot := update.Transaction.Slot
			events, perr := solparser.ParseSubscribeTransaction(slot, info, nil, time.Now().UnixMicro())
			if perr != nil {
				fmt.Fprintf(os.Stderr, "ParseSubscribeTransaction: %v\n", perr)
				return
			}
			for _, ev := range events {
				if ev.Type == "" {
					continue
				}
				key := string(ev.Type)
				stats[key]++
				sigStr := ev.GetMetadata().Signature
				printEventLikeRust(ev, slot, sigStr)
			}
		},
		OnError: func(err error) {
			fmt.Fprintf(os.Stderr, "stream: %v\n", err)
		},
		OnEnd: func() {
			select {
			case <-done:
			default:
				close(done)
			}
		},
	}

	sub, err := client.SubscribeTransactions(filter, callbacks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Subscribe: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("subscribed id=%s (PumpFun + PumpSwap programs)\n", sub.ID)
	fmt.Println("Ctrl+C to stop")
	fmt.Println()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-done:
	case <-interrupt:
	}
	client.Unsubscribe(sub.ID)

	fmt.Println("\nFinal stats:")
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return stats[keys[i]] > stats[keys[j]] })
	for _, k := range keys {
		fmt.Printf("  %-35s: %d\n", k, stats[k])
	}
}

func grpcEndpointFromEnv() string {
	s := firstNonEmpty(os.Getenv("GRPC_URL"), os.Getenv("GEYSER_ENDPOINT"))
	if s == "" {
		return "solana-yellowstone-grpc.publicnode.com:443"
	}
	u, err := url.Parse(s)
	if err == nil && u.Host != "" {
		return u.Host
	}
	return s
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
