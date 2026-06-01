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

func clmmOpenPositionInstruction(disc uint64, lower, upper int32, liquidity uint64) []byte {
	data := make([]byte, 8+4+4+4+4+8+8+8)
	binary.LittleEndian.PutUint64(data[:8], disc)
	binary.LittleEndian.PutUint32(data[8:12], uint32(lower))
	binary.LittleEndian.PutUint32(data[12:16], uint32(upper))
	binary.LittleEndian.PutUint64(data[24:32], liquidity)
	return data
}

func TestParseRaydiumClmmDecreaseUsesRustV2InstructionDiscriminator(t *testing.T) {
	ev := ParseRaydiumClmmInstruction(
		clmmU64Instruction(instrClmmDecLiqV2, 111, 222, 333),
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
	if data.Pool != "account_0" || data.PositionNftMint != "account_1" || data.User != "account_2" {
		t.Fatalf("unexpected accounts: %+v", data)
	}
	if data.Liquidity != "111" || data.Amount0Min != 222 || data.Amount1Min != 333 {
		t.Fatalf("unexpected amounts: %+v", data)
	}

	oldLogDisc := disc8(160, 38, 208, 111, 104, 91, 44, 1)
	if got := ParseRaydiumClmmInstruction(clmmU64Instruction(oldLogDisc, 111, 222, 333), raydiumClmmTestAccounts(4), "sig", 1, 0, nil, 10); got.Type != "" {
		t.Fatalf("old log discriminator should not parse as instruction, got %q", got.Type)
	}
}

func TestParseRaydiumClmmOpenAndClosePosition(t *testing.T) {
	openIx := clmmOpenPositionInstruction(instrClmmOpenPositionV2, -10, 20, 123)
	open := ParseRaydiumClmmInstruction(
		openIx,
		raydiumClmmTestAccounts(4),
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
	if openData.Pool != "account_0" || openData.User != "account_1" || openData.PositionNftMint != "account_2" {
		t.Fatalf("unexpected open accounts: %+v", openData)
	}
	if openData.TickLowerIndex != -10 || openData.TickUpperIndex != 20 || openData.Liquidity != "123" {
		t.Fatalf("unexpected open values: %+v", openData)
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeRaydiumClmmOpenPosition}) {
		t.Fatalf("include_only prefilter should allow Raydium CLMM open-position instructions")
	}
	unified := ParseInstructionUnified(
		openIx,
		raydiumClmmTestAccounts(4),
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
	if closeEvent.Pool != "account_0" || closeEvent.User != "account_1" || closeEvent.PositionNftMint != "account_2" {
		t.Fatalf("unexpected close accounts: %+v", closeEvent)
	}
}

func TestMeteoraDbcLogEventsDoNotEnableInstructionPrefilter(t *testing.T) {
	if EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeMeteoraDbcSwap}) {
		t.Fatalf("DBC log-only events should not enable instruction parsing")
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
	dlmmIx := make([]byte, 17)
	dlmmIx[0] = 11
	binary.LittleEndian.PutUint64(dlmmIx[1:9], 333)
	binary.LittleEndian.PutUint64(dlmmIx[9:17], 444)
	dlmm := ParseInstructionUnified(
		dlmmIx,
		raydiumClmmTestAccounts(3),
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
	if dlmmSwap.Pool != "account_0" || dlmmSwap.From != "account_1" || dlmmSwap.AmountIn != 333 {
		t.Fatalf("unexpected Meteora DLMM swap values: %+v", dlmmSwap)
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
