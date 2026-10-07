# PumpFun create/create_v2 offline validation

This directory preserves 36 finalized mainnet transaction responses: 3 successful legacy create, 25 successful create_v2, 6 failed legacy create and 2 failed create_v2. The corpus covers creation-account selection, historical CreateEvent layouts and failed-transaction suppression.

## Retained inputs

- `transactions/`: original RPC responses used by the replay and regression tests.
- `pump_public_idl.json`: official IDL snapshot.
- `sources.json`: source URL, collection provenance and IDL SHA-256.
- `replay.go`: offline instruction/log/RPC/reconstructed-Yellowstone comparison.
- `summarize.py`: checks the replay output and optionally saves a compact summary.

Derived results, discovery responses and pre-fix reports are excluded from maintained inputs. Replay does not contact RPC, sign or submit transactions. Its outputs are written to a directory you choose rather than overwriting this guide.

## Replay

Run from the repository root:

```sh
go run validation/pumpfun_create_20261007/replay.go validation/pumpfun_create_20261007/transactions > /tmp/pumpfun-create-results.json
python3 validation/pumpfun_create_20261007/summarize.py --results /tmp/pumpfun-create-results.json --output /tmp/pumpfun-create-summary.json
```

Successful samples must emit exactly one creation event with account fields selected by the actual create/create_v2 discriminator and mint. RPC and reconstructed Yellowstone creation fields must agree, excluding metadata. Failed samples must emit no executed creation events; raw instruction/log parsing is retained only as intent evidence.

The sampled create/create_v2 instructions have 14/16 fixed accounts. Unrelated buy instructions and arbitrary remaining accounts must not supply creation quote fields. Historical logs can omit fields; unavailable values remain unspecified.

## Limits

This is a bounded historical corpus, not live-stream acceptance or a guarantee of complete protocol coverage. It contains no confirmed USDC-quoted launch. Yellowstone messages are reconstructed from RPC data, not independently captured; balance fields and other execution details are not a full bank replay. TokenProgram comparisons cover the creating instruction only. `sources.json` records the capture baseline; replay always uses the parser checked out locally.
