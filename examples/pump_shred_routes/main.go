// Native signed-wire replay. Run: go run ./examples/pump_shred_routes
package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	sol "github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

type expected struct {
	Kind    string     `json:"kind"`
	Index   int        `json:"instruction_index"`
	User    string     `json:"user"`
	Input   string     `json:"input"`
	Output  string     `json:"output"`
	Amount  *uint64    `json:"amount_in"`
	Minimum *uint64    `json:"min_amount_out"`
	Hops    [][]string `json:"hops"`
}
type fixture struct {
	Signature string     `json:"signature"`
	Wire      string     `json:"wire"`
	Slot      uint64     `json:"slot"`
	Writable  []string   `json:"loaded_writable"`
	Readonly  []string   `json:"loaded_readonly"`
	Expected  []expected `json:"expected"`
}

func check(ok bool, message string) {
	if !ok {
		panic(message)
	}
}
func runReplay() int {
	_, source, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../fixtures/pump_shred_routes.json"))
	if err != nil {
		panic(err)
	}
	var fixtures []fixture
	if err = json.Unmarshal(data, &fixtures); err != nil {
		panic(err)
	}
	var baseline []sdk.DexEvent
	for pass := 0; pass < 2; pass++ {
		var trace []sdk.DexEvent
		for _, f := range fixtures {
			raw, err := base64.StdEncoding.DecodeString(f.Wire)
			if err != nil {
				panic(err)
			}
			tx, _, err := sdk.DecodeWireTransaction(raw, 0, true)
			if err != nil {
				panic(err)
			}
			check(tx.Signatures[0] == f.Signature, "signature mismatch")
			signed := raw[1+64*len(tx.Signatures):]
			if tx.Version == 1 {
				signed = raw[:len(raw)-64*len(tx.Signatures)]
			}
			for i, s := range tx.Signatures {
				key, _ := base58.Decode(tx.Message.AccountKeys[i])
				sig, _ := base58.Decode(s)
				check(ed25519.Verify(key, signed, sig), "invalid signature")
			}
			parse := func(filter sdk.EventTypeFilter) []sdk.DexEvent {
				e, err := sdk.DexEventsFromShredTransactionWireWithLoadedAddresses(raw, f.Writable, f.Readonly, "", f.Slot, 888, nil, 999, filter)
				if err != nil {
					panic(err)
				}
				return e
			}
			if f.Expected[0].Kind == "RapidBuy" || f.Expected[0].Kind == "BuyPumpDelta" {
				events := parse(nil)
				check(len(events) == len(f.Expected), "raw count mismatch")
				keys := append(append(append([]string{}, tx.Message.AccountKeys...), f.Writable...), f.Readonly...)
				for _, event := range events {
					e, ok := event.Data.(*sdk.RawInstructionEvent)
					check(ok, "expected raw instruction")
					ix := tx.Message.Instructions[e.InstructionIndex]
					check(e.ProgramID == keys[ix.ProgramIDIndex], "raw program")
					check(reflect.DeepEqual(e.AccountIndices, ix.Accounts), "raw indices")
					for i, k := range ix.Accounts {
						check(e.Accounts[i] != nil && *e.Accounts[i] == keys[k], "raw account")
					}
					data, _ := base58.Decode(ix.Data)
					check(len(e.InstructionData) == len(data), "raw data length")
					for i, b := range data {
						check(e.InstructionData[i] == int(b), "raw byte")
					}
				}
				selected := parse(&sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpFunTrade}})
				check(len(selected) > 0, "PumpFun wrapper was dropped")
				for _, event := range selected {
					raw, ok := event.Data.(*sdk.RawInstructionEvent)
					check(ok, "wrapper must remain raw")
					pump := false
					for _, dex := range raw.Dexes {
						if dex.DexType == sdk.RawInstructionDexTypePumpFun {
							pump = true
						}
					}
					check(pump, "unrelated raw instruction leaked into PumpFun subscription")
				}
				trace = append(trace, events...)
				continue
			}
			events := parse(&sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpFunTrade, sdk.EventTypePumpFunCreate}})
			check(len(events) == len(f.Expected), "event count mismatch")
			count := len(events)
			if len(f.Writable)+len(f.Readonly) > 0 {
				count = 0
			}
			check(len(sdk.DexEventsFromShredTransactionWire(raw, "", 0, 0, nil, 0, nil)) == count, "strict ALT parsing failed")
			keys := append(append(append([]string{}, tx.Message.AccountKeys...), f.Writable...), f.Readonly...)
			for i, event := range events {
				e, ok := event.Data.(*sdk.PumpRouteIntentEvent)
				check(ok, "expected route intent")
				if e.Kind == "RapidBuy" || e.Kind == "BuyPumpDelta" {
					check(e.IsBuy != nil && *e.IsBuy, "missing buy intent")
				} else {
					check(e.IsBuy == nil, "multi-hop direction must be unknown")
				}
				if e.Kind == "RapidBuy" {
					check(e.IxName != nil && *e.IxName == "buy_exact_quote_in_v2", "IDL method mismatch")
				} else {
					check(e.IxName == nil, "unknown method must remain nil")
				}
				w := f.Expected[i]
				payload, err := json.Marshal(event)
				if err != nil {
					panic(err)
				}
				var tagged map[string]struct {
					Data []int `json:"instruction_data"`
				}
				if err = json.Unmarshal(payload, &tagged); err != nil {
					panic(err)
				}
				check(len(tagged["PumpRouteIntent"].Data) == len(e.InstructionData), "instruction byte JSON schema")
				var roundtrip map[string]sdk.PumpRouteIntentEvent
				check(json.Unmarshal(payload, &roundtrip) == nil, "route JSON roundtrip")
				restored := roundtrip["PumpRouteIntent"]
				check(reflect.DeepEqual(restored.IsBuy, e.IsBuy) && reflect.DeepEqual(restored.IxName, e.IxName), "route JSON hints")
				var legacy sdk.PumpRouteIntentEvent
				check(json.Unmarshal([]byte(`{"kind":"RapidBuy"}`), &legacy) == nil, "legacy JSON")
				check(legacy.IsBuy == nil && legacy.IxName == nil, "legacy hints stay unknown")
				check(e.Kind == w.Kind && e.InstructionIndex == w.Index && e.User == w.User && e.InputAccount == w.Input && e.OutputAccount == w.Output, "route fields mismatch")
				check(reflect.DeepEqual(e.AmountIn, w.Amount) && reflect.DeepEqual(e.MinAmountOut, w.Minimum), "amount mismatch")
				var hops [][]string
				for _, h := range e.Hops {
					hops = append(hops, []string{h.BaseMint, h.QuoteMint, h.Venue, h.BaseVault, h.QuoteVault})
				}
				check(reflect.DeepEqual(hops, w.Hops), "hop mismatch")
				check(e.Metadata.Signature == f.Signature && e.Metadata.TxIndex == 888 && e.Metadata.GrpcRecvUs == 999 && e.Metadata.RecentBlockhash == tx.Message.RecentBlockhash, "metadata mismatch")
				ix := tx.Message.Instructions[e.InstructionIndex]
				data, _ := base58.Decode(ix.Data)
				a := make([]string, len(ix.Accounts))
				for j, k := range ix.Accounts {
					a[j] = keys[k]
				}
				for _, bad := range [][]byte{data[:3], append(append([]byte{}, data...), 1)} {
					check(sdk.ParsePumpRoute(e.Program, bad, a, e.Metadata, 0, nil) == nil, "malformed layout accepted")
				}
				check(sdk.ParsePumpRoute("11111111111111111111111111111111", data, a, e.Metadata, 0, nil) == nil, "unknown program accepted")
				check(sdk.ParsePumpRoute(e.Program, data, a[:len(a)-1], e.Metadata, 0, nil) == nil, "truncated accounts accepted")
			}
			check(len(parse(&sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpRouteIntent}})) == len(events), "explicit filter")
			for _, filter := range []sdk.EventTypeFilter{&sdk.ExcludeFilter{ExcludeTypes: []sdk.EventType{sdk.EventTypePumpRouteIntent}}, &sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpFunBuy}}, &sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpFunCreate}}} {
				check(len(parse(filter)) == 0, "route filter")
			}
			_, err = sdk.DexEventsFromShredTransactionWireWithLoadedAddresses(raw, append(append([]string{}, f.Writable...), "11111111111111111111111111111111"), f.Readonly, "", 0, 0, nil, 0, nil)
			check(err != nil, "invalid ALT count accepted")
			calls := 0
			resolved, err := sdk.DexEventsFromShredTransactionWireWithResolver(raw, "", f.Slot, 888, nil, 999, nil, func(decoded *sdk.NativeWireTransaction, slot uint64) ([]string, []string, error) {
				calls++
				check(slot == f.Slot && reflect.DeepEqual(decoded, tx), "cache callback context")
				return f.Writable, f.Readonly, nil
			})
			check(err == nil && reflect.DeepEqual(resolved, events), "cache resolved events")
			if len(f.Writable)+len(f.Readonly) > 0 {
				check(calls == 1, "cache not consulted")
			}
			trace = append(trace, events...)
		}
		if pass > 0 {
			check(reflect.DeepEqual(trace, baseline), "non-repeatable replay")
		}
		baseline = trace
	}
	checkBoundaries(fixtures, source)
	return len(fixtures)
}
func main() {
	fmt.Printf("%d signed ShredStream transactions verified twice; offline replay, zero RPC.\n", runReplay())
}

func checkBoundaries(fixtures []fixture, source string) {
	data, e := os.ReadFile(filepath.Join(filepath.Dir(source), "../fixtures/pump_curve_instruction.json"))
	if e != nil {
		panic(e)
	}
	var f struct {
		Program  string
		Accounts []string
		Data     []byte
	}
	if e = json.Unmarshal(data, &f); e != nil {
		panic(e)
	}
	wallet := sol.PrivateKey(ed25519.NewKeyFromSeed(bytes.Repeat([]byte{117}, 32)))
	a := append([]string{}, f.Accounts...)
	a[3] = wallet.PublicKey().String()
	for _, amount := range []uint64{1, ^uint64(0)} {
		wireData := append([]byte{}, f.Data...)
		binary.LittleEndian.PutUint64(wireData[8:], amount)
		metas := make(sol.AccountMetaSlice, len(a))
		for i, k := range a {
			metas[i] = &sol.AccountMeta{PublicKey: sol.MustPublicKeyFromBase58(k), IsSigner: i == 3, IsWritable: true}
		}
		tx, e := sol.NewTransaction([]sol.Instruction{sol.NewInstruction(sol.MustPublicKeyFromBase58(f.Program), metas, wireData)}, sol.Hash{}, sol.TransactionPayer(wallet.PublicKey()))
		if e != nil {
			panic(e)
		}
		_, e = tx.Sign(func(k sol.PublicKey) *sol.PrivateKey { return &wallet })
		if e != nil {
			panic(e)
		}
		if e = tx.VerifySignatures(); e != nil {
			panic(e)
		}
		raw, e := tx.MarshalBinary()
		if e != nil {
			panic(e)
		}
		events := sdk.DexEventsFromShredTransactionWire(raw, "", 0, 0, nil, 0, &sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpFunTrade}})
		check(len(events) == 1, "signed curve intent count")
		intent, ok := events[0].Data.(*sdk.PumpRouteIntentEvent)
		check(ok && intent.Kind == "MultiHopCurveSwap" && *intent.AmountIn == amount, "signed curve amount")
	}
	zero := append([]byte{}, f.Data...)
	clear(zero[8:16])
	check(sdk.ParsePumpRoute(f.Program, zero, a, sdk.EventMetadata{}, 0, nil) == nil, "zero curve input")
	for _, f := range fixtures {
		if f.Expected[0].Kind != "MultiHopSwap" {
			continue
		}
		raw, _ := base64.StdEncoding.DecodeString(f.Wire)
		tx, _, e := sdk.DecodeWireTransaction(raw, 0, true)
		if e != nil {
			panic(e)
		}
		keys := append(append(tx.Message.AccountKeys, f.Writable...), f.Readonly...)
		ix := tx.Message.Instructions[f.Expected[0].Index]
		a := make([]string, len(ix.Accounts))
		for i, k := range ix.Accounts {
			a[i] = keys[k]
		}
		data, _ := base58.Decode(ix.Data)
		program := keys[ix.ProgramIDIndex]
		zero := append([]byte{}, data...)
		clear(zero[16:24])
		intent := sdk.ParsePumpRoute(program, zero, a, sdk.EventMetadata{}, 0, nil)
		check(intent != nil && *intent.MinAmountOut == 0, "zero min output")
		clear(zero[8:16])
		check(sdk.ParsePumpRoute(program, zero, a, sdk.EventMetadata{}, 0, nil) == nil, "zero input")
		a[7] = "11111111111111111111111111111111"
		check(sdk.ParsePumpRoute(program, data, a, sdk.EventMetadata{}, 0, nil) == nil, "wrong fixed token program")
		break
	}
}
