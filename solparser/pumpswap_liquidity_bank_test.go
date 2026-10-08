package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestPumpSwapLiquidityBank(t *testing.T) {
	b, err := os.ReadFile("testdata/pumpswap_liquidity_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name     string
			Raw      json.RawMessage
			Expected []map[string]string
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 15 {
		t.Fatal("scenario count")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			tx := pumpUpgradeRPC(t, c.Raw)
			events, err := ParseRpcTransaction(tx, "simulation", nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			if tx.Meta.Err != nil {
				if len(events) != 0 {
					t.Fatal("rolled-back events retained")
				}
				return
			}
			var lp []map[string]any
			for _, e := range events {
				if e.Type != EventTypePumpSwapLiquidityAdded && e.Type != EventTypePumpSwapLiquidityRemoved {
					continue
				}
				b, err := json.Marshal(e.Data)
				if err != nil {
					t.Fatal(err)
				}
				d := json.NewDecoder(bytes.NewReader(b))
				d.UseNumber()
				var fields map[string]any
				if err = d.Decode(&fields); err != nil {
					t.Fatal(err)
				}
				fields["type"] = string(e.Type)
				lp = append(lp, fields)
			}
			if len(lp) != len(c.Expected) {
				t.Fatal("LP occurrences", len(lp), len(c.Expected))
			}
			for i, w := range c.Expected {
				for k, v := range w {
					if fmt.Sprint(lp[i][k]) != v {
						t.Fatal(k, lp[i][k], v)
					}
				}
			}
		})
	}
}

func TestPumpSwapLpTruncatedEvents(t *testing.T) {
	for _, parser := range []func([]byte, EventMetadata) DexEvent{parsePSAddLiqFromData, parsePSRemoveLiqFromData} {
		for n := 0; n < 248; n++ {
			if parser(make([]byte, n), EventMetadata{}).Data != nil {
				t.Fatal("truncated LP event", n)
			}
		}
	}
}
func TestPumpSwapLpInstructionAccountIdentities(t *testing.T) {
	keys := make([]string, 15)
	for i := range keys {
		keys[i] = fmt.Sprintf("account_%d", i)
	}
	for _, parser := range []func([]string, EventMetadata) DexEvent{parsePumpSwapDepositInstr, parsePumpSwapWithdrawInstr} {
		e := parser(keys, EventMetadata{})
		var got []string
		switch v := e.Data.(type) {
		case *PumpSwapLiquidityAddedEvent:
			got = []string{v.Pool, v.User, v.UserBaseTokenAccount, v.UserQuoteTokenAccount, v.UserPoolTokenAccount}
		case *PumpSwapLiquidityRemovedEvent:
			got = []string{v.Pool, v.User, v.UserBaseTokenAccount, v.UserQuoteTokenAccount, v.UserPoolTokenAccount}
		default:
			t.Fatal("missing LP intent")
		}
		for i, k := range []int{0, 2, 6, 7, 8} {
			if got[i] != keys[k] {
				t.Fatal("IDL account", i, got[i], keys[k])
			}
		}
		if parser(keys[:14], EventMetadata{}).Data != nil {
			t.Fatal("incomplete IDL accounts")
		}
	}
}
