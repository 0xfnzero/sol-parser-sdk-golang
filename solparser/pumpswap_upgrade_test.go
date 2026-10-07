package solparser

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func appendPumpSwapU64(data []byte, value uint64) []byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], value)
	return append(data, buf[:]...)
}

func appendPumpSwapI128FromI64(data []byte, value int64) []byte {
	var buf [16]byte
	binary.LittleEndian.PutUint64(buf[:8], uint64(value))
	if value < 0 {
		binary.LittleEndian.PutUint64(buf[8:], ^uint64(0))
	}
	return append(data, buf[:]...)
}

func appendCurrentPumpSwapTradeTail(data []byte) []byte {
	data = appendPumpSwapU64(data, 177)
	data = appendPumpSwapU64(data, 188)
	data = appendPumpSwapU64(data, 199)
	data = appendPumpSwapU64(data, 211)
	data = appendPumpSwapI128FromI64(data, -987654321)
	data = append(data, 1)
	data = appendPumpSwapU64(data, 222)
	data = appendPumpSwapU64(data, 233)
	return appendPumpSwapU64(data, 244)
}

func pumpSwapBuyPayload(withTail bool) []byte {
	data := make([]byte, 393)
	data[352] = 1
	binary.LittleEndian.PutUint64(data[385:393], 22)
	var nameLen [4]byte
	binary.LittleEndian.PutUint32(nameLen[:], 3)
	data = append(data, nameLen[:]...)
	data = append(data, "buy"...)
	if withTail {
		data = appendCurrentPumpSwapTradeTail(data)
	}
	return data
}

func TestPumpSwapCurrentTradeTailParity(t *testing.T) {
	buy := parsePSBuyFromData(pumpSwapBuyPayload(true), EventMetadata{})
	if buy.Type != EventTypePumpSwapBuy {
		t.Fatalf("current buy payload did not parse: %q", buy.Type)
	}
	b := buy.Data.(*PumpSwapBuyEvent)
	if b.MinBaseAmountOut != 22 || b.IxName != "buy" || b.CashbackFeeBasisPoints != 177 ||
		b.Cashback != 188 || b.BuybackFeeBasisPoints != 199 || b.BuybackFee != 211 ||
		b.VirtualQuoteReserves != "-987654321" || !b.CanBoost || b.BaseSupply != 222 ||
		b.HolderRewardsBps != 233 || b.HolderRewards != 244 {
		t.Fatalf("unexpected current buy fields: %+v", b)
	}

	sellPayload := appendCurrentPumpSwapTradeTail(make([]byte, 352))
	sell := parsePSSellFromData(sellPayload, EventMetadata{})
	if sell.Type != EventTypePumpSwapSell {
		t.Fatalf("current sell payload did not parse: %q", sell.Type)
	}
	s := sell.Data.(*PumpSwapSellEvent)
	if s.CashbackFeeBasisPoints != 177 || s.Cashback != 188 ||
		s.BuybackFeeBasisPoints != 199 || s.BuybackFee != 211 ||
		s.VirtualQuoteReserves != "-987654321" || !s.CanBoost || s.BaseSupply != 222 ||
		s.HolderRewardsBps != 233 || s.HolderRewards != 244 {
		t.Fatalf("unexpected current sell fields: %+v", s)
	}
}

func TestPumpSwapTradeLayoutValidation(t *testing.T) {
	legacyBuy := make([]byte, 385)
	if ev := parsePSBuyFromData(legacyBuy, EventMetadata{}); ev.Type != EventTypePumpSwapBuy {
		t.Fatalf("legacy buy payload did not parse: %q", ev.Type)
	}
	for tailLen := 0; tailLen <= 80; tailLen++ {
		expected := tailLen == 0 || tailLen == 16 || tailLen == 32 || tailLen == 57 || tailLen == 73 || tailLen >= 81
		ev := parsePSSellFromData(make([]byte, 352+tailLen), EventMetadata{})
		if (ev.Type != "") != expected {
			t.Fatalf("sell tail length %d acceptance mismatch: got %q", tailLen, ev.Type)
		}
	}
	for _, partial := range []int{1, 15, 17, 31, 33, 56} {
		buy := append(pumpSwapBuyPayload(false), make([]byte, partial)...)
		if ev := parsePSBuyFromData(buy, EventMetadata{}); ev.Type != "" {
			t.Fatalf("buy partial tail length %d parsed as %q", partial, ev.Type)
		}
		sell := append(make([]byte, 352), make([]byte, partial)...)
		if ev := parsePSSellFromData(sell, EventMetadata{}); ev.Type != "" {
			t.Fatalf("sell partial tail length %d parsed as %q", partial, ev.Type)
		}
	}

	invalidTrack := pumpSwapBuyPayload(false)
	invalidTrack[352] = 2
	if ev := parsePSBuyFromData(invalidTrack, EventMetadata{}); ev.Type != "" {
		t.Fatalf("invalid track_volume parsed as %q", ev.Type)
	}
	invalidUTF8 := pumpSwapBuyPayload(false)
	invalidUTF8[397] = 0xff
	if ev := parsePSBuyFromData(invalidUTF8, EventMetadata{}); ev.Type != "" {
		t.Fatalf("invalid ix_name UTF-8 parsed as %q", ev.Type)
	}
	invalidBoost := appendCurrentPumpSwapTradeTail(make([]byte, 352))
	invalidBoost[400] = 2
	if ev := parsePSSellFromData(invalidBoost, EventMetadata{}); ev.Type != "" {
		t.Fatalf("invalid can_boost parsed as %q", ev.Type)
	}
}

func TestPumpSwapSignedI128Extremes(t *testing.T) {
	tests := []struct {
		name string
		raw  [16]byte
		want string
	}{
		{name: "minimum", raw: [16]byte{15: 0x80}, want: "-170141183460469231731687303715884105728"},
		{name: "negative one", raw: [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, want: "-1"},
		{name: "maximum", raw: [16]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f}, want: "170141183460469231731687303715884105727"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tail := appendCurrentPumpSwapTradeTail(nil)
			copy(tail[32:48], test.raw[:])
			ev := parsePSSellFromData(append(make([]byte, 352), tail...), EventMetadata{})
			if ev.Type != EventTypePumpSwapSell {
				t.Fatalf("payload did not parse: %q", ev.Type)
			}
			if got := ev.Data.(*PumpSwapSellEvent).VirtualQuoteReserves; got != test.want {
				t.Fatalf("virtual quote reserves = %s, want %s", got, test.want)
			}
		})
	}
}

func TestPumpSwapCreatePoolEventReadsCurrentCreatorFeeFields(t *testing.T) {
	data := make([]byte, 336)
	binary.LittleEndian.PutUint16(data[8:10], 42)
	data[325] = 1
	binary.LittleEndian.PutUint64(data[326:334], 250)
	data[334] = 1
	data[335] = 1

	event := parsePSCreatePoolFromData(data, EventMetadata{})
	if event.Type != EventTypePumpSwapCreatePool {
		t.Fatalf("expected PumpSwapCreatePool, got %q", event.Type)
	}
	create := event.Data.(*PumpSwapCreatePoolEvent)
	if create.Index != 42 || !create.IsMayhemMode || create.CreatorFeeBps != 250 ||
		!create.CanEditCreatorFee || !create.IsHolderReward {
		t.Fatalf("unexpected create pool event fields: %+v", create)
	}

	for bodyLen := 326; bodyLen <= 336; bodyLen++ {
		expected := bodyLen == 326 || bodyLen >= 335
		event := parsePSCreatePoolFromData(make([]byte, bodyLen), EventMetadata{})
		if (event.Type != "") != expected {
			t.Fatalf("create pool body length %d acceptance mismatch: got %q", bodyLen, event.Type)
		}
	}
}

func TestPumpSwapPoolVirtualQuoteReserves(t *testing.T) {
	const bodyOffset = 8
	legacy := make([]byte, bodyOffset+244)
	copy(legacy, []byte{241, 154, 109, 4, 17, 177, 109, 188})
	legacyEvent := ParsePumpswapPool(&AccountData{Data: legacy}, EventMetadata{})
	if legacyEvent.Type != EventTypeAccountPumpSwapPool ||
		legacyEvent.Data.(*PumpSwapPoolAccountEvent).Pool.VirtualQuoteReserves != "0" {
		t.Fatalf("legacy pool did not default virtual reserves: %+v", legacyEvent)
	}

	boost := make([]byte, bodyOffset+253)
	copy(boost, legacy[:bodyOffset+237])
	boost = appendPumpSwapI128FromI64(boost[:bodyOffset+237], -987654321)
	boostEvent := ParsePumpswapPool(&AccountData{Data: boost}, EventMetadata{})
	if boostEvent.Type != EventTypeAccountPumpSwapPool ||
		boostEvent.Data.(*PumpSwapPoolAccountEvent).Pool.VirtualQuoteReserves != "-987654321" {
		t.Fatalf("boost pool virtual reserves mismatch: %+v", boostEvent)
	}

	current := appendPumpSwapU64(boost, 250)
	current = append(current, 1, 1)
	currentEvent := ParsePumpswapPool(&AccountData{Data: current}, EventMetadata{})
	pool := currentEvent.Data.(*PumpSwapPoolAccountEvent).Pool
	if currentEvent.Type != EventTypeAccountPumpSwapPool || pool.CreatorFeeBps != 250 ||
		!pool.CanEditCreatorFee || !pool.IsHolderReward {
		t.Fatalf("current pool creator fee fields mismatch: %+v", currentEvent)
	}

	for _, bounds := range [][2]int{{245, 253}, {254, 262}} {
		for bodyLen := bounds[0]; bodyLen < bounds[1]; bodyLen++ {
			partial := make([]byte, bodyOffset+bodyLen)
			copy(partial, current)
			if ev := ParsePumpswapPool(&AccountData{Data: partial}, EventMetadata{}); ev.Type != "" {
				t.Fatalf("partial upgraded pool body %d parsed as %q", bodyLen, ev.Type)
			}
		}
	}
	creatorFee := current[:bodyOffset+262]
	if ev := ParsePumpswapPool(&AccountData{Data: creatorFee}, EventMetadata{}); ev.Type == "" {
		t.Fatal("creator-fee pool layout should parse")
	}
}

func TestPumpSwapCreatePoolInstructionReadsCurrentArgs(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, instrPumpSwapCreatePool)
	var index [2]byte
	binary.LittleEndian.PutUint16(index[:], 7)
	data = append(data, index[:]...)
	data = appendPumpSwapU64(data, 100)
	data = appendPumpSwapU64(data, 200)
	data = append(data, pumpfunTestPubkey(10)...)
	data = append(data, 1, 1)
	data = appendPumpSwapU64(data, 250)
	data = append(data, 1, 1)

	ev := ParsePumpswapInstruction(data, pumpfunV2TestAccounts(18), "sig", 1, 0, nil, 10)
	if ev.Type != EventTypePumpSwapCreatePool {
		t.Fatalf("expected PumpSwapCreatePool, got %q", ev.Type)
	}
	create := ev.Data.(*PumpSwapCreatePoolEvent)
	if create.Index != 7 || create.BaseAmountIn != 100 || create.QuoteAmountIn != 200 ||
		create.Pool != "account_A" || create.Creator != "account_C" ||
		create.BaseMint != "account_D" || create.QuoteMint != "account_E" ||
		create.CoinCreator != ReadPubkey(pumpfunTestPubkey(10), 0) ||
		create.CreatorFeeBps != 250 || !create.CanEditCreatorFee || !create.IsHolderReward {
		t.Fatalf("unexpected create_pool fields: %+v", create)
	}
}

func TestMergeRpcInstructionEventsPreservesPumpSwapUpgradeFields(t *testing.T) {
	innerIdx := 0
	buyEvents := mergeRpcInstructionEvents([]rpcIndexedEvent{
		{
			OuterIdx: 0,
			Event: DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{
				PoolV2:               "pool-v2",
				VirtualQuoteReserves: "0",
			}},
		},
		{
			OuterIdx: 0,
			InnerIdx: &innerIdx,
			Event: DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{
				BuybackFeeBasisPoints: 199,
				BuybackFee:            211,
				VirtualQuoteReserves:  "-987654321",
				CanBoost:              true,
				BaseSupply:            222,
			}},
		},
	})
	if len(buyEvents) != 1 {
		t.Fatalf("expected one merged buy event, got %d", len(buyEvents))
	}
	buy := buyEvents[0].Data.(*PumpSwapBuyEvent)
	if buy.BuybackFeeBasisPoints != 199 || buy.BuybackFee != 211 ||
		buy.VirtualQuoteReserves != "-987654321" || !buy.CanBoost || buy.BaseSupply != 222 ||
		buy.PoolV2 != "pool-v2" {
		t.Fatalf("merged buy lost upgrade fields: %+v", buy)
	}

	sellEvents := mergeRpcInstructionEvents([]rpcIndexedEvent{
		{
			OuterIdx: 0,
			Event: DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{
				PoolV2:               "pool-v2",
				VirtualQuoteReserves: "0",
			}},
		},
		{
			OuterIdx: 0,
			InnerIdx: &innerIdx,
			Event: DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{
				BuybackFeeBasisPoints: 199,
				BuybackFee:            211,
				VirtualQuoteReserves:  "-987654321",
				CanBoost:              true,
				BaseSupply:            222,
			}},
		},
	})
	if len(sellEvents) != 1 {
		t.Fatalf("expected one merged sell event, got %d", len(sellEvents))
	}
	sell := sellEvents[0].Data.(*PumpSwapSellEvent)
	if sell.BuybackFeeBasisPoints != 199 || sell.BuybackFee != 211 ||
		sell.VirtualQuoteReserves != "-987654321" || !sell.CanBoost || sell.BaseSupply != 222 ||
		sell.PoolV2 != "pool-v2" {
		t.Fatalf("merged sell lost upgrade fields: %+v", sell)
	}
}

func TestPumpSwapInstructionJSONDefaultsVirtualQuoteReservesToZero(t *testing.T) {
	accounts := make([]string, 13)
	events := []DexEvent{
		parsePumpSwapBuyInstr(make([]byte, 8), accounts, EventMetadata{}, false),
		parsePumpSwapSellInstr(make([]byte, 8), accounts, EventMetadata{}),
	}
	for _, event := range events {
		encoded, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("marshal %s event: %v", event.Type, err)
		}
		if !bytes.Contains(encoded, []byte(`"virtual_quote_reserves":"0"`)) {
			t.Fatalf("%s instruction JSON has non-canonical virtual reserves: %s", event.Type, encoded)
		}
	}
}
