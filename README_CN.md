<div align="center">
    <h1>⚡ Sol Parser SDK - Go</h1>
    <h3><em>高性能 Solana DEX 事件解析器，专为 Go 设计</em></h3>
</div>

<p align="center">
    <strong>通过 Yellowstone gRPC 实时解析 Solana DEX 事件的 Go 库</strong>
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
    <a href="https://fnzero.dev/">官网</a> |
    <a href="https://t.me/fnzero_group">Telegram</a> |
    <a href="https://discord.gg/vuazbGkqQE">Discord</a>
</p>

---

## 📦 SDK 版本

本 SDK 提供多种语言版本：

| 语言 | 仓库 | 描述 |
|------|------|------|
| **Rust** | [sol-parser-sdk](https://github.com/0xfnzero/sol-parser-sdk) | 超低延迟，SIMD 优化 |
| **Node.js** | [sol-parser-sdk-nodejs](https://github.com/0xfnzero/sol-parser-sdk-nodejs) | TypeScript/JavaScript，Node.js 支持 |
| **Python** | [sol-parser-sdk-python](https://github.com/0xfnzero/sol-parser-sdk-python) | 原生 async/await 支持 |
| **Go** | [sol-parser-sdk-golang](https://github.com/0xfnzero/sol-parser-sdk-golang) | 并发安全，goroutine 支持 |

---

## 📊 性能亮点

### ⚡ 实时解析
- **零延迟** 基于日志的事件解析
- **gRPC 流式传输** 支持 Yellowstone/Geyser 协议
- **多协议** 单次订阅同时监听多个 DEX
- **并发安全** 原子计数器与 goroutine 统计

### 🏗️ 支持的协议
- ✅ **PumpFun** - Meme 代币交易
- ✅ **PumpSwap** - PumpFun 交换协议
- ✅ **Raydium AMM V4** - 自动做市商
- ✅ **Raydium CLMM** - 集中流动性
- ✅ **Raydium CPMM** - 集中池
- ✅ **Orca Whirlpool** - 集中流动性 AMM
- ✅ **Meteora DAMM V2** - 动态 AMM
- ✅ **Meteora DLMM** - 动态流动性做市商
- ✅ **Bonk Launchpad** - 代币发射平台

---

## 🔥 快速开始

### 安装

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

### 运行示例

```bash
# PumpFun 交易过滤（Buy/Sell/BuyExactSolIn/Create）
GEYSER_API_TOKEN=your_token go run examples/pumpfun_trade_filter.go

# PumpSwap 超低延迟，附带性能指标
GEYSER_API_TOKEN=your_token go run examples/pumpswap_low_latency.go

# 同时订阅所有协议
GEYSER_API_TOKEN=your_token go run examples/multi_protocol_grpc.go

# Meteora DAMM V2 事件
GEYSER_API_TOKEN=your_token go run examples/meteora_damm_grpc.go

# ShredStream（Jito 风格 gRPC）：SubscribeEntries + DecodeGRPCEntry
export SHRED_URL="http://127.0.0.1:10800"
go run examples/shredstream_entries.go

# Yellowstone gRPC（与 GRPC_URL / GRPC_TOKEN 环境变量）
export GRPC_URL="https://solana-yellowstone-grpc.publicnode.com:443"
export GRPC_TOKEN="your_token"
go run examples/yellowstone_grpc_parse.go
```

### 示例列表

| 描述 | 运行命令 | 源码 |
|------|----------|------|
| **PumpFun** | | |
| PumpFun 交易过滤，附带延迟指标 | `go run examples/pumpfun_trade_filter.go` | [examples/pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) |
| **PumpSwap** | | |
| PumpSwap 超低延迟，含统计 | `go run examples/pumpswap_low_latency.go` | [examples/pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) |
| **多协议** | | |
| 同时订阅所有 DEX 协议 | `go run examples/multi_protocol_grpc.go` | [examples/multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) |
| **Meteora** | | |
| Meteora DAMM V2 事件 | `go run examples/meteora_damm_grpc.go` | [examples/meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) |
| **ShredStream** | | |
| gRPC 订阅 + `entries` 解码（对齐 Rust `shredstream_example`） | `SHRED_URL=http://host:port go run examples/shredstream_entries.go` | [examples/shredstream_entries.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/shredstream_entries.go) |
| **Yellowstone** | | |
| Geyser 订阅 + `ParseLogsOnly`（可用 `GRPC_URL` / `GRPC_TOKEN`） | `GRPC_URL=... GRPC_TOKEN=... go run examples/yellowstone_grpc_parse.go` | [examples/yellowstone_grpc_parse.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/yellowstone_grpc_parse.go) |

### 基本用法

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
            solparser.PUMPFUN_PROGRAM_ID,
            solparser.PUMPSWAP_PROGRAM_ID,
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

## 🏗️ 支持的协议与事件

### 事件类型
每个协议均支持：
- 📈 **交易/兑换事件** - 买入/卖出交易
- 💧 **流动性事件** - 存入/提取
- 🏊 **池子事件** - 池子创建/初始化
- 🎯 **仓位事件** - 开仓/平仓（CLMM）

### PumpFun 事件
- `PumpFunBuy` - 买入代币
- `PumpFunSell` - 卖出代币
- `PumpFunBuyExactSolIn` - 指定 SOL 数量买入
- `PumpFunCreate` - 创建新代币
- `PumpFunTrade` - 通用交易（兜底）

### PumpSwap 事件
- `PumpSwapBuy` - 通过池子买入代币
- `PumpSwapSell` - 通过池子卖出代币
- `PumpSwapCreatePool` - 创建流动性池
- `PumpSwapLiquidityAdded` - 添加流动性
- `PumpSwapLiquidityRemoved` - 移除流动性

### Raydium 事件
- `RaydiumAmmV4Swap` - AMM V4 兑换
- `RaydiumClmmSwap` - CLMM 兑换
- `RaydiumCpmmSwap` - CPMM 兑换

### Orca 事件
- `OrcaWhirlpoolSwap` - Whirlpool 兑换

### Meteora 事件
- `MeteoraDammV2Swap` - DAMM V2 兑换
- `MeteoraDammV2AddLiquidity` - 添加流动性
- `MeteoraDammV2RemoveLiquidity` - 移除流动性
- `MeteoraDammV2CreatePosition` - 创建仓位
- `MeteoraDammV2ClosePosition` - 关闭仓位

### Bonk 事件
- `BonkTrade` - Bonk Launchpad 交易

---

## 📁 项目结构

```
sol-parser-sdk-golang/
├── solparser/
│   ├── grpc_client.go      # GrpcClient（连接、订阅、认证）
│   ├── parser.go           # ParseLogsOnly、ParseTransactionEvents
│   ├── types.go            # DexEvent、TransactionFilter、TransactionUpdate
│   └── ...                 # 协议专用解析器
├── proto/
│   ├── geyser.proto        # Yellowstone gRPC proto
│   └── generated/          # 生成的 Go proto 文件
├── examples/
│   ├── pumpfun_trade_filter.go
│   ├── pumpswap_low_latency.go
│   ├── multi_protocol_grpc.go
│   └── meteora_damm_grpc.go
├── go.mod
└── go.sum
```

---

## 🔧 高级用法

### 自定义 gRPC 端点

```go
endpoint := os.Getenv("GEYSER_ENDPOINT")
if endpoint == "" {
    endpoint = "solana-yellowstone-grpc.publicnode.com:443"
}
token := os.Getenv("GEYSER_API_TOKEN")
client := solparser.NewGrpcClient(endpoint, token)
```

### 并发统计（原子计数器）

```go
import "sync/atomic"

var totalEvents int64

// 在回调中：
atomic.AddInt64(&totalEvents, int64(len(events)))

// 在 goroutine 中：
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        count := atomic.LoadInt64(&totalEvents)
        fmt.Printf("总事件数: %d\n", count)
    }
}()
```


---

## 与 Rust `sol-parser-sdk` 的 RPC 解析对齐说明

- **日志热路径**：`ParseLogOptimized` / `matcher.go` 与 Rust `parse_log_optimized` 对齐。
- **`ParseRpcTransaction`**：已支持 **ALT 全量账户表**、**外层 8 字节 / 内层 16 字节** 指令解析、**merge**、Pump 系 **账户补全** 与 **PumpSwap `is_pump_pool`（费用程序 CPI）**。
- **ShredStream**：见 `shredstream` 包（gRPC + `pb`，proto 与 `0xfnzero/solana-streamer` 中 prost 生成物对齐）。`Client` 提供 `SendHeartbeat` 与 `SubscribeEntries`；`Entry.entries` 解码请用 **`DecodeGRPCEntry` / `DecodeEntriesBincode`**（实现参考 [shredstream-sdk-go](https://github.com/shredstream/shredstream-sdk-go) 的线格式解析，得到 `DecodedTransaction`）。


---

## 📄 许可证

MIT License

## 📞 联系我们

- **仓库**: https://github.com/0xfnzero/sol-parser-sdk-golang
- **官网**: https://fnzero.dev/
- **Telegram**: https://t.me/fnzero_group
- **Discord**: https://discord.gg/vuazbGkqQE
