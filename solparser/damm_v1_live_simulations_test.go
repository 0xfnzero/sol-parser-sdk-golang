package solparser

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRealDammV1Execution(t *testing.T) {
	raw, err := os.ReadFile("testdata/damm_v1_live_simulations_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name     string
			Logs     []string
			Expected []struct {
				In    uint64 `json:"in_amount"`
				Out   uint64 `json:"out_amount"`
				Trade uint64 `json:"trade_fee"`
				Admin uint64 `json:"admin_fee"`
				Host  uint64 `json:"host_fee"`
			}
			SellSourceDebits []uint64 `json:"sell_source_debits"`
			Error            any
			Balances         []uint64 `json:"bank_output_balances"`
		}
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if len(c.Logs) != len(c.Expected) {
				t.Fatal("log/event count mismatch")
			}
			for i, log := range c.Logs {
				e, ok := ParseLogUnified(log, "simulation", 1, nil).Data.(*MeteoraPoolsSwapEvent)
				if !ok {
					t.Fatal("missing event")
				}
				w := c.Expected[i]
				if e.InAmount != w.In || e.OutAmount != w.Out || e.TradeFee != w.Trade || e.AdminFee != w.Admin || e.HostFee != w.Host || e.AmountIn != 0 || e.MinimumOutAmount != 0 {
					t.Fatal("executed event differs")
				}
			}
			if c.Error == nil {
				if len(c.Expected) == 1 {
					if c.Expected[0].Out != c.Balances[0] {
						t.Fatal("credit mismatch")
					}
				} else if len(c.Expected) == 2 {
					// Vault-share rounding differs from event input: check SPL source debits.
					if len(c.SellSourceDebits) == 0 {
						t.Fatal("missing source debit evidence")
					}
					var debit uint64
					for _, amount := range c.SellSourceDebits {
						debit += amount
					}
					if c.Expected[0].Out-debit != c.Balances[0] || 9900000+c.Expected[1].Out != c.Balances[1] {
						t.Fatal("roundtrip balance mismatch")
					}
				} else {
					t.Fatal("event count")
				}
			}
		})
	}
}
