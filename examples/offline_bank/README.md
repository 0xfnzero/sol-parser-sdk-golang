# Native Go Pump regression examples

Run from the repository root:

```sh
go mod download
(cd examples/pump_offline_bank/engine && go mod download)
go run ./examples/pump_shred_routes
go run ./examples/pump_offline_bank
go test ./...
```

Install dependencies once before offline execution. After that both examples use only bundled files; no endpoint, funded wallet or broadcast is required. The replay example verifies all signatures on six original mainnet wires, including V1 wrappers, with strict ALT resolution, metadata, hops, filters, malformed layouts and repeatability. It also signs direct curve route boundary cases. This is signed-wire replay, not wrapper-program execution.

The bank example uses `examples/offline_bank/config.json` by default. An optional first argument selects another JSON config. Snapshot paths remain relative to `examples/offline_bank`; `repetitions` must be at least 2. Tests call the same native example functions. Every repetition builds and signs 20 new V0 transactions and executes SBF locally: curve→curve buy/sell, pool→curve, curve→pool, pool→pool × static/ALT × success/slippage failure. It checks signature rejection, invalid writable/readonly ALT indices, parent rollback with fee debit, route-intent arguments and per-hop log fields against an independent oracle. Repeated signed wires and execution logs must match. Each example also feeds local execution status into the offline RPC-format decoder and verifies parent failures produce no settled fills; this conversion makes no RPC request.

The Go bank engine is an isolated module (Go 1.25.7+), pinned to `litesvm-go v0.3.0`. The library retains Go 1.22 and its existing runtime dependencies. The example launches only a native Go engine process, then parses its signed outputs in the SDK process. This avoids a protobuf namespace collision between the VM dependency's Yellowstone definitions and the SDK; protobuf conflict checks remain enabled. The pinned Go VM reports parent error index 0 for this fixture while Rust/Python/Node report 1; the SDK independently checks the wire instruction index is 1. The Go engine asserts non-mutation, state changes, rollback and fee debit; Python/Node also compare committed accounts to simulated post-accounts.

## Parsing contract

`DexEventsFromShredTransactionWire` is strict by default. Use `DexEventsFromShredTransactionWireWithLoadedAddresses` or `DexEventsFromShredTransactionWireWithResolver` for cached ALT addresses. `ShredAddressLookupResolver` receives the decoded transaction and slot; return an error on a cache miss. `ResolveShredAddressesFromCache` resolves runtime writable-then-readonly ordering from an already populated table map, checking bounds. The caller must select a valid slot-aware cache snapshot. The resolver must never fetch RPC data. `DexEventsFromShredTransactionWireBestEffort` explicitly opts into legacy placeholders; route adapters always require resolved keys.

`PumpRouteIntent` describes the outer instruction, not an executed fill. Supported adapters bind to the exact official Pump/Pump AMM, Rapid and Delta program IDs and account layouts. The existing `PumpFunTrade` filter includes these intents; `PumpSwapTrade` includes only AMM routes. Create/directional-only filters do not include intents; explicitly excluding `PumpRouteIntent` wins. Delta's input depends on runtime balances, so unknown amounts stay null; opaque instruction bytes are retained. A zero minimum output is valid; zero declared input is rejected.

Strict mode skips instructions containing unresolved accounts or programs; it does not guess addresses. Cache callbacks are synchronous and must never issue RPC. Loaded addresses are validated before producing events. Raw shreds lack CPI results and transaction status; they can describe an intent even when execution later fails. Log events from failed simulations describe attempted work and must not be booked as settled fills.

## Frozen input provenance

`snapshot.json.gz` contains pre-route account state, transaction inputs and unmodified deployed SBF bytes from the 2026-10-09 Pump validation fork. Users were locally funded; quote depth and synthetic token-paired coins/pools were prepared locally. The example rekeys users, ATAs and user-volume PDAs to a deterministic public test key. It stores no captured execution results. Each run executes programs again.

`config.json` pins the compressed SHA-256; `programs.json` records program hashes; `expected.json` is an independently decoded official-IDL oracle. The public deterministic key must never receive funds. These checks cover the pinned deployment, not arbitrary wrappers or production latency. Third-party wrapper programs are verified by signed-wire replay only.
