<div align="center">
    <h1>Sol Parser SDK — Go</h1>
    <p>Solana DEX event parsing for Go</p>
    <a href="https://pkg.go.dev/github.com/0xfnzero/sol-parser-sdk-golang"><img src="https://pkg.go.dev/badge/github.com/0xfnzero/sol-parser-sdk-golang.svg" alt="sol-parser-sdk-golang"></a>
    <a href="#license"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT License"></a>
</div>

<p align="center">
    <a href="README.md">English</a> · <a href="README_CN.md">中文</a> ·
    <a href="#examples">Examples</a> ·
    <a href="docs/USAGE.md">Documentation</a> ·
    <a href="RELEASE_NOTES.md">Release and migration</a> ·
    <a href="#sdk-versions-and-related-sdks">Related SDKs</a>
</p>

Parse typed events from gRPC, ShredStream, RPC transactions and protocol accounts for indexing, monitoring and trading systems.

Current version: **0.9.1**. Read the [migration notes](RELEASE_NOTES.md) before upgrading; see the [changelog](CHANGELOG.md) for historical changes.

## Contents

- [SDK Versions and Related SDKs](#sdk-versions-and-related-sdks)
- [How to use](#how-to-use)
- [What This SDK Is For](#what-this-sdk-is-for)
- [Trade provenance and parsing boundaries](#trade-provenance-and-parsing-boundaries)
- [Examples](#examples)
- [Protocols](#protocols)
- [Documentation](#documentation)

<a id="sdk-versions-and-related-sdks"></a>

## 📦 SDK Versions and Related SDKs

Repositories are versioned independently; follow each SDK’s installation guide for compatible dependencies.

| Language | Repository | Purpose | Version |
|----------|------------|-------------|---------|
| Rust | [sol-parser-sdk](https://github.com/0xfnzero/sol-parser-sdk) | DEX transaction and account event parsing | `v0.9.1` |
| Node.js | [sol-parser-sdk-nodejs](https://github.com/0xfnzero/sol-parser-sdk-nodejs) | DEX transaction and account event parsing | `v0.9.1` |
| Python | [sol-parser-sdk-python](https://github.com/0xfnzero/sol-parser-sdk-python) | DEX transaction and account event parsing | `v0.9.1` |
| Go | [sol-parser-sdk-golang](https://github.com/0xfnzero/sol-parser-sdk-golang) | DEX transaction and account event parsing | `v0.9.1` |
| Rust | [sol-trade-sdk](https://github.com/0xfnzero/sol-trade-sdk) | Transaction building, signing and multi-channel submission | `v9.0.1` |
| Node.js | [sol-trade-sdk-nodejs](https://github.com/0xfnzero/sol-trade-sdk-nodejs) | Transaction building, signing and multi-channel submission | `v9.0.1` |
| Python | [sol-trade-sdk-python](https://github.com/0xfnzero/sol-trade-sdk-python) | Transaction building, signing and multi-channel submission | `v9.0.1` |
| Go | [sol-trade-sdk-golang](https://github.com/0xfnzero/sol-trade-sdk-golang) | Transaction building, signing and multi-channel submission | `v9.0.1` |
| Rust | [sol-shred-sdk](https://github.com/0xfnzero/sol-shred-sdk) | Raw shred decoding and ShredStream instruction events | `v6.0.1` |
| Rust | [solana-streamer](https://github.com/0xfnzero/solana-streamer) | gRPC / ShredStream subscriptions and event streaming | `v5.0.0` |
| Rust | [sol-trade-router-sdk](https://github.com/0xfnzero/sol-trade-router-sdk) | On-chain multi-hop router and client SDK | `v0.3.0` |

<a id="how-to-use"></a>

## How to use

### 1. Install

This repo’s `go.mod` module path is **`github.com/0xfnzero/sol-parser-sdk-golang`** (see [`go.mod`](go.mod)). Examples import `github.com/0xfnzero/sol-parser-sdk-golang/solparser`.

**From source** (recommended)

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

**Use in another module**

```bash
go get github.com/0xfnzero/sol-parser-sdk-golang@v0.9.1
```

(Or use `replace github.com/0xfnzero/sol-parser-sdk-golang => ../sol-parser-sdk-golang` for local development.)

### 2. Environment (examples)

**Yellowstone / Geyser gRPC** (all examples that subscribe over gRPC use the same two names):

| Variable | Meaning |
|----------|---------|
| **`GRPC_URL`** | Endpoint host or full URL (e.g. `https://solana-yellowstone-grpc.publicnode.com:443` or `host:443`). Parsed to `host:port` where needed. |
| **`GRPC_TOKEN`** | `x-token` (or empty if the node allows unauthenticated access). |

**ShredStream** (separate binary / HTTP-style endpoint — **not** the same as `GRPC_URL`):

| Variable | Meaning |
|----------|---------|
| **`SHRED_URL`** | e.g. `http://127.0.0.1:10800` (plain-text gRPC to ShredStream proxy). |

Optional ShredStream tuning: `SHRED_PARSE_DEX`, `SHRED_MAX_JSON_PER_ENTRY`, `SHRED_JSON_COMPACT`, `SHREDSTREAM_QUIET`, `SHRED_MAX_MSG` — see [examples/shredstream_entries.go](examples/shredstream_entries.go).

**RPC utility** [parse_tx_by_signature.go](examples/parse_tx_by_signature.go): `TX_SIGNATURE`, `RPC_URL`.

### 3. Smoke test

```bash
go test ./...
```

### 4. Full gRPC transaction parse (recommended)

Use **`ParseSubscribeTransaction`** (Geyser `SubscribeUpdateTransactionInfo` → RPC-shaped tx + meta) for **instruction accounts + Program data logs + merge + Pump fills**, aligned with Rust `parse_rpc_transaction` behavior.

```go
import "github.com/0xfnzero/sol-parser-sdk-golang/solparser" // module path: see go.mod

events, err := solparser.ParseSubscribeTransaction(slot, txInfo, nil, grpcRecvUs)
if err != nil {
    // handle
}
for _, ev := range events {
    // ev.Type, ev.Data — JSON via json.Marshal(ev)
}
```

**Lighter path:** `ParseLogOptimized` / logs-only helpers when you do not have full transaction + meta.

### 5. ShredStream (HTTP endpoint — not Yellowstone gRPC)

Uses **`SHRED_URL`** only (e.g. `http://127.0.0.1:10800`). This is **not** `GRPC_URL` (different service).

```bash
export SHRED_URL="http://127.0.0.1:10800"
go run examples/shredstream_entries.go
```

The example decodes `Entry.entries`, optionally parses outer instructions to **`DexEvent` JSON** via **`DexEventsFromShredTransactionWire`**. This hot path strictly parses static accounts by default. Supply cached loaded addresses or a slot-aware cache resolver for ALT instructions; unresolved instructions are skipped. The explicit `DexEventsFromShredTransactionWireBestEffort` API retains legacy placeholder behavior. Unknown Pump.fun wrappers emit unresolved PumpFunTrade candidates with third_program; arbitrary CPI cannot be reconstructed from raw wires. Inner CPI/log-only events still require Yellowstone/RPC paths.

<a id="what-this-sdk-is-for"></a>

## What This SDK Is For

Run `go run ./examples/hotpath_latency` from the repository root for a native,
offline hot-path benchmark. Edit `examples/hotpath_latency/config.json` to set
warmup, samples and batch size; an alternate config path is optional. It reports
p50/p95/p99 nanoseconds and allocation counts per batch of six saved signed
wrapped-route/ALT transactions. Fixture loading is outside measurement. No RPC
or transaction submission occurs; CPU timings do not measure network landing.

This is the Go implementation of the FnZero Solana DEX parser SDK for bots, indexers, copy-trading systems, sniper pipelines, and backend services that need normalized real-time DEX events.

| Area | Coverage |
|------|----------|
| Parser inputs | Yellowstone gRPC, ShredStream, RPC transactions, encoded transactions, protocol account data |
| DEX protocols | PumpFun, PumpSwap, Pump Fees, Raydium LaunchLab, Raydium CPMM, Raydium CLMM, Raydium AMM V4, Meteora DAMM v2, Meteora DLMM, Meteora DBC, Orca Whirlpool |
| Use cases | Real-time DEX event parsing, token launch monitoring, copy trading, sniper bots, account filling, JSON event pipelines |
| Runtime | Go 1.22+, backend services, workers, low-latency bot infrastructure |

<a id="trade-provenance-and-parsing-boundaries"></a>

## Trade provenance and parsing boundaries

gRPC/RPC use supplied execution metadata and actual CPI instructions; Shred sees outer instructions and locally resolved accounts. An unknown third-party instruction referencing Pump.fun in its own accounts can produce a `PumpFunTrade` candidate with the original input in `third_program` and `Unresolved` status. This does not establish execution, direction or amounts. Buy/sell-only filters do not treat an unknown direction as a fill.

The parsing hot path performs no RPC. Check `parse_status` before interpreting defaults as amounts. Original account order and instruction bytes remain in `third_program`; transaction position and receive time remain in event metadata. A local Shred index is not confirmed slot-wide execution order. See [wrapped transactions](docs/wrapped-pumpfun.md) and [cross-language contracts](docs/cross-language-sync.md).

<a id="examples"></a>

## Examples

Run from the **repository root** after `go mod tidy`. One row per source file (links point to GitHub `main`).

| Description | Run command | Source |
|-------------|-------------|--------|
| **PumpFun** | | |
| PumpFun `DexEvent` + metrics | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_with_metrics.go` | [pumpfun_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_with_metrics.go) | `v0.9.1` |
| PumpFun trade filter | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_trade_filter.go` | [pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) | `v0.9.1` |
| Quick connection test | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_quick_test.go` | [pumpfun_quick_test.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_quick_test.go) | `v0.9.1` |
| **PumpSwap** | | |
| PumpSwap + metrics | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_with_metrics.go` | [pumpswap_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_with_metrics.go) | `v0.9.1` |
| Ultra-low latency | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_low_latency.go` | [pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) | `v0.9.1` |
| **Meteora DAMM** | | |
| Meteora DAMM V2 | `GRPC_URL=… GRPC_TOKEN=… go run examples/meteora_damm_grpc.go` | [meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) | `v0.9.1` |
| **ShredStream** (see **step 5** above) | | |
| SubscribeEntries + decode + optional `DexEvent` JSON | `SHRED_URL=http://host:port go run examples/shredstream_entries.go` | [shredstream_entries.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/shredstream_entries.go) | `v0.9.1` |
| **Yellowstone** | | |
| Geyser subscribe + `ParseSubscribeTransaction` | `GRPC_URL=… GRPC_TOKEN=… go run examples/yellowstone_grpc_parse.go` | [yellowstone_grpc_parse.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/yellowstone_grpc_parse.go) | `v0.9.1` |
| **Multi-protocol** | | |
| All supported DEX programs | `GRPC_URL=… GRPC_TOKEN=… go run examples/multi_protocol_grpc.go` | [multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) | `v0.9.1` |
| **Utility** | | |
| Parse tx by signature (RPC, not gRPC stream) | `TX_SIGNATURE=… RPC_URL=… go run examples/parse_tx_by_signature.go` | [parse_tx_by_signature.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/parse_tx_by_signature.go) | `v0.9.1` |

<a id="protocols"></a>

## Protocols

PumpFun, PumpSwap, Raydium AMM V4 / CLMM / CPMM, Orca Whirlpool, Meteora DAMM V2 / DLMM, Raydium LaunchLab (see `solparser/`).

<a id="useful-exports"></a>

## Useful exports

- **`ParseSubscribeTransaction`** — Geyser single-tx → `[]DexEvent` (instructions + logs + merge + Pump account fill).
- **`ParseRpcTransaction`** / **`ParseTransactionFromRpc`** — HTTP RPC JSON → events.
- **`ParseInstructionUnified`** / **`ParseInnerInstructionUnified`** — outer 8-byte / inner 16-byte discriminators.
- **`DexEventsFromShredTransactionWire`** — wire tx bytes → outer `ParseInstructionUnified` (strict static keys; cached ALT variants available).
- **`DecodeGRPCEntry`** / **`DecodeEntriesBincode`** — ShredStream `Entry.entries` bytes → `DecodedTransaction` slices.
- **`DexEvent`** — `json.Marshal` for `{ "PumpSwapBuy": { … } }` style output.

<a id="development"></a>

## Development

```bash
go test ./...
go build ./...
go vet ./...
```

<a id="api-compatibility"></a>

## API compatibility

Full cross-language API parity is still in progress. See the [API migration guide](docs/USAGE.md#api-compatibility) for public behavior and support boundaries, and the examples directory for usage.

### Offline simulation route evidence

`solparser.AnalyzeSimulationRoutes(wire, simulationResponseJSON, graduatedPools)` accepts original transaction bytes and a complete `simulateTransaction` JSON response. It makes no RPC calls. Parsed SPL transfers, compiled CPI, execution failure, and unknown Token-2022 net credits are preserved. See [docs/USAGE.md#simulation-routes](docs/USAGE.md#simulation-routes) and `go run ./examples/simulation_routes solparser/testdata/cached_tip_routes_20261002.json route-buy`.


Live parser-to-trade integration uses Yellowstone **gRPC**. See [gRPC cache integration and language examples](docs/USAGE.md#grpc-cache). Quoting and construction read cached state without hot-path RPC.

Shared PumpFun create/create_v2 mainnet fixtures, replay instructions and verification limits are documented in the [validation guide](https://github.com/0xfnzero/sol-parser-sdk-golang/tree/main/validation/pumpfun_create_20261007).

### Offline Pump ShredStream regressions

User examples live in examples; automated regression tests are maintained separately. Obsolete route-intent simulation entry points and their unused snapshots have been removed. See [wrapped transaction replay and filters](docs/wrapped-pumpfun.md). Shred decodes resolved outer instructions; third-party candidates retain original input in PumpFunTrade.third_program and are not confirmed fills. No RPC runs in the parser hot path.

<a id="documentation"></a>

## Documentation

- [Release and migration](RELEASE_NOTES.md)
- [Usage guide](docs/USAGE.md)
- [Cross-language support and boundaries](docs/cross-language-sync.md)


<a id="license"></a>

## License

MIT — https://github.com/0xfnzero/sol-parser-sdk-golang


Shared stock bridges: 33 official stock/ETF mints and 58 pools. See [eleven-SDK catalog and routing alignment](docs/STOCK_BRIDGE_SYNC.md) for configuration, USDC routing and cache-only boundaries.
