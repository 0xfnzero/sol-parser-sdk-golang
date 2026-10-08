package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestDbcBoundaryBankEventsAndRollback(t *testing.T) {
	data, err := os.ReadFile("testdata/dbc_boundary_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name               string
			Raw                json.RawMessage
			Expected           map[string]any `json:"expected_swap"`
			RolledBackComplete int            `json:"rolled_back_curve_complete_payloads"`
			Validation         struct {
				Consumed   uint64 `json:"bank_consumed_quote"`
				Output     uint64 `json:"bank_base_credit"`
				Referral   uint64 `json:"bank_referral_credit"`
				Unused     uint64 `json:"bank_unconsumed_quote"`
				InitialFee uint64 `json:"initial_requested_input_fee"`
				ActualFee  uint64 `json:"recalculated_fill_fee"`
			}
		}
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 7 {
		t.Fatal("boundary scenario count")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			tx := pumpUpgradeRPC(t, c.Raw)
			events, parseError := ParseRpcTransaction(tx, "simulation", nil, 0)
			if parseError != nil {
				t.Fatal(parseError)
			}
			if tx.Meta.Err != nil {
				if len(events) != 0 {
					t.Fatal("rolled-back curve completion retained")
				}
				if (c.Name == "buy_after_curve_completed" || c.Name == "sell_after_curve_completed") && c.RolledBackComplete == 0 {
					t.Fatal("missing earlier rolled-back completion payload")
				}
				return
			}
			var swaps []*MeteoraDbcSwapEvent
			var completed []*MeteoraDbcCurveCompleteEvent
			for _, ev := range events {
				switch body := ev.Data.(type) {
				case *MeteoraDbcSwapEvent:
					swaps = append(swaps, body)
				case *MeteoraDbcCurveCompleteEvent:
					completed = append(completed, body)
				}
			}
			if len(swaps) != 1 || len(completed) != 1 {
				t.Fatal("swap/curve completion occurrence", len(swaps), len(completed))
			}
			e := swaps[0]
			encoded, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			decoder := json.NewDecoder(bytes.NewReader(encoded))
			decoder.UseNumber()
			var fields map[string]any
			if err = decoder.Decode(&fields); err != nil {
				t.Fatal(err)
			}
			for key, want := range c.Expected {
				if fmt.Sprint(fields[key]) != fmt.Sprint(want) {
					t.Fatal(key, fields[key], want)
				}
			}
			v := c.Validation
			if e.IncludedFeeInputAmount != v.Consumed || e.OutputAmount != v.Output || e.ReferralFee != v.Referral {
				t.Fatal("bank credits mismatch")
			}
			if e.AmountLeft == v.Unused || e.AmountLeft+v.InitialFee-v.ActualFee != v.Unused {
				t.Fatal("curve net remainder is not gross unused funding")
			}
			if completed[0].QuoteReserve != e.QuoteReserveAmount || e.QuoteReserveAmount < e.MigrationThreshold {
				t.Fatal("completion reserve")
			}
		})
	}
}
