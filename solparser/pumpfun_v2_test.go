package solparser

import (
	"encoding/binary"
	"testing"
)

func pumpfunV2TestAccounts(n int) []string {
	accounts := make([]string, n)
	for i := range accounts {
		accounts[i] = "account_" + string(rune('A'+i))
	}
	return accounts
}

func pumpfunV2Instruction(disc uint64, first, second uint64) []byte {
	data := make([]byte, 24)
	binary.LittleEndian.PutUint64(data[:8], disc)
	binary.LittleEndian.PutUint64(data[8:16], first)
	binary.LittleEndian.PutUint64(data[16:24], second)
	return data
}

func TestParsePumpfunBuyV2UsesRustAccountIndexes(t *testing.T) {
	ev := ParsePumpfunInstruction(
		pumpfunV2Instruction(instrPumpOuterBuyV2, 123, 456),
		pumpfunV2TestAccounts(27),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if ev.Type != EventTypePumpFunTrade {
		t.Fatalf("expected PumpFunTrade, got %q", ev.Type)
	}
	tr, ok := ev.Data.(*PumpFunTradeEvent)
	if !ok {
		t.Fatalf("expected PumpFunTradeEvent, got %T", ev.Data)
	}
	if tr.IxName != "buy_v2" {
		t.Fatalf("ix_name: %q", tr.IxName)
	}
	if tr.Mint != "account_B" || tr.FeeRecipient != "account_G" ||
		tr.BondingCurve != "account_K" || tr.AssociatedBondingCurve != "account_L" ||
		tr.User != "account_N" || tr.TokenProgram != "account_D" || tr.CreatorVault != "account_Q" {
		t.Fatalf("unexpected v2 accounts: %+v", tr)
	}
	if tr.TokenAmount != 123 || tr.SolAmount != 456 {
		t.Fatalf("amounts: token=%d sol=%d", tr.TokenAmount, tr.SolAmount)
	}
}

func TestPumpfunPostMergeEnrichesCreateV2AndTradeFlags(t *testing.T) {
	events := []DexEvent{
		{
			Type: EventTypePumpFunCreateV2,
			Data: &PumpFunCreateV2TokenEvent{
				Metadata:          EventMetadata{Signature: "sig", Slot: 1, TxIndex: 0, GrpcRecvUs: 10},
				Mint:              "mint",
				IsCashbackEnabled: true,
				IsMayhemMode:      true,
			},
		},
		{
			Type: EventTypePumpFunBuy,
			Data: &PumpFunTradeEvent{
				Metadata:     EventMetadata{Signature: "sig", Slot: 1, TxIndex: 0, GrpcRecvUs: 10},
				Mint:         "mint",
				FeeRecipient: "fee",
				IsBuy:        true,
				IxName:       "buy_v2",
			},
		},
	}

	enrichPumpfunSameTxPostMerge(events)

	create := events[0].Data.(*PumpFunCreateV2TokenEvent)
	trade := events[1].Data.(*PumpFunTradeEvent)
	if create.ObservedFeeRecipient != "fee" {
		t.Fatalf("observed fee: %q", create.ObservedFeeRecipient)
	}
	if !trade.MayhemMode || !trade.IsCashbackCoin || !trade.TrackVolume {
		t.Fatalf("trade flags not enriched: %+v", trade)
	}
}
