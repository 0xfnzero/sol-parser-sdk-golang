# Cross-language synchronization (v0.10.0)

The SDK and examples use their own language. All replays below run offline; no parser, price tracker, or cached quote hot path requests RPC.

## Trade evidence and wrappers

RPC/gRPC uses supplied execution metadata and CPI instructions. Supported trade events expose `parse_status` and optional `third_program`. `ExecutionConfirmed` means a successful transaction supplied execution context; it does not guarantee every amount is known. Instruction budgets and minimum outputs are not fills. `InstructionDecoded` denotes decoded instructions without execution confirmation.

For unknown Shred wrappers whose own account list references Pump.fun, the default callback is `PumpFunTrade` with `Unresolved`. Only a candidate is known: direction, method, mint, pool and amounts remain defaults. A Pump.fun account reference does not prove a CPI occurred. Count candidates separately from confirmed trades. A missing ALT address remains null and cannot establish a Pump.fun reference.

`third_program` contains the outer `program_id`, `instruction_index`, ordered `accounts` (including duplicates/nulls), original `account_indices`, and byte-array `instruction_data`. Event metadata retains transaction position and receive time. Entries positions are local to a message, not slot-wide execution ordering.

Direction-only subscriptions do not receive unresolved candidates. RawInstruction and PumpRouteIntent callbacks have been removed; subscribe to PumpFunTrade and inspect third_program and parse_status. Unknown non-Pump routers do not become confirmed swaps from account references alone. RPC/gRPC decodes supported native CPI swaps normally.

Wrapper input remains opaque. The SDK does not reconstruct third-party instruction arguments or install wrapper rules. See [callback and migration details](wrapped-pumpfun.md).

Native Pump.fun ALT inputs remain parseable without resolving every declared address. Missing quote/token-program fields stay unknown. The existing `PumpFunCreateV2` language variant retains `ix_name=create_v2`; a raw create scan preserves `is_created_buy` even with a trade-only filter. Official PumpSwap `multi_hop_swap` emits one input-only PumpFunTrade per canonical bonding-curve hop. A fixed Pump program account alone is insufficient: all-AMM routes must not emit curve trades. Direction follows unambiguous connected mint pairs; ambiguous direction remains Unresolved. Route budgets and minimums are not per-hop fills. Original route inputs remain in third_program.

## Stock catalog and prices

Both parser and trade SDKs ship the same historical stock-token/pool catalog as Rust. Opt in through the screened catalog API and extend your owned copy. The catalog records fee/reserve observations; it is not live liquidity, a guaranteed executable depth, or a current tax quote. Pool preference reflects historical fees. Revalidate current state before trading.

The parser's `StockPriceTracker` consumes existing transaction events and produces notifications containing mint, symbol, pool, DEX, quote mint, reported execution price, optional SOL/USDC conversion, and optional supplied liquidity. It requires confirmed execution evidence and validates protocol-specific fields. Unresolved Shred placeholders never produce prices.

The SOL/USDC reference must precede the stock transaction, be fresh, and match the configured reference pool/protocol. Same-transaction conversion is withheld. A stale or missing reference leaves the unavailable denomination null. Supply coherent liquidity separately; unknown liquidity stays null. Reset the tracker on disconnect, dropped events, or a selected-fork change. Offline capture timestamps are replay inputs, not current market time.

Conversion-leg matching checks successful execution, protocol, program ID, pool and both mints. It returns independent matching legs, not a guaranteed connected path. StonkFun naming is retained. Pump.fun/StonkFun/PumpSwap launch origin does not restrict the conversion venue to Raydium or PumpSwap.

## Cached trade preparation

Trade SDKs expose a reusable `CandidateRouteGraph`: construct once when pool identities change and reuse for multiple searches. It owns topology, yields independent paths, enforces budgets and preserves deterministic breadth-first ordering. The compatibility route functions still build a graph per call.

Observed snapshots provide a separate observation watermark without changing actual account slot/write version. Only attest observation after a coherent baseline and continuous provider coverage of every dependency. Account replay, local time and unrelated slot heartbeats do not establish this. Readiness loss invalidates snapshots. Ordinary snapshots retain their conservative age checks.

## Native examples

Pump.fun ALT and official multi-hop replay: `go run ./examples/pump_execution_replay`. The offline default includes the three reported `create_v2` transactions and the official `multi_hop_swap` transaction. Pass a signature to fetch a different transaction. Static-key Shred, resolved-ALT Shred, and execution replay are printed separately; `tx_index=888` and receive time in this example are caller-supplied demonstration values. Creation callbacks precede their buys, including when the subscription uses the generic create/trade filter. This language retains the `PumpFunCreateV2` event name with `ix_name=create_v2`. Identifiable curve hops emit `PumpFunTrade`, `ix_name=multi_hop_swap`, and the original route in `third_program`. Shred has no per-hop execution amounts, so those fields stay zero. gRPC/RPC retain actual fills and non-SOL `quote_amount`; `sol_amount=0` is valid for a non-SOL quote.

Run from the repository root after installing its development dependencies:

```sh
go run ./examples/stock_price_replay
go run ./examples/stock_route_replay
go run ./examples/pump_execution_replay
```

Stock examples accept `--fixture FILE`. Parser captures are dated 2026-10-20; trade examples replay normalized legs from those same signatures. Historical wrapper captures remain regression evidence and are not labeled live transactions.
