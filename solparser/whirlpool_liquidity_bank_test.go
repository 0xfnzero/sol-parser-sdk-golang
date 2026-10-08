package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestWhirlpoolLiquidityBank(t *testing.T) {
	b, err := os.ReadFile("testdata/whirlpool_liquidity_20261008.json")
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
	if len(f.Cases) != 30 {
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
				if e.Type != EventTypeOrcaWhirlpoolLiquidityIncreased && e.Type != EventTypeOrcaWhirlpoolLiquidityDecreased {
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

func TestWhirlpoolLiquidityOccurrencesAndMissingLogs(t *testing.T) {
	b, err := os.ReadFile("testdata/whirlpool_liquidity_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name string
			Raw  json.RawMessage
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	var log *OrcaWhirlpoolLiquidityIncreasedEvent
	for _, c := range f.Cases {
		if c.Name == "increase" {
			events, err := ParseRpcTransaction(pumpUpgradeRPC(t, c.Raw), "simulation", nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range events {
				if v, ok := e.Data.(*OrcaWhirlpoolLiquidityIncreasedEvent); ok {
					log = v
				}
			}
		}
	}
	if log == nil {
		t.Fatal("missing bank event")
	}
	newLog := func() DexEvent { v := *log; return DexEvent{Type: EventTypeOrcaWhirlpoolLiquidityIncreased, Data: &v} }
	newIx := func() DexEvent {
		v := *log
		v.TokenAAmount = ^uint64(0)
		v.TokenBAmount = ^uint64(0)
		v.TickLowerIndex = 0
		v.TickUpperIndex = 0
		return DexEvent{Type: EventTypeOrcaWhirlpoolLiquidityIncreased, Data: &v}
	}
	out := DedupeLogInstructionEvents([]DexEvent{newLog(), newLog()}, []DexEvent{newIx(), newIx()})
	if len(out) != 2 {
		t.Fatal("repeated occurrences", len(out))
	}
	for _, e := range out {
		v := e.Data.(*OrcaWhirlpoolLiquidityIncreasedEvent)
		if v.TokenAAmount != log.TokenAAmount || v.TickLowerIndex != log.TickLowerIndex {
			t.Fatal("actual quantities overwritten")
		}
	}
	if out := DedupeLogInstructionEvents([]DexEvent{newLog()}, []DexEvent{newIx(), newIx()}); len(out) != 3 {
		t.Fatal("ambiguous pairing", len(out))
	}
	ix := newIx()
	ix.Data.(*OrcaWhirlpoolLiquidityIncreasedEvent).Position = "11111111111111111111111111111111"
	if out := DedupeLogInstructionEvents([]DexEvent{newLog()}, []DexEvent{ix}); len(out) != 2 {
		t.Fatal("different positions merged")
	}
}
