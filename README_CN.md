<div align="center">
    <h1>Sol Parser SDK — Go</h1>
    <p>面向 Solana DEX 的 Go 事件解析 SDK</p>
    <a href="https://pkg.go.dev/github.com/0xfnzero/sol-parser-sdk-golang"><img src="https://pkg.go.dev/badge/github.com/0xfnzero/sol-parser-sdk-golang.svg" alt="sol-parser-sdk-golang"></a>
    <a href="#许可证"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</div>

<p align="center">
    <a href="README.md">English</a> · <a href="README_CN.md">中文</a> ·
    <a href="#示例列表">示例</a> ·
    <a href="docs/USAGE.md">文档</a> ·
    <a href="RELEASE_NOTES.md">发布与迁移</a> ·
    <a href="#sdk-版本与相关项目">相关 SDK</a>
</p>

从 gRPC、ShredStream、RPC 交易及协议账户解析类型化事件，用于索引、监控和交易系统。

当前版本：**0.9.1**。升级前请阅读 [迁移说明](RELEASE_NOTES.md)；历史变更见 [版本记录](CHANGELOG_CN.md)。

## 目录

- [SDK 版本与相关项目](#sdk-版本与相关项目)
- [怎么用](#怎么用)
- [这个 SDK 适合什么场景](#这个-sdk-适合什么场景)
- [交易来源与解析边界](#交易来源与解析边界)
- [协议](#协议)
- [文档导航](#文档导航)

<a id="sdk-版本与相关项目"></a>

## 📦 SDK 版本与相关项目

各仓库版本独立，依赖兼容要求以对应 SDK 的安装说明为准。

| 语言 | 仓库 | 用途 | 版本 |
|----------|------------|-------------|---------|
| Rust | [sol-parser-sdk](https://github.com/0xfnzero/sol-parser-sdk) | DEX 交易与账户事件解析 | `v0.9.1` |
| Node.js | [sol-parser-sdk-nodejs](https://github.com/0xfnzero/sol-parser-sdk-nodejs) | DEX 交易与账户事件解析 | `v0.9.1` |
| Python | [sol-parser-sdk-python](https://github.com/0xfnzero/sol-parser-sdk-python) | DEX 交易与账户事件解析 | `v0.9.1` |
| Go | [sol-parser-sdk-golang](https://github.com/0xfnzero/sol-parser-sdk-golang) | DEX 交易与账户事件解析 | `v0.9.1` |
| Rust | [sol-trade-sdk](https://github.com/0xfnzero/sol-trade-sdk) | 交易构建、签名与多通道提交 | `v9.0.1` |
| Node.js | [sol-trade-sdk-nodejs](https://github.com/0xfnzero/sol-trade-sdk-nodejs) | 交易构建、签名与多通道提交 | `v9.0.1` |
| Python | [sol-trade-sdk-python](https://github.com/0xfnzero/sol-trade-sdk-python) | 交易构建、签名与多通道提交 | `v9.0.1` |
| Go | [sol-trade-sdk-golang](https://github.com/0xfnzero/sol-trade-sdk-golang) | 交易构建、签名与多通道提交 | `v9.0.1` |
| Rust | [sol-shred-sdk](https://github.com/0xfnzero/sol-shred-sdk) | 原始 Shred 解码与 ShredStream 指令事件 | `v6.0.1` |
| Rust | [solana-streamer](https://github.com/0xfnzero/solana-streamer) | gRPC / ShredStream 订阅与事件流封装 | `v5.0.0` |
| Rust | [sol-trade-router-sdk](https://github.com/0xfnzero/sol-trade-router-sdk) | 链上多跳 Router 与客户端 SDK | `v0.3.0` |

<a id="怎么用"></a>

## 怎么用

### 1. 安装

本仓库 `go.mod` 的 module 路径为 **`github.com/0xfnzero/sol-parser-sdk-golang`**（见 [`go.mod`](go.mod)），示例里使用 `import "github.com/0xfnzero/sol-parser-sdk-golang/solparser"`。

**源码克隆**（推荐）

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

**在其他 Go 工程引用**

```bash
go get github.com/0xfnzero/sol-parser-sdk-golang@v0.9.1
```

（本地开发可用 `replace github.com/0xfnzero/sol-parser-sdk-golang => ../sol-parser-sdk-golang` 指向克隆目录。）

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
import "github.com/0xfnzero/sol-parser-sdk-golang/solparser" // module 路径见 go.mod

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

示例对 `Entry.entries` 解码，并可选择将线格式交易走 **`DexEventsFromShredTransactionWire`** 输出 **`DexEvent` JSON**。该热路径默认严格解析静态账户；ALT 指令需传入缓存加载地址或 slot-aware cache resolver，未解析完整时跳过。旧占位行为需显式调用 `DexEventsFromShredTransactionWireBestEffort`。未知 Pump.fun 包装指令通过 PumpFunTrade 输出未解析候选并保留 third_program；原始 wire 无法还原任意 CPI。Inner CPI / 仅日志事件仍需 Yellowstone/RPC 路径。

<a id="这个-sdk-适合什么场景"></a>

## 这个 SDK 适合什么场景

这是 FnZero Solana DEX Parser SDK 的 Go 实现，适合机器人、索引器、跟单系统、狙击交易管线和后端服务，把实时链上交易解析成统一的 DEX 事件。

| 范围 | 覆盖内容 |
|------|----------|
| 解析输入 | Yellowstone gRPC、ShredStream、RPC transactions、encoded transactions、协议账户数据 |
| DEX 协议 | PumpFun、PumpSwap、Pump Fees、Raydium LaunchLab、Raydium CPMM、Raydium CLMM、Raydium AMM V4、Meteora DAMM v2、Meteora DLMM、Meteora DBC、Orca Whirlpool |
| 使用场景 | 实时 DEX 事件解析、代币发射监控、跟单交易、狙击机器人、账户数据补全、JSON 事件管线 |
| 运行环境 | Go 1.22+、后端服务、worker、低延迟机器人基础设施 |

<a id="交易来源与解析边界"></a>

## 交易来源与解析边界

gRPC/RPC 使用提供的执行元数据和实际 CPI 指令；Shred 只能看到外层指令与本地已解析的账户。未知第三方指令在自身账户中引用 Pump.fun 时，可回调 `PumpFunTrade` 候选，携带 `third_program` 原始输入与 `Unresolved` 状态；这不证明交易已执行，也不确定方向或金额。买入／卖出专用过滤器不会把未知方向当作成交。

解析热路径不调用 RPC。按 `parse_status` 判断证据，缺失字段的默认值不代表实际零成交。账户顺序与原始指令数据保留在 `third_program`；交易序号和接收时间仍在事件元数据中，本地 Shred 序号不等于全 slot 执行顺序。详见 [第三方包装交易](docs/wrapped-pumpfun.md)及[跨语言接口与证据约定](docs/cross-language-sync.md)。

<a id="示例列表"></a>

## 示例列表

在**仓库根目录**执行，`go mod tidy` 后即可。下表 **源码** 列链接到 GitHub `main` 上对应文件。

| 描述 | 运行命令 | 源码 |
|------|----------|------|
| **PumpFun** | | |
| PumpFun 事件 + 指标 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_with_metrics.go` | [pumpfun_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_with_metrics.go) | `v0.9.1` |
| PumpFun 交易过滤 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_trade_filter.go` | [pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) | `v0.9.1` |
| 快速连接测试 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_quick_test.go` | [pumpfun_quick_test.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_quick_test.go) | `v0.9.1` |
| **PumpSwap** | | |
| PumpSwap + 指标 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_with_metrics.go` | [pumpswap_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_with_metrics.go) | `v0.9.1` |
| 超低延迟 | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_low_latency.go` | [pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) | `v0.9.1` |
| **Meteora DAMM** | | |
| Meteora DAMM V2 | `GRPC_URL=… GRPC_TOKEN=… go run examples/meteora_damm_grpc.go` | [meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) | `v0.9.1` |
| **ShredStream**（端点见上文步骤 5） | | |
| SubscribeEntries + 解码 + 可选 DexEvent JSON | `SHRED_URL=http://host:port go run examples/shredstream_entries.go` | [shredstream_entries.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/shredstream_entries.go) | `v0.9.1` |
| **Yellowstone** | | |
| Geyser 订阅 + `ParseSubscribeTransaction` | `GRPC_URL=… GRPC_TOKEN=… go run examples/yellowstone_grpc_parse.go` | [yellowstone_grpc_parse.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/yellowstone_grpc_parse.go) | `v0.9.1` |
| **多协议** | | |
| 同时订阅多 DEX | `GRPC_URL=… GRPC_TOKEN=… go run examples/multi_protocol_grpc.go` | [multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) | `v0.9.1` |
| **工具** | | |
| 按签名拉 RPC 解析（非 gRPC 流） | `TX_SIGNATURE=… RPC_URL=… go run examples/parse_tx_by_signature.go` | [parse_tx_by_signature.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/parse_tx_by_signature.go) | `v0.9.1` |

<a id="协议"></a>

## 协议

PumpFun、PumpSwap、Raydium AMM V4 / CLMM / CPMM、Orca Whirlpool、Meteora DAMM V2 / DLMM、Raydium LaunchLab（见 `solparser/`）。

<a id="常用-api"></a>

## 常用 API

- **`ParseSubscribeTransaction`** — Geyser 单笔交易 → `[]DexEvent`（指令 + 日志 + 合并 + Pump 补全）。
- **`ParseRpcTransaction`** / **`ParseTransactionFromRpc`** — HTTP RPC JSON → 事件。
- **`ParseInstructionUnified`** / **`ParseInnerInstructionUnified`** — 外层 8 字节 / 内层 16 字节 discriminator。
- **`DexEventsFromShredTransactionWire`** — 线格式交易字节 → 外层 `ParseInstructionUnified`（默认严格解析静态账户，另提供缓存 ALT 入口）。
- **`DecodeGRPCEntry`** / **`DecodeEntriesBincode`** — ShredStream `Entry.entries` → `DecodedTransaction`。
- **`DexEvent`** — `json.Marshal` 输出 `{ "PumpSwapBuy": { … } }` 形式。

<a id="开发"></a>

## 开发

```bash
go test ./...
go build ./...
go vet ./...
```

<a id="api-兼容性"></a>

## API 兼容性

完整跨语言 API 对齐仍在进行；公开行为和支持边界见 [API 迁移说明](docs/USAGE.md#api-compatibility)，使用方法见 examples 目录。

### Offline simulation route evidence

`solparser.AnalyzeSimulationRoutes(wire, simulationResponseJSON, graduatedPools)` accepts original transaction bytes and a complete `simulateTransaction` JSON response. It makes no RPC calls. Parsed SPL transfers, compiled CPI, execution failure, and unknown Token-2022 net credits are preserved. See [docs/USAGE.md#simulation-routes](docs/USAGE.md#simulation-routes) and `go run ./examples/simulation_routes solparser/testdata/cached_tip_routes_20261002.json route-buy`.

PumpFun create/create_v2 的共享主网样本、重放方法和验证边界见 [验证说明](https://github.com/0xfnzero/sol-parser-sdk-golang/tree/main/validation/pumpfun_create_20261007)。

### Pump ShredStream 离线回归

用户示例位于 examples，自动回归测试单独维护。旧路由意图模拟入口及其无引用快照已移除；当前包装交易回放与过滤语义见 [包装交易说明](docs/wrapped-pumpfun.md)。Shred 只解码已解析完整的外层账户；第三方候选通过 PumpFunTrade 的 third_program 透传，不能当作已确认成交，热路径不调用 RPC。

<a id="文档导航"></a>

## 文档导航

- [发布与迁移](RELEASE_NOTES.md)
- [使用指南](docs/USAGE.md)
- [跨语言功能与边界](docs/cross-language-sync.md)


<a id="许可证"></a>

## 许可证

MIT — https://github.com/0xfnzero/sol-parser-sdk-golang


股票池同步：33 种官方股票/ETF、58 个股票池；配置扩展、USDC 起点、缓存选路与使用边界见 [11 SDK 同步说明](docs/STOCK_BRIDGE_SYNC.md)。
