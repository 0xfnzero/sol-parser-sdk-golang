package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/mr-tron/base58"
	"os"
	"reflect"
	"testing"
)

func executionTrades(events []DexEvent) []map[string]any {
	var result []map[string]any
	for _, e := range events {
		_, pump := e.Data.(*PumpFunTradeEvent)
		if !pump && e.Type != EventTypePumpSwapBuy && e.Type != EventTypePumpSwapSell {
			continue
		}
		b, _ := json.Marshal(e.Data)
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		var all map[string]any
		_ = dec.Decode(&all)
		row := map[string]any{}
		for _, k := range []string{"mint", "user", "pool", "is_buy", "is_created_buy", "ix_name", "token_amount", "quote_amount", "sol_amount", "base_amount_in", "base_amount_out", "quote_amount_in", "quote_amount_out"} {
			row[k] = all[k]
		}
		result = append(result, row)
	}
	return result
}

func TestPumpExecution(t *testing.T) {
	for _, path := range []string{"testdata/pump_execution_20261019.json", "testdata/pump_routes_20261009.json", "testdata/wrapped_pumpfun_reported.json", "testdata/pump_create_alt_20261026.json", "testdata/pump_multi_hop_20261025.json"} {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var fixture struct {
			Cases []struct {
				Name string
				Raw  json.RawMessage
			}
		}
		if err := json.Unmarshal(b, &fixture); err != nil {
			t.Fatal(err)
		}
		for _, c := range fixture.Cases {
			t.Run(c.Name, func(t *testing.T) {
				tx := pumpUpgradeRPC(t, c.Raw)
				events, err := ParseRpcTransaction(tx, "replay", nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				want := executionTrades(events)
				if len(want) == 0 {
					t.Fatal("no trades")
				}
				for mode := 0; mode < 3; mode++ {
					if mode == 1 {
						tx.Meta.LogMessages = tx.Meta.LogMessages[:len(tx.Meta.LogMessages)/2]
					}
					if mode == 2 {
						tx.Meta.LogMessages = nil
					}
					events, err = ParseRpcTransaction(tx, "replay", nil, 0)
					if err != nil {
						t.Fatal(err)
					}
					if got := executionTrades(events); !reflect.DeepEqual(got, want) {
						t.Fatalf("mode %d: got %v want %v", mode, got, want)
					}
					meta, transaction, err := ConvertRpcToGrpc(tx)
					if err != nil {
						t.Fatal(err)
					}
					grpcEvents, parseErr := ParseSubscribeTransaction(tx.Slot, &SubscribeUpdateTransactionInfo{Transaction: transaction, Meta: meta}, nil, 0)
					if parseErr != nil {
						t.Fatal(parseErr)
					}
					if !reflect.DeepEqual(executionTrades(grpcEvents), want) {
						t.Fatal("gRPC mismatch")
					}
				}
				for i := range tx.Meta.InnerInstructions {
					group := &tx.Meta.InnerInstructions[i]
					kept := group.Instructions[:0]
					for _, ix := range group.Instructions {
						if !bytes.HasPrefix(ix.Data, []byte{228, 69, 165, 46, 81, 203, 154, 29}) {
							kept = append(kept, ix)
						}
					}
					group.Instructions = kept
				}
				events, err = ParseRpcTransaction(tx, "replay", nil, 0)
				if err != nil {
					t.Fatal(err)
				}
				for _, e := range events {
					if p, ok := e.Data.(*PumpFunTradeEvent); ok && (p.SolAmount != 0 || p.QuoteAmount != 0 || p.TokenAmount != 0) {
						t.Fatal("instruction limit leaked into execution")
					}
				}
				tx.Meta.Err = map[string]any{"InstructionError": []any{0, "InvalidArgument"}}
				events, err = ParseRpcTransaction(tx, "replay", nil, 0)
				if err != nil || len(events) != 0 {
					t.Fatal("failed transaction retained")
				}
			})
		}
	}
}

func TestSignedPumpShredInputs(t *testing.T) {
	for _, file := range []string{"pump_create_alt_20261026.json", "pump_multi_hop_20261025.json"} {
		b, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		var fixture struct {
			Cases []struct {
				Name, Signature string
				Raw             json.RawMessage
				WireRPC         struct {
					Slot        uint64
					Transaction []json.RawMessage `json:"transaction"`
					Meta        struct {
						LoadedAddresses struct{ Writable, Readonly []string }
					} `json:"meta"`
				} `json:"wire_rpc"`
			}
		}
		if err = json.Unmarshal(b, &fixture); err != nil {
			t.Fatal(err)
		}
		for i, c := range fixture.Cases {
			t.Run(c.Name, func(t *testing.T) {
				var encoded string
				if err = json.Unmarshal(c.WireRPC.Transaction[0], &encoded); err != nil {
					t.Fatal(err)
				}
				wire, err := base64.StdEncoding.DecodeString(encoded)
				if err != nil {
					t.Fatal(err)
				}
				filter := &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypePumpFunCreate, EventTypePumpFunTrade}}
				events := DexEventsFromShredTransactionWire(wire, "", c.WireRPC.Slot, 888, nil, 123456, filter)
				var trades []*PumpFunTradeEvent
				for _, e := range events {
					if tr, ok := e.Data.(*PumpFunTradeEvent); ok {
						trades = append(trades, tr)
					}
				}
				if len(trades) != 1 {
					t.Fatalf("expected one trade, got %+v", events)
				}
				tr := trades[0]
				if !tr.IsCreatedBuy || tr.Metadata.Signature != c.Signature || tr.Metadata.TxIndex != 888 || tr.Metadata.GrpcRecvUs != 123456 {
					t.Fatalf("lost creation or metadata: %+v", tr)
				}
				loaded := c.WireRPC.Meta.LoadedAddresses
				full, err := DexEventsFromShredTransactionWireWithLoadedAddresses(wire, loaded.Writable, loaded.Readonly, "", c.WireRPC.Slot, 888, nil, 123456, filter)
				if err != nil {
					t.Fatal(err)
				}
				if len(full) != len(events) {
					t.Fatal("ALT event count mismatch")
				}
				if file == "pump_create_alt_20261026.json" {
					if len(events) != 2 || events[0].Type != EventTypePumpFunCreateV2 || events[1].Type != EventTypePumpFunTrade {
						t.Fatal("create must precede trade")
					}
					create := events[0].Data.(*PumpFunCreateV2TokenEvent)
					resolved := full[0].Data.(*PumpFunCreateV2TokenEvent)
					if create.IxName != "create_v2" || create.Mint != tr.Mint || create.Name != resolved.Name || create.Symbol != resolved.Symbol || create.Uri != resolved.Uri || create.Creator != resolved.Creator {
						t.Fatalf("create mismatch: %+v", create)
					}
					if tr.IxName != []string{"buy_exact_sol_in", "buy_v2", "buy_v3"}[i] {
						t.Fatal(tr.IxName)
					}
					if i > 0 && tr.QuoteMint != zeroPubkey {
						t.Fatal("unresolved quote was guessed")
					}
					only := DexEventsFromShredTransactionWire(wire, "", 1, 0, nil, 1, &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypePumpFunTrade}})
					if len(only) != 1 || !only[0].Data.(*PumpFunTradeEvent).IsCreatedBuy {
						t.Fatal("creation flag depends on create subscription")
					}
				} else {
					if tr.IxName != "multi_hop_swap" || !tr.IsBuy || tr.ParseStatus != TradeParseInstructionDecoded || tr.Mint != "Aijxkeq5L9iaDYZxfaXUhnHQj4cYeo8um6a3t5vzpump" || tr.QuoteMint != "6Vk3fAQWYEot6JbQb8Nx8j5wZeDnMhzk8FajjHhZpump" {
						t.Fatalf("invalid curve hop %+v", tr)
					}
					if tr.SolAmount != 0 || tr.TokenAmount != 0 || tr.QuoteAmount != 0 || tr.ThirdProgram == nil {
						t.Fatal("route limits used as hop fills")
					}
					decoded, _, err := DecodeWireTransaction(wire, 0, true)
					if err != nil {
						t.Fatal(err)
					}
					for _, ix := range decoded.Message.Instructions {
						data, _ := base58.Decode(ix.Data)
						if !bytes.HasPrefix(data, pumpMultiHopSwap) {
							continue
						}
						indices := append([]int{}, ix.Accounts...)
						indices[23] = ix.Accounts[0]
						if len(parseShredPumpCurveHops(data, indices, decoded.Message.AccountKeys, 0, EventMetadata{})) != 0 {
							t.Fatal("all-AMM route emitted a curve trade")
						}
						if len(parseShredPumpCurveHops(data[:23], ix.Accounts, decoded.Message.AccountKeys, 0, EventMetadata{})) != 0 {
							t.Fatal("truncated route accepted")
						}
						indices = append([]int{}, ix.Accounts...)
						indices[24] = 255
						hops := parseShredPumpCurveHops(data, indices, decoded.Message.AccountKeys, 0, EventMetadata{})
						if len(hops) != 1 || hops[0].Data.(*PumpFunTradeEvent).ThirdProgram.Accounts[24] != nil {
							t.Fatal("unresolved vault was guessed")
						}
					}
					if len(DexEventsFromShredTransactionWire(wire, "", 1, 0, nil, 1, &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypePumpFunBuy}})) != 0 {
						t.Fatal("direction filter should not infer execution")
					}
					rpc := pumpUpgradeRPC(t, c.Raw)
					executed, parseErr := ParseRpcTransaction(rpc, c.Signature, filter, 0)
					if parseErr != nil {
						t.Fatal(parseErr)
					}
					var fill *PumpFunTradeEvent
					for _, e := range executed {
						if x, ok := e.Data.(*PumpFunTradeEvent); ok {
							fill = x
						}
					}
					if fill == nil || fill.IxName != "multi_hop_swap" || fill.SolAmount != 0 || fill.TokenAmount != 255705136795915 || fill.QuoteAmount != 11728308720756 {
						t.Fatalf("execution fill %+v", fill)
					}
				}
			})
		}
	}
}
