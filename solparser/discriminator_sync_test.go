package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"testing"
)

// 回归：Program log 的 8 字节 disc（binary.go）须与 inner 16 字节的后 8 字节及 magic 前缀一致（Rust pump_amm_inner / pump_inner）。

func TestPumpSwapInnerDiscMatchesLogDiscPlusMagic(t *testing.T) {
	magic := []byte{228, 69, 165, 46, 81, 203, 154, 29}
	var buf8 [8]byte
	binary.LittleEndian.PutUint64(buf8[:], discPSBuy)
	full := append(magic, buf8[:]...)
	if !bytes.Equal(full, pumpswapInnerBuy) {
		t.Fatalf("buy inner disc mismatch")
	}
	binary.LittleEndian.PutUint64(buf8[:], discPSSell)
	full = append(magic, buf8[:]...)
	if !bytes.Equal(full, pumpswapInnerSell) {
		t.Fatalf("sell inner disc mismatch")
	}
}

func TestPumpfunInnerTradeDiscMatchesLogDiscPlusSuffix(t *testing.T) {
	var d16 [16]byte
	binary.LittleEndian.PutUint64(d16[:8], discPumpTrade)
	copy(d16[8:], []byte{155, 167, 108, 32, 122, 76, 173, 64})
	if !bytes.Equal(d16[:], pumpfunInnerTradeEvent) {
		t.Fatalf("pumpfun inner trade disc mismatch")
	}
}

func TestParseInnerInstructionAppliesActualEventTypeFilter(t *testing.T) {
	const pumpswapBuyPayloadLen = 16*8 + 7*32 + 1 + 5*8 + 4
	ix := append(append([]byte{}, pumpswapInnerBuy...), make([]byte, pumpswapBuyPayloadLen)...)

	ev := ParseInnerInstructionUnified(ix, nil, "sig", 1, 0, nil, 10, nil, PUMPSWAP_PROGRAM_ID, false)
	if ev.Type != EventTypePumpSwapBuy {
		t.Fatalf("expected unfiltered PumpSwapBuy, got %q", ev.Type)
	}

	ev = ParseInnerInstructionUnified(
		ix,
		nil,
		"sig",
		1,
		0,
		nil,
		10,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpSwapCreatePool}),
		PUMPSWAP_PROGRAM_ID,
		false,
	)
	if ev.Type != "" {
		t.Fatalf("include-only PumpSwapCreatePool should drop buy CPI, got %q", ev.Type)
	}

	ev = ParseInnerInstructionUnified(
		ix,
		nil,
		"sig",
		1,
		0,
		nil,
		10,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpSwapTrade}),
		PUMPSWAP_PROGRAM_ID,
		false,
	)
	if ev.Type != EventTypePumpSwapBuy {
		t.Fatalf("PumpSwapTrade include-only should keep buy CPI, got %q", ev.Type)
	}
}

func innerCPI(eventDisc uint64, payload []byte) []byte {
	var d [8]byte
	binary.LittleEndian.PutUint64(d[:], eventDisc)
	ix := make([]byte, 0, 16+len(payload))
	ix = append(ix, d[:]...)
	ix = append(ix, eventCPISuffix...)
	ix = append(ix, payload...)
	return ix
}

func currentAnchorEventCPI(eventDisc uint64, payload []byte) []byte {
	var d [8]byte
	binary.LittleEndian.PutUint64(d[:], eventDisc)
	ix := make([]byte, 0, 16+len(payload))
	ix = append(ix, eventCPIPrefix...)
	ix = append(ix, d[:]...)
	ix = append(ix, payload...)
	return ix
}

func TestParseMeteoraDlmmCurrentAnchorEventCPI(t *testing.T) {
	payload := make([]byte, 147)
	payload[72] = 1
	binary.LittleEndian.PutUint64(payload[89:97], 100)
	binary.LittleEndian.PutUint64(payload[105:113], 90)

	ev := ParseInnerInstructionUnified(
		currentAnchorEventCPI(dlmmSwap2, payload),
		nil,
		"sig",
		1,
		0,
		nil,
		10,
		EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDlmmSwap}),
		METEORA_DLMM_PROGRAM_ID,
		false,
	)
	if ev.Type != EventTypeMeteoraDlmmSwap {
		t.Fatalf("expected MeteoraDlmmSwap, got %q", ev.Type)
	}
	swap := ev.Data.(*MeteoraDlmmSwapEvent)
	if swap.AmountIn != 100 || swap.AmountOut != 90 {
		t.Fatalf("unexpected swap values: %+v", swap)
	}
}

func TestParseInnerInstructionCoversRustAdvancedPrograms(t *testing.T) {
	clmmPayload := make([]byte, 32+32+1+4+8+8)
	clmmPayload[64] = 1
	binary.LittleEndian.PutUint32(clmmPayload[65:69], uint32(42))
	binary.LittleEndian.PutUint64(clmmPayload[69:77], 1000)
	binary.LittleEndian.PutUint64(clmmPayload[77:85], 7)
	ev := ParseInnerInstructionUnified(innerCPI(discClmmOpenLimitOrder, clmmPayload), nil, "sig", 1, 0, nil, 10, nil, RAYDIUM_CLMM_PROGRAM_ID, false)
	if ev.Type != EventTypeRaydiumClmmOpenLimitOrder {
		t.Fatalf("expected CLMM open limit order CPI, got %q", ev.Type)
	}

	cpmmPayload := make([]byte, 32+32+32+32+8+8)
	cpmmPayload[0] = 1
	cpmmPayload[96] = 2
	binary.LittleEndian.PutUint64(cpmmPayload[128:136], 11)
	binary.LittleEndian.PutUint64(cpmmPayload[136:144], 22)
	ev = ParseInnerInstructionUnified(innerCPI(discCpmmCreatePool, cpmmPayload), nil, "sig", 1, 0, nil, 10, nil, RAYDIUM_CPMM_PROGRAM_ID, false)
	if ev.Type != EventTypeRaydiumCpmmInitialize {
		t.Fatalf("expected CPMM initialize CPI, got %q", ev.Type)
	}

	orPayload := make([]byte, 32*5+2+1+1+16)
	binary.LittleEndian.PutUint16(orPayload[128:130], 64)
	ev = ParseInnerInstructionUnified(innerCPI(discOrcaPoolInit, orPayload), nil, "sig", 1, 0, nil, 10, nil, ORCA_WHIRLPOOL_PROGRAM_ID, false)
	if ev.Type != EventTypeOrcaWhirlpoolPoolInitialized {
		t.Fatalf("expected Orca pool init CPI, got %q", ev.Type)
	}

	meteoraPayload := make([]byte, 8+8+8+8+32)
	binary.LittleEndian.PutUint64(meteoraPayload[:8], 1)
	ev = ParseInnerInstructionUnified(innerCPI(discMeteoraSetPoolFees, meteoraPayload), nil, "sig", 1, 0, nil, 10, nil, METEORA_POOLS_PROGRAM_ID, false)
	if ev.Type != EventTypeMeteoraPoolsSetPoolFees {
		t.Fatalf("expected Meteora Pools set fees CPI, got %q", ev.Type)
	}
}

func TestParseInnerCompiledInstructionFallback(t *testing.T) {
	ix := make([]byte, 8+8+8+8)
	binary.LittleEndian.PutUint64(ix[:8], instrCpmmInitialize)
	binary.LittleEndian.PutUint64(ix[8:16], 111)
	binary.LittleEndian.PutUint64(ix[16:24], 222)
	accounts := []string{"pool", "creator"}
	ev := ParseInnerCompiledInstructionIfSupported(ix, accounts, "sig", 1, 0, nil, 10, nil, RAYDIUM_CPMM_PROGRAM_ID)
	if ev.Type != EventTypeRaydiumCpmmInitialize {
		t.Fatalf("expected normal inner CPMM initialize, got %q", ev.Type)
	}
	init := ev.Data.(*RaydiumCpmmInitializeEvent)
	if init.Pool != "pool" || init.Creator != "creator" || init.InitAmount0 != 111 || init.InitAmount1 != 222 {
		t.Fatalf("unexpected CPMM initialize fallback data: %+v", init)
	}

	eventCPI := innerCPI(discDammSwap, make([]byte, 48))
	if normalInstructionDataMayParse(METEORA_DAMM_V2_PROGRAM_ID, eventCPI) {
		t.Fatalf("DAMM event CPI must not pass normal inner fallback")
	}
}

func TestMeteoraDammInitializePoolNormalGateUsesRustDiscriminator(t *testing.T) {
	ix := make([]byte, 8+16+16+1)
	binary.LittleEndian.PutUint64(ix[:8], discDammInit)
	if !normalInstructionDataMayParse(METEORA_DAMM_V2_PROGRAM_ID, ix) {
		t.Fatalf("DAMM InitializePool discriminator should pass normal instruction gate")
	}

	dlmmIx := make([]byte, 8+16+16+1)
	binary.LittleEndian.PutUint64(dlmmIx[:8], dlmmInitPool)
	if normalInstructionDataMayParse(METEORA_DAMM_V2_PROGRAM_ID, dlmmIx) {
		t.Fatalf("DLMM InitializePool discriminator must not pass DAMM normal instruction gate")
	}

	accounts := []string{"creator", "1", "2", "3", "4", "5", "pool", "7", "tokenA", "tokenB"}
	ev := ParseMeteoraDammInstruction(ix, accounts, "sig", 1, 0, nil, 10)
	if ev.Type != EventTypeMeteoraDammV2InitializePool {
		t.Fatalf("expected DAMM InitializePool instruction, got %q", ev.Type)
	}
	init := ev.Data.(*MeteoraDammV2InitializePoolEvent)
	if init.Creator != "creator" || init.Pool != "pool" || init.TokenAMint != "tokenA" || init.TokenBMint != "tokenB" {
		t.Fatalf("unexpected DAMM InitializePool accounts: %+v", init)
	}
}

func TestParseLogOptimizedCoversCpmmCreatePoolProgramData(t *testing.T) {
	payload := make([]byte, 32+32+32+32+8+8)
	payload[0] = 1
	payload[96] = 2
	binary.LittleEndian.PutUint64(payload[128:136], 33)
	binary.LittleEndian.PutUint64(payload[136:144], 44)
	buf := make([]byte, 8+len(payload))
	binary.LittleEndian.PutUint64(buf[:8], discCpmmCreatePool)
	copy(buf[8:], payload)
	log := "Program data: " + base64.StdEncoding.EncodeToString(buf)
	ev := ParseLogOptimizedWithProgramID(log, "sig", 1, 0, nil, 10, EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumCpmmInitialize}), false, "", RAYDIUM_CPMM_PROGRAM_ID)
	if ev.Type != EventTypeRaydiumCpmmInitialize {
		t.Fatalf("expected scoped CPMM create-pool log, got %q", ev.Type)
	}
	init := ev.Data.(*RaydiumCpmmInitializeEvent)
	if init.InitAmount0 != 33 || init.InitAmount1 != 44 {
		t.Fatalf("unexpected CPMM create-pool log data: %+v", init)
	}
}

func TestParseLogOptimizedRoutesSharedDiscriminatorsWithProgramScope(t *testing.T) {
	launchPayload := make([]byte, 139)
	copy(launchPayload[0:32], pkForTest(10))
	binary.LittleEndian.PutUint64(launchPayload[88:96], 111)
	binary.LittleEndian.PutUint64(launchPayload[96:104], 222)
	launchPayload[136] = 0
	launchPayload[138] = 1
	launchBuf := make([]byte, 8+len(launchPayload))
	binary.LittleEndian.PutUint64(launchBuf[:8], discRaydiumLaunchlabTrade)
	copy(launchBuf[8:], launchPayload)
	launchLog := "Program data: " + base64.StdEncoding.EncodeToString(launchBuf)
	ev := ParseLogOptimized(launchLog, "sig", 1, 0, nil, 10, EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumLaunchlabTrade}), false, "")
	if ev.Type != EventTypeRaydiumLaunchlabTrade {
		t.Fatalf("expected unscoped LaunchLab trade with LaunchLab filter, got %q", ev.Type)
	}
	launch := ev.Data.(*RaydiumLaunchlabTradeEvent)
	if launch.AmountIn != 111 || launch.AmountOut != 222 || !launch.IsBuy || !launch.ExactIn {
		t.Fatalf("unexpected LaunchLab trade: %+v", launch)
	}

	dlmmPayload := make([]byte, 32+32+4+4+8+8+1+8+8+16+8)
	copy(dlmmPayload[0:32], pkForTest(11))
	copy(dlmmPayload[32:64], pkForTest(12))
	binary.LittleEndian.PutUint64(dlmmPayload[72:80], 333)
	binary.LittleEndian.PutUint64(dlmmPayload[80:88], 444)
	dlmmPayload[88] = 1
	binary.LittleEndian.PutUint64(dlmmPayload[89:97], 5)
	binary.LittleEndian.PutUint64(dlmmPayload[97:105], 6)
	binary.LittleEndian.PutUint64(dlmmPayload[121:129], 7)
	dlmmBuf := make([]byte, 8+len(dlmmPayload))
	binary.LittleEndian.PutUint64(dlmmBuf[:8], dlmmSwap)
	copy(dlmmBuf[8:], dlmmPayload)
	dlmmLog := "Program data: " + base64.StdEncoding.EncodeToString(dlmmBuf)
	ev = ParseLogOptimizedWithProgramID(dlmmLog, "sig", 1, 0, nil, 10, EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDlmmSwap}), false, "", METEORA_DLMM_PROGRAM_ID)
	if ev.Type != EventTypeMeteoraDlmmSwap {
		t.Fatalf("expected scoped DLMM swap, got %q", ev.Type)
	}
	swap := ev.Data.(*MeteoraDlmmSwapEvent)
	if swap.AmountIn != 333 || swap.AmountOut != 444 {
		t.Fatalf("unexpected DLMM swap: %+v", swap)
	}
}

func TestParsePumpSwapBuyRejectsTruncatedMinBasePayload(t *testing.T) {
	if ev := parsePSBuyFromData(make([]byte, 396), EventMetadata{}); ev.Type != "" {
		t.Fatalf("truncated buy payload should not parse, got %q", ev.Type)
	}
	if ev := parsePSBuyFromData(make([]byte, 397), EventMetadata{}); ev.Type != EventTypePumpSwapBuy {
		t.Fatalf("full minimum buy payload should parse, got %q", ev.Type)
	}
}
