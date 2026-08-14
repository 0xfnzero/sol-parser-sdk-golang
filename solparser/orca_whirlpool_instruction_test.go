package solparser

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func orcaInstructionAccounts(n int) []string {
	accounts := make([]string, n)
	for i := range accounts {
		accounts[i] = fmt.Sprintf("account_%d", i)
	}
	return accounts
}

func orcaSwapInstruction(disc uint64, amount, threshold uint64, sqrt [16]byte, inputSpecified, aToB bool) []byte {
	data := make([]byte, 8+8+8+16+1+1)
	binary.LittleEndian.PutUint64(data[:8], disc)
	binary.LittleEndian.PutUint64(data[8:16], amount)
	binary.LittleEndian.PutUint64(data[16:24], threshold)
	copy(data[24:40], sqrt[:])
	if inputSpecified {
		data[40] = 1
	}
	if aToB {
		data[41] = 1
	}
	return data
}

func orcaLiquidityInstruction(disc uint64, liquidity [16]byte, amountA, amountB uint64) []byte {
	data := make([]byte, 8+16+8+8)
	binary.LittleEndian.PutUint64(data[:8], disc)
	copy(data[8:24], liquidity[:])
	binary.LittleEndian.PutUint64(data[24:32], amountA)
	binary.LittleEndian.PutUint64(data[32:40], amountB)
	return data
}

func orcaInitPoolInstruction(tickSpacing uint16, initialSqrtPrice [16]byte) []byte {
	data := make([]byte, 8+2+16)
	binary.LittleEndian.PutUint64(data[:8], disc8(17, 43, 80, 74, 168, 202, 6, 113))
	binary.LittleEndian.PutUint16(data[8:10], tickSpacing)
	copy(data[10:26], initialSqrtPrice[:])
	return data
}

func TestParseOrcaWhirlpoolSwapAndSwapV2InstructionFields(t *testing.T) {
	swapDisc := disc8(248, 198, 158, 145, 225, 117, 135, 200)
	swapV2Disc := disc8(43, 4, 237, 11, 26, 201, 30, 98)
	sqrt := u128ForTest(80, 123)
	ev := ParseOrcaWhirlpoolInstruction(
		orcaSwapInstruction(swapDisc, 111, 222, sqrt, true, false),
		orcaInstructionAccounts(4),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if ev.Type != EventTypeOrcaWhirlpoolSwap {
		t.Fatalf("expected OrcaWhirlpoolSwap, got %q", ev.Type)
	}
	swap, ok := ev.Data.(*OrcaWhirlpoolSwapEvent)
	if !ok {
		t.Fatalf("expected OrcaWhirlpoolSwapEvent, got %T", ev.Data)
	}
	if swap.Whirlpool != "account_2" || swap.AToB ||
		swap.PreSqrtPrice != u128LEDecimalString(sqrt) || swap.PostSqrtPrice != "" ||
		swap.InputAmount != 111 || swap.OutputAmount != 222 {
		t.Fatalf("unexpected swap fields: %+v", swap)
	}

	sqrtV2 := u128ForTest(80, 124)
	unified := ParseInstructionUnified(
		orcaSwapInstruction(swapV2Disc, 333, 444, sqrtV2, false, true),
		orcaInstructionAccounts(5),
		"sig",
		1,
		0,
		nil,
		10,
		&IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeOrcaWhirlpoolSwap}},
		ORCA_WHIRLPOOL_PROGRAM_ID,
	)
	swapV2, ok := unified.Data.(*OrcaWhirlpoolSwapEvent)
	if unified.Type != EventTypeOrcaWhirlpoolSwap || !ok {
		t.Fatalf("expected unified OrcaWhirlpoolSwap, got %q %+v", unified.Type, unified.Data)
	}
	if swapV2.Whirlpool != "account_4" || !swapV2.AToB ||
		swapV2.PreSqrtPrice != u128LEDecimalString(sqrtV2) ||
		swapV2.InputAmount != 0 || swapV2.OutputAmount != 333 {
		t.Fatalf("unexpected swap_v2 fields: %+v", swapV2)
	}
}

func TestParseOrcaWhirlpoolLiquidityInstructionFields(t *testing.T) {
	incDisc := disc8(46, 156, 243, 118, 13, 205, 251, 178)
	decDisc := disc8(160, 38, 208, 111, 104, 91, 44, 1)
	inc := ParseOrcaWhirlpoolInstruction(
		orcaLiquidityInstruction(incDisc, u128ForTest(80, 1), 222, 333),
		orcaInstructionAccounts(5),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if inc.Type != EventTypeOrcaWhirlpoolLiquidityIncreased {
		t.Fatalf("expected OrcaWhirlpoolLiquidityIncreased, got %q", inc.Type)
	}
	incData, ok := inc.Data.(*OrcaWhirlpoolLiquidityIncreasedEvent)
	if !ok {
		t.Fatalf("expected OrcaWhirlpoolLiquidityIncreasedEvent, got %T", inc.Data)
	}
	if incData.Whirlpool != "account_1" || incData.Position != "account_3" ||
		incData.Liquidity != u128LEDecimalString(u128ForTest(80, 1)) ||
		incData.TokenAAmount != 222 || incData.TokenBAmount != 333 {
		t.Fatalf("unexpected increase liquidity fields: %+v", incData)
	}

	dec := ParseOrcaWhirlpoolInstruction(
		orcaLiquidityInstruction(decDisc, u128ForTest(80, 2), 444, 555),
		orcaInstructionAccounts(5),
		"sig",
		1,
		0,
		nil,
		10,
	)
	if dec.Type != EventTypeOrcaWhirlpoolLiquidityDecreased {
		t.Fatalf("expected OrcaWhirlpoolLiquidityDecreased, got %q", dec.Type)
	}
	decData, ok := dec.Data.(*OrcaWhirlpoolLiquidityDecreasedEvent)
	if !ok {
		t.Fatalf("expected OrcaWhirlpoolLiquidityDecreasedEvent, got %T", dec.Data)
	}
	if decData.Whirlpool != "account_1" || decData.Position != "account_3" ||
		decData.Liquidity != u128LEDecimalString(u128ForTest(80, 2)) ||
		decData.TokenAAmount != 444 || decData.TokenBAmount != 555 {
		t.Fatalf("unexpected decrease liquidity fields: %+v", decData)
	}
}

func TestParseOrcaWhirlpoolInitializePoolInstructionAndFilters(t *testing.T) {
	initialSqrt := u128ForTest(96, 99)
	ev := ParseInstructionUnified(
		orcaInitPoolInstruction(128, initialSqrt),
		orcaInstructionAccounts(10),
		"sig",
		1,
		0,
		nil,
		10,
		&IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeOrcaWhirlpoolPoolInitialized}},
		ORCA_WHIRLPOOL_PROGRAM_ID,
	)
	if ev.Type != EventTypeOrcaWhirlpoolPoolInitialized {
		t.Fatalf("expected OrcaWhirlpoolPoolInitialized, got %q", ev.Type)
	}
	init, ok := ev.Data.(*OrcaWhirlpoolPoolInitializedEvent)
	if !ok {
		t.Fatalf("expected OrcaWhirlpoolPoolInitializedEvent, got %T", ev.Data)
	}
	if init.Whirlpool != "account_1" || init.WhirlpoolsConfig != "account_2" ||
		init.TokenMintA != "account_3" || init.TokenMintB != "account_4" ||
		init.TickSpacing != 128 || init.TokenProgramA != "account_8" ||
		init.TokenProgramB != "account_9" || init.InitialSqrtPrice != u128LEDecimalString(initialSqrt) {
		t.Fatalf("unexpected initialize pool fields: %+v", init)
	}

	excluded := ParseInstructionUnified(
		orcaInitPoolInstruction(128, initialSqrt),
		orcaInstructionAccounts(10),
		"sig",
		1,
		0,
		nil,
		10,
		&ExcludeFilter{ExcludeTypes: []EventType{EventTypeOrcaWhirlpoolPoolInitialized}},
		ORCA_WHIRLPOOL_PROGRAM_ID,
	)
	if excluded.Type != "" {
		t.Fatalf("exclude filter should drop Orca pool initialized, got %q", excluded.Type)
	}
}
