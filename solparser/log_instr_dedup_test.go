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
