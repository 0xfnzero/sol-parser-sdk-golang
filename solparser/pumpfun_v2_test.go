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

func pumpfunCreateV2Instruction(mayhem, cashback bool) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data[:8], instrPumpOuterCreateV2)
	data = appendPumpfunString(data, "name")
	data = appendPumpfunString(data, "SYM")
	data = appendPumpfunString(data, "uri")
	data = appendPumpfunPubkey(data, 120)
	if mayhem {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	if cashback {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	return data
}

func pumpfunTestPubkey(seed byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = seed + byte(i)
	}
	return out
}

func appendPumpfunU16(out []byte, v uint16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	return append(out, b[:]...)
}

func appendPumpfunU32(out []byte, v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return append(out, b[:]...)
}

func appendPumpfunU64(out []byte, v uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return append(out, b[:]...)
}

func appendPumpfunI64(out []byte, v int64) []byte {
	return appendPumpfunU64(out, uint64(v))
}

func appendPumpfunString(out []byte, s string) []byte {
	out = appendPumpfunU32(out, uint32(len(s)))
	return append(out, []byte(s)...)
}

func appendPumpfunPubkey(out []byte, seed byte) []byte {
	return append(out, pumpfunTestPubkey(seed)...)
}

func pumpfunGlobalAccountData(initialVirtualQuoteReserves uint64) []byte {
	data := []byte{167, 232, 232, 177, 200, 108, 114, 127}
	body := []byte{1}
	body = appendPumpfunPubkey(body, 1)
	body = appendPumpfunPubkey(body, 2)
	body = appendPumpfunU64(body, 3)
	body = appendPumpfunU64(body, 4)
	body = appendPumpfunU64(body, 5)
	body = appendPumpfunU64(body, 6)
	body = appendPumpfunU64(body, 7)
	body = appendPumpfunPubkey(body, 8)
	body = append(body, 1)
	body = appendPumpfunU64(body, 9)
	body = appendPumpfunU64(body, 10)
	for i := 0; i < 7; i++ {
		body = appendPumpfunPubkey(body, byte(20+i))
	}
	body = appendPumpfunPubkey(body, 40)
	body = appendPumpfunPubkey(body, 41)
	body = append(body, 1)
	body = appendPumpfunPubkey(body, 42)
	body = appendPumpfunPubkey(body, 43)
	body = append(body, 0)
	for i := 0; i < 7; i++ {
		body = appendPumpfunPubkey(body, byte(50+i))
	}
	body = append(body, 1)
	for i := 0; i < 8; i++ {
		body = appendPumpfunPubkey(body, byte(70+i))
	}
	body = appendPumpfunU64(body, 11)
	body = appendPumpfunU64(body, initialVirtualQuoteReserves)
	body = appendPumpfunPubkey(body, 90)
	return append(data, body...)
}

func TestParsePumpfunGlobalReadsOfficialTailAndRejectsTruncatedData(t *testing.T) {
	data := pumpfunGlobalAccountData(4_292_000_000)
	account := &AccountData{Pubkey: "global", Data: data}

	ev := ParsePumpfunGlobal(account, EventMetadata{Signature: "sig"})
	if ev.Type != EventTypeAccountPumpFunGlobal {
		t.Fatalf("expected PumpFunGlobal account, got %q", ev.Type)
	}
	global := ev.Data.(*PumpFunGlobalAccountEvent).Global
	if len(global.FeeRecipients) != 7 || len(global.BuybackFeeRecipients) != 8 ||
		global.InitialVirtualQuoteReserves != 4_292_000_000 {
		t.Fatalf("unexpected global tail: %+v", global)
	}

	truncated := &AccountData{Pubkey: "global", Data: data[:len(data)-1]}
	if ev := ParsePumpfunGlobal(truncated, EventMetadata{}); ev.Type != "" {
		t.Fatalf("truncated global account should not parse: %+v", ev)
	}
}

func pumpfunBondingCurveAccountData(creator, quoteMint []byte) []byte {
	data := []byte{23, 183, 248, 55, 96, 216, 172, 96}
	data = appendPumpfunU64(data, 100)
	data = appendPumpfunU64(data, 4_292_000_000)
	data = appendPumpfunU64(data, 200)
	data = appendPumpfunU64(data, 3_000_000_000)
	data = appendPumpfunU64(data, 1_000)
	data = append(data, 1)
	data = append(data, creator...)
	data = append(data, 1)
	data = append(data, 0)
	data = append(data, quoteMint...)
	return data
}

func TestParsePumpfunBondingCurveReadsQuoteFields(t *testing.T) {
	creator := pumpfunTestPubkey(7)
	quoteMint := pumpfunTestPubkey(8)
	account := &AccountData{
		Pubkey: "bonding_curve",
		Owner:  PUMPFUN_PROGRAM_ID,
		Data:   pumpfunBondingCurveAccountData(creator, quoteMint),
	}

	ev := ParseAccountUnified(
		account,
		EventMetadata{Signature: "sig"},
		EventTypeFilterIncludeOnly([]EventType{EventTypeAccountPumpFunBondingCurve}),
	)
	if ev.Type != EventTypeAccountPumpFunBondingCurve {
		t.Fatalf("expected PumpFunBondingCurve account, got %q", ev.Type)
	}
	curve := ev.Data.(*PumpFunBondingCurveAccountEvent).BondingCurve
	if curve.VirtualQuoteReserves != 4_292_000_000 ||
		curve.RealQuoteReserves != 3_000_000_000 ||
		curve.Creator != ReadPubkey(creator, 0) ||
		curve.QuoteMint != ReadPubkey(quoteMint, 0) ||
		!curve.Complete || !curve.IsMayhemMode || curve.IsCashbackCoin {
		t.Fatalf("unexpected bonding curve account: %+v", curve)
	}

	truncated := &AccountData{
		Pubkey: "bonding_curve",
		Owner:  PUMPFUN_PROGRAM_ID,
		Data:   account.Data[:len(account.Data)-1],
	}
	if ev := ParsePumpfunBondingCurve(truncated, EventMetadata{}); ev.Type != "" {
		t.Fatalf("truncated bonding curve should not parse: %+v", ev)
	}
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
	if ev.Type != EventTypePumpFunBuy {
		t.Fatalf("expected PumpFunBuy, got %q", ev.Type)
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

func TestParsePumpfunCreateV2ReadsOfficialArgsAndAccounts(t *testing.T) {
	ev := ParsePumpfunInstruction(
		pumpfunCreateV2Instruction(true, true),
		pumpfunV2TestAccounts(16),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if ev.Type != EventTypePumpFunCreateV2 {
		t.Fatalf("expected PumpFunCreateV2, got %q", ev.Type)
	}
	create, ok := ev.Data.(*PumpFunCreateV2TokenEvent)
	if !ok {
		t.Fatalf("expected PumpFunCreateV2TokenEvent, got %T", ev.Data)
	}
	if create.Mint != "account_A" || create.BondingCurve != "account_C" ||
		create.User != "account_F" || create.Creator != ReadPubkey(pumpfunTestPubkey(120), 0) ||
		!create.IsMayhemMode || !create.IsCashbackEnabled {
		t.Fatalf("unexpected create_v2 fields: %+v", create)
	}
}

func TestParsePumpfunLegacyBuyExactAndSellInstructionParity(t *testing.T) {
	buy := ParsePumpfunInstruction(
		pumpfunV2Instruction(instrPumpOuterBuy, 111, 222),
		pumpfunV2TestAccounts(18),
		"sig",
		1,
		0,
		nil,
		10,
	)
	tr, ok := buy.Data.(*PumpFunTradeEvent)
	if buy.Type != EventTypePumpFunBuy || !ok {
		t.Fatalf("expected PumpFunBuy event, got %q %T", buy.Type, buy.Data)
	}
	if tr.IxName != "buy" || tr.Mint != "account_C" || tr.FeeRecipient != "account_B" ||
		tr.TokenProgram != "account_I" || tr.CreatorVault != "account_J" ||
		tr.BondingCurveV2 != "account_Q" || tr.BuybackFeeRecipient != "account_R" ||
		tr.Amount != 111 || tr.MaxSolCost != 222 {
		t.Fatalf("unexpected legacy buy fields: %+v", tr)
	}

	exact := ParsePumpfunInstruction(
		pumpfunV2Instruction(instrPumpOuterBuyExactSolIn, 333, 444),
		pumpfunV2TestAccounts(16),
		"sig",
		1,
		0,
		nil,
		10,
	)
	tr, ok = exact.Data.(*PumpFunTradeEvent)
	if exact.Type != EventTypePumpFunBuyExactSolIn || !ok {
		t.Fatalf("expected PumpFunBuyExactSolIn event, got %q %T", exact.Type, exact.Data)
	}
	if tr.IxName != "buy_exact_sol_in" || tr.SpendableSolIn != 333 ||
		tr.MinTokensOut != 444 || tr.Amount != 444 {
		t.Fatalf("unexpected legacy exact buy fields: %+v", tr)
	}

	sell := ParsePumpfunInstruction(
		pumpfunV2Instruction(instrPumpOuterSell, 555, 666),
		pumpfunV2TestAccounts(17),
		"sig",
		1,
		0,
		nil,
		10,
	)
	tr, ok = sell.Data.(*PumpFunTradeEvent)
	if sell.Type != EventTypePumpFunSell || !ok {
		t.Fatalf("expected PumpFunSell event, got %q %T", sell.Type, sell.Data)
	}
	if tr.IxName != "sell" || tr.Mint != "account_C" || tr.CreatorVault != "account_I" ||
		tr.TokenProgram != "account_J" || tr.UserVolumeAccumulator != "account_O" ||
		tr.BondingCurveV2 != "account_P" || tr.BuybackFeeRecipient != "account_Q" ||
		tr.Amount != 555 || tr.MinSolOutput != 666 {
		t.Fatalf("unexpected legacy sell fields: %+v", tr)
	}
}

func TestParsePumpfunBuyExactQuoteInV2UsesQuoteAmountFields(t *testing.T) {
	ev := ParsePumpfunInstruction(
		pumpfunV2Instruction(instrPumpOuterBuyExactQuoteInV2, 777, 888),
		pumpfunV2TestAccounts(27),
		"sig",
		1,
		0,
		nil,
		10,
	)
	tr, ok := ev.Data.(*PumpFunTradeEvent)
	if ev.Type != EventTypePumpFunBuyExactSolIn || !ok {
		t.Fatalf("expected PumpFunBuyExactSolIn event, got %q %T", ev.Type, ev.Data)
	}
	if tr.IxName != "buy_exact_quote_in_v2" || tr.Amount != 888 ||
		tr.MaxSolCost != 0 || tr.QuoteAmount != 777 ||
		tr.SpendableQuoteIn != 777 || tr.MinTokensOut != 888 ||
		tr.QuoteMint != "account_C" {
		t.Fatalf("unexpected exact quote fields: %+v", tr)
	}
}

func TestParsePumpfunSellV2ReturnsSellEventType(t *testing.T) {
	ev := ParsePumpfunInstruction(
		pumpfunV2Instruction(instrPumpOuterSellV2, 333, 444),
		pumpfunV2TestAccounts(26),
		"sig",
		1,
		0,
		nil,
		10,
	)
	tr, ok := ev.Data.(*PumpFunTradeEvent)
	if ev.Type != EventTypePumpFunSell || !ok {
		t.Fatalf("expected PumpFunSell event, got %q %T", ev.Type, ev.Data)
	}
	if tr.IxName != "sell_v2" || tr.Amount != 333 || tr.MinSolOutput != 444 || tr.MaxSolCost != 0 {
		t.Fatalf("unexpected sell_v2 fields: %+v", tr)
	}
}

func TestEnrichPumpFunTradeFromAccountsUsesV2IndexesForShortBuyExactQuoteIn(t *testing.T) {
	accounts := pumpfunV2TestAccounts(27)
	accounts[1] = "mint"
	accounts[9] = "legacy_creator_vault"
	tr := &PumpFunTradeEvent{
		IxName:                      "buy_exact_quote_in",
		IsBuy:                       true,
		Mint:                        "mint",
		QuoteMint:                   zeroPubkey,
		TokenProgram:                zeroPubkey,
		FeeRecipient:                zeroPubkey,
		BondingCurve:                zeroPubkey,
		AssociatedBondingCurve:      zeroPubkey,
		User:                        zeroPubkey,
		CreatorVault:                zeroPubkey,
		AssociatedQuoteUser:         zeroPubkey,
		AssociatedCreatorVault:      zeroPubkey,
		GlobalVolumeAccumulator:     zeroPubkey,
		UserVolumeAccumulator:       zeroPubkey,
		AssociatedTokenProgram:      zeroPubkey,
		AssociatedQuoteBondingCurve: zeroPubkey,
	}

	enrichPumpFunTradeFromAccounts(tr, accounts)

	if tr.QuoteMint != "account_C" || tr.TokenProgram != "account_D" ||
		tr.FeeRecipient != "account_G" || tr.BondingCurve != "account_K" ||
		tr.AssociatedBondingCurve != "account_L" || tr.User != "account_N" ||
		tr.CreatorVault != "account_Q" {
		t.Fatalf("short exact quote trade did not use v2 indexes: %+v", tr)
	}
	if tr.CreatorVault == "legacy_creator_vault" {
		t.Fatalf("short exact quote trade used legacy creator vault")
	}
}

func TestEnrichPumpFunTradeFromAccountsKeepsLegacyIndexesForLegacyShortBuyExactQuoteIn(t *testing.T) {
	accounts := make([]string, 18)
	for i := range accounts {
		accounts[i] = zeroPubkey
	}
	accounts[2] = "mint"
	accounts[8] = "spl_token"
	accounts[9] = "legacy_creator_vault"
	tr := &PumpFunTradeEvent{
		IxName:       "buy_exact_quote_in",
		IsBuy:        true,
		Mint:         "mint",
		TokenProgram: zeroPubkey,
		CreatorVault: zeroPubkey,
	}

	enrichPumpFunTradeFromAccounts(tr, accounts)

	if tr.TokenProgram != "spl_token" || tr.CreatorVault != "legacy_creator_vault" {
		t.Fatalf("legacy short exact quote trade used wrong indexes: %+v", tr)
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

func TestParsePumpfunTradeFromDataKeepsQuoteTailFields(t *testing.T) {
	quoteMintBytes := pumpfunTestPubkey(90)
	shareholderBytes := pumpfunTestPubkey(120)
	data := []byte{}
	data = append(data, pumpfunTestPubkey(1)...)
	data = appendPumpfunU64(data, 10)
	data = appendPumpfunU64(data, 20)
	data = append(data, 1)
	data = append(data, pumpfunTestPubkey(2)...)
	data = appendPumpfunI64(data, 30)
	data = appendPumpfunU64(data, 40)
	data = appendPumpfunU64(data, 50)
	data = appendPumpfunU64(data, 60)
	data = appendPumpfunU64(data, 70)
	data = append(data, pumpfunTestPubkey(3)...)
	data = appendPumpfunU64(data, 80)
	data = appendPumpfunU64(data, 90)
	data = append(data, pumpfunTestPubkey(4)...)
	data = appendPumpfunU64(data, 100)
	data = appendPumpfunU64(data, 110)
	data = append(data, 0)
	data = appendPumpfunU64(data, 120)
	data = appendPumpfunU64(data, 130)
	data = appendPumpfunU64(data, 140)
	data = appendPumpfunI64(data, 150)
	data = appendPumpfunString(data, "buy_exact_quote_in_v2")
	data = append(data, 0)
	data = appendPumpfunU64(data, 30)
	data = appendPumpfunU64(data, 170)
	data = appendPumpfunU64(data, 500)
	data = appendPumpfunU64(data, 600)
	data = appendPumpfunU32(data, 1)
	data = append(data, shareholderBytes...)
	data = appendPumpfunU16(data, 2500)
	data = append(data, quoteMintBytes...)
	data = appendPumpfunU64(data, 700)
	data = appendPumpfunU64(data, 800)
	data = appendPumpfunU64(data, 900)

	ev := parseTradeFromData(data, EventMetadata{Signature: "sig", Slot: 1}, false)
	if ev.Type != EventTypePumpFunBuyExactSolIn {
		t.Fatalf("expected exact quote trade, got %q", ev.Type)
	}
	tr, ok := ev.Data.(*PumpFunTradeEvent)
	if !ok {
		t.Fatalf("expected PumpFunTradeEvent, got %T", ev.Data)
	}
	if tr.QuoteMint != ReadPubkey(quoteMintBytes, 0) || tr.QuoteAmount != 700 ||
		tr.VirtualQuoteReserves != 800 || tr.RealQuoteReserves != 900 {
		t.Fatalf("quote tail fields not preserved: %+v", tr)
	}
	if tr.BuybackFeeBasisPoints != 500 || tr.BuybackFee != 600 {
		t.Fatalf("buyback fields not preserved: %+v", tr)
	}
	if len(tr.Shareholders) != 1 || tr.Shareholders[0].Address != ReadPubkey(shareholderBytes, 0) ||
		tr.Shareholders[0].ShareBps != 2500 {
		t.Fatalf("shareholders not preserved: %+v", tr.Shareholders)
	}
}

func TestMergePumpfunTradeDoesNotClobberQuoteTailWithDefaults(t *testing.T) {
	base := DexEvent{
		Type: EventTypePumpFunTrade,
		Data: &PumpFunTradeEvent{
			QuoteMint:            "USDC_MINT",
			QuoteAmount:          10,
			VirtualQuoteReserves: 20,
			RealQuoteReserves:    30,
		},
	}
	inner := DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{}}

	mergeDexEvents(&base, inner)

	tr := base.Data.(*PumpFunTradeEvent)
	if tr.QuoteMint != "USDC_MINT" || tr.QuoteAmount != 10 ||
		tr.VirtualQuoteReserves != 20 || tr.RealQuoteReserves != 30 {
		t.Fatalf("quote tail fields were clobbered: %+v", tr)
	}
}
