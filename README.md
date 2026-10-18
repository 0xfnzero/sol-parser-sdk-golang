<div align="center">

## v0.5.15 — Pump execution replay

Current Anchor TradeEvent CPI decoding is supported alongside the historical encoding. RPC/gRPC instruction budgets and minimum outputs remain in instruction fields. Execution fields `sol_amount`, `quote_amount`, and `token_amount` come from execution events; without logs or event CPIs they remain zero (unavailable, not a proven zero fill).

Run `go run ./examples/pump_execution_replay` from this repository to replay real transactions fetched on 2026-10-19 offline. The example prints full/no-log/truncated-log results. Append a recent signature for online replay. Set `SOLANA_RPC_URL` for your RPC; transaction v1 is supported.

The current Delta sell signature is `4ARos3Yq1SMNE3izDpy9YdKrRgMmD7UCDxvnVrBRaGCJHiifPHSmQWc4DPfjUaEUcxS7puGmEN64K7EZdeXLJaj2` (slot 458542546). The dated fixture also includes Rapid create/buy, BuyV3, and PumpSwap buy with their full signatures and slots. Older five-combination multi-hop fixtures are retained as regression tests, not labeled as new mainnet captures.

## v0.5.15 — raw Shred instructions

Default Shred parsing emits `RawInstruction` for unsupported third-party programs, including Rapid/Delta, without inferring direction, amounts, or internal buy methods. Infrastructure programs (System, Token, ATA, ComputeBudget, etc.) are omitted by default. Supported official DEX parsers are unchanged; official Pump/PumpSwap multi-hop instructions retain `PumpRouteIntent`.

The payload contains `program_id`, `instruction_index`, `accounts`, `account_indices`, `instruction_data`, and transaction metadata. Account order and duplicates are preserved. Unresolved ALT addresses are null/None, with original indices retained; supply cached loaded addresses or a resolver. No hot-path RPC is performed. Data is copied verbatim without requiring a known layout.

No filter returns unknown third-party instructions. Explicitly including `RawInstruction` returns every outer instruction, including infrastructure. A `PumpFunTrade`-only filter does not include raw events: subscribe to both types when needed. Streamer forwards `DexEvent::RawInstructionEvent` as Common. A Pump address in accounts does not prove a CPI or a particular buy method. Event count is not fill count.

Wrapper adapter functions remain explicit compatibility helpers only; default Shred parsing does not call them. The released-version history below describes previous behavior.

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

## 📦 SDK Versions and Related SDKs

Parser SDK language versions and related Rust SDKs:

| Language | Repository | Description | Version |
|----------|------------|-------------|---------|
| Rust | [sol-parser-sdk](https://github.com/0xfnzero/sol-parser-sdk) | Solana DEX transaction and account event parsing | `v0.7.16` |
| Node.js | [sol-parser-sdk-nodejs](https://github.com/0xfnzero/sol-parser-sdk-nodejs) | TypeScript/JavaScript for Node.js | `v0.5.22` |
| Python | [sol-parser-sdk-python](https://github.com/0xfnzero/sol-parser-sdk-python) | Async/await native support | `v0.5.15` |
| Go | [sol-parser-sdk-golang](https://github.com/0xfnzero/sol-parser-sdk-golang) | Go transaction and account event parsing | `v0.5.15` |
| **Rust** | [sol-shred-sdk](https://github.com/0xfnzero/sol-shred-sdk) | Raw Solana shred decoding and ShredStream DEX event parsing | `v4.0.7` |
| **Rust** | [sol-trade-sdk](https://github.com/0xfnzero/sol-trade-sdk) | Solana DEX trade construction and transaction execution | `v6.0.2` |

## v0.5.14 — Pump route direction and instruction names

Route intents now expose `is_buy` and `ix_name`, following existing trade field names and official IDL instruction names. RapidBuy reports a buy intent and `buy_exact_quote_in_v2`; BuyPumpDelta reports only a buy intent. Multi-hop direction and unknown methods remain null. Read currencies from `hops[].base_mint` / `quote_mint`. These hints do not confirm execution, and filters do not convert intents into trades.

Older JSON without the fields remains readable as unknown. Rust struct literals must supply the new fields; Go uses pointers and Node.js/Python use null/None for unknown values. Streamer now forwards `DexEvent::PumpRouteIntentEvent` with protocol and event filtering.

## v0.5.13 — Pump routes and offline replay

Adds native Pump route-intent parsing for official multi-hop instructions and the exact RapidBuy/BuyPumpDelta wrapper layouts captured in reported transactions. Adds locally cached ALT resolution and rejects unresolved account identities by default, with explicit best-effort compatibility APIs where supported. Route intents expose instruction constraints and hop identities; confirmed execution amounts require execution metadata.

Includes the latest complete implementation, signed transaction fixtures, native offline-bank regression examples and reproducible local hot-path benchmarks. Unknown wrapper layouts are not inferred from embedded bytes. Local tests do not measure production network latency.

## v0.5.12 — Pump depth-2 compatibility

Adds signed Pump depth-2 create-event replay coverage, including ALT-loaded accounts and executed logs. Existing event layouts and public parsing interfaces remain unchanged.

Validation includes official Pump SDK 4.1.0 instruction comparisons, signed local LiteSVM simulations and SDK regressions. Local simulation state/config is synthetic; no funded transactions were broadcast and production latency was not measured.

## v0.5.11 — Signed transaction and hot-path hardening

Hardens signed transaction sanitization, loaded-address boundaries, ordered stream filtering and parser lifecycle. Aligns CLMM and DEX instruction/event layouts, nested CPI route attribution and liquidity/reward accounting. Adds independently signed wire fixtures, ALT failure cases and offline bank regressions.

Validation includes local CPU benchmarks and offline signed-bank scenarios. Measured hot paths use cached inputs without RPC. Benchmarks do not establish production network or transaction-landing latency. No funded mainnet transactions were broadcast.

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

## Release notes

### v0.5.10

- Matches PumpFun create/create_v2 accounts by the actual instruction discriminator and event mint, including CPI calls; ambiguous matches are left unspecified.
- Stops inferring quote mint/vault/token-program fields from arbitrary remaining create_v2 accounts and preserves authoritative decoded fields.
- Restores historical CreateEvent decoding and verifies successful and failed transactions using saved mainnet fixtures.
- Suppresses rolled-back DEX events from failed transactions and preserves failure status through RPC/gRPC adapters.

### v0.5.9

- Adds native gRPC lifecycle and block metadata support, instruction-level route analysis and StonkFun registry/snapshot examples.
- Improves exact metadata serialization, account matching and wire transaction bounds validation.
- Includes native alignment evidence and simulation route fixtures. These language releases do not yet implement the new CPMM creator-fee collection event API.

### v0.5.8

- Syncs with Rust sol-parser-sdk **0.7.3**, including the current PumpFun, PumpSwap, and Pump Fees protocol definitions.
- Adds Meteora DAMM v2 events: UpdateDelegatePermission, WithdrawDeadLiquidityReward, CreateConfig, CreateDynamicConfig.
- Aligns EvtSwap2 180-byte layout with slot-gated claiming/compounding fees and swap modes; routes EvtLiquidityChange to Add/Remove.
- Adds DLMM swap account context fields, PumpFun pre/post balance snapshots, and LaunchLab trade quote context.
- Adds current PumpFun/PumpSwap creator-fee and holder-reward fields across instruction, log, account, gRPC, and JSON event paths.
- Rejects truncated known PumpSwap trade and CreatePool tails while retaining complete historical layouts.

### v0.5.6

- Adds Meteora DBC log parsing with program-context routing and filter parity.
- Adds Raydium CLMM/CPMM and Orca account parsers.
- Preserves block transaction indexes in RPC and Geyser-to-RPC conversion.
- Skips ShredStream instruction parsing early for account-only or empty include-only filters.
- Tightens RPC log active-program tracking for low-latency parser correctness.

### v0.5.5

- Aligns ShredStream wire-transaction parsing with Rust/Node.js/Python.
- Uses default pubkey placeholders for V0 ALT-loaded instruction accounts instead of dropping the instruction.
- Adds discriminator fallback when the ShredStream outer program id is ALT-loaded.
- Improves Pump.fun v2 short-account parsing, same-transaction post-merge enrichment, and multi-protocol routing parity.
- Adds regression coverage for ShredStream ALT account and ALT program-id fallback.

---

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
go get github.com/0xfnzero/sol-parser-sdk-golang@v0.5.15
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

The example decodes `Entry.entries`, optionally parses outer instructions to **`DexEvent` JSON** via **`DexEventsFromShredTransactionWire`**. This hot path strictly parses static accounts by default. Supply cached loaded addresses or a slot-aware cache resolver for ALT instructions; unresolved instructions are skipped. The explicit `DexEventsFromShredTransactionWireBestEffort` API retains legacy placeholder behavior. Recognized routes emit `PumpRouteIntent`; arbitrary CPI cannot be reconstructed from raw wires. Inner CPI/log-only events still require Yellowstone/RPC paths.

---

## Examples

Run from the **repository root** after `go mod tidy`. One row per source file (links point to GitHub `main`).

| Description | Run command | Source |
|-------------|-------------|--------|
| **PumpFun** | | |
| PumpFun `DexEvent` + metrics | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_with_metrics.go` | [pumpfun_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_with_metrics.go) | `v0.5.15` |
| PumpFun trade filter | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_trade_filter.go` | [pumpfun_trade_filter.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_trade_filter.go) | `v0.5.15` |
| Quick connection test | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpfun_quick_test.go` | [pumpfun_quick_test.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpfun_quick_test.go) | `v0.5.15` |
| **PumpSwap** | | |
| PumpSwap + metrics | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_with_metrics.go` | [pumpswap_with_metrics.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_with_metrics.go) | `v0.5.15` |
| Ultra-low latency | `GRPC_URL=… GRPC_TOKEN=… go run examples/pumpswap_low_latency.go` | [pumpswap_low_latency.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/pumpswap_low_latency.go) | `v0.5.15` |
| **Meteora DAMM** | | |
| Meteora DAMM V2 | `GRPC_URL=… GRPC_TOKEN=… go run examples/meteora_damm_grpc.go` | [meteora_damm_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/meteora_damm_grpc.go) | `v0.5.15` |
| **ShredStream** (see **step 5** above) | | |
| SubscribeEntries + decode + optional `DexEvent` JSON | `SHRED_URL=http://host:port go run examples/shredstream_entries.go` | [shredstream_entries.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/shredstream_entries.go) | `v0.5.15` |
| **Yellowstone** | | |
| Geyser subscribe + `ParseSubscribeTransaction` | `GRPC_URL=… GRPC_TOKEN=… go run examples/yellowstone_grpc_parse.go` | [yellowstone_grpc_parse.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/yellowstone_grpc_parse.go) | `v0.5.15` |
| **Multi-protocol** | | |
| All supported DEX programs | `GRPC_URL=… GRPC_TOKEN=… go run examples/multi_protocol_grpc.go` | [multi_protocol_grpc.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/multi_protocol_grpc.go) | `v0.5.15` |
| **Utility** | | |
| Parse tx by signature (RPC, not gRPC stream) | `TX_SIGNATURE=… RPC_URL=… go run examples/parse_tx_by_signature.go` | [parse_tx_by_signature.go](https://github.com/0xfnzero/sol-parser-sdk-golang/blob/main/examples/parse_tx_by_signature.go) | `v0.5.15` |

---

## Protocols

PumpFun, PumpSwap, Raydium AMM V4 / CLMM / CPMM, Orca Whirlpool, Meteora DAMM V2 / DLMM, Raydium LaunchLab (see `solparser/`).

---

## Useful exports

- **`ParseSubscribeTransaction`** — Geyser single-tx → `[]DexEvent` (instructions + logs + merge + Pump account fill).
- **`ParseRpcTransaction`** / **`ParseTransactionFromRpc`** — HTTP RPC JSON → events.
- **`ParseInstructionUnified`** / **`ParseInnerInstructionUnified`** — outer 8-byte / inner 16-byte discriminators.
- **`DexEventsFromShredTransactionWire`** — wire tx bytes → outer `ParseInstructionUnified` (strict static keys; cached ALT variants available).
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

## API compatibility

Full cross-language API parity is still in progress. See the [API migration guide](docs/USAGE.md#api-compatibility) for public behavior and support boundaries, and the examples directory for usage.

### Offline simulation route evidence

`solparser.AnalyzeSimulationRoutes(wire, simulationResponseJSON, graduatedPools)` accepts original transaction bytes and a complete `simulateTransaction` JSON response. It makes no RPC calls. Parsed SPL transfers, compiled CPI, execution failure, and unknown Token-2022 net credits are preserved. See [docs/USAGE.md#simulation-routes](docs/USAGE.md#simulation-routes) and `go run ./examples/simulation_routes solparser/testdata/cached_tip_routes_20261002.json route-buy`.


实时 parser → trade 接入使用 Yellowstone **gRPC**；见 [gRPC 缓存接入与三语言示例](docs/USAGE.md#grpc-cache)。此路径不使用 WebSocket，报价和构建热路径不调用 RPC。

Shared PumpFun create/create_v2 mainnet fixtures, replay instructions and verification limits are documented in the [validation guide](https://github.com/0xfnzero/sol-parser-sdk-golang/tree/main/validation/pumpfun_create_20261007).

### Offline Pump ShredStream regressions

Pump route intents, strict cached ALT resolution and native signed replay/SBF simulation examples are documented in [examples/offline_bank/README.md](examples/offline_bank/README.md). Run the language-native examples and tests from this repository; no other-language test driver or RPC endpoint is needed. Unresolved instructions now skip by default; use the explicit best-effort API only for legacy behavior. Route intents must not be treated as confirmed fills.
