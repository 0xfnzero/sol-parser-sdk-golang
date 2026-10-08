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
