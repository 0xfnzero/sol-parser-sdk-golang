"""Validate offline replay results without overwriting maintained documentation."""
import argparse
import collections
import hashlib
import json
from pathlib import Path


def summarize(results_path: Path) -> dict:
    corpus = Path(__file__).resolve().parent / "transactions"
    rows = json.loads(results_path.read_text())
    if not rows:
        raise ValueError("Replay results are empty")
    counts = collections.Counter()
    transactions = []
    signatures = set()
    unspecified = {"", None, "11111111111111111111111111111111"}
    for row in rows:
        signature = row["signature"]
        if signature in signatures:
            raise ValueError(f"Duplicate signature: {signature}")
        signatures.add(signature)
        raw_bytes = (corpus / f"{signature}.json").read_bytes()
        raw = json.loads(raw_bytes)
        if row["transaction_error"] != raw["meta"]["err"]:
            raise ValueError(f"Failure status differs from original transaction: {signature}")
        if row["rpc_error"] is not None or row["grpc_error"] is not None:
            raise ValueError(f"Parser error: {signature}")
        instruction = row["instructions"][0]
        kind, direct = next(iter(instruction["event"].items()))
        success = row["transaction_error"] is None
        counts[f"{kind} {'success' if success else 'failed'}"] += 1
        if not success:
            if row["rpc"] or row["grpc_reconstructed"]:
                raise ValueError(f"Failed transaction emitted an executed create event: {signature}")
        else:
            if len(row["rpc"]) != 1 or len(row["grpc_reconstructed"]) != 1 or len(row["logs"]) != 1:
                raise ValueError(f"Expected one creation event in each parsing path: {signature}")
            rpc_kind, rpc = next(iter(row["rpc"][0].items()))
            grpc_kind, grpc = next(iter(row["grpc_reconstructed"][0].items()))
            clean = lambda event: {k: v for k, v in event.items() if k != "metadata"}
            if rpc_kind != grpc_kind or clean(rpc) != clean(grpc):
                raise ValueError(f"RPC/Yellowstone creation fields differ: {signature}")
            v2 = kind == "PumpFunCreateV2"
            accounts = instruction["accounts"]
            expected = {"mint": 0, "bonding_curve": 2, "user": 5 if v2 else 7,
                        "system_program": 6 if v2 else 8, "associated_token_program": 8 if v2 else 10,
                        "mint_authority": 1, "token_program": 7 if v2 else 9}
            for field, index in expected.items():
                if rpc.get(field) != accounts[index] or direct.get(field) != accounts[index]:
                    raise ValueError(f"Creating instruction account mismatch: {signature}: {field}")
            if rpc.get("quote_mint") not in unspecified and direct.get("quote_mint") in unspecified:
                raise ValueError(f"Unverified quote mint inferred from another instruction: {signature}")
        transactions.append({"signature": signature, "kind": kind, "success": success,
                             "slot": row["slot"], "create_accounts": instruction["account_count"],
                             "raw_sha256": hashlib.sha256(raw_bytes).hexdigest()})
    expected_signatures = {p.stem for p in corpus.glob("*.json")}
    if signatures != expected_signatures:
        raise ValueError("Replay results do not cover the complete saved corpus")
    return {"counts": dict(sorted(counts.items())), "transactions": transactions}


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--results", required=True, type=Path, help="JSON output from replay.go")
    parser.add_argument("--output", type=Path, help="Optional summary path; defaults to stdout")
    args = parser.parse_args()
    output = json.dumps(summarize(args.results), indent=2) + "\n"
    if args.output:
        args.output.write_text(output)
        print(f"Validated replay; saved summary to {args.output}")
    else:
        print(output, end="")


if __name__ == "__main__":
    main()
