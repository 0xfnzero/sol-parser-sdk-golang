//go:build ignore

// ShredStream gRPC：订阅 SubscribeEntries，对 Entry.entries 做 DecodeGRPCEntry 解码（对齐 Rust shredstream_example / solana-streamer）。
//
// 环境变量：
//   SHRED_URL      必填，如 http://127.0.0.1:10800 或 http://70.40.184.37:10800（会解析为 host:port，明文 gRPC）
//   SHRED_MAX_MSG  可选，接收消息最大字节，默认 1073741824（1GiB，与 Rust 示例接近）
//
// 运行（在 sol-parser-sdk-golang 目录下）：
//
//	export SHRED_URL="http://70.40.184.37:10800"
//	go run examples/shredstream_entries.go
//
// 说明：与 Rust 文档一致——仅静态账户、无 inner instructions、无 block_time；本示例只演示解码与签名打印。

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"sol-parser-sdk-golang/shredstream"
)

func main() {
	raw := os.Getenv("SHRED_URL")
	if raw == "" {
		raw = os.Getenv("SHRED_GRPC_ADDR")
	}
	ep := dialTargetFromURL(raw)
	if ep == "" {
		fmt.Fprintf(os.Stderr, "请设置 SHRED_URL，例如: export SHRED_URL=\"http://127.0.0.1:10800\"\n")
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

	fmt.Println("ShredStream SubscribeEntries + DecodeGRPCEntry")
	fmt.Println("==============================================")
	fmt.Printf("dial target: %s\n\n", ep)

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
		batches    uint64
		txTotal    uint64
		decodeErrs uint64
		printCap   = uint64(5) // 仅打印前几条样本
		printed    uint64
	)

	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				b := atomic.LoadUint64(&batches)
				tx := atomic.LoadUint64(&txTotal)
				de := atomic.LoadUint64(&decodeErrs)
				fmt.Printf("[stats] batches=%d txs=%d decode_errors=%d\n", b, tx, de)
			}
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			fmt.Fprintf(os.Stderr, "recv: %v\n", err)
			break
		}
		atomic.AddUint64(&batches, 1)

		slot, txs, err := shredstream.DecodeGRPCEntry(msg)
		if err != nil {
			atomic.AddUint64(&decodeErrs, 1)
			fmt.Fprintf(os.Stderr, "DecodeGRPCEntry: %v (slot=%d len=%d)\n", err, msg.GetSlot(), len(msg.GetEntries()))
			continue
		}
		n := uint64(len(txs))
		atomic.AddUint64(&txTotal, n)

		for _, tx := range txs {
			if atomic.LoadUint64(&printed) >= printCap {
				break
			}
			atomic.AddUint64(&printed, 1)
			sig := tx.Signature()
			short := sig
			if len(short) > 20 {
				short = short[:20] + "…"
			}
			fmt.Printf("sample slot=%d sig=%s raw_len=%d\n", slot, short, len(tx.Raw))
		}
	}

	b := atomic.LoadUint64(&batches)
	tx := atomic.LoadUint64(&txTotal)
	de := atomic.LoadUint64(&decodeErrs)
	fmt.Printf("\n done: batches=%d txs=%d decode_errors=%d\n", b, tx, de)
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
