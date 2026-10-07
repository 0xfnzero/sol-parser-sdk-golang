# PumpFun create/create_v2 mainnet verification — 2026-10-07

Baseline parser commit: `b057e687d384e7a4cd912c086fc509f5d4f4fbfb` (v0.5.9). Current outputs include the local fix; pre-fix outputs are preserved as `results_before_fix.json` and `summary_before_fix.json`.
RPC source: `https://api.mainnet-beta.solana.com`; finalized `getTransaction`, `encoding=json`, `maxSupportedTransactionVersion=0`. Only read-only requests were made.

Fetched and replayed 36 real transactions: 3 successful legacy create, 25 successful create_v2, 6 failed legacy create and 2 failed create_v2. Failed transactions are retained as evidence but excluded from successful creation statistics.

Before the fix: 9 of 25 successful create_v2 transactions acquired an unrelated QuoteMint in full RPC and reconstructed Yellowstone parsing. After the fix: 0. All sampled direct create_v2 instructions have 16 accounts and no quote accounts. Before the fix, the erroneous QuoteMint was exactly account index 16 of the longest PumpFun invocation, rather than the create_v2 invocation.

Root cause (before fix): PumpFunCreateV2 used the generic account getter, which chose the invocation with the largest account list without checking its discriminator. A later buy could win and its accounts were interpreted using create_v2 indices. The fix selects create/create_v2 by discriminator and event mint, supports inner invocations, declines ambiguous or unrelated matches, and removes unverified quote inference from remaining accounts. Authoritative decoded quote fields are preserved. The actual all-zero System Program address is retained as a valid account.

The direct instruction path also no longer interprets arbitrary remaining accounts as quote fields. None of the 36 real creation instructions has more than its 14/16 fixed accounts; a synthetic regression covers that additional trigger.

TokenProgram: all successful samples agree with the actual creating instruction token-program account (legacy SPL Token; create_v2 Token-2022). This dataset does not reproduce issue #4.

Historical logs: before the fix, 2 successful legacy create transactions (2024) had an unrecognized CreateEvent containing only the three strings and mint/bonding_curve/user keys. After the fix, every successful sample emits exactly one create log; omissions: 0. Fields unavailable in the historical layout remain unspecified rather than being invented.

Limitations: no confirmed USDC-quoted launch was found in this collection, so this does not verify the reported USDC-specific transaction or issue #5. Default/zero QuoteMint values in the log path are treated as unspecified, not evidence of a USDC mint. The gRPC test reconstructs Yellowstone protobuf transaction/account/log data from the raw RPC response; it is not an independently captured live Yellowstone stream. Metadata block time differs by API; comparisons exclude metadata. Token balances and failed-status protobuf details are not reconstructed, so those are outside this create-account comparison.

Replay (from repository root):
```sh
go run validation/pumpfun_create_20261007/replay.go validation/pumpfun_create_20261007/transactions > validation/pumpfun_create_20261007/results.json
python3 validation/pumpfun_create_20261007/summarize.py
```

Raw transaction responses, account mappings, parser outputs and SHA-256 hashes are saved alongside this report. Official public IDL snapshot is saved as `pump_public_idl.json` with its source URL in `sources.json`.

## Successful transaction comparisons

| Kind | Signature | Slot | Create accounts | Log create | RPC quote corruption | RPC QuoteMint |
|---|---|---:|---:|---:|---|---|
| PumpFunCreateV2 | [2G2YFBmJBjeUMjBW73RvpNJqKq9eEsZ6DtASecvgTQvKM6JZPQQtc3teHDtLvxUgycJxxygyXy2bpcKc4RtAcYw8](https://solscan.io/tx/2G2YFBmJBjeUMjBW73RvpNJqKq9eEsZ6DtASecvgTQvKM6JZPQQtc3teHDtLvxUgycJxxygyXy2bpcKc4RtAcYw8) | 389243770 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreate | [2MfmZSU8mWEbskNHAxQGphAyB5x72SC2HYuAkLeSyH6MjttUHqUjm24wPMeH7zrCoQoYFwbLeiFE571mc8WvgZyL](https://solscan.io/tx/2MfmZSU8mWEbskNHAxQGphAyB5x72SC2HYuAkLeSyH6MjttUHqUjm24wPMeH7zrCoQoYFwbLeiFE571mc8WvgZyL) | 299999990 | 14 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [2WpWrAHQrnp9R5VjeJDLTQ3oKeDqBreTtFhSi9cgSKjzkiht3gpoNoqmnQASJNMySBXwrsZqjvxx8Pim3fve2j14](https://solscan.io/tx/2WpWrAHQrnp9R5VjeJDLTQ3oKeDqBreTtFhSi9cgSKjzkiht3gpoNoqmnQASJNMySBXwrsZqjvxx8Pim3fve2j14) | 454059406 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [2eVpZNvQYPMBVNT5357Qt7HVAUSv9gdwqxKStvdwQ498rcJLLCHVcYPaGfFqVBz44ndYvkA1H4CwjBuDSviAj7jp](https://solscan.io/tx/2eVpZNvQYPMBVNT5357Qt7HVAUSv9gdwqxKStvdwQ498rcJLLCHVcYPaGfFqVBz44ndYvkA1H4CwjBuDSviAj7jp) | 454059438 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [2xWgdL54hfHN6zjZTckUJX8fQtRrRAJLCa8NJFYjWACH1WFo7reYJmGN4oBe1JbZF8drgVD3boL2W6TXnsvxptc9](https://solscan.io/tx/2xWgdL54hfHN6zjZTckUJX8fQtRrRAJLCa8NJFYjWACH1WFo7reYJmGN4oBe1JbZF8drgVD3boL2W6TXnsvxptc9) | 389243740 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [38j2VLxLmKvytugUhgMv8tanS3ai18znhpkALcgJFFegyWBYKYsrkW9NdUdUDhW2sBatQ8NJEHzzD5rHTDnrGxDJ](https://solscan.io/tx/38j2VLxLmKvytugUhgMv8tanS3ai18znhpkALcgJFFegyWBYKYsrkW9NdUdUDhW2sBatQ8NJEHzzD5rHTDnrGxDJ) | 454059356 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [3AgrmuGjTL4tThqNambGywNyZ6gYhro7hRBpUFCLGr7J62yMD4LHdSwqkFgCJir72HdsN62H8bYkjX4ipZDryDxD](https://solscan.io/tx/3AgrmuGjTL4tThqNambGywNyZ6gYhro7hRBpUFCLGr7J62yMD4LHdSwqkFgCJir72HdsN62H8bYkjX4ipZDryDxD) | 389243699 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [3MwqHs2YKFFAkr6cAaDPd2ewjserQdm2kxFjkpwikjofHx8sZxRLVsjtvRgV6uJ337edbwNXBQ3yugidhnozXar4](https://solscan.io/tx/3MwqHs2YKFFAkr6cAaDPd2ewjserQdm2kxFjkpwikjofHx8sZxRLVsjtvRgV6uJ337edbwNXBQ3yugidhnozXar4) | 454059398 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [3TZNP9rM7fNdbNth4cjPWRrpZQRHUrfZsjwQH3JE42FrpkMqmjYUzf4UwLF6uTiupig6ZSsDxhZJAyhfN1E3jz4A](https://solscan.io/tx/3TZNP9rM7fNdbNth4cjPWRrpZQRHUrfZsjwQH3JE42FrpkMqmjYUzf4UwLF6uTiupig6ZSsDxhZJAyhfN1E3jz4A) | 454059415 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [3YWdxPUWjRDPZV69t79Z5yiphknL78RtNWV1cZULxJ3QyybptcwwBxteoGmwn7NhhCWsGxrrCo3ii8KdveTwUsVQ](https://solscan.io/tx/3YWdxPUWjRDPZV69t79Z5yiphknL78RtNWV1cZULxJ3QyybptcwwBxteoGmwn7NhhCWsGxrrCo3ii8KdveTwUsVQ) | 389243783 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreate | [3ejAydePsZTUfeZaU9eiAM37gV5AYPG9hhyMtUobJLQtSMgroDeiE8njXu2JiwMyX5YJLp3vrfNPy89S7EDdUCAy](https://solscan.io/tx/3ejAydePsZTUfeZaU9eiAM37gV5AYPG9hhyMtUobJLQtSMgroDeiE8njXu2JiwMyX5YJLp3vrfNPy89S7EDdUCAy) | 389243886 | 14 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [4DqRrNJWhS3snWEAr64Mzefq6kjipzxLWVRFebETM5KgAqRK4k5j3YYfp22Bwi1XQLFLS6erYhSnZp9LCrfuEGU5](https://solscan.io/tx/4DqRrNJWhS3snWEAr64Mzefq6kjipzxLWVRFebETM5KgAqRK4k5j3YYfp22Bwi1XQLFLS6erYhSnZp9LCrfuEGU5) | 389243884 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [4LErdWwuJ2CnNB6ftbPKLAmVh1vLz5yrProFNqL6SEibRAV9Yf2dQ8juCXH6n8XbaogxPjs4YNuWYgpKVNx3bdT4](https://solscan.io/tx/4LErdWwuJ2CnNB6ftbPKLAmVh1vLz5yrProFNqL6SEibRAV9Yf2dQ8juCXH6n8XbaogxPjs4YNuWYgpKVNx3bdT4) | 389243803 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [4WSwDTuvnpcXsFTSjxborTSAUbqmBNiTza9756Jo6N5Wdc7zzz9UdXuDovHkRTeCKaGd44fhtUFksoSbLYTioBFs](https://solscan.io/tx/4WSwDTuvnpcXsFTSjxborTSAUbqmBNiTza9756Jo6N5Wdc7zzz9UdXuDovHkRTeCKaGd44fhtUFksoSbLYTioBFs) | 454059392 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [56L2Dtm4jeeYLtFqfVYMvzuvquMBMrLmUmWAKMjHBgSG398RnWU38dkLthbNm5Qs7v5ziATCNoLdnseo4Dvc8xLQ](https://solscan.io/tx/56L2Dtm4jeeYLtFqfVYMvzuvquMBMrLmUmWAKMjHBgSG398RnWU38dkLthbNm5Qs7v5ziATCNoLdnseo4Dvc8xLQ) | 454059390 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [5PgkFgTLSEZPZhxgWNToPzTZct95SycuaH9H3g82YwcKXudtiynEZmiaoQrhGv2gdZLrvsVZ5Wn88KGirbFYpLrt](https://solscan.io/tx/5PgkFgTLSEZPZhxgWNToPzTZct95SycuaH9H3g82YwcKXudtiynEZmiaoQrhGv2gdZLrvsVZ5Wn88KGirbFYpLrt) | 454059335 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [5TszGhT2Qd6ivqkp7WKedhaXbB6f93PdoBSS8yLMcWZseTqRAzbDrL9BfvXfv68YyvVZUC5RYKb2spBfQsp8MMdg](https://solscan.io/tx/5TszGhT2Qd6ivqkp7WKedhaXbB6f93PdoBSS8yLMcWZseTqRAzbDrL9BfvXfv68YyvVZUC5RYKb2spBfQsp8MMdg) | 389243816 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [5kkFU81HLnQHrzd6bwUT92G4DiYvKbh14TawvEbuDgrptZ1ZQedyoXxGWGax2Kcx4GuNCLTPiy2jK5jVeXoWQe1](https://solscan.io/tx/5kkFU81HLnQHrzd6bwUT92G4DiYvKbh14TawvEbuDgrptZ1ZQedyoXxGWGax2Kcx4GuNCLTPiy2jK5jVeXoWQe1) | 454059425 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [5km9pHejwA7EJqsLKfkRZ4ELQnbmCFR3kMWbeSSr8fj9pYNHymq64ky1vpyBGfcGHL24FWiCFbsBjg7be6Mix1jZ](https://solscan.io/tx/5km9pHejwA7EJqsLKfkRZ4ELQnbmCFR3kMWbeSSr8fj9pYNHymq64ky1vpyBGfcGHL24FWiCFbsBjg7be6Mix1jZ) | 389243904 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [5tAzKPx3jv1vke1PygJW31LBVnyF86rPd9CTyfJEYMgxRoihFE8VwWFT98X6b5RDP4NjSUyo9v7V2ojJovTr4D6q](https://solscan.io/tx/5tAzKPx3jv1vke1PygJW31LBVnyF86rPd9CTyfJEYMgxRoihFE8VwWFT98X6b5RDP4NjSUyo9v7V2ojJovTr4D6q) | 389243877 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreate | [661yA5s6uxUsggTcT4cZ6yyMfNmF4mCZzxLnCb1zBrSY3YhAiouYXBitiPvkngBEKxaeGcpXiek6qXDajiRr9wuY](https://solscan.io/tx/661yA5s6uxUsggTcT4cZ6yyMfNmF4mCZzxLnCb1zBrSY3YhAiouYXBitiPvkngBEKxaeGcpXiek6qXDajiRr9wuY) | 299999988 | 14 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [67TBoA9v7vjGBjPbCmtKJdLwLZdroEctf3a7UCzziCdqKV2ne79myRBcfxG3vQ9TWNZb4yvNHw3ViZ2DdZh6W11p](https://solscan.io/tx/67TBoA9v7vjGBjPbCmtKJdLwLZdroEctf3a7UCzziCdqKV2ne79myRBcfxG3vQ9TWNZb4yvNHw3ViZ2DdZh6W11p) | 389243731 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [CCQnP4wEUuNE915jqo7b2LFVhzayiWMuDGhmLiYKkx8YcajiPWEhBayFhZmMe9DZfbEHpsmgq8hAcEuZWargyhM](https://solscan.io/tx/CCQnP4wEUuNE915jqo7b2LFVhzayiWMuDGhmLiYKkx8YcajiPWEhBayFhZmMe9DZfbEHpsmgq8hAcEuZWargyhM) | 389243832 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [HMRPdoNjLHc6mGwWSFGd7ST995iMwbWkkG2bxVgsHKXeJpR9hxnpd1kWX9HdjDSqSwKkEZqt76MRtDzVEB9DQkt](https://solscan.io/tx/HMRPdoNjLHc6mGwWSFGd7ST995iMwbWkkG2bxVgsHKXeJpR9hxnpd1kWX9HdjDSqSwKkEZqt76MRtDzVEB9DQkt) | 454059370 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [Kk6VJVTHFhHw2z4dmHNfpmP1mhSxvNY2v6PBjLDUvRP5BXBUuZBDAKnhAjYgFwrY2UrAQU8Ht2KmPiRhi4jTkAa](https://solscan.io/tx/Kk6VJVTHFhHw2z4dmHNfpmP1mhSxvNY2v6PBjLDUvRP5BXBUuZBDAKnhAjYgFwrY2UrAQU8Ht2KmPiRhi4jTkAa) | 454059356 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [MrD2Xcb9n6noXCFcihLrZsqYRNkQfm8VCetHYXi3DwYpfxAD9ug4d6CEMegqRZPdpDZtWnicogeNEq5UKihBNwm](https://solscan.io/tx/MrD2Xcb9n6noXCFcihLrZsqYRNkQfm8VCetHYXi3DwYpfxAD9ug4d6CEMegqRZPdpDZtWnicogeNEq5UKihBNwm) | 454059428 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [WZa4Q4JXs4qN6NT7L6FtXdyTsQa1LcgaHcozTzwv5frnpFDb4LH1Umwhtqt2iu1X3HET4z6HrneicMUkUU4eFcB](https://solscan.io/tx/WZa4Q4JXs4qN6NT7L6FtXdyTsQa1LcgaHcozTzwv5frnpFDb4LH1Umwhtqt2iu1X3HET4z6HrneicMUkUU4eFcB) | 389243709 | 16 | 1 | no | `11111111111111111111111111111111` |
| PumpFunCreateV2 | [tDo6DVYrj6q3i8y532dryX2JAtkmQXWgH8hdEU9G2E87QCaRHRJLHpJHpRYJnLEeabQ7s7yJBobdUyTxv9Y9yYr](https://solscan.io/tx/tDo6DVYrj6q3i8y532dryX2JAtkmQXWgH8hdEU9G2E87QCaRHRJLHpJHpRYJnLEeabQ7s7yJBobdUyTxv9Y9yYr) | 454059395 | 16 | 1 | no | `11111111111111111111111111111111` |


---

# PumpFun create / create_v2 cross-language regression verification

Date: 2026-10-07. At verification time, changes were local and uncommitted. Publication is tracked by the repository release.

## Corrected behavior

- Select creation account context by the actual create/create_v2 discriminator and event mint, including CPI invocations. Require the corresponding minimum IDL account count (14/16). Decline ambiguous matches; never select a larger buy invocation as creation context.
- The current official public Pump IDL defines 16 fixed create_v2 accounts and **no quote mint/vault/token-program accounts**. Positions 16/17/18 are not interpreted as quote accounts. Preserve quote fields from decoded events and existing authoritative enrichment.
- Recognize the historical CreateEvent layout: three Borsh strings followed by exactly 96 bytes (mint, curve, user). Reject partial historical/modern layouts. Rust handles both direct logs and pre-decoded log/CPI event paths under both parser features.
- Rust and sol-shred-sdk also recognize historical create instruction arguments ending after the three strings, leaving unavailable creator metadata unspecified.
- Node.js and Python fill the complete legacy creation account context. Python declares the canonical create account fields so dataclass serialization retains them.
- sol-shred-sdk and the Rust parser SDK apply the quote fix to their ShredStream instruction paths as well.

## Mainnet evidence

Source: `sol-parser-sdk-golang/validation/pumpfun_create_20261007/transactions` and its `sources.json`, `pump_public_idl.json`, and `README.md`.
Official IDL source: https://raw.githubusercontent.com/pump-fun/pump-public-docs/main/idl/pump.json

The shared corpus contains 36 signed getTransaction responses: 3 successful legacy creates, 25 successful create_v2 transactions, and 8 failed transactions. All four SDKs replayed the corpus. For each successful transaction, regression assertions compare creation mint/user/token program against the actual creating instruction and reject unrelated quote account values. Failed transactions are excluded from successful-launch assertions.

Rust RPC verification reconstructs the original signed binary transaction from the unparsed JSON response because its RPC parser expects binary encoding. sol-shred-sdk replays that same reconstructed signed transaction with the actual loaded addresses. These are offline RPC-derived replays, not independent live gRPC or raw-shred captures. This corpus does not establish USDC-launch coverage.

Each repository also includes five self-contained RPC fixtures: two historical legacy creates, a newer legacy create, and two recent create_v2 transactions. Regression tests additionally cover CPI creates, multiple mints, ambiguous duplicate creates, a longer unrelated buy, arbitrary remaining accounts, preservation of decoded quote fields, and historical payload bounds.

## Validation

| SDK | Full test results |
| --- | --- |
| sol-parser-sdk (Rust) | Default: 476 passed / 1 ignored; zero-copy: 474 passed / 1 ignored |
| sol-shred-sdk | Default: 398 passed / 2 ignored; zero-copy: 396 passed / 2 ignored |
| sol-parser-sdk-nodejs | TypeScript build passed; full suite with shared corpus: 302 passed / 8 skipped |
| sol-parser-sdk-python | Full suite with shared corpus: 275 passed |

All repositories passed `git diff --check`. Existing unrelated sol-shred-sdk working changes were preserved.

Run the new regression tests from each repository with `PUMPFUN_CREATE_CORPUS` pointing to the shared transactions directory to repeat all 36 samples; without that variable they use the repository's five saved fixtures.

- Rust: `cargo test --test pumpfun_create_regression`; also `cargo test --no-default-features --features parse-zero-copy --test pumpfun_create_regression`.
- sol-shred-sdk: the same Cargo test commands. On this macOS host, Cargo required `DYLD_LIBRARY_PATH` and `LIBCLANG_PATH` to point to the Xcode toolchain `usr/lib` directory for its existing RocksDB build dependency.
- Node.js: `npm test -- --run src/pumpfun_create_regression.test.ts`.
- Python: `.venv/bin/python -m pytest -q tests/test_pumpfun_create_regression.py`.


---

# Failed-transaction and skipped-test audit — 2026-10-07

## Findings and fix

The eight failed mainnet transactions failed on-chain: six attempted to allocate an already-used mint account; two had insufficient SOL. No parser fix can change their historical chain outcome. A new launch must use an unused mint account and sufficient SOL for the transfer/rent/fees.

Reviewing these previously excluded samples exposed an SDK defect: Go, Node.js and Python could emit instruction/log Create or Buy events for failed transactions, even though all of the transaction's program effects were rolled back. Python RPC metadata omitted `err`; Go and Node.js Yellowstone adapters also lost the failure flag.

The corrected high-level RPC/gRPC event paths suppress all DEX events when the transaction has a failure status. RPC metadata retains the original error; RPC-to-Yellowstone conversion preserves error **presence**, without pretending the JSON error is a Yellowstone binary error enum. Native protobuf errors remain usable as failure indicators even when their encoded payload is empty. Python's full subscription callback suppresses both instruction and log events.

The Rust RPC parser already suppressed failed transactions; the mainnet regression now explicitly asserts this behavior for the eight failed fixtures instead of skipping them. Raw ShredStream instruction decoding in sol-shred-sdk has no execution status and reports instruction intent; callers must use confirmed RPC/gRPC status to decide whether a launch committed.

## Actual failed transactions

| Signature | Slot | Confirmed chain failure |
| --- | ---: | --- |
| [2dabmQiykzC1tsNmgtRFbsRJjEKMfUySgScJc2J2uq8CQtoXWZ3YyT99p3nVuCXifEhLkyUe4Sf4nfe2UkFT8AJr](https://solscan.io/tx/2dabmQiykzC1tsNmgtRFbsRJjEKMfUySgScJc2J2uq8CQtoXWZ3YyT99p3nVuCXifEhLkyUe4Sf4nfe2UkFT8AJr) | 299999997 | Mint account already in use |
| [3GufGVVRrnrH1Yuto57jDeTKXeki6ooE9w6h4qYRvNm5FcCpxwDHNvzHXdpAMbxRmukWJpSPHUKHEjnFTCC979ka](https://solscan.io/tx/3GufGVVRrnrH1Yuto57jDeTKXeki6ooE9w6h4qYRvNm5FcCpxwDHNvzHXdpAMbxRmukWJpSPHUKHEjnFTCC979ka) | 454059364 | Insufficient SOL / lamports |
| [3LQK2JHVbVSSfyCYwTwzoYTY9NEN4gZXzqw3TvAvMJHsvPsopCHtsa5ALhj4tFn8xWbmJ5DsYMMcnPukTBkSjtQf](https://solscan.io/tx/3LQK2JHVbVSSfyCYwTwzoYTY9NEN4gZXzqw3TvAvMJHsvPsopCHtsa5ALhj4tFn8xWbmJ5DsYMMcnPukTBkSjtQf) | 299999997 | Mint account already in use |
| [3MRWKDUcD5CGn5Xz4qWyGA8Y9udw4SsSr4gcxCjTUqCFt6Zc6uiwHF1wLZXKoqBcGEM2kruBS85ymYxmYk8FLGBk](https://solscan.io/tx/3MRWKDUcD5CGn5Xz4qWyGA8Y9udw4SsSr4gcxCjTUqCFt6Zc6uiwHF1wLZXKoqBcGEM2kruBS85ymYxmYk8FLGBk) | 299999997 | Mint account already in use |
| [3NBSCgAXpNqg7JBUzp6RsfZYzf2QbbHi6r2wcoBF4Lpuzozae7ExUJUGoqpVCXXs8cbibr6aWySH1bsg86Enuhez](https://solscan.io/tx/3NBSCgAXpNqg7JBUzp6RsfZYzf2QbbHi6r2wcoBF4Lpuzozae7ExUJUGoqpVCXXs8cbibr6aWySH1bsg86Enuhez) | 299999991 | Mint account already in use |
| [3QVpcUq64D4DUV9empFr3G98pCEa1Cp2fAcFJKyiJGsQ2ir2S4yBLhHYXPNmLfCE8KVDgLApfBYTFM7pf9JVPbsb](https://solscan.io/tx/3QVpcUq64D4DUV9empFr3G98pCEa1Cp2fAcFJKyiJGsQ2ir2S4yBLhHYXPNmLfCE8KVDgLApfBYTFM7pf9JVPbsb) | 389243699 | Insufficient SOL / lamports |
| [3REXMDkvZsWvgc8czYpfn24avz18APMuzbV2VgQBeymArdJvfQcfcpKr46f3j1H2bAiymFneaNocPH7KzPUPbR1S](https://solscan.io/tx/3REXMDkvZsWvgc8czYpfn24avz18APMuzbV2VgQBeymArdJvfQcfcpKr46f3j1H2bAiymFneaNocPH7KzPUPbR1S) | 299999991 | Mint account already in use |
| [3zC4bw1osyebZ77KagXYP3b5rEBHvJCgcwCriCLc5gAUp9Ev6Bx5ZMbqnYpkKE2VgZgMFcZN5FYKQfcKL45MWpyu](https://solscan.io/tx/3zC4bw1osyebZ77KagXYP3b5rEBHvJCgcwCriCLc5gAUp9Ev6Bx5ZMbqnYpkKE2VgZgMFcZN5FYKQfcKL45MWpyu) | 299999991 | Mint account already in use |

The detailed JSON audit retains `meta.err` and the actual failure log message for each signature. The eight actual failures are now checked by Go/Node.js/Python/Rust RPC regressions; they produce no committed DEX events.

## Previously skipped or ignored tests

- The eight Node.js tests in `current_mainnet_transactions.test.ts` were opt-in network tests. Enabled them with `RUN_MAINNET_TESTS=1`: all eight passed against the public mainnet RPC, including PumpFun/PumpSwap, Meteora, Orca and Raydium fixtures.
- The Rust parser's ignored route timing test passed when run with `cargo test --lib -- --ignored`.
- sol-shred-sdk's ignored route timing and decoder microbenchmark tests both passed with the same command. These are manual timing tests, not previously failed tests.

## Final validation

- Go: `go test ./...` and `go vet ./...` passed, including the eight real failed transactions and native gRPC failure status tests.
- Node.js: build passed; full suite with `RUN_MAINNET_TESTS=1` and the shared 36-transaction corpus: **311 passed, zero skipped**.
- Python: full suite with the shared 36-transaction corpus: **277 passed**. Tests include RPC conversion, instruction-only gRPC parsing and the full subscription callback on failed transactions.
- Rust: shared-corpus RPC regression: **2 passed** (covering every successful sample plus all eight actual failed transactions); the ignored timing test passed.
- sol-shred-sdk: both ignored manual tests passed; its existing default/zero-copy correctness suites passed in the earlier synchronized-fix verification.

At verification time, changes were local and uncommitted. No transactions were submitted to Solana. Publication is tracked by the repository release.
