package solparser

import (
	"bytes"
	"encoding/json"
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
		for _, k := range []string{"mint", "user", "pool", "is_buy", "token_amount", "quote_amount", "sol_amount", "base_amount_in", "base_amount_out", "quote_amount_in", "quote_amount_out"} {
			row[k] = all[k]
		}
		result = append(result, row)
	}
	return result
}

func TestPumpExecution(t *testing.T) {
	for _, path := range []string{"testdata/pump_execution_20261019.json", "testdata/pump_routes_20261009.json"} {
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
