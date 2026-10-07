import collections, hashlib, json
from pathlib import Path
root=Path(__file__).resolve().parent
rows=json.loads((root/'results.json').read_text())
counts=collections.Counter()
summary=[]
for r in rows:
 i=r['instructions'][0]
 kind,direct=next(iter(i['event'].items()))
 ok=r['transaction_error'] is None
 counts[(kind,'success' if ok else 'failed')]+=1
 rpc=next(iter(r['rpc'][0].values())) if r['rpc'] else {}
 grpc=next(iter(r['grpc_reconstructed'][0].values())) if r['grpc_reconstructed'] else {}
 def clean(d):return {k:v for k,v in d.items() if k!='metadata'}
 assert clean(rpc)==clean(grpc),(r['signature'],'RPC/gRPC create-field mismatch')
 expected_program=i['accounts'][7 if kind=='PumpFunCreateV2' else 9]
 assert direct['token_program']==expected_program
 assert rpc['token_program']==expected_program
 if ok:
  accounts=i['accounts'];v2=kind=='PumpFunCreateV2'
  for field,index in {'mint':0,'bonding_curve':2,'user':5 if v2 else 7,'system_program':6 if v2 else 8,'associated_token_program':8 if v2 else 10,'mint_authority':1}.items():
   assert rpc.get(field)==accounts[index],(r['signature'],field,rpc.get(field),accounts[index])
  assert len(r['rpc'])==1 and len(r['grpc_reconstructed'])==1
  assert len(r['logs'])==1,(r['signature'],'missing create log')
 unspecified={'',None,'11111111111111111111111111111111'}
 bogus=rpc.get('quote_mint') not in unspecified and direct.get('quote_mint') in unspecified
 raw=json.loads((root/'transactions'/(r['signature']+'.json')).read_text())
 keys=raw['transaction']['message']['accountKeys']+raw['meta'].get('loadedAddresses',{}).get('writable',[])+raw['meta'].get('loadedAddresses',{}).get('readonly',[])
 inv=list(raw['transaction']['message']['instructions'])
 for g in raw['meta'].get('innerInstructions',[]):inv+=g['instructions']
 inv=[x for x in inv if keys[x['programIdIndex']]=='6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P']
 longest=max(inv,key=lambda x:len(x['accounts']))
 longest_quote=keys[longest['accounts'][16]] if len(longest['accounts'])>16 else None
 if bogus:assert longest_quote==rpc['quote_mint']
 summary.append({'signature':r['signature'],'slot':r['slot'],'kind':kind,'success':ok,'create_accounts':i['account_count'],'loaded_accounts':r['loaded_account_count'],'logs_create_count':len(r['logs']),'instruction_quote_mint':direct.get('quote_mint'),'rpc_quote_mint':rpc.get('quote_mint'),'rpc_quote_token_program':rpc.get('quote_token_program'),'quote_mint_corrupted':bogus,'longest_pump_instruction_accounts':len(longest['accounts']),'longest_pump_instruction_data':longest['data'],'rpc_grpc_create_fields_equal':True,'raw_sha256':hashlib.sha256((root/'transactions'/(r['signature']+'.json')).read_bytes()).hexdigest()})
(root/'summary.json').write_text(json.dumps({'counts':{f'{k[0]} {k[1]}':v for k,v in counts.items()},'transactions':summary},indent=2)+'\n')
bad=[r for r in summary if r['success'] and r['quote_mint_corrupted']]
old=[r for r in summary if r['success'] and r['kind']=='PumpFunCreate' and not r['logs_create_count']]
lines=['# PumpFun create/create_v2 mainnet verification — 2026-10-07','',
'Baseline parser commit: `b057e687d384e7a4cd912c086fc509f5d4f4fbfb` (v0.5.9). Current outputs include the local fix; pre-fix outputs are preserved as `results_before_fix.json` and `summary_before_fix.json`.',
'RPC source: `https://api.mainnet-beta.solana.com`; finalized `getTransaction`, `encoding=json`, `maxSupportedTransactionVersion=0`. Only read-only requests were made.',
'',f'Fetched and replayed {len(rows)} real transactions: 3 successful legacy create, 25 successful create_v2, 6 failed legacy create and 2 failed create_v2. Failed transactions are retained as evidence but excluded from successful creation statistics.',
'',f'Before the fix: 9 of 25 successful create_v2 transactions acquired an unrelated QuoteMint in full RPC and reconstructed Yellowstone parsing. After the fix: {len(bad)}. All sampled direct create_v2 instructions have 16 accounts and no quote accounts. Before the fix, the erroneous QuoteMint was exactly account index 16 of the longest PumpFun invocation, rather than the create_v2 invocation.',
'',
'Root cause (before fix): PumpFunCreateV2 used the generic account getter, which chose the invocation with the largest account list without checking its discriminator. A later buy could win and its accounts were interpreted using create_v2 indices. The fix selects create/create_v2 by discriminator and event mint, supports inner invocations, declines ambiguous or unrelated matches, and removes unverified quote inference from remaining accounts. Authoritative decoded quote fields are preserved. The actual all-zero System Program address is retained as a valid account.',
'',
'The direct instruction path also no longer interprets arbitrary remaining accounts as quote fields. None of the 36 real creation instructions has more than its 14/16 fixed accounts; a synthetic regression covers that additional trigger.',
'',
'TokenProgram: all successful samples agree with the actual creating instruction token-program account (legacy SPL Token; create_v2 Token-2022). This dataset does not reproduce issue #4.',
'',f'Historical logs: before the fix, 2 successful legacy create transactions (2024) had an unrecognized CreateEvent containing only the three strings and mint/bonding_curve/user keys. After the fix, every successful sample emits exactly one create log; omissions: {len(old)}. Fields unavailable in the historical layout remain unspecified rather than being invented.',
'',
'Limitations: no confirmed USDC-quoted launch was found in this collection, so this does not verify the reported USDC-specific transaction or issue #5. Default/zero QuoteMint values in the log path are treated as unspecified, not evidence of a USDC mint. The gRPC test reconstructs Yellowstone protobuf transaction/account/log data from the raw RPC response; it is not an independently captured live Yellowstone stream. Metadata block time differs by API; comparisons exclude metadata. Token balances and failed-status protobuf details are not reconstructed, so those are outside this create-account comparison.',
'',
'Replay (from repository root):','```sh','go run validation/pumpfun_create_20261007/replay.go validation/pumpfun_create_20261007/transactions > validation/pumpfun_create_20261007/results.json','python3 validation/pumpfun_create_20261007/summarize.py','```','',
'Raw transaction responses, account mappings, parser outputs and SHA-256 hashes are saved alongside this report. Official public IDL snapshot is saved as `pump_public_idl.json` with its source URL in `sources.json`.',
'', '## Successful transaction comparisons','',
'| Kind | Signature | Slot | Create accounts | Log create | RPC quote corruption | RPC QuoteMint |','|---|---|---:|---:|---:|---|---|']
for r in summary:
 if not r['success']:continue
 sig=r['signature'];lines.append(f"| {r['kind']} | [{sig}](https://solscan.io/tx/{sig}) | {r['slot']} | {r['create_accounts']} | {r['logs_create_count']} | {'YES' if r['quote_mint_corrupted'] else 'no'} | `{r['rpc_quote_mint']}` |")
(root/'REPORT.md').write_text('\n'.join(lines)+'\n')
assert not bad and not old, 'real-transaction regression remains'
print('transactions',len(rows),'successful quote corruption',len(bad),'successful historical logs missing',len(old))
