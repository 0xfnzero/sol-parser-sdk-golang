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
