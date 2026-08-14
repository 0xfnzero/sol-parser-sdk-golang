package solparser

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func raydiumClmmTestAccounts(n int) []string {
	accounts := make([]string, n)
	for i := range accounts {
		accounts[i] = fmt.Sprintf("account_%d", i)
	}
	return accounts
}

func clmmU64Instruction(disc uint64, values ...uint64) []byte {
	data := make([]byte, 8+len(values)*8)
	binary.LittleEndian.PutUint64(data[:8], disc)
	for i, value := range values {
		binary.LittleEndian.PutUint64(data[8+i*8:8+(i+1)*8], value)
	}
	return data
}

func clmmLiquidityInstruction(disc uint64, liquidity [16]byte, amount0, amount1 uint64) []byte {
	data := make([]byte, 8+16+8+8)
	binary.LittleEndian.PutUint64(data[:8], disc)
	copy(data[8:24], liquidity[:])
	binary.LittleEndian.PutUint64(data[24:32], amount0)
	binary.LittleEndian.PutUint64(data[32:40], amount1)
	return data
}

func clmmOpenPositionInstruction(disc uint64, lower, upper int32, liquidity [16]byte) []byte {
	data := make([]byte, 8+4+4+4+4+16+8+8)
	binary.LittleEndian.PutUint64(data[:8], disc)
	binary.LittleEndian.PutUint32(data[8:12], uint32(lower))
	binary.LittleEndian.PutUint32(data[12:16], uint32(upper))
	copy(data[24:40], liquidity[:])
	return data
}

func clmmCreateCustomizablePoolInstruction(sqrtPriceX64 [16]byte) []byte {
	data := make([]byte, 8+16)
	binary.LittleEndian.PutUint64(data[:8], instrClmmCreateCustomizablePool)
	copy(data[8:24], sqrtPriceX64[:])
	return data
}

func TestParseRaydiumClmmDecreaseUsesRustV2InstructionDiscriminator(t *testing.T) {
	liquidity := u128ForTest(80, 111)
	ev := ParseRaydiumClmmInstruction(
		clmmLiquidityInstruction(instrClmmDecLiqV2, liquidity, 222, 333),
		raydiumClmmTestAccounts(4),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if ev.Type != EventTypeRaydiumClmmDecreaseLiquidity {
		t.Fatalf("expected RaydiumClmmDecreaseLiquidity, got %q", ev.Type)
	}
	data, ok := ev.Data.(*RaydiumClmmDecreaseLiquidityEvent)
	if !ok {
		t.Fatalf("expected RaydiumClmmDecreaseLiquidityEvent, got %T", ev.Data)
	}
	if data.Pool != "account_3" || data.PositionNftMint != "account_1" || data.User != "account_0" {
		t.Fatalf("unexpected accounts: %+v", data)
	}
	if data.Liquidity != u128LEDecimalString(liquidity) || data.Amount0Min != 222 || data.Amount1Min != 333 {
		t.Fatalf("unexpected amounts: %+v", data)
	}

	oldLogDisc := disc8(160, 38, 208, 111, 104, 91, 44, 1)
	if got := ParseRaydiumClmmInstruction(clmmLiquidityInstruction(oldLogDisc, u128ForTest(0, 111), 222, 333), raydiumClmmTestAccounts(4), "sig", 1, 0, nil, 10); got.Type != "" {
		t.Fatalf("old log discriminator should not parse as instruction, got %q", got.Type)
	}
}

func TestParseRaydiumClmmOpenAndClosePosition(t *testing.T) {
	liquidity := u128ForTest(80, 123)
	openIx := clmmOpenPositionInstruction(instrClmmOpenPositionV2, -10, 20, liquidity)
	open := ParseRaydiumClmmInstruction(
		openIx,
		raydiumClmmTestAccounts(7),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if open.Type != EventTypeRaydiumClmmOpenPosition {
		t.Fatalf("expected RaydiumClmmOpenPosition, got %q", open.Type)
	}
	openData, ok := open.Data.(*RaydiumClmmOpenPositionEvent)
	if !ok {
		t.Fatalf("expected RaydiumClmmOpenPositionEvent, got %T", open.Data)
	}
	if openData.Pool != "account_5" || openData.User != "account_1" || openData.PositionNftMint != "account_2" {
		t.Fatalf("unexpected open accounts: %+v", openData)
	}
	if openData.TickLowerIndex != -10 || openData.TickUpperIndex != 20 || openData.Liquidity != u128LEDecimalString(liquidity) {
		t.Fatalf("unexpected open values: %+v", openData)
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeRaydiumClmmOpenPosition}) {
		t.Fatalf("include_only prefilter should allow Raydium CLMM open-position instructions")
	}
	unified := ParseInstructionUnified(
		openIx,
		raydiumClmmTestAccounts(7),
		"sig",
		1,
		0,
		nil,
		10,
		&IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeRaydiumClmmOpenPosition}},
		GrpcRaydiumClmmProgramID,
	)
	if unified.Type != EventTypeRaydiumClmmOpenPosition {
		t.Fatalf("unified parser should route Raydium CLMM, got %q", unified.Type)
	}

	legacyOpen := ParseRaydiumClmmInstruction(
		clmmOpenPositionInstruction(instrClmmOpenPosition, -11, 21, u128ForTest(0, 456)),
		raydiumClmmTestAccounts(7),
		"sig",
		1,
		0,
		nil,
		10,
	)
	legacyOpenData, ok := legacyOpen.Data.(*RaydiumClmmOpenPositionEvent)
	if legacyOpen.Type != EventTypeRaydiumClmmOpenPosition || !ok || legacyOpenData.Pool != "account_5" {
		t.Fatalf("legacy open-position should use pool account 5, got %q %+v", legacyOpen.Type, legacyOpen.Data)
	}

	token22Open := ParseRaydiumClmmInstruction(
		clmmOpenPositionInstruction(instrClmmOpenPositionWithToken22Nft, -12, 22, u128ForTest(0, 789)),
		raydiumClmmTestAccounts(7),
		"sig",
		1,
		0,
		nil,
		10,
	)
	token22OpenData, ok := token22Open.Data.(*RaydiumClmmOpenPositionEvent)
	if token22Open.Type != EventTypeRaydiumClmmOpenPosition || !ok || token22OpenData.Pool != "account_4" {
		t.Fatalf("Token22 open-position should use pool account 4, got %q %+v", token22Open.Type, token22Open.Data)
	}

	closeData := make([]byte, 8)
	binary.LittleEndian.PutUint64(closeData, instrClmmClosePosition)
	close := ParseRaydiumClmmInstruction(closeData, raydiumClmmTestAccounts(4), "sig", 1, 0, nil, 10)
	if close.Type != EventTypeRaydiumClmmClosePosition {
		t.Fatalf("expected RaydiumClmmClosePosition, got %q", close.Type)
	}
	closeEvent, ok := close.Data.(*RaydiumClmmClosePositionEvent)
	if !ok {
		t.Fatalf("expected RaydiumClmmClosePositionEvent, got %T", close.Data)
	}
	if closeEvent.Pool != zeroPubkey || closeEvent.User != "account_0" || closeEvent.PositionNftMint != "account_1" {
		t.Fatalf("unexpected close accounts: %+v", closeEvent)
	}
}

func TestParseRaydiumClmmCreateCustomizablePoolInstruction(t *testing.T) {
	sqrtPriceX64 := u128ForTest(80, 999)
	ev := ParseRaydiumClmmInstruction(
		clmmCreateCustomizablePoolInstruction(sqrtPriceX64),
		raydiumClmmTestAccounts(7),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if ev.Type != EventTypeRaydiumClmmCreatePool {
		t.Fatalf("expected RaydiumClmmCreatePool, got %q", ev.Type)
	}
	data, ok := ev.Data.(*RaydiumClmmCreatePoolEvent)
	if !ok {
		t.Fatalf("expected RaydiumClmmCreatePoolEvent, got %T", ev.Data)
	}
	if data.Pool != "account_2" || data.Creator != "account_0" ||
		data.Token0Mint != "account_3" || data.Token1Mint != "account_4" ||
		data.TokenVault0 != "account_5" || data.TokenVault1 != "account_6" {
		t.Fatalf("unexpected create-customizable accounts: %+v", data)
	}
	if data.SqrtPriceX64 != u128LEDecimalString(sqrtPriceX64) || data.OpenTime != 0 {
		t.Fatalf("unexpected create-customizable values: %+v", data)
	}
}

func TestMeteoraDbcLogEventsDoNotEnableInstructionPrefilter(t *testing.T) {
	if EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeMeteoraDbcSwap}) {
		t.Fatalf("DBC log-only events should not enable instruction parsing")
	}
}

func TestParseRaydiumCpmmNormalInstructionUsesRustAccountsAndDefaults(t *testing.T) {
	deposit := ParseRaydiumCpmmInstruction(
		clmmU64Instruction(discCpmmDeposit, 111, 222, 333),
		raydiumClmmTestAccounts(4),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if deposit.Type != EventTypeRaydiumCpmmDeposit {
		t.Fatalf("expected RaydiumCpmmDeposit, got %q", deposit.Type)
	}
	dep := deposit.Data.(*RaydiumCpmmDepositEvent)
	if dep.Pool != "account_0" || dep.User != "account_1" ||
		dep.LpTokenAmount != 111 || dep.Token0Amount != 222 || dep.Token1Amount != 333 {
		t.Fatalf("unexpected CPMM deposit: %+v", dep)
	}

	withdraw := ParseRaydiumCpmmInstruction(
		clmmU64Instruction(discCpmmWithdraw, 444, 555, 666),
		raydiumClmmTestAccounts(4),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if withdraw.Type != EventTypeRaydiumCpmmWithdraw {
		t.Fatalf("expected RaydiumCpmmWithdraw, got %q", withdraw.Type)
	}
	wit := withdraw.Data.(*RaydiumCpmmWithdrawEvent)
	if wit.Pool != "account_0" || wit.User != "account_1" ||
		wit.LpTokenAmount != 444 || wit.Token0Amount != 555 || wit.Token1Amount != 666 {
		t.Fatalf("unexpected CPMM withdraw: %+v", wit)
	}
}

func TestParseMeteoraPoolsAndDlmmOuterInstructionsAreRouted(t *testing.T) {
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeMeteoraPoolsSwap}) {
		t.Fatalf("include_only prefilter should allow Meteora Pools instructions")
	}
	pools := ParseInstructionUnified(
		clmmU64Instruction(instrMeteoraPoolsSwap, 111, 222),
		raydiumClmmTestAccounts(2),
		"sig",
		1,
		0,
		nil,
		10,
		&IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeMeteoraPoolsSwap}},
		METEORA_POOLS_PROGRAM_ID,
	)
	if pools.Type != EventTypeMeteoraPoolsSwap {
		t.Fatalf("expected MeteoraPoolsSwap, got %q", pools.Type)
	}
	poolsSwap, ok := pools.Data.(*MeteoraPoolsSwapEvent)
	if !ok {
		t.Fatalf("expected MeteoraPoolsSwapEvent, got %T", pools.Data)
	}
	if poolsSwap.InAmount != 111 || poolsSwap.OutAmount != 222 {
		t.Fatalf("unexpected Meteora Pools swap values: %+v", poolsSwap)
	}

	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeMeteoraDlmmSwap}) {
		t.Fatalf("include_only prefilter should allow Meteora DLMM instructions")
	}
	dlmmIx := clmmU64Instruction(instrDlmmSwap, 333, 444)
	dlmm := ParseInstructionUnified(
		dlmmIx,
		raydiumClmmTestAccounts(11),
		"sig",
		1,
		0,
		nil,
		10,
		&IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeMeteoraDlmmSwap}},
		METEORA_DLMM_PROGRAM_ID,
	)
	if dlmm.Type != EventTypeMeteoraDlmmSwap {
		t.Fatalf("expected MeteoraDlmmSwap, got %q", dlmm.Type)
	}
	dlmmSwap, ok := dlmm.Data.(*MeteoraDlmmSwapEvent)
	if !ok {
		t.Fatalf("expected MeteoraDlmmSwapEvent, got %T", dlmm.Data)
	}
	if dlmmSwap.Pool != "account_0" || dlmmSwap.From != "account_10" || dlmmSwap.AmountIn != 333 {
		t.Fatalf("unexpected Meteora DLMM swap values: %+v", dlmmSwap)
	}
}

func TestMeteoraDlmmVersionedAccountLayoutsUseCurrentIDL(t *testing.T) {
	add := ParseMeteoraDlmmInstruction(
		clmmU64Instruction(instrDlmmAddLiquidity2),
		raydiumClmmTestAccounts(14),
		"sig", 1, 0, nil, 10,
	)
	liquidity, ok := add.Data.(*MeteoraDlmmAddLiquidityEvent)
	if !ok || liquidity.From != "account_9" {
		t.Fatalf("unexpected add_liquidity2 accounts: %+v", add.Data)
	}

	closeEvent := ParseMeteoraDlmmInstruction(
		clmmU64Instruction(instrDlmmClosePosition2),
		raydiumClmmTestAccounts(5),
		"sig", 1, 0, nil, 10,
	)
	closePosition, ok := closeEvent.Data.(*MeteoraDlmmClosePositionEvent)
	if !ok || closePosition.Owner != "account_1" || closePosition.Pool != zeroPubkey {
		t.Fatalf("unexpected close_position2 accounts: %+v", closeEvent.Data)
	}
}

func TestParseRaydiumLaunchlabBuyExactInUsesRustLayout(t *testing.T) {
	ix := clmmU64Instruction(instrRaydiumLaunchlabBuyExactIn, 111, 222)
	ev := ParseRaydiumLaunchlabInstruction(ix, raydiumClmmTestAccounts(6), "sig", 1, 0, nil, 10)
	if ev.Type != EventTypeRaydiumLaunchlabTrade {
		t.Fatalf("expected RaydiumLaunchlabTrade, got %q", ev.Type)
	}
	trade, ok := ev.Data.(*RaydiumLaunchlabTradeEvent)
	if !ok {
		t.Fatalf("expected RaydiumLaunchlabTradeEvent, got %T", ev.Data)
	}
	if trade.PoolState != "account_4" || trade.User != "account_0" {
		t.Fatalf("unexpected accounts: %+v", trade)
	}
	if trade.AmountIn != 111 || trade.AmountOut != 222 || !trade.IsBuy || !trade.ExactIn {
		t.Fatalf("unexpected trade values: %+v", trade)
	}

	routed := ParseInstructionUnified(
		ix,
		raydiumClmmTestAccounts(6),
		"sig",
		1,
		0,
		nil,
		10,
		nil,
		RAYDIUM_LAUNCHLAB_PROGRAM_ID,
	)
	if routed.Type != EventTypeRaydiumLaunchlabTrade {
		t.Fatalf("unified parser should route Raydium LaunchLab, got %q", routed.Type)
	}
}

func TestNonPumpAccountEventNamesDoNotEnableInstructionPrefilter(t *testing.T) {
	events := map[EventType]bool{}
	for _, eventType := range AllEventTypes() {
		events[eventType] = true
	}
	if !events[EventTypeAccountRaydiumCpmmPoolState] || !events[EventTypeAccountOrcaWhirlpool] {
		t.Fatalf("expected new non-Pump account event names in AllEventTypes")
	}
	if EventTypeFilterAllowsInstructionParsing([]EventType{
		EventTypeAccountRaydiumCpmmPoolState,
		EventTypeAccountOrcaWhirlpool,
	}) {
		t.Fatalf("account-only event filters should not enable instruction parsing")
	}
}
