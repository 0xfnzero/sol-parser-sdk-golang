package solparser

import "testing"

func TestDedupeLogInstructionEventsKeepsLogTradeValues(t *testing.T) {
	logEvent := DexEvent{
		Type: EventTypePumpFunTrade,
		Data: &PumpFunTradeEvent{
			Mint:                   "Mint111111111111111111111111111111111111",
			User:                   "User111111111111111111111111111111111111",
			IsBuy:                  true,
			SolAmount:              50,
			TokenAmount:            10,
			FeeRecipient:           zeroPubkey,
			BondingCurve:           zeroPubkey,
			IsCreatedBuy:           false,
			TokenProgram:           zeroPubkey,
			CreatorVault:           zeroPubkey,
			Creator:                zeroPubkey,
			AssociatedBondingCurve: zeroPubkey,
		},
	}
	ixEvent := DexEvent{
		Type: EventTypePumpFunBuy,
		Data: &PumpFunTradeEvent{
			Mint:                   "Mint111111111111111111111111111111111111",
			User:                   "User111111111111111111111111111111111111",
			IsBuy:                  true,
			SolAmount:              999,
			TokenAmount:            999,
			FeeRecipient:           "Fee1111111111111111111111111111111111111",
			BondingCurve:           "Curve11111111111111111111111111111111111",
			AssociatedBondingCurve: "Assoc11111111111111111111111111111111111",
			TokenProgram:           "Token11111111111111111111111111111111111",
			CreatorVault:           "Vault11111111111111111111111111111111111",
			IsCreatedBuy:           true,
		},
	}

	out := DedupeLogInstructionEvents([]DexEvent{logEvent}, []DexEvent{ixEvent})
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
	trade, ok := out[0].Data.(*PumpFunTradeEvent)
	if !ok {
		t.Fatalf("expected PumpFunTradeEvent, got %T", out[0].Data)
	}
	if trade.SolAmount != 50 {
		t.Fatalf("log sol amount was overwritten: got %d", trade.SolAmount)
	}
	if trade.FeeRecipient != "Fee1111111111111111111111111111111111111" {
		t.Fatalf("fee recipient was not filled: %s", trade.FeeRecipient)
	}
	if trade.BondingCurve != "Curve11111111111111111111111111111111111" {
		t.Fatalf("bonding curve was not filled: %s", trade.BondingCurve)
	}
	if !trade.IsCreatedBuy {
		t.Fatalf("is_created_buy was not OR-merged")
	}
}

func TestDedupeLogInstructionEventsKeepsV2BuyLanesDistinct(t *testing.T) {
	base := PumpFunTradeEvent{
		Mint:         "Mint222222222222222222222222222222222222",
		User:         "User222222222222222222222222222222222222",
		IsBuy:        true,
		SolAmount:    1,
		TokenAmount:  1,
		BondingCurve: zeroPubkey,
	}
	buyLogTrade := base
	buyLogTrade.IxName = "buy_v2"
	exactLogTrade := base
	exactLogTrade.IxName = "buy_exact_quote_in_v2"
	exactIxTrade := base
	exactIxTrade.IxName = "buy_exact_quote_in_v2"
	exactIxTrade.BondingCurve = "ExactCurve2222222222222222222222222222222"
	buyIxTrade := base
	buyIxTrade.IxName = "buy_v2"
	buyIxTrade.BondingCurve = "BuyCurve22222222222222222222222222222222"

	out := DedupeLogInstructionEvents(
		[]DexEvent{
			{Type: EventTypePumpFunTrade, Data: &buyLogTrade},
			{Type: EventTypePumpFunTrade, Data: &exactLogTrade},
		},
		[]DexEvent{
			{Type: EventTypePumpFunBuy, Data: &exactIxTrade},
			{Type: EventTypePumpFunBuy, Data: &buyIxTrade},
		},
	)
	if len(out) != 2 {
		t.Fatalf("expected 2 events, got %d", len(out))
	}
	first := out[0].Data.(*PumpFunTradeEvent)
	second := out[1].Data.(*PumpFunTradeEvent)
	if first.BondingCurve != "BuyCurve22222222222222222222222222222222" {
		t.Fatalf("buy_v2 lane was mispaired: %s", first.BondingCurve)
	}
	if second.BondingCurve != "ExactCurve2222222222222222222222222222222" {
		t.Fatalf("buy_exact_quote_in_v2 lane was mispaired: %s", second.BondingCurve)
	}
}

func TestDedupeLogInstructionEventsFillsPumpfunUpgradeFields(t *testing.T) {
	logEvent := DexEvent{
		Type: EventTypePumpFunTrade,
		Data: &PumpFunTradeEvent{
			Mint:                  "Mint333333333333333333333333333333333333",
			User:                  "User333333333333333333333333333333333333",
			IsBuy:                 true,
			SolAmount:             50,
			TokenAmount:           10,
			BondingCurveV2:        zeroPubkey,
			BuybackFeeRecipient:   zeroPubkey,
			SpendableQuoteIn:      0,
			MinTokensOut:          0,
			UserVolumeAccumulator: zeroPubkey,
		},
	}
	ixEvent := DexEvent{
		Type: EventTypePumpFunBuy,
		Data: &PumpFunTradeEvent{
			Mint:                  "Mint333333333333333333333333333333333333",
			User:                  "User333333333333333333333333333333333333",
			IsBuy:                 true,
			SolAmount:             999,
			TokenAmount:           999,
			BondingCurveV2:        "CurveV233333333333333333333333333333333",
			BuybackFeeRecipient:   "Buyback333333333333333333333333333333333",
			SpendableQuoteIn:      123,
			MinTokensOut:          456,
			UserVolumeAccumulator: "Volume3333333333333333333333333333333333",
		},
	}

	out := DedupeLogInstructionEvents([]DexEvent{logEvent}, []DexEvent{ixEvent})
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
	trade := out[0].Data.(*PumpFunTradeEvent)
	if trade.SolAmount != 50 {
		t.Fatalf("log sol amount was overwritten: got %d", trade.SolAmount)
	}
	if trade.BondingCurveV2 != "CurveV233333333333333333333333333333333" {
		t.Fatalf("bonding_curve_v2 was not filled: %s", trade.BondingCurveV2)
	}
	if trade.BuybackFeeRecipient != "Buyback333333333333333333333333333333333" {
		t.Fatalf("buyback fee recipient was not filled: %s", trade.BuybackFeeRecipient)
	}
	if trade.SpendableQuoteIn != 123 || trade.MinTokensOut != 456 {
		t.Fatalf("v2 instruction args were not filled: spendable=%d min=%d", trade.SpendableQuoteIn, trade.MinTokensOut)
	}
}
