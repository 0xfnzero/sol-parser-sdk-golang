package solparser

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

func TestPumpUpgradeOfficialEvents(t *testing.T) {
	raw, e := os.ReadFile("../tests/fixtures/pump_upgrade/events.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixtures []struct{ Variant, Program, Body, DiscString string }
	if e = json.Unmarshal(raw, &fixtures); e != nil {
		t.Fatal(e)
	}
	for _, f := range fixtures {
		body, _ := hex.DecodeString(f.Body)
		disc, _ := strconv.ParseUint(f.DiscString, 10, 64)
		ev := parsePumpUpgradeEvent(disc, body, EventMetadata{}, f.Program)
		if string(ev.Type) != f.Variant {
			t.Fatal(f.Variant, ev.Type)
		}
		if parsePumpUpgradeEvent(disc, body[:len(body)-1], EventMetadata{}, f.Program).Type != "" {
			t.Fatal("truncation accepted")
		}
		if parsePumpUpgradeEvent(disc, body, EventMetadata{}, "bad").Type != "" {
			t.Fatal("foreign program accepted")
		}
		cpi := []byte{228, 69, 165, 46, 81, 203, 154, 29}
		cpi = binary.LittleEndian.AppendUint64(cpi, disc)
		cpi = append(cpi, body...)
		if ParseInnerInstructionUnified(cpi, nil, "sig", 1, 0, nil, 0, nil, f.Program, false).Type != ev.Type {
			t.Fatal("CPI missing")
		}
	}
	ev := parsePumpUpgradeEvent(619296439455019615, make([]byte, 104), EventMetadata{}, PUMPFUN_PROGRAM_ID)
	if ev.Data.(*PumpFunCompleteEvent).QuoteMint != "So11111111111111111111111111111111111111112" {
		t.Fatal("legacy SOL completion")
	}
}

func TestDistinctMultiHopVenuesNotMerged(t *testing.T) {
	a := DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: "a"}}
	b := DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: "b"}}
	if tryMergeDexEvents(&a, b) {
		t.Fatal("route legs collapsed")
	}
}
