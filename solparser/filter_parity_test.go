package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"testing"

	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
)

func pumpFeesUpdateAdminLogForTest() string {
	buf := make([]byte, 8+8+32+32)
	binary.LittleEndian.PutUint64(buf[:8], discPumpFeesUpdateAdmin)
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func pumpfunBuyExactSolInTradeLogForTest() string {
	buf := make([]byte, 0, 260)
	putU64 := func(v uint64) {
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], v)
		buf = append(buf, b[:]...)
	}
	putI64 := func(v int64) {
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], uint64(v))
		buf = append(buf, b[:]...)
	}
	putU32 := func(v uint32) {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		buf = append(buf, b[:]...)
	}
	putPubkey := func(seed byte) {
		for i := 0; i < 32; i++ {
			buf = append(buf, seed+byte(i))
		}
	}
	putString := func(value string) {
		putU32(uint32(len(value)))
		buf = append(buf, []byte(value)...)
	}

	putU64(discPumpTrade)
	putPubkey(1)
	putU64(10)
	putU64(20)
	buf = append(buf, 1)
	putPubkey(2)
	putI64(30)
	for _, v := range []uint64{40, 50, 60, 70} {
		putU64(v)
	}
	putPubkey(3)
	putU64(80)
	putU64(90)
	putPubkey(4)
	putU64(100)
	putU64(110)
	buf = append(buf, 0)
	for _, v := range []uint64{120, 130, 140} {
		putU64(v)
	}
	putI64(150)
	putString("buy_exact_sol_in")
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func pumpFeesLargeUpdateFeeSharesLogForTest(t *testing.T) string {
	buf := make([]byte, 0, 8+8+32+32+32+4+64*34)
	putU64 := func(v uint64) {
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], v)
		buf = append(buf, b[:]...)
	}
	putU16 := func(v uint16) {
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], v)
		buf = append(buf, b[:]...)
	}
	putU32 := func(v uint32) {
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], v)
		buf = append(buf, b[:]...)
	}
	putPubkey := func(seed byte) {
		for i := 0; i < 32; i++ {
			buf = append(buf, seed+byte(i))
		}
	}

	putU64(discPumpFeesUpdateFeeShares)
	putU64(uint64(1777920719))
	putPubkey(1)
	putPubkey(2)
	putPubkey(3)
	putU32(64)
	for i := 0; i < 64; i++ {
		putPubkey(byte(40 + i))
		putU16(uint16(1000 + i))
	}
	if len(buf) <= 2048 {
		t.Fatalf("large Program data fixture should exceed 2048 bytes, got %d", len(buf))
	}
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func TestParseLogOptimizedAppliesEventTypeFilter(t *testing.T) {
	log := pumpFeesUpdateAdminLogForTest()

	ev := ParseLogOptimized(log, "sig", 1, 0, nil, 1, nil, false, "")
	if ev.Type != EventTypePumpFeesUpdateAdmin {
		t.Fatalf("expected unfiltered UpdateAdmin, got %q", ev.Type)
	}

	ev = ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpFeesUpdateAdmin}),
		false,
		"",
	)
	if ev.Type != EventTypePumpFeesUpdateAdmin {
		t.Fatalf("expected include-only UpdateAdmin, got %q", ev.Type)
	}

	ev = ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunCreate}),
		false,
		"",
	)
	if ev.Type != "" {
		t.Fatalf("expected non-matching include-only filter to drop event, got %q", ev.Type)
	}

	ev = ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterExclude([]EventType{EventTypePumpFeesUpdateAdmin}),
		false,
		"",
	)
	if ev.Type != "" {
		t.Fatalf("expected exclude filter to drop event, got %q", ev.Type)
	}
}

func TestParseLogOptimizedAcceptsLargeProgramDataPayloads(t *testing.T) {
	ev := ParseLogOptimized(
		pumpFeesLargeUpdateFeeSharesLogForTest(t),
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpFeesUpdateFeeShares}),
		false,
		"",
	)
	if ev.Type != EventTypePumpFeesUpdateFeeShares {
		t.Fatalf("expected large update fee shares event, got %q", ev.Type)
	}
	data, ok := ev.Data.(*PumpFeesUpdateFeeSharesEvent)
	if !ok {
		t.Fatalf("expected PumpFeesUpdateFeeSharesEvent, got %T", ev.Data)
	}
	if len(data.NewShareholders) != 64 || data.NewShareholders[63].ShareBps != 1063 {
		t.Fatalf("unexpected shareholders: %+v", data.NewShareholders)
	}
}

func TestScopedPumpfunTradePrefilterAcceptsBuyFamilyFilters(t *testing.T) {
	ev := ParseLogOptimizedWithProgramID(
		pumpfunBuyExactSolInTradeLogForTest(),
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunBuy}),
		false,
		"",
		PUMPFUN_PROGRAM_ID,
	)
	if ev.Type != EventTypePumpFunBuyExactSolIn {
		t.Fatalf("expected scoped PumpFun buy-exact-sol-in, got %q", ev.Type)
	}
}

func dbcSwapLogForTest() string {
	buf := make([]byte, 8+32+32+2+8*9+16)
	o := 0
	binary.LittleEndian.PutUint64(buf[o:o+8], discDbcSwap)
	o += 8
	for i := 0; i < 32; i++ {
		buf[o+i] = 1
		buf[o+32+i] = 2
	}
	o += 64
	buf[o] = 1
	o++
	buf[o] = 1
	o++
	for _, v := range []uint64{10, 9, 10, 8} {
		binary.LittleEndian.PutUint64(buf[o:o+8], v)
		o += 8
	}
	binary.LittleEndian.PutUint64(buf[o:o+8], 0)
	binary.LittleEndian.PutUint64(buf[o+8:o+16], 1)
	o += 16
	for _, v := range []uint64{1, 2, 3, 10, 123} {
		binary.LittleEndian.PutUint64(buf[o:o+8], v)
		o += 8
	}
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func cpmmCreatePoolLogForTest() string {
	buf := make([]byte, 8+32*4+8+8)
	o := 0
	binary.LittleEndian.PutUint64(buf[o:o+8], discCpmmCreatePool)
	o += 8
	for seed := byte(10); seed < 14; seed++ {
		for i := 0; i < 32; i++ {
			buf[o+i] = seed + byte(i)
		}
		o += 32
	}
	binary.LittleEndian.PutUint64(buf[o:o+8], 1000)
	o += 8
	binary.LittleEndian.PutUint64(buf[o:o+8], 2000)
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func dammAddLiquidityLogForTest() string {
	buf := make([]byte, 8+32*3+16+8*6)
	o := 0
	binary.LittleEndian.PutUint64(buf[o:o+8], discDammAdd)
	o += 8
	for seed := byte(20); seed < 23; seed++ {
		for i := 0; i < 32; i++ {
			buf[o+i] = seed + byte(i)
		}
		o += 32
	}
	binary.LittleEndian.PutUint64(buf[o:o+8], 123)
	o += 16
	for _, v := range []uint64{1, 2, 3, 4, 5, 6} {
		binary.LittleEndian.PutUint64(buf[o:o+8], v)
		o += 8
	}
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func TestParseLogOptimizedUsesProgramContextForMeteoraDbc(t *testing.T) {
	log := dbcSwapLogForTest()
	filter := EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDbcSwap})
	if ev := ParseLogOptimized(log, "sig", 1, 0, nil, 1, filter, false, ""); ev.Type != "" {
		t.Fatalf("unscoped DBC shared discriminator should be dropped, got %q", ev.Type)
	}
	ev := ParseLogOptimizedWithProgramID(log, "sig", 1, 0, nil, 1, filter, false, "", METEORA_DBC_PROGRAM_ID)
	if ev.Type != EventTypeMeteoraDbcSwap {
		t.Fatalf("expected scoped DBC swap, got %q", ev.Type)
	}
	swap := ev.Data.(*MeteoraDbcSwapEvent)
	if swap.OutputAmount != 8 || swap.CurrentTimestamp != 123 {
		t.Fatalf("unexpected DBC swap payload: %+v", swap)
	}
}

func TestParseLogOptimizedRoutesScopedCpmmCreatePoolWithoutClmmLeak(t *testing.T) {
	log := cpmmCreatePoolLogForTest()
	if ev := ParseLogOptimized(log, "sig", 1, 0, nil, 1, EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumCpmmInitialize}), false, ""); ev.Type != "" {
		t.Fatalf("unscoped CPMM create pool should be dropped, got %q", ev.Type)
	}

	cpmm := ParseLogOptimizedWithProgramID(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumCpmmInitialize}),
		false,
		"",
		RAYDIUM_CPMM_PROGRAM_ID,
	)
	if cpmm.Type != EventTypeRaydiumCpmmInitialize {
		t.Fatalf("expected scoped CPMM initialize, got %q", cpmm.Type)
	}
	if data, ok := cpmm.Data.(*RaydiumCpmmInitializeEvent); !ok || data.InitAmount0 != 1000 {
		t.Fatalf("unexpected scoped CPMM initialize payload: %#v", cpmm.Data)
	}

	clmm := ParseLogOptimizedWithProgramID(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumClmmCreatePool}),
		false,
		"",
		RAYDIUM_CLMM_PROGRAM_ID,
	)
	if clmm.Type != "" {
		t.Fatalf("CLMM scope must not parse CPMM create pool, got %q", clmm.Type)
	}
}

func TestParseLogOptimizedParsesScopedDammNonSwapProgramData(t *testing.T) {
	ev := ParseLogOptimizedWithProgramID(
		dammAddLiquidityLogForTest(),
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDammV2AddLiquidity}),
		false,
		"",
		METEORA_DAMM_V2_PROGRAM_ID,
	)
	if ev.Type != EventTypeMeteoraDammV2AddLiquidity {
		t.Fatalf("expected DAMM add liquidity, got %q", ev.Type)
	}
	data, ok := ev.Data.(*MeteoraDammV2AddLiquidityEvent)
	if !ok || data.LiquidityDelta != "123" || data.TokenBAmount != 4 {
		t.Fatalf("unexpected DAMM add liquidity payload: %#v", ev.Data)
	}
}

func TestProtocolHelperExcludeMatchesRustSemantics(t *testing.T) {
	filter := EventTypeFilterExclude([]EventType{EventTypePumpFeesUpdateAdmin})
	if !EventTypeFilterIncludesPumpFees(filter) {
		t.Fatalf("PumpFees route should stay enabled for exclude filters")
	}
	if !EventTypeFilterIncludesRaydiumCpmm(EventTypeFilterExclude([]EventType{EventTypeRaydiumCpmmSwap})) {
		t.Fatalf("partial CPMM excludes should keep CPMM route enabled")
	}
	if EventTypeFilterIncludesRaydiumCpmm(EventTypeFilterExclude([]EventType{
		EventTypeRaydiumCpmmSwap,
		EventTypeRaydiumCpmmDeposit,
		EventTypeRaydiumCpmmWithdraw,
		EventTypeRaydiumCpmmInitialize,
	})) {
		t.Fatalf("excluding every CPMM event should skip CPMM route")
	}
	if EventTypeFilterIncludesRaydiumLaunchlab(EventTypeFilterExclude([]EventType{
		EventTypeRaydiumLaunchlabTrade,
		EventTypeRaydiumLaunchlabPoolCreate,
		EventTypeRaydiumLaunchlabMigrateAmm,
	})) {
		t.Fatalf("excluding every LaunchLab event should skip LaunchLab route")
	}
	if EventTypeFilterIncludesPumpfun(EventTypeFilterIncludeOnly([]EventType{EventTypeAccountPumpFunGlobal})) {
		t.Fatalf("PumpFun account-only filters should not enable instruction routes")
	}
	if EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeAccountPumpFunGlobal}) {
		t.Fatalf("PumpFun account-only filters should not pass instruction prefilter")
	}
	if EventTypeFilterIncludesRaydiumClmm(EventTypeFilterIncludeOnly([]EventType{EventTypeAccountRaydiumClmmPoolState})) {
		t.Fatalf("Raydium CLMM account-only filters should not enable instruction routes")
	}
	if EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeAccountRaydiumClmmPoolState}) {
		t.Fatalf("Raydium CLMM account-only filters should not pass instruction prefilter")
	}
	if EventTypeFilterIncludesPumpfun(EventTypeFilterIncludeOnly([]EventType{EventTypePumpFeesUpdateAdmin})) {
		t.Fatalf("PumpFees filters should not enable PumpFun route")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypePumpFeesUpdateAdmin}) {
		t.Fatalf("PumpFees filters should pass instruction prefilter")
	}
	if !EventTypeFilterIncludesPumpswap(EventTypeFilterIncludeOnly([]EventType{EventTypePumpSwapTrade})) {
		t.Fatalf("PumpSwapTrade should enable PumpSwap route")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypePumpSwapTrade}) {
		t.Fatalf("PumpSwapTrade should pass instruction prefilter")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeMeteoraDammV2InitializePool}) {
		t.Fatalf("Meteora DAMM initialize pool should pass instruction prefilter")
	}
	if !EventTypeFilterIncludesMeteoraPools(EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraPoolsSwap})) {
		t.Fatalf("Meteora Pools swap should enable Meteora Pools route helper")
	}
	if EventTypeFilterIncludesMeteoraPools(EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunTrade})) {
		t.Fatalf("PumpFun filters should not enable Meteora Pools route helper")
	}
	if !EventTypeFilterIncludesMeteoraDlmm(EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDlmmSwap})) {
		t.Fatalf("Meteora DLMM swap should enable Meteora DLMM route helper")
	}
	if EventTypeFilterIncludesMeteoraDlmm(EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunTrade})) {
		t.Fatalf("PumpFun filters should not enable Meteora DLMM route helper")
	}
	if !EventTypeFilterIncludesRaydiumAmmV4(EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumAmmV4Deposit})) {
		t.Fatalf("Raydium AMM V4 deposit should enable AMM V4 route")
	}
	if !EventTypeFilterIncludesRaydiumClmm(EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumClmmOpenLimitOrder})) {
		t.Fatalf("Raydium CLMM open limit order should enable CLMM route")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeRaydiumClmmOpenLimitOrder}) {
		t.Fatalf("Raydium CLMM open limit order should pass instruction prefilter")
	}
	if !EventTypeFilterIncludeOnly([]EventType{EventTypePumpSwapTrade}).ShouldInclude(EventTypePumpSwapBuy) {
		t.Fatalf("PumpSwapTrade should include concrete buy events")
	}
	if EventTypeFilterExclude([]EventType{EventTypePumpSwapTrade}).ShouldInclude(EventTypePumpSwapSell) {
		t.Fatalf("excluding PumpSwapTrade should drop concrete sell events")
	}
}

func TestAllEventTypesMatchesRustInventory(t *testing.T) {
	want := []EventType{
		EventTypeBlockMeta,
		EventTypeRaydiumLaunchlabTrade,
		EventTypeRaydiumLaunchlabPoolCreate,
		EventTypeRaydiumLaunchlabMigrateAmm,
		EventTypePumpFunTrade,
		EventTypePumpFunBuy,
		EventTypePumpFunSell,
		EventTypePumpFunBuyExactSolIn,
		EventTypePumpFunCreate,
		EventTypePumpFunCreateV2,
		EventTypePumpFunComplete,
		EventTypePumpFunMigrate,
		EventTypePumpFeesCreateFeeSharingConfig,
		EventTypePumpFeesInitializeFeeConfig,
		EventTypePumpFeesResetFeeSharingConfig,
		EventTypePumpFeesRevokeFeeSharingAuthority,
		EventTypePumpFeesTransferFeeSharingAuthority,
		EventTypePumpFeesUpdateAdmin,
		EventTypePumpFeesUpdateFeeConfig,
		EventTypePumpFeesUpdateFeeShares,
		EventTypePumpFeesUpsertFeeTiers,
		EventTypePumpFunMigrateBondingCurveCreator,
		EventTypePumpSwapTrade,
		EventTypePumpSwapBuy,
		EventTypePumpSwapSell,
		EventTypePumpSwapCreatePool,
		EventTypePumpSwapLiquidityAdded,
		EventTypePumpSwapLiquidityRemoved,
		EventTypeRaydiumCpmmSwap,
		EventTypeRaydiumCpmmDeposit,
		EventTypeRaydiumCpmmWithdraw,
		EventTypeRaydiumCpmmInitialize,
		EventTypeRaydiumClmmSwap,
		EventTypeRaydiumClmmCreatePool,
		EventTypeRaydiumClmmOpenPosition,
		EventTypeRaydiumClmmClosePosition,
		EventTypeRaydiumClmmIncreaseLiquidity,
		EventTypeRaydiumClmmDecreaseLiquidity,
		EventTypeRaydiumClmmLiquidityChange,
		EventTypeRaydiumClmmConfigChange,
		EventTypeRaydiumClmmCreatePersonalPosition,
		EventTypeRaydiumClmmLiquidityCalculate,
		EventTypeRaydiumClmmOpenLimitOrder,
		EventTypeRaydiumClmmIncreaseLimitOrder,
		EventTypeRaydiumClmmDecreaseLimitOrder,
		EventTypeRaydiumClmmSettleLimitOrder,
		EventTypeRaydiumClmmUpdateRewardInfos,
		EventTypeRaydiumClmmOpenPositionWithTokenExtNft,
		EventTypeRaydiumClmmCollectFee,
		EventTypeRaydiumAmmV4Swap,
		EventTypeRaydiumAmmV4Deposit,
		EventTypeRaydiumAmmV4Withdraw,
		EventTypeRaydiumAmmV4Initialize2,
		EventTypeRaydiumAmmV4WithdrawPnl,
		EventTypeOrcaWhirlpoolSwap,
		EventTypeOrcaWhirlpoolLiquidityIncreased,
		EventTypeOrcaWhirlpoolLiquidityDecreased,
		EventTypeOrcaWhirlpoolPoolInitialized,
		EventTypeMeteoraPoolsSwap,
		EventTypeMeteoraPoolsAddLiquidity,
		EventTypeMeteoraPoolsRemoveLiquidity,
		EventTypeMeteoraPoolsBootstrapLiquidity,
		EventTypeMeteoraPoolsPoolCreated,
		EventTypeMeteoraPoolsSetPoolFees,
		EventTypeMeteoraDammV2Swap,
		EventTypeMeteoraDammV2AddLiquidity,
		EventTypeMeteoraDammV2RemoveLiquidity,
		EventTypeMeteoraDammV2InitializePool,
		EventTypeMeteoraDammV2CreatePosition,
		EventTypeMeteoraDammV2ClosePosition,
		EventTypeMeteoraDbcSwap,
		EventTypeMeteoraDbcInitializePool,
		EventTypeMeteoraDbcCurveComplete,
		EventTypeMeteoraDlmmSwap,
		EventTypeMeteoraDlmmAddLiquidity,
		EventTypeMeteoraDlmmRemoveLiquidity,
		EventTypeMeteoraDlmmInitializePool,
		EventTypeMeteoraDlmmInitializeBinArray,
		EventTypeMeteoraDlmmCreatePosition,
		EventTypeMeteoraDlmmClosePosition,
		EventTypeMeteoraDlmmClaimFee,
		EventTypeTokenAccount,
		EventTypeTokenInfo,
		EventTypeNonceAccount,
		EventTypeAccountPumpFunGlobal,
		EventTypeAccountPumpFunBondingCurve,
		EventTypeAccountPumpFunFeeConfig,
		EventTypeAccountPumpFunSharingConfig,
		EventTypeAccountPumpFunGlobalVolumeAccumulator,
		EventTypeAccountPumpFunUserVolumeAccumulator,
		EventTypeAccountPumpSwapGlobalConfig,
		EventTypeAccountPumpSwapPool,
		EventTypeAccountRaydiumClmmAmmConfig,
		EventTypeAccountRaydiumClmmPoolState,
		EventTypeAccountRaydiumClmmTickArrayState,
		EventTypeAccountRaydiumCpmmAmmConfig,
		EventTypeAccountRaydiumCpmmPoolState,
		EventTypeAccountOrcaWhirlpool,
		EventTypeAccountOrcaPosition,
		EventTypeAccountOrcaTickArray,
		EventTypeAccountOrcaFeeTier,
		EventTypeAccountOrcaWhirlpoolsConfig,
	}
	got := AllEventTypes()
	if len(got) != len(want) {
		t.Fatalf("event type count mismatch: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event type %d mismatch: got %q want %q", i, got[i], want[i])
		}
	}
}

func clmmOpenLimitOrderLogForTest() string {
	buf := make([]byte, 8+32+32+1+4+8+8)
	o := 0
	binary.LittleEndian.PutUint64(buf[o:o+8], discClmmOpenLimitOrder)
	o += 8
	for i := 0; i < 32; i++ {
		buf[o+i] = 1
		buf[o+32+i] = 2
	}
	o += 64
	buf[o] = 1
	o++
	binary.LittleEndian.PutUint32(buf[o:o+4], 0xffffff85)
	o += 4
	binary.LittleEndian.PutUint64(buf[o:o+8], 456)
	o += 8
	binary.LittleEndian.PutUint64(buf[o:o+8], 7)
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func pkForTest(seed byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = seed + byte(i)
	}
	return out
}

func pkStringForTest(seed byte) string {
	return Base58Encode(pkForTest(seed))
}

func u128ForTest(bit uint, lo uint64) [16]byte {
	var out [16]byte
	if lo != 0 {
		binary.LittleEndian.PutUint64(out[:8], lo)
	}
	if bit >= 64 {
		binary.LittleEndian.PutUint64(out[8:], uint64(1)<<(bit-64))
	} else if bit > 0 {
		binary.LittleEndian.PutUint64(out[:8], binary.LittleEndian.Uint64(out[:8])|(uint64(1)<<bit))
	}
	return out
}

func appendPkForTest(buf []byte, seed byte) []byte {
	return append(buf, pkForTest(seed)...)
}

func appendU128ForTest(buf []byte, value [16]byte) []byte {
	return append(buf, value[:]...)
}

func TestParseRaydiumClmmBaseProgramDataLayouts(t *testing.T) {
	swap := make([]byte, 0, 8+32*4+8*4+1+16+16+4)
	var disc [8]byte
	binary.LittleEndian.PutUint64(disc[:], discClmmSwap)
	swap = append(swap, disc[:]...)
	for _, seed := range []byte{1, 2, 3, 4} {
		swap = appendPkForTest(swap, seed)
	}
	for _, value := range []uint64{10, 1, 20, 2} {
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], value)
		swap = append(swap, b[:]...)
	}
	swap = append(swap, 1)
	swap = appendU128ForTest(swap, u128ForTest(80, 30))
	swap = appendU128ForTest(swap, u128ForTest(96, 40))
	var i32 [4]byte
	binary.LittleEndian.PutUint32(i32[:], uint32(^uint32(76)))
	swap = append(swap, i32[:]...)

	ev := ParseLogOptimized("Program data: "+base64.StdEncoding.EncodeToString(swap), "sig", 1, 0, nil, 1, nil, false, "")
	if ev.Type != EventTypeRaydiumClmmSwap {
		t.Fatalf("expected CLMM swap, got %q", ev.Type)
	}
	s := ev.Data.(*RaydiumClmmSwapEvent)
	if s.PoolState != pkStringForTest(1) || s.Sender != pkStringForTest(2) ||
		s.TokenAccount0 != pkStringForTest(3) || s.TokenAccount1 != pkStringForTest(4) {
		t.Fatalf("unexpected swap pubkeys: %+v", s)
	}
	if s.Amount0 != 10 || s.TransferFee0 != 1 || s.Amount1 != 20 || s.TransferFee1 != 2 ||
		!s.ZeroForOne || s.SqrtPriceX64 != u128LEDecimalString(u128ForTest(80, 30)) ||
		s.Liquidity != u128LEDecimalString(u128ForTest(96, 40)) || s.Tick != -77 {
		t.Fatalf("unexpected swap payload: %+v", s)
	}

	create := make([]byte, 0, 8+32+32+2+32+16+4+32+32)
	binary.LittleEndian.PutUint64(disc[:], discClmmCreate)
	create = append(create, disc[:]...)
	create = appendPkForTest(create, 5)
	create = appendPkForTest(create, 6)
	var u16 [2]byte
	binary.LittleEndian.PutUint16(u16[:], 64)
	create = append(create, u16[:]...)
	create = appendPkForTest(create, 7)
	create = appendU128ForTest(create, u128ForTest(72, 55))
	binary.LittleEndian.PutUint32(i32[:], uint32(int32(88)))
	create = append(create, i32[:]...)
	create = appendPkForTest(create, 8)
	create = appendPkForTest(create, 9)

	ev = ParseLogOptimized("Program data: "+base64.StdEncoding.EncodeToString(create), "sig", 1, 0, nil, 1, nil, false, "")
	if ev.Type != EventTypeRaydiumClmmCreatePool {
		t.Fatalf("expected CLMM create pool, got %q", ev.Type)
	}
	c := ev.Data.(*RaydiumClmmCreatePoolEvent)
	if c.Token0Mint != pkStringForTest(5) || c.Token1Mint != pkStringForTest(6) ||
		c.TickSpacing != 64 || c.Pool != pkStringForTest(7) ||
		c.SqrtPriceX64 != u128LEDecimalString(u128ForTest(72, 55)) ||
		c.Tick != 88 || c.TokenVault0 != pkStringForTest(8) || c.TokenVault1 != pkStringForTest(9) {
		t.Fatalf("unexpected create payload: %+v", c)
	}

	oldInstructionDisc := []byte{248, 198, 158, 145, 225, 117, 135, 200}
	if dropped := ParseLogOptimized("Program data: "+base64.StdEncoding.EncodeToString(append(oldInstructionDisc, pkForTest(1)...)), "sig", 1, 0, nil, 1, nil, false, ""); dropped.Type != "" {
		t.Fatalf("old instruction discriminator should not parse as CLMM log, got %q", dropped.Type)
	}
}

func TestParseRaydiumClmmCollectFeeLogs(t *testing.T) {
	var disc [8]byte
	var amount [8]byte
	personal := make([]byte, 0, 8+32+32+32+8+8)
	binary.LittleEndian.PutUint64(disc[:], discClmmCollectPersonal)
	personal = append(personal, disc[:]...)
	personal = appendPkForTest(personal, 10)
	personal = appendPkForTest(personal, 11)
	personal = appendPkForTest(personal, 12)
	binary.LittleEndian.PutUint64(amount[:], 70)
	personal = append(personal, amount[:]...)
	binary.LittleEndian.PutUint64(amount[:], 80)
	personal = append(personal, amount[:]...)

	ev := ParseLogOptimized("Program data: "+base64.StdEncoding.EncodeToString(personal), "sig", 1, 0, nil, 1, EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumClmmCollectFee}), false, "")
	if ev.Type != EventTypeRaydiumClmmCollectFee {
		t.Fatalf("expected collect fee, got %q", ev.Type)
	}
	p := ev.Data.(*RaydiumClmmCollectFeeEvent)
	if p.PositionNftMint != pkStringForTest(10) || p.RecipientTokenAccount0 != pkStringForTest(11) ||
		p.RecipientTokenAccount1 != pkStringForTest(12) || p.Amount0 != 70 || p.Amount1 != 80 {
		t.Fatalf("unexpected personal collect: %+v", p)
	}

	protocol := make([]byte, 0, 8+32+32+32+8+8)
	binary.LittleEndian.PutUint64(disc[:], discClmmCollectProtocol)
	protocol = append(protocol, disc[:]...)
	protocol = appendPkForTest(protocol, 13)
	protocol = appendPkForTest(protocol, 14)
	protocol = appendPkForTest(protocol, 15)
	binary.LittleEndian.PutUint64(amount[:], 90)
	protocol = append(protocol, amount[:]...)
	binary.LittleEndian.PutUint64(amount[:], 100)
	protocol = append(protocol, amount[:]...)

	ev = ParseLogOptimized("Program data: "+base64.StdEncoding.EncodeToString(protocol), "sig", 1, 0, nil, 1, EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumClmmCollectFee}), false, "")
	if ev.Type != EventTypeRaydiumClmmCollectFee {
		t.Fatalf("expected collect fee, got %q", ev.Type)
	}
	pr := ev.Data.(*RaydiumClmmCollectFeeEvent)
	if pr.PoolState != pkStringForTest(13) || pr.RecipientTokenAccount0 != pkStringForTest(14) ||
		pr.RecipientTokenAccount1 != pkStringForTest(15) || pr.Amount0 != 90 || pr.Amount1 != 100 {
		t.Fatalf("unexpected protocol collect: %+v", pr)
	}
}

func TestParseRaydiumClmmAdvancedLogEventAndFilter(t *testing.T) {
	log := clmmOpenLimitOrderLogForTest()
	ev := ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumClmmOpenLimitOrder}),
		false,
		"",
	)
	if ev.Type != EventTypeRaydiumClmmOpenLimitOrder {
		t.Fatalf("expected OpenLimitOrder, got %q", ev.Type)
	}
	data := ev.Data.(*RaydiumClmmOpenLimitOrderEvent)
	if data.ZeroForOne != true || data.TickIndex != -123 || data.TotalAmount != 456 || data.TransferFee != 7 {
		t.Fatalf("unexpected open limit order payload: %+v", data)
	}
	if dropped := ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterExclude([]EventType{EventTypeRaydiumClmmOpenLimitOrder}),
		false,
		"",
	); dropped.Type != "" {
		t.Fatalf("expected exclude filter to drop open limit order, got %q", dropped.Type)
	}
}

func TestSubscribeRequestBuilderIncludesAccountFilters(t *testing.T) {
	client := NewYellowstoneGrpc("127.0.0.1:10000")
	txFilter := TransactionFilterForProtocols([]Protocol{ProtocolPumpFun})
	accFilter := AccountFilterForProtocols([]Protocol{ProtocolPumpSwap})
	accFilter.Filters = append(accFilter.Filters, AccountFilterMemcmp(32, []byte{1, 2, 3}))

	req := client.buildSubscribeRequestMulti([]TransactionFilter{txFilter}, []AccountFilter{accFilter})
	if req.Commitment == nil || *req.Commitment != pb.CommitmentLevel_PROCESSED {
		t.Fatalf("expected processed commitment, got %v", req.Commitment)
	}
	if got := req.Transactions["tx_0"].AccountInclude; len(got) != 1 || got[0] != PUMPFUN_PROGRAM_ID {
		t.Fatalf("unexpected transaction account_include: %#v", got)
	}

	accountReq := req.Accounts["acc_0"]
	if accountReq == nil {
		t.Fatalf("expected account filter acc_0")
	}
	if len(accountReq.Owner) == 0 || accountReq.Owner[0] != PUMPSWAP_PROGRAM_ID {
		t.Fatalf("unexpected account owners: %#v", accountReq.Owner)
	}
	if len(accountReq.Filters) != 1 {
		t.Fatalf("expected one account memcmp filter, got %d", len(accountReq.Filters))
	}
	memcmp := accountReq.Filters[0].GetMemcmp()
	if memcmp == nil || memcmp.Offset != 32 || !bytes.Equal(memcmp.GetBytes(), []byte{1, 2, 3}) {
		t.Fatalf("unexpected memcmp filter: %#v", memcmp)
	}
}
