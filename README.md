<div align="center">
    <h1>⚡ Sol Parser SDK - Go</h1>
    <h3><em>High-performance Solana DEX event parser for Go</em></h3>
</div>

<p align="center">
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang"><img src="https://img.shields.io/badge/go-sol--parser--sdk--golang-00ADD8.svg" alt="Go"></a>
    <a href="https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

<p align="center">
    <a href="./README_CN.md">中文</a> |
    <a href="./README.md">English</a> |
    <a href="https://fnzero.dev/">Website</a> |
    <a href="https://t.me/fnzero_group">Telegram</a> |
    <a href="https://discord.gg/vuazbGkqQE">Discord</a>
</p>

---

## Other language SDKs

| Language | Repository |
|----------|------------|
| Rust | [sol-parser-sdk](https://github.com/0xfnzero/sol-parser-sdk) |
| Node.js | [sol-parser-sdk-nodejs](https://github.com/0xfnzero/sol-parser-sdk-nodejs) |
| Python | [sol-parser-sdk-python](https://github.com/0xfnzero/sol-parser-sdk-python) |
| Go | [sol-parser-sdk-golang](https://github.com/0xfnzero/sol-parser-sdk-golang) |

---

## How to use

### 1. Install

This repo’s `go.mod` module path is **`sol-parser-sdk-golang`** (see [`go.mod`](go.mod)). Examples import `sol-parser-sdk-golang/solparser`.

**From source** (recommended)

```bash
git clone https://github.com/0xfnzero/sol-parser-sdk-golang
cd sol-parser-sdk-golang
go mod tidy
```

**Use in another module** — add a `replace` to this GitHub repo (or a local clone), for example:

```go
require sol-parser-sdk-golang v0.0.0

replace sol-parser-sdk-golang => github.com/0xfnzero/sol-parser-sdk-golang v0.0.0
```

(Use a real tagged version once published, or `replace ... => ../sol-parser-sdk-golang` for local dev.)

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
import "sol-parser-sdk-golang/solparser" // module path: see go.mod

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

The example decodes `Entry.entries`, optionally parses outer instructions to **`DexEvent` JSON** via **`DexEventsFromShredTransactionWire`** (static account keys only; V0 + ALT may need TS **`shredstream_pumpfun_json`** + RPC for full keys).

---

## Examples

Run from the **repository root** after `go mod tidy`. One row per source file (links point to GitHub `main`).

| Description | Run command | Source |
|-------------|-------------|--------|
| **PumpFun** | | |
| PumpFun `DexEvent` + metrics | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_with_metrics.go` | [pumpfun_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_with_metrics.go) |
| PumpFun trade filter | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_trade_filter.go` | [pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) |
| Quick connection test | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_quick_test.go` | [pumpfun_quick_test.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_quick_test.go) |
| **PumpSwap** | | |
| PumpSwap + metrics | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_with_metrics.go` | [pumpswap_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_with_metrics.go) |
| Ultra-low latency | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_low_latency.go` | [pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) |
| **Meteora DAMM** | | |
| Meteora DAMM V2 | `GRPC_URL=… GRPC_TOKEN=… go run examples/meteora_damm_grpc.go` | [meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) |
| **ShredStream** (see **step 5** above) | | |
| SubscribeEntries + decode + optional `DexEvent` JSON | `SHRED_URL=http://host:port go run examples/shredstream_entries.go` | [shredstream_entries.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/shredstream_entries.go) |
| **Yellowstone** | | |
| Geyser subscribe + `ParseSubscribeTransaction` | `GRPC_URL=… GRPC_TOKEN=… go run examples/yellowstone_grpc_parse.go` | [yellowstone_grpc_parse.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/yellowstone_grpc_parse.go) |
| **Multi-protocol** | | |
| All supported DEX programs | `GRPC_URL=… GRPC_TOKEN=… go run examples/multi_protocol_grpc.go` | [multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) |
| **Utility** | | |
| Parse tx by signature (RPC, not gRPC stream) | `TX_SIGNATURE=… RPC_URL=… go run examples/parse_tx_by_signature.go` | [parse_tx_by_signature.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/parse_tx_by_signature.go) |

---

## Protocols

PumpFun, PumpSwap, Raydium AMM V4 / CLMM / CPMM, Orca Whirlpool, Meteora DAMM V2 / DLMM, Bonk Launchpad (see `solparser/`).

---

## Useful exports

- **`ParseSubscribeTransaction`** — Geyser single-tx → `[]DexEvent` (instructions + logs + merge + Pump account fill).
- **`ParseRpcTransaction`** / **`ParseTransactionFromRpc`** — HTTP RPC JSON → events.
- **`ParseInstructionUnified`** / **`ParseInnerInstructionUnified`** — outer 8-byte / inner 16-byte discriminators.
- **`DexEventsFromShredTransactionWire`** — wire tx bytes → outer `ParseInstructionUnified` (Shred static keys).
- **`DecodeGRPCEntry`** / **`DecodeEntriesBincode`** — ShredStream `Entry.entries` bytes → `DecodedTransaction` slices.
- **`DexEvent`** — `json.Marshal` for `{ "PumpSwapBuy": { … } }` style output.

---

## Development

```bash
go test ./...
go build ./...
go vet ./...
```

---

## License

MIT — https://github.com/0xfnzero/sol-parser-sdk-golang
