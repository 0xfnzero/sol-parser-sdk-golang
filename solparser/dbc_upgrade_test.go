package solparser

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func TestCurrentDbcEventCPI(t *testing.T) {
	raw, err := os.ReadFile("testdata/dbc_swap2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name string
			Mode uint8
			Data string
		}
	}
	if err = json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, c := range fixture.Cases {
		data, err := base64.StdEncoding.DecodeString(c.Data)
		if err != nil {
			t.Fatal(err)
		}
		cpi := append([]byte{228, 69, 165, 46, 81, 203, 154, 29}, data...)
		parse := func(data []byte) DexEvent {
			return ParseInnerInstructionUnified(data, nil, "sig", 1, 0, nil, 0, EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDbcSwap}), METEORA_DBC_PROGRAM_ID, false)
		}
		event := parse(cpi)
		e, ok := event.Data.(*MeteoraDbcSwapEvent)
		if !ok {
			t.Fatal("missing", c.Name, c.Mode)
		}
		if e.EventVersion != 2 || e.SwapMode != c.Mode || e.HasTransferHook != (c.Name == "EvtSwap2WithTransferHook") {
			t.Fatal("variant/mode")
		}
		gross := uint64(90000)
		if c.Mode == 0 {
			gross = 100000
		}
		if e.AmountIn != gross || e.ActualInputAmount != gross-1000 || e.OutputAmount != 80000 || e.QuoteReserveAmount != 9007199254740993 || e.MigrationThreshold != 9007199254740995 {
			t.Fatal("executed values")
		}
		if c.Mode == 2 {
			if e.MinimumAmountOut != 0 || e.MaximumAmountIn != 100000 {
				t.Fatal("exact-out limits")
			}
		} else {
			if e.MinimumAmountOut != 79000 || e.MaximumAmountIn != 0 {
				t.Fatal("exact-in limits")
			}
		}
		for size := 0; size < len(cpi); size++ {
			if parse(cpi[:size]).Type != "" {
				t.Fatal("accepted truncated data", size)
			}
		}
		invalid := append([]byte{}, cpi...)
		invalid[16+82] = 3
		if parse(invalid).Type != "" {
			t.Fatal("accepted invalid mode")
		}
		oldBody := *e
		oldBody.EventVersion = 0
		old := DexEvent{Type: EventTypeMeteoraDbcSwap, Data: &oldBody}
		for _, test := range []struct {
			events []DexEvent
			want   int
		}{{[]DexEvent{old, event}, 1}, {[]DexEvent{old, old, event, event}, 2}, {[]DexEvent{old, old, event}, 3}} {
			if len(DedupeLogInstructionEvents(nil, test.events)) != test.want {
				t.Fatal("wrong compatibility occurrence count")
			}
		}
	}
}

func TestRealBankDbcEvents(t *testing.T) {
	raw, err := os.ReadFile("testdata/dbc_live_simulations_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name               string
			Slot               uint64
			Program            string
			Error              any
			DedupCount         int      `json:"dedup_count"`
			BankOutputBalances []uint64 `json:"bank_output_balances"`
			Events             []struct {
				Data     string
				Expected struct {
					OutputAmount      uint64 `json:"output_amount"`
					ActualInputAmount uint64 `json:"actual_input_amount"`
					SwapMode          uint8  `json:"swap_mode"`
					EventVersion      uint8  `json:"event_version"`
					TradeDirection    uint8  `json:"trade_direction"`
					Amount0           uint64 `json:"amount_0"`
					Amount1           uint64 `json:"amount_1"`
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var events []DexEvent
			for _, row := range c.Events {
				data, _ := base64.StdEncoding.DecodeString(row.Data)
				ev := ParseInnerInstructionUnified(data, nil, "simulation", c.Slot, 0, nil, 0, EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDbcSwap}), c.Program, false)
				e, ok := ev.Data.(*MeteoraDbcSwapEvent)
				if !ok {
					t.Fatal("missing DBC event")
				}
				w := row.Expected
				if e.OutputAmount != w.OutputAmount || e.ActualInputAmount != w.ActualInputAmount || e.SwapMode != w.SwapMode || e.EventVersion != w.EventVersion || e.TradeDirection != w.TradeDirection || e.Amount0 != w.Amount0 || e.Amount1 != w.Amount1 {
					t.Fatal("emitted fields mismatch")
				}
				events = append(events, ev)
			}
			current := DedupeLogInstructionEvents(nil, events)
			if len(current) != c.DedupCount {
				t.Fatal("dedup count")
			}
			if c.Error == nil && len(current) == 1 {
				if len(current) == 0 || current[len(current)-1].Data.(*MeteoraDbcSwapEvent).OutputAmount != c.BankOutputBalances[0] {
					t.Fatal("bank output mismatch")
				}
			}
		})
	}
}

func TestDbcReferralBankCredit(t *testing.T) {
	raw, err := os.ReadFile("testdata/dbc_referral_live_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name, Program string
			Error         any
			Credit        uint64 `json:"referral_bank_credit"`
			Count         int    `json:"dedup_count"`
			Events        []struct {
				Data string
				Fee  uint64 `json:"expected_referral_fee"`
				Has  bool   `json:"expected_has_referral"`
			}
		}
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var events []DexEvent
			for _, row := range c.Events {
				data, _ := base64.StdEncoding.DecodeString(row.Data)
				e := ParseInnerInstructionUnified(data, nil, "simulation", 1, 0, nil, 0, nil, c.Program, false)
				body, ok := e.Data.(*MeteoraDbcSwapEvent)
				if !ok || body.ReferralFee != row.Fee || body.HasReferral != row.Has {
					t.Fatal("referral fields differ")
				}
				events = append(events, e)
			}
			current := DedupeLogInstructionEvents(nil, events)
			if len(current) != c.Count {
				t.Fatal("dedup count")
			}
			var credit uint64
			for _, e := range current {
				credit += e.Data.(*MeteoraDbcSwapEvent).ReferralFee
			}
			if c.Error == nil && credit != c.Credit {
				t.Fatal("bank credit differs")
			}
			if c.Error != nil && len(current) != 0 {
				t.Fatal("rolled back fills")
			}
		})
	}
}
