package solparser

import (
	"encoding/binary"
	"testing"
)

func TestCurrentCpmmFieldsAndLegacy(t *testing.T) {
	b := make([]byte, 162)
	for i := 81; i < 113; i++ {
		b[i] = 1
	}
	for i := 113; i < 145; i++ {
		b[i] = 2
	}
	binary.LittleEndian.PutUint64(b[145:], 9007199254740993)
	binary.LittleEndian.PutUint64(b[153:], 77)
	b[161] = 1
	e := parseCpmmSwapEventFromData(b, EventMetadata{}).Data.(*RaydiumCpmmSwapEvent)
	if e.TradeFee != 9007199254740993 || e.CreatorFee != 77 || !e.CreatorFeeOnInput || e.InputMint == e.OutputMint {
		t.Fatalf("current fields: %+v", e)
	}
	if parseCpmmSwapEventFromData(b[:81], EventMetadata{}).Data == nil {
		t.Fatal("legacy")
	}
	for n := 0; n < 162; n++ {
		if n != 81 && parseCpmmSwapEventFromData(b[:n], EventMetadata{}).Data != nil {
			t.Fatalf("truncation %d", n)
		}
	}
	b[161] = 2
	if parseCpmmSwapEventFromData(b, EventMetadata{}).Data != nil {
		t.Fatal("invalid bool")
	}
}
func TestCurrentDlmmFeeComponents(t *testing.T) {
	b := make([]byte, 147)
	binary.LittleEndian.PutUint64(b[97:], 9007199254740993)
	binary.LittleEndian.PutUint64(b[113:], 17531)
	binary.LittleEndian.PutUint64(b[121:], 1947)
	b[145] = 1
	e := parseDlmmSwap2Data(b, EventMetadata{}).Data.(*MeteoraDlmmSwapEvent)
	if e.Fee != 19478 || e.MmFee != 17531 || e.AmountLeft != 9007199254740993 || !e.FeesOnInput || e.FeesOnTokenX {
		t.Fatalf("current fields: %+v", e)
	}
	for n := 0; n < 147; n++ {
		if parseDlmmSwap2Data(b[:n], EventMetadata{}).Data != nil {
			t.Fatalf("truncation %d", n)
		}
	}
	binary.LittleEndian.PutUint64(b[113:], ^uint64(0))
	if parseDlmmSwap2Data(b, EventMetadata{}).Data != nil {
		t.Fatal("overflow")
	}
	binary.LittleEndian.PutUint64(b[113:], 3)
	b[146] = 2
	if parseDlmmSwap2Data(b, EventMetadata{}).Data != nil {
		t.Fatal("invalid bool")
	}
}
func TestPumpSwapMergePreservesCreatorFeeUnclaimed(t *testing.T) {
	buy := &PumpSwapBuyEvent{}
	supplementPumpSwapBuy(buy, &PumpSwapBuyEvent{CreatorFeeUnclaimed: 15158069})
	if buy.CreatorFeeUnclaimed != 15158069 {
		t.Fatal("buy loses current creator fee")
	}
	supplementPumpSwapBuy(buy, &PumpSwapBuyEvent{CreatorFeeUnclaimed: 99})
	if buy.CreatorFeeUnclaimed != 15158069 {
		t.Fatal("buy overwrites authoritative fee")
	}
	sell := &PumpSwapSellEvent{}
	supplementPumpSwapSell(sell, &PumpSwapSellEvent{CreatorFeeUnclaimed: 77})
	if sell.CreatorFeeUnclaimed != 77 {
		t.Fatal("sell loses current creator fee")
	}
}
