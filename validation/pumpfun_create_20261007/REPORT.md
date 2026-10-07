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
