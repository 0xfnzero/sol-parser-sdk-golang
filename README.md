<div align="center">
    <h1>⚡ Sol Parser SDK - Go</h1>
    <h3><em>High-performance Solana DEX event parser for Go</em></h3>
</div>

<p align="center">
    <strong>High-performance Go library for parsing Solana DEX events in real-time via Yellowstone gRPC</strong>
</p>

<p align="center">
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang">
        <img src="https://img.shields.io/badge/go-sol--parser--sdk--golang-00ADD8.svg" alt="Go">
    </a>
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/LICENSE">
        <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License">
    </a>
</p>

<p align="center">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/Solana-9945FF?style=for-the-badge&logo=solana&logoColor=white" alt="Solana">
    <img src="https://img.shields.io/badge/gRPC-4285F4?style=for-the-badge&logo=grpc&logoColor=white" alt="gRPC">
</p>

<p align="center">
    <a href="./README_CN.md">中文</a> |
    <a href="./README.md">English</a> |
    <a href="https://fnzero.dev/">Website</a> |
    <a href="https://t.me/fnzero_group">Telegram</a> |
    <a href="https://discord.gg/vuazbGkqQE">Discord</a>
</p>

<p align="center">
    <a href="https://github.com/0xfnzero/sol-parser-sdk">Rust</a> |
    <a href="https://github.com/0xfnzero/sol-parser-sdk-nodejs">Node.js</a> |
    <a href="https://github.com/0xfnzero/sol-parser-sdk-python">Python</a> |
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang"><strong>Go</strong></a>
</p>

---

## 📊 Performance Highlights

### ⚡ Real-Time Parsing
- **Sub-millisecond** log-based event parsing
- **gRPC streaming** with Yellowstone/Geyser protocol
- **Concurrent-safe** with goroutine support
- **Event type filtering** for targeted parsing
- **Zero-allocation** parsing on hot paths where possible

### 🎚️ Flexible Order Modes
| Mode | Latency | Description |
|------|---------|-------------|
| **Unordered** | <1ms | Immediate output, ultra-low latency |
| **MicroBatch** | 1-5ms | Micro-batch ordering with time window |
| **StreamingOrdered** | 5-20ms | Stream ordering with continuous sequence release |
| **Ordered** | 10-100ms | Full slot ordering, wait for complete slot |

### 🚀 Optimization Highlights
- ✅ **Concurrent-safe** atomic counters and goroutine-based stats
- ✅ **Optimized pattern matching** for protocol detection
- ✅ **Event type filtering** for targeted parsing
- ✅ **Conditional Create detection** (only when needed)
- ✅ **Multiple order modes** for latency vs ordering trade-off
- ✅ **Efficient memory usage** with buffer pooling

---

## 🔥 Quick Start

### Installation

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

### Use Go Modules

```bash
go get github.com/0xfnzero/sol-parser-sdk-golang
```

### Performance Testing

Test parsing with the optimized examples:

```bash
# PumpFun trade filter (Buy/Sell/BuyExactSolIn/Create)
GEYSER_API_TOKEN=your_token go run examples/pumpfun_trade_filter.go

# PumpSwap low-latency with performance metrics
GEYSER_API_TOKEN=your_token go run examples/pumpswap_low_latency.go

# All protocols simultaneously
GEYSER_API_TOKEN=your_token go run examples/multi_protocol_grpc.go

# Expected output:
# gRPC接收时间: 1234567890 μs
# 事件接收时间: 1234567900 μs
# 延迟时间: 10 μs  <-- Ultra-low latency!
```

### Examples

| Description | Command | Source Code |
|-------------|---------|-------------|
| **PumpFun** | | |
| PumpFun trade filtering (Buy/Sell/BuyExactSolIn/Create) with latency metrics | `go run examples/pumpfun_trade_filter.go` | [examples/pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) |
| Quick PumpFun connection test (first 10 events) | `go run examples/pumpfun_quick_test.go` | [examples/pumpfun_quick_test.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_quick_test.go) |
| **PumpSwap** | | |
| PumpSwap ultra-low latency with per-event + 10s stats | `go run examples/pumpswap_low_latency.go` | [examples/pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) |
| PumpSwap events with performance metrics | `go run examples/pumpswap_with_metrics.go` | [examples/pumpswap_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_with_metrics.go) |
| **Meteora DAMM** | | |
| Meteora DAMM V2 (Swap/AddLiquidity/RemoveLiquidity/CreatePosition/ClosePosition) | `go run examples/meteora_damm_grpc.go` | [examples/meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) |
| **Multi-Protocol** | | |
| Subscribe to all DEX protocols simultaneously | `go run examples/multi_protocol_grpc.go` | [examples/multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) |

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "sol-parser-sdk-golang/solparser"
    "github.com/mr-tron/base58"
)

func main() {
    endpoint := "solana-yellowstone-grpc.publicnode.com:443"
    token := os.Getenv("GEYSER_API_TOKEN")

    client := solparser.NewGrpcClient(endpoint, token)
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    filter := &solparser.TransactionFilter{
        AccountInclude: []string{
            "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P", // PumpFun
            "pAMMBay6oceH9fJKBRdGP4LmT4saRGfEE7xmrCaGWpZ", // PumpSwap
        },
        Vote:   false,
        Failed: false,
    }

    err := client.SubscribeTransactions(context.Background(), filter, func(update *solparser.TransactionUpdate) {
        txInfo := update.Transaction
        if txInfo == nil {
            return
        }

        sigStr := base58.Encode(txInfo.Signature)
        logs := txInfo.LogMessages

        events, err := solparser.ParseLogsOnly(logs, sigStr, update.Slot, nil)
        if err != nil || len(events) == 0 {
            return
        }

        for _, ev := range events {
            fmt.Printf("[%s] %+v\n", ev.EventType(), ev)
        }
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## 🏗️ Supported Protocols

### DEX Protocols
- ✅ **PumpFun** - Meme coin trading
- ✅ **PumpSwap** - PumpFun swap protocol
- ✅ **Raydium AMM V4** - Automated Market Maker
- ✅ **Raydium CLMM** - Concentrated Liquidity
- ✅ **Raydium CPMM** - Concentrated Pool
- ✅ **Orca Whirlpool** - Concentrated liquidity AMM
- ✅ **Meteora DAMM V2** - Dynamic AMM
- ✅ **Meteora DLMM** - Dynamic Liquidity Market Maker
- ✅ **Bonk Launchpad** - Token launch platform

### Event Types
Each protocol supports:
- 📈 **Trade/Swap Events** - Buy/sell transactions
- 💧 **Liquidity Events** - Deposits/withdrawals
- 🏊 **Pool Events** - Pool creation/initialization
- 🎯 **Position Events** - Open/close positions (CLMM)

---

## ⚡ Performance Features

### Optimized Pattern Matching
```go
import "strings"

// Pre-defined protocol identifiers
const PumpFunProgram = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"

// Fast check before full parsing
if strings.Contains(logString, PumpFunProgram) {
    return parsePumpFunEvent(logs, signature, slot)
}
```

### Event Type Filtering
```go
// Filter specific event types for targeted parsing
eventFilter := &solparser.EventTypeFilter{
    IncludeOnly: []solparser.EventType{
        solparser.EventTypePumpFunTrade,
        solparser.EventTypePumpSwapBuy,
    },
}
```

### Concurrent Stats with Atomic Counters
```go
import "sync/atomic"

var totalEvents int64

// In callback:
atomic.AddInt64(&totalEvents, int64(len(events)))

// In goroutine:
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        count := atomic.LoadInt64(&totalEvents)
        fmt.Printf("Total events: %d\n", count)
    }
}()
```

---

## 🎯 Event Filtering

Reduce processing overhead by filtering specific events:

### Example: Trading Bot
```go
eventFilter := &solparser.EventTypeFilter{
    IncludeOnly: []solparser.EventType{
        solparser.EventTypePumpFunTrade,
        solparser.EventTypeRaydiumAmmV4Swap,
        solparser.EventTypeRaydiumClmmSwap,
        solparser.EventTypeOrcaWhirlpoolSwap,
    },
}
```

### Example: Pool Monitor
```go
eventFilter := &solparser.EventTypeFilter{
    IncludeOnly: []solparser.EventType{
        solparser.EventTypePumpFunCreate,
        solparser.EventTypePumpSwapCreatePool,
    },
}
```

**Performance Impact:**
- 60-80% reduction in processing
- Lower memory usage
- Reduced network bandwidth

---

## 🔧 Advanced Features

### Create+Buy Detection
Automatically detects when a token is created and immediately bought in the same transaction:

```go
// Automatically detects "Program data: GB7IKAUcB3c..." pattern
events, err := solparser.ParseLogsOnly(logs, signature, slot, nil)

// Sets IsCreatedBuy flag on Trade events
for _, ev := range events {
    if trade, ok := ev.(*solparser.PumpFunTrade); ok && trade.IsCreatedBuy {
        fmt.Println("Create+Buy detected!")
    }
}
```

### Custom gRPC Endpoint

```go
endpoint := os.Getenv("GEYSER_ENDPOINT")
if endpoint == "" {
    endpoint = "solana-yellowstone-grpc.publicnode.com:443"
}
token := os.Getenv("GEYSER_API_TOKEN")
client := solparser.NewGrpcClient(endpoint, token)
```

### Unsubscribe

```go
// Context-based cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := client.SubscribeTransactions(ctx, filter, callback)
```

---

## 📁 Project Structure

```
sol-parser-sdk-golang/
├── solparser/
│   ├── grpc_client.go      # GrpcClient (connect, subscribe, auth)
│   ├── parser.go           # ParseLogsOnly, ParseTransactionEvents
│   ├── types.go            # DexEvent, TransactionFilter, TransactionUpdate
│   └── ...                 # Protocol-specific parsers
├── proto/
│   ├── geyser.proto        # Yellowstone gRPC proto
│   └── generated/          # Generated Go proto files
├── examples/
│   ├── pumpfun_trade_filter.go
│   ├── pumpfun_quick_test.go
│   ├── pumpswap_low_latency.go
│   ├── pumpswap_with_metrics.go
│   ├── meteora_damm_grpc.go
│   └── multi_protocol_grpc.go
├── go.mod
└── go.sum
```

---

## 🚀 Optimization Techniques

### 1. **Concurrent-Safe Design**
- Atomic counters for stats
- Goroutine-safe callbacks
- Lock-free event delivery where possible

### 2. **Optimized Pattern Matching**
- Pre-defined protocol identifiers
- Fast string matching with strings.Contains
- Minimal string allocations

### 3. **Event Type Filtering**
- Early filtering at protocol level
- Conditional Create detection
- Single-type ultra-fast path

### 4. **Efficient Memory Usage**
- Buffer pooling where possible
- Minimal heap allocations
- Reusable buffers for parsing

### 5. **Context Support**
- Graceful cancellation
- Timeout handling
- Resource cleanup

---

## 📄 License

MIT License

## 📞 Contact

- **Repository**: https://github.com/0xfnzero/sol-parser-sdk-golang
- **Website**: https://fnzero.dev/
- **Telegram**: https://t.me/fnzero_group
- **Discord**: https://discord.gg/vuazbGkqQE

---

## ⚠️ Performance Tips

1. **Use Event Filtering** — Filter by program ID for 60-80% performance gain
2. **Run with race detector** — `go run -race` to verify concurrent safety
3. **Monitor goroutines** — Keep track of goroutine count in production
4. **Use atomic counters** — For thread-safe statistics
5. **Tune buffer sizes** — Adjust channel buffers based on throughput

## 🔬 Development

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Build
go build ./...

# Format code
go fmt ./...

# Vet code
go vet ./...
```
