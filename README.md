<div align="center">
    <h1>⚡ Sol Parser SDK - Go</h1>
    <h3><em>High-performance Solana DEX event parser for Go</em></h3>
</div>

<p align="center">
    <strong>Go library for parsing Solana DEX events in real-time via Yellowstone gRPC</strong>
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

---

## 📊 Performance Highlights

### ⚡ Real-Time Parsing
- **Zero-latency** log-based event parsing
- **gRPC streaming** with Yellowstone/Geyser protocol
- **Multi-protocol** support in a single subscription
- **Concurrent-safe** atomic counters and goroutine-based stats

### 🏗️ Supported Protocols
- ✅ **PumpFun** - Meme coin trading
- ✅ **PumpSwap** - PumpFun swap protocol
- ✅ **Raydium AMM V4** - Automated Market Maker
- ✅ **Raydium CLMM** - Concentrated Liquidity
- ✅ **Raydium CPMM** - Concentrated Pool
- ✅ **Orca Whirlpool** - Concentrated liquidity AMM
- ✅ **Meteora DAMM V2** - Dynamic AMM
- ✅ **Meteora DLMM** - Dynamic Liquidity Market Maker
- ✅ **Bonk Launchpad** - Token launch platform

---

## 🔥 Quick Start

### Installation

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

### Run Examples

```bash
# PumpFun trade filter (Buy/Sell/BuyExactSolIn/Create)
GEYSER_API_TOKEN=your_token go run examples/pumpfun_trade_filter.go

# PumpSwap low-latency with performance metrics
GEYSER_API_TOKEN=your_token go run examples/pumpswap_low_latency.go

# All protocols simultaneously
GEYSER_API_TOKEN=your_token go run examples/multi_protocol_grpc.go

# Meteora DAMM V2 events
GEYSER_API_TOKEN=your_token go run examples/meteora_damm_grpc.go
```

### Examples

| Example | Description | Command |
|---------|-------------|---------|
| **PumpFun** | | |
| `pumpfun_trade_filter` | PumpFun trade filtering (Buy/Sell/BuyExactSolIn/Create) with latency metrics | `go run examples/pumpfun_trade_filter.go` |
| **PumpSwap** | | |
| `pumpswap_low_latency` | PumpSwap ultra-low latency with per-event + 10s stats | `go run examples/pumpswap_low_latency.go` |
| **Multi-Protocol** | | |
| `multi_protocol_grpc` | Subscribe to all DEX protocols simultaneously | `go run examples/multi_protocol_grpc.go` |
| **Meteora** | | |
| `meteora_damm_grpc` | Meteora DAMM V2 (Swap/AddLiquidity/RemoveLiquidity/CreatePosition/ClosePosition) | `go run examples/meteora_damm_grpc.go` |

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

### Parse Logs Only (No gRPC)

```go
package main

import (
    "fmt"
    "sol-parser-sdk-golang/solparser"
)

func main() {
    logs := []string{
        "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P invoke [1]",
        "Program data: vdt/pQ8AAA...", // base64 encoded event
        "Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success",
    }

    events, err := solparser.ParseLogsOnly(logs, "tx_signature", 123456789, nil)
    if err != nil {
        panic(err)
    }

    for _, ev := range events {
        fmt.Printf("[%s] %+v\n", ev.EventType(), ev)
    }
}
```

---

## 🏗️ Supported Protocols & Events

### Event Types
Each protocol supports:
- 📈 **Trade/Swap Events** - Buy/sell transactions
- 💧 **Liquidity Events** - Deposits/withdrawals
- 🏊 **Pool Events** - Pool creation/initialization
- 🎯 **Position Events** - Open/close positions (CLMM)

### PumpFun Events
- `PumpFunBuy` - Buy token
- `PumpFunSell` - Sell token
- `PumpFunBuyExactSolIn` - Buy with exact SOL amount
- `PumpFunCreate` - Create new token
- `PumpFunTrade` - Generic trade (fallback)

### PumpSwap Events
- `PumpSwapBuy` - Buy token via pool
- `PumpSwapSell` - Sell token via pool
- `PumpSwapCreatePool` - Create liquidity pool
- `PumpSwapLiquidityAdded` - Add liquidity
- `PumpSwapLiquidityRemoved` - Remove liquidity

### Raydium Events
- `RaydiumAmmV4Swap` - AMM V4 swap
- `RaydiumClmmSwap` - CLMM swap
- `RaydiumCpmmSwap` - CPMM swap

### Orca Events
- `OrcaWhirlpoolSwap` - Whirlpool swap

### Meteora Events
- `MeteoraDammV2Swap` - DAMM V2 swap
- `MeteoraDammV2AddLiquidity` - Add liquidity
- `MeteoraDammV2RemoveLiquidity` - Remove liquidity
- `MeteoraDammV2CreatePosition` - Create position
- `MeteoraDammV2ClosePosition` - Close position

### Bonk Events
- `BonkTrade` - Bonk Launchpad trade

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
│   ├── pumpswap_low_latency.go
│   ├── multi_protocol_grpc.go
│   └── meteora_damm_grpc.go
├── go.mod
└── go.sum
```

---

## 🔧 Advanced Usage

### Custom gRPC Endpoint

```go
endpoint := os.Getenv("GEYSER_ENDPOINT")
if endpoint == "" {
    endpoint = "solana-yellowstone-grpc.publicnode.com:443"
}
token := os.Getenv("GEYSER_API_TOKEN")
client := solparser.NewGrpcClient(endpoint, token)
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

## 📄 License

MIT License

## 📞 Contact

- **Repository**: https://github.com/0xfnzero/sol-parser-sdk-golang
- **Website**: https://fnzero.dev/
- **Telegram**: https://t.me/fnzero_group
- **Discord**: https://discord.gg/vuazbGkqQE
