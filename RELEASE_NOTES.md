# sol-parser-sdk-golang v0.11.0

Released November 3, 2026.

## Changes

Remove TradeParseStatus and related parse_status fields. Preserve third_program provenance without inferring executed amounts from third-party wrapper inputs. Carry transaction_version and ordered ALT addresses into Pump.fun creation events. Expand Shred ALT resolver context with receive time, transaction index and event filters; support x-token authentication. Parsing performs no RPC.

Synchronize V1 transaction decoding and real-transaction replay coverage. Node.js, Go and Python retain CLMM instruction limits separately from execution data and preserve token balances through public RPC/gRPC conversions. Go Shred preserves HTTPS/TLS; Python shares one conversion implementation and rejects invalid CLMM boolean bytes.

## Migration

Remove references to TradeParseStatus / PumpFunParseStatus / parse_status. Use third_program to inspect wrapper provenance and distinguish instruction inputs from execution fields. Update custom ALT resolver signatures to accept the expanded context; see the language-specific Shred example. Rust event/config struct literals must initialize new fields or use Default. Node.js consumers must use @solana/web3.js 1.99+ for V1 decoding.

## Validation

Native-language regression suites, captured-mainnet transaction replay, package builds and cross-language review passed. Live provider subscriptions were not exercised during this release.
