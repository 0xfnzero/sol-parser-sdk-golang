# Changelog

## v0.9.0 — 2026-10-22

Synchronize SDK code, examples, version navigation and AI release rules. Read the [release and migration notes](RELEASE_NOTES.md) for callback removals, V3 defaults and the Go /v9 migration.

Release and integration notes preserved from the previous README. APIs, dependencies and validation results describe those releases, not the current contract. See the [README](README.md) and [release and migration notes](RELEASE_NOTES.md) for current usage.

## v0.6.0

See [release and migration notes](RELEASE_NOTES.md).

## v0.5.17 — Shred metadata passthrough and ordering scope

Preserve and verify transaction position and receive time through raw instruction callbacks. Entries indexes are message-local, unlike Yellowstone node indexes; use signatures for cross-message deduplication. This release clarifies API documentation and adds regression protection without changing metadata fields or numbering behavior.

## v0.5.17 — DEX-scoped RawInstruction and typed DEX identifiers

Raw instructions now respect DEX subscriptions using programs referenced by each instruction. `RawInstructionDexType` provides typed identifiers with unchanged JSON strings. Unknown router layouts do not establish pool addresses, trade direction, or method versions.

## v0.5.17 — Pump execution replay

Current Anchor TradeEvent CPI decoding is supported alongside the historical encoding. RPC/gRPC instruction budgets and minimum outputs remain in instruction fields. Execution fields `sol_amount`, `quote_amount`, and `token_amount` come from execution events; without logs or event CPIs they remain zero (unavailable, not a proven zero fill).

Run `go run ./examples/pump_execution_replay` from this repository to replay real transactions fetched on 2026-10-19 offline. The example prints full/no-log/truncated-log results. Append a recent signature for online replay. Set `SOLANA_RPC_URL` for your RPC; transaction v1 is supported.

The current Delta sell signature is `4ARos3Yq1SMNE3izDpy9YdKrRgMmD7UCDxvnVrBRaGCJHiifPHSmQWc4DPfjUaEUcxS7puGmEN64K7EZdeXLJaj2` (slot 458542546). The dated fixture also includes Rapid create/buy, BuyV3, and PumpSwap buy with their full signatures and slots. Older five-combination multi-hop fixtures are retained as regression tests, not labeled as new mainnet captures.

## v0.5.17 — raw Shred instructions

Default Shred parsing emits unresolved `PumpFunTrade` candidates for unknown wrappers referencing Pump.fun. Their `third_program` preserves raw input; direction, amounts and internal methods are not inferred. Official Pump/PumpSwap multi-hop instructions retain `PumpRouteIntent`.

The payload contains `program_id`, `instruction_index`, `accounts`, `account_indices`, `instruction_data`, and transaction metadata. Account order and duplicates are preserved. Unresolved ALT addresses are null/None, with original indices retained; supply cached loaded addresses or a resolver. No hot-path RPC is performed. Data is copied verbatim without requiring a known layout.

Run `go run ./examples/raw_instruction_filter` to compare PumpFun-only, Raydium CPMM-only, and all-raw subscriptions against signed offline transactions; the output includes `dexes`.

The raw event retains `metadata.tx_index` and `metadata.grpc_recv_us` from the parser arguments (streamer names the receive field `metadata.recv_us`). The default Entries client index is local to one received Entries message and restarts at zero for the next message; it is not a confirmed slot-wide transaction index. `instruction_index` orders outer instructions within a transaction. Receive time is a local microsecond timestamp, not an execution-order guarantee. Standalone replay APIs may default these values to zero; pass the upstream values when available.

Default and `PumpFunTrade` subscriptions receive unresolved Pump.fun candidates. Direction-only subscriptions do not. Explicit `RawInstruction` requests all outer instructions for compatibility; excluding it does not suppress Pump.fun callbacks. Unknown wrappers referencing other DEXes require explicit raw subscription in Shred; gRPC uses actual supported CPI instructions. See the [current evidence and filtering contract](docs/cross-language-sync.md).

Wrapper adapter functions remain explicit compatibility helpers only; default Shred parsing does not call them. The released-version history below describes previous behavior.

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
