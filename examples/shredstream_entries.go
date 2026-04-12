//go:build ignore

// ShredStream gRPC：订阅 SubscribeEntries，对 Entry.entries 做 DecodeGRPCEntry 解码（对齐 sol-parser-sdk-ts `shredstream/client.ts` + examples/shredstream_example.ts）。
//
// 环境变量：
//   SHRED_URL      必填，如 http://127.0.0.1:10800（解析为 host:port，明文 gRPC；与 GRPC_URL 不同服务勿混用）
//   SHRED_MAX_MSG                可选，最大接收字节，默认 1GiB
//   SHREDSTREAM_QUIET=1          关闭每条 gRPC Entry 的 [shredstream] 行（仅保留 10s 统计）
//   SHRED_PRINT_SIGS=N           每条 Entry 额外打印最多 N 笔签名样本（默认 0）
//   SHRED_PARSE_DEX=0            关闭线格式交易 → ParseInstructionUnified → DexEvent JSON（默认开启）
//   SHRED_MAX_JSON_PER_ENTRY=50   每条 gRPC Entry 最多打印多少条 DexEvent JSON（默认 50，防刷屏）
//   SHRED_JSON_COMPACT=1          DexEvent 单行 JSON（默认缩进多行，便于阅读）
//
// 输出：每条 gRPC 消息用分隔线包围；摘要一行对齐；DexEvent 带序号与类型，JSON 默认缩进。
//
// 运行：cd sol-parser-sdk-golang && export SHRED_URL=http://host:10800 && go run examples/shredstream_entries.go
//
// 完整账户+ALT 与主网 RPC 路径见 sol-parser-sdk-ts `shredstream_pumpfun_json.ts`。

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"sol-parser-sdk-golang/shredstream"
	"sol-parser-sdk-golang/solparser"
)

const lineW = 80

func hr(ch byte) string {
	return strings.Repeat(string(ch), lineW)
}

func main() {
	raw := os.Getenv("SHRED_URL")
	ep := dialTargetFromURL(raw)
	if ep == "" {
		fmt.Fprintf(os.Stderr, "请设置 SHRED_URL（ShredStream 端点，与 GRPC_URL 不同），例如: export SHRED_URL=\"http://127.0.0.1:10800\"\n")
		os.Exit(1)
	}

	cfg := shredstream.DefaultShredStreamConfig()
	if s := os.Getenv("SHRED_MAX_MSG"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			cfg.MaxDecodingMessageSize = n
		}
	} else {
		cfg.MaxDecodingMessageSize = 1024 * 1024 * 1024 // 1 GiB
	}

	quiet := os.Getenv("SHREDSTREAM_QUIET") == "1"
	printSigN := uint64(0)
	if s := os.Getenv("SHRED_PRINT_SIGS"); s != "" {
		if n, err := strconv.ParseUint(s, 10, 64); err == nil {
			printSigN = n
		}
	}

	parseDex := os.Getenv("SHRED_PARSE_DEX") != "0"
	jsonCompact := os.Getenv("SHRED_JSON_COMPACT") == "1"
	maxJsonPerEntry := 50
	if s := os.Getenv("SHRED_MAX_JSON_PER_ENTRY"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			maxJsonPerEntry = n
		}
	}

	fmt.Println(hr('='))
	fmt.Println(" ShredStream  ·  SubscribeEntries + DecodeGRPCEntry + DexEvent (static keys)")
	fmt.Println(hr('='))
	fmt.Printf(" Endpoint     %s\n", ep)
	if parseDex {
		mode := "indented"
		if jsonCompact {
			mode = "compact one-line"
		}
		fmt.Printf(" Dex JSON     ON  (%s)  max %d events / gRPC message  (SHRED_PARSE_DEX=0 to off)\n", mode, maxJsonPerEntry)
	} else {
		fmt.Println(" Dex JSON     OFF  (summary + tx sig samples only)")
	}
	fmt.Println(" Quiet        SHREDSTREAM_QUIET=1 hides batch blocks below")
	fmt.Println(hr('-'))
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client, err := shredstream.Dial(ctx, ep, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	stream, err := client.SubscribeEntries(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SubscribeEntries: %v\n", err)
		os.Exit(1)
	}

	var (
		entryMsgs uint64 // gRPC Entry 消息条数（对应 Node grpc_entry_msgs）
		txTotal   uint64 // 解码出的交易总数（对应 Node txs_decoded）
		decodeErrs uint64 // 对应 Node entryDecodeFailures
		dexQueued uint64 // 对应 Node dexEventsQueued（静态账户解析出的 DexEvent 条数）
	)

	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				em := atomic.LoadUint64(&entryMsgs)
				tx := atomic.LoadUint64(&txTotal)
				df := atomic.LoadUint64(&decodeErrs)
				dq := atomic.LoadUint64(&dexQueued)
				fmt.Println()
				fmt.Println(hr('.'))
				fmt.Printf(" [stats / 10s]  grpc_messages=%-8d  txs_decoded=%-10d  decode_fail=%-6d  dex_events_total=%d\n",
					em, tx, df, dq)
				fmt.Println(hr('.'))
				fmt.Println()
			}
		}
	}()

	var batchNO uint64
	for {
		msg, err := stream.Recv()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			fmt.Fprintf(os.Stderr, "recv: %v\n", err)
			break
		}
		atomic.AddUint64(&entryMsgs, 1)
		batchNO++

		rawBytes := msg.GetEntries()
		var solanaEntries uint64
		if ne, err := shredstream.BincodeVecEntryCount(rawBytes); err == nil {
			solanaEntries = ne
		}

		slot, txs, err := shredstream.DecodeGRPCEntry(msg)
		if err != nil {
			atomic.AddUint64(&decodeErrs, 1)
			fmt.Fprintf(os.Stderr, "DecodeGRPCEntry: %v (slot=%d len=%d)\n", err, msg.GetSlot(), len(rawBytes))
			continue
		}
		n := uint64(len(txs))
		atomic.AddUint64(&txTotal, n)

		recvUs := time.Now().UnixMicro()
		var dexInBatch int
		var toPrint []solparser.DexEvent
		if parseDex {
			for ti := range txs {
				tx := txs[ti]
				sig := tx.Signature()
				evs := solparser.DexEventsFromShredTransactionWire(tx.Raw, sig, slot, uint32(ti), nil, recvUs, nil)
				dexInBatch += len(evs)
				toPrint = append(toPrint, evs...)
			}
		}

		if parseDex {
			atomic.AddUint64(&dexQueued, uint64(dexInBatch))
		}

		if !quiet {
			fmt.Println(hr('='))
			fmt.Printf(" Batch #%4d  |  slot %-12d  |  payload %6d B\n", batchNO, slot, len(rawBytes))
			fmt.Printf(" Solana Vec   |  entries %-6d  |  wire txs %-6d  |  dex_events %-4d  (static keys only)\n",
				solanaEntries, n, dexInBatch)
			fmt.Println(hr('-'))
		}

		if parseDex && len(toPrint) > 0 {
			limit := len(toPrint)
			if maxJsonPerEntry > 0 && maxJsonPerEntry < limit {
				limit = maxJsonPerEntry
			}
			if limit < len(toPrint) && !quiet {
				fmt.Fprintf(os.Stderr, " note: printing %d/%d dex events (SHRED_MAX_JSON_PER_ENTRY)\n", limit, len(toPrint))
			}
			for i := 0; i < limit; i++ {
				ev := toPrint[i]
				if !quiet {
					fmt.Printf(" --- [%d/%d]  %s\n", i+1, dexInBatch, ev.Type)
				}
				var b []byte
				var err error
				if jsonCompact {
					b, err = json.Marshal(ev)
				} else {
					b, err = json.MarshalIndent(ev, "", "  ")
				}
				if err != nil {
					continue
				}
				fmt.Println(string(b))
				fmt.Println()
			}
		} else if parseDex && !quiet && dexInBatch == 0 {
			fmt.Println(" (no DexEvent: no matching outer ix, or V0 account indices need ALT + RPC)")
			fmt.Println()
		}

		if !quiet {
			fmt.Println(hr('='))
			fmt.Println()
		}

		if printSigN > 0 {
			fmt.Println(" Tx signatures (sample)")
			for i := range txs {
				if uint64(i) >= printSigN {
					break
				}
				tx := txs[i]
				sig := tx.Signature()
				short := sig
				if len(short) > 36 {
					short = short[:36] + "…"
				}
				fmt.Printf("   [%d] %s  (raw %d B)\n", i, short, len(tx.Raw))
			}
			fmt.Println()
		}
	}

	b := atomic.LoadUint64(&entryMsgs)
	tx := atomic.LoadUint64(&txTotal)
	de := atomic.LoadUint64(&decodeErrs)
	dq := atomic.LoadUint64(&dexQueued)
	fmt.Println(hr('='))
	fmt.Printf(" Stopped.  grpc_messages=%d  txs_decoded=%d  decode_fail=%d  dex_events_total=%d\n", b, tx, de, dq)
	fmt.Println(hr('='))
}

// dialTargetFromURL 将 http(s)://host:port 转为 gRPC Dial 用的 host:port；已是 host:port 则原样返回。
func dialTargetFromURL(raw string) string {
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err == nil && u.Host != "" {
		return u.Host
	}
	return raw
}
