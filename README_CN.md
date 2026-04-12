<div align="center">
    <h1>⚡ Sol Parser SDK - Go</h1>
    <h3><em>高性能 Solana DEX 事件解析（Go）</em></h3>
</div>

<p align="center">
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang"><img src="https://img.shields.io/badge/go-sol--parser--sdk--golang-00ADD8.svg" alt="Go"></a>
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

<p align="center">
    <a href="./README_CN.md">中文</a> |
    <a href="./README.md">English</a> |
    <a href="https://fnzero.dev/">官网</a> |
    <a href="https://t.me/fnzero_group">Telegram</a> |
    <a href="https://discord.gg/vuazbGkqQE">Discord</a>
</p>

---

## 其他语言 SDK

| 语言 | 仓库 |
|------|------|
| Rust | [sol-parser-sdk](https://github.com/0xfnzero/sol-parser-sdk) |
| Node.js | [sol-parser-sdk-nodejs](https://github.com/0xfnzero/sol-parser-sdk-nodejs) |
| Python | [sol-parser-sdk-python](https://github.com/0xfnzero/sol-parser-sdk-python) |
| Go | [sol-parser-sdk-golang](https://github.com/0xfnzero/sol-parser-sdk-golang) |

---

## 怎么用

### 1. 安装

本仓库 `go.mod` 的 module 路径为 **`sol-parser-sdk-golang`**（见 [`go.mod`](go.mod)），示例里使用 `import "sol-parser-sdk-golang/solparser"`。

**源码克隆**（推荐）

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

**在其他 Go 工程引用** — 在 `go.mod` 中用 `replace` 指向上游或本地目录，例如：

```go
require sol-parser-sdk-golang v0.0.0

replace sol-parser-sdk-golang => github.com/0xfnzero/sol-parser-sdk-golang v0.0.0
```

（发布 tag 后改用真实版本号，或 `replace ... => ../sol-parser-sdk-golang` 做本地开发。）

### 2. 环境变量（示例）

**Yellowstone / Geyser gRPC**（所有通过 gRPC 订阅的示例统一只用下面两个名字）：

| 变量 | 含义 |
|------|------|
| **`GRPC_URL`** | 节点地址，可为完整 URL 或 `host:443`（示例内会解析为 `host:port`）。 |
| **`GRPC_TOKEN`** | `x-token`（若节点允许无 token 可留空）。 |

**ShredStream**（独立服务，**不是** `GRPC_URL`）：

| 变量 | 含义 |
|------|------|
| **`SHRED_URL`** | 如 `http://127.0.0.1:10800`（明文 gRPC 连 ShredStream 代理）。 |

ShredStream 可选：`SHRED_PARSE_DEX`、`SHRED_MAX_JSON_PER_ENTRY`、`SHRED_JSON_COMPACT`、`SHREDSTREAM_QUIET`、`SHRED_MAX_MSG` 等，见 [examples/shredstream_entries.go](examples/shredstream_entries.go)。

**RPC 工具** [parse_tx_by_signature.go](examples/parse_tx_by_signature.go)：`TX_SIGNATURE`、`RPC_URL`。

### 3. 冒烟

```bash
go test ./...
```

### 4. 全量 gRPC 交易解析（推荐）

使用 **`ParseSubscribeTransaction`**（Geyser `SubscribeUpdateTransactionInfo` → 与 RPC 对齐的 message/meta），得到 **指令账户 + Program data 日志 + merge + Pump 系账户补全**，行为对齐 Rust `parse_rpc_transaction`。

```go
import "sol-parser-sdk-golang/solparser" // module 路径见 go.mod

events, err := solparser.ParseSubscribeTransaction(slot, txInfo, nil, grpcRecvUs)
if err != nil {
    // 处理错误
}
for _, ev := range events {
    // ev.Type、ev.Data — json.Marshal(ev) 输出 JSON
}
```

更轻量：仅有日志时用 `ParseLogOptimized` 等日志路径。

### 5. ShredStream（HTTP 端点，不是 Yellowstone gRPC）

只使用 **`SHRED_URL`**（如 `http://127.0.0.1:10800`），与 **`GRPC_URL`** 不是同一服务。

```bash
export SHRED_URL="http://127.0.0.1:10800"
go run examples/shredstream_entries.go
```

示例对 `Entry.entries` 解码，并可选择将线格式交易走 **`DexEventsFromShredTransactionWire`** 输出 **`DexEvent` JSON**（仅静态账户；V0+ALT 完整账户需 Node 版 **`shredstream_pumpfun_json`** + RPC）。

---

## 示例列表

在**仓库根目录**执行，`go mod tidy` 后即可。下表 **源码** 列链接到 GitHub `main` 上对应文件。

| 描述 | 运行命令 | 源码 |
|------|----------|------|
| **PumpFun** | | |
| PumpFun 事件 + 指标 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_with_metrics.go` | [pumpfun_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_with_metrics.go) |
| PumpFun 交易过滤 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_trade_filter.go` | [pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) |
| 快速连接测试 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_quick_test.go` | [pumpfun_quick_test.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_quick_test.go) |
| **PumpSwap** | | |
| PumpSwap + 指标 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_with_metrics.go` | [pumpswap_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_with_metrics.go) |
| 超低延迟 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_low_latency.go` | [pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) |
| **Meteora DAMM** | | |
| Meteora DAMM V2 | `GRPC_URL=… GRPC_TOKEN=… go run examples/meteora_damm_grpc.go` | [meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) |
| **ShredStream**（端点见上文步骤 5） | | |
| SubscribeEntries + 解码 + 可选 DexEvent JSON | `SHRED_URL=http://host:port go run examples/shredstream_entries.go` | [shredstream_entries.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/shredstream_entries.go) |
| **Yellowstone** | | |
| Geyser 订阅 + `ParseSubscribeTransaction` | `GRPC_URL=… GRPC_TOKEN=… go run examples/yellowstone_grpc_parse.go` | [yellowstone_grpc_parse.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/yellowstone_grpc_parse.go) |
| **多协议** | | |
| 同时订阅多 DEX | `GRPC_URL=… GRPC_TOKEN=… go run examples/multi_protocol_grpc.go` | [multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) |
| **工具** | | |
| 按签名拉 RPC 解析（非 gRPC 流） | `TX_SIGNATURE=… RPC_URL=… go run examples/parse_tx_by_signature.go` | [parse_tx_by_signature.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/parse_tx_by_signature.go) |

---

## 协议

PumpFun、PumpSwap、Raydium AMM V4 / CLMM / CPMM、Orca Whirlpool、Meteora DAMM V2 / DLMM、Bonk Launchpad（见 `solparser/`）。

---

## 常用 API

- **`ParseSubscribeTransaction`** — Geyser 单笔交易 → `[]DexEvent`（指令 + 日志 + 合并 + Pump 补全）。
- **`ParseRpcTransaction`** / **`ParseTransactionFromRpc`** — HTTP RPC JSON → 事件。
- **`ParseInstructionUnified`** / **`ParseInnerInstructionUnified`** — 外层 8 字节 / 内层 16 字节 discriminator。
- **`DexEventsFromShredTransactionWire`** — 线格式交易字节 → 外层 `ParseInstructionUnified`（Shred 静态账户）。
- **`DecodeGRPCEntry`** / **`DecodeEntriesBincode`** — ShredStream `Entry.entries` → `DecodedTransaction`。
- **`DexEvent`** — `json.Marshal` 输出 `{ "PumpSwapBuy": { … } }` 形式。

---

## 开发

```bash
go test ./...
go build ./...
go vet ./...
```

---

## 许可证

MIT — https://github.com/0xfnzero/sol-parser-sdk-golang
