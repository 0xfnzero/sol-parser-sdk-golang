package solparser

import (
	"bytes"
	"encoding/binary"
)

// 外层**指令** discriminator（Rust `instr/pump.rs` / `pump_amm.rs`）。Program log 里的 Buy/Sell 等 Event disc 仍见 `binary.go` / `matcher.go`，二者不可混用。
var (
	instrPumpOuterCreate            = disc8(24, 30, 200, 40, 5, 28, 7, 119)
	instrPumpOuterCreateV2          = disc8(214, 144, 76, 236, 95, 139, 49, 180)
	instrPumpOuterBuy               = disc8(102, 6, 61, 18, 1, 218, 235, 234)
	instrPumpOuterSell              = disc8(51, 230, 133, 164, 1, 127, 131, 173)
	instrPumpOuterBuyExactSolIn     = disc8(56, 252, 116, 8, 158, 223, 205, 95)
	instrPumpOuterBuyV2             = disc8(184, 23, 238, 97, 103, 197, 211, 61)
	instrPumpOuterSellV2            = disc8(93, 246, 130, 60, 231, 233, 64, 178)
	instrPumpOuterBuyExactQuoteInV2 = disc8(194, 171, 28, 70, 104, 77, 91, 47)
	instrPumpMigrateCPI             = disc8(189, 233, 93, 185, 92, 148, 234, 148)
)

var (
	instrPumpSwapBuy           = disc8(102, 6, 61, 18, 1, 218, 235, 234)
	instrPumpSwapSell          = disc8(51, 230, 133, 164, 1, 127, 131, 173)
	instrPumpSwapCreatePool    = disc8(233, 146, 209, 142, 207, 104, 64, 188)
	instrPumpSwapBuyExactQuote = disc8(198, 46, 21, 82, 180, 217, 232, 112)
	instrPumpSwapDeposit       = disc8(242, 35, 198, 137, 82, 225, 242, 182)
	instrPumpSwapWithdraw      = disc8(183, 18, 70, 156, 148, 109, 161, 34)
)

var (
	instrRaydiumLaunchlabBuyExactIn          = disc8(250, 234, 13, 123, 213, 156, 19, 236)
	instrRaydiumLaunchlabBuyExactOut         = disc8(24, 211, 116, 40, 105, 3, 153, 56)
	instrRaydiumLaunchlabInitialize          = disc8(175, 175, 109, 31, 13, 152, 155, 237)
	instrRaydiumLaunchlabInitializeV2        = disc8(67, 153, 175, 39, 218, 16, 38, 32)
	instrRaydiumLaunchlabInitializeToken2022 = disc8(37, 190, 126, 222, 44, 154, 171, 17)
	instrRaydiumLaunchlabMigrateToAmm        = disc8(207, 82, 192, 145, 254, 207, 145, 223)
	instrRaydiumLaunchlabMigrateToCpswap     = disc8(136, 92, 200, 103, 28, 218, 144, 140)
	instrRaydiumLaunchlabSellExactIn         = disc8(149, 39, 222, 155, 211, 124, 152, 26)
	instrRaydiumLaunchlabSellExactOut        = disc8(95, 200, 71, 34, 8, 9, 11, 166)
)

var (
	instrClmmSwap                       = disc8(248, 198, 158, 145, 225, 117, 135, 200)
	instrClmmSwapV2                     = disc8(43, 4, 237, 11, 26, 201, 30, 98)
	instrClmmIncLiqV2                   = disc8(133, 29, 89, 223, 69, 238, 176, 10)
	instrClmmDecLiqV2                   = disc8(58, 127, 188, 62, 79, 82, 196, 96)
	instrClmmCreatePool                 = disc8(233, 146, 209, 142, 207, 104, 64, 188)
	instrClmmCreateCustomizablePool     = disc8(43, 68, 212, 167, 89, 47, 164, 1)
	instrClmmOpenPosition               = disc8(135, 128, 47, 77, 15, 152, 240, 49)
	instrClmmOpenPositionV2             = disc8(77, 184, 74, 214, 112, 86, 241, 199)
	instrClmmOpenPositionWithToken22Nft = disc8(77, 255, 174, 82, 125, 29, 201, 46)
	instrClmmClosePosition              = disc8(123, 134, 81, 0, 49, 68, 98, 98)
)

var (
	instrCpmmInitialize = disc8(175, 175, 109, 31, 13, 152, 155, 237)
)

// InstructionData 指令数据
type InstructionData struct {
	ProgramIDIndex uint32
	Accounts       []uint32
	Data           []byte
}

// parseInstructionUnifiedPreFilterRust 执行统一入口的 include_only 快速预检：
// 若 `EventTypeFilter` 为 `IncludeOnlyFilter` 且 `include_only` 非空，且其中**无一**与
// `EventTypeFilterAllowsInstructionParsing` 所列类型相交，则整条入口不解析（返回空）。
// `IncludeOnlyFilter` 且 `IncludeOnly` 长度为 0 时与 Rust `Some([])` 一致：不允许解析。
func parseInstructionUnifiedPreFilterRust(filter EventTypeFilter) bool {
	if filter == nil {
		return false
	}
	only, ok := filter.(*IncludeOnlyFilter)
	if !ok {
		return false
	}
	if len(only.IncludeOnly) == 0 {
		return true
	}
	return !EventTypeFilterAllowsInstructionParsing(only.IncludeOnly)
}

// ParseInstructionUnified 统一的指令解析入口
// 覆盖本包已有的外层指令解析器：PumpFun、PumpSwap、Pump Fees、Meteora DAMM V2、Raydium、Orca、RaydiumLaunchlab。
func ParseInstructionUnified(
	instructionData []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
	filter EventTypeFilter,
	programID string,
) DexEvent {
	if len(instructionData) == 0 {
		return DexEvent{}
	}
	if parseInstructionUnifiedPreFilterRust(filter) {
		return DexEvent{}
	}

	switch programID {
	case PUMPFUN_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpfun(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParsePumpfunInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case PUMPSWAP_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpswap(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParsePumpswapInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case METEORA_DAMM_V2_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraDammV2(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseMeteoraDammInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case METEORA_POOLS_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraPools(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseMeteoraPoolsInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case METEORA_DLMM_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraDlmm(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseMeteoraDlmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case PUMP_FEES_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpFees(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParsePumpFeesInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case RAYDIUM_CLMM_PROGRAM_ID, GrpcRaydiumClmmProgramID:
		if filter != nil && !EventTypeFilterIncludesRaydiumClmm(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseRaydiumClmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case RAYDIUM_CPMM_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumCpmm(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseRaydiumCpmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case RAYDIUM_AMM_V4_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumAmmV4(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseRaydiumAmmV4Instruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case ORCA_WHIRLPOOL_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesOrcaWhirlpool(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseOrcaWhirlpoolInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)

	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumLaunchlab(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseRaydiumLaunchlabInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs), filter)
	}

	return DexEvent{}
}

// ParsePumpFeesInstruction 对齐 Rust `pump_fees::parse_instruction`（pfeeUx...）。
func ParsePumpFeesInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	return parsePumpFeesInstruction(data, accounts, meta)
}

// PumpFun / PumpSwap **inner** CPI 事件：16 字节 discriminator（与 Rust `pump_inner.rs` / `pump_amm_inner.rs` 一致）。
var (
	pumpfunInnerTradeEvent      = []byte{189, 219, 127, 211, 78, 230, 97, 238, 155, 167, 108, 32, 122, 76, 173, 64}
	pumpfunInnerCreateToken     = []byte{27, 114, 169, 77, 222, 235, 99, 118, 155, 167, 108, 32, 122, 76, 173, 64}
	pumpfunInnerMigrateComplete = []byte{189, 233, 93, 185, 92, 148, 234, 148, 155, 167, 108, 32, 122, 76, 173, 64}

	pumpswapInnerBuy             = []byte{228, 69, 165, 46, 81, 203, 154, 29, 103, 244, 82, 31, 44, 245, 119, 119}
	pumpswapInnerSell            = []byte{228, 69, 165, 46, 81, 203, 154, 29, 62, 47, 55, 10, 165, 3, 220, 42}
	pumpswapInnerCreatePool      = []byte{228, 69, 165, 46, 81, 203, 154, 29, 177, 49, 12, 210, 160, 118, 167, 116}
	pumpswapInnerAddLiquidity    = []byte{228, 69, 165, 46, 81, 203, 154, 29, 120, 248, 61, 83, 31, 142, 107, 144}
	pumpswapInnerRemoveLiquidity = []byte{228, 69, 165, 46, 81, 203, 154, 29, 22, 9, 133, 26, 160, 44, 71, 192}
)

var (
	eventCPIPrefix = []byte{228, 69, 165, 46, 81, 203, 154, 29}
	eventCPISuffix = []byte{155, 167, 108, 32, 122, 76, 173, 64}
)

func disc16HasPrefix(disc []byte) bool {
	return len(disc) >= 16 && bytes.Equal(disc[:8], eventCPIPrefix)
}

func disc16HasSuffix(disc []byte) bool {
	return len(disc) >= 16 && bytes.Equal(disc[8:16], eventCPISuffix)
}

func eventCPIDisc8(disc []byte) (uint64, bool) {
	if len(disc) < 16 {
		return 0, false
	}
	if disc16HasPrefix(disc) {
		return binary.LittleEndian.Uint64(disc[8:16]), true
	}
	if disc16HasSuffix(disc) {
		return binary.LittleEndian.Uint64(disc[:8]), true
	}
	return 0, false
}

func headInDiscs(data []byte, discs ...uint64) bool {
	if len(data) < 8 {
		return false
	}
	disc := binary.LittleEndian.Uint64(data[:8])
	for _, want := range discs {
		if disc == want {
			return true
		}
	}
	return false
}

func firstByteIn(data []byte, allowed ...byte) bool {
	if len(data) == 0 {
		return false
	}
	for _, want := range allowed {
		if data[0] == want {
			return true
		}
	}
	return false
}

func normalInstructionDataMayParse(programID string, data []byte) bool {
	if len(data) == 0 {
		return false
	}
	switch programID {
	case RAYDIUM_AMM_V4_PROGRAM_ID:
		return firstByteIn(data, 1, 3, 4, 7, 9, 11)
	case METEORA_DLMM_PROGRAM_ID:
		return firstByteIn(data, 0, 1, 2, 7, 8, 11, 13, 14)
	case METEORA_DAMM_V2_PROGRAM_ID:
		return headInDiscs(data, discDammInit)
	case PUMPFUN_PROGRAM_ID:
		return headInDiscs(data,
			instrPumpOuterCreate,
			instrPumpOuterCreateV2,
			instrPumpOuterBuy,
			instrPumpOuterSell,
			instrPumpOuterBuyExactSolIn,
			instrPumpOuterBuyV2,
			instrPumpOuterBuyExactQuoteInV2,
			instrPumpOuterSellV2,
		)
	case PUMPSWAP_PROGRAM_ID:
		return headInDiscs(data,
			instrPumpSwapBuy,
			instrPumpSwapSell,
			instrPumpSwapCreatePool,
			instrPumpSwapBuyExactQuote,
			instrPumpSwapDeposit,
			instrPumpSwapWithdraw,
		)
	case PUMP_FEES_PROGRAM_ID:
		return headInDiscs(data,
			instrPumpFeesCreateFeeSharingConfig,
			instrPumpFeesInitializeFeeConfig,
			instrPumpFeesResetFeeSharingConfig,
			instrPumpFeesResetFeeSharingConfigV2,
			instrPumpFeesRevokeFeeSharingAuthority,
			instrPumpFeesTransferFeeSharingAuthority,
			instrPumpFeesUpdateAdmin,
			instrPumpFeesUpdateFeeConfig,
			instrPumpFeesUpdateFeeShares,
			instrPumpFeesUpdateFeeSharesV2,
			instrPumpFeesUpsertFeeTiers,
		)
	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		return headInDiscs(data,
			instrRaydiumLaunchlabBuyExactIn,
			instrRaydiumLaunchlabBuyExactOut,
			instrRaydiumLaunchlabSellExactIn,
			instrRaydiumLaunchlabSellExactOut,
			instrRaydiumLaunchlabInitialize,
			instrRaydiumLaunchlabInitializeV2,
			instrRaydiumLaunchlabInitializeToken2022,
		)
	case RAYDIUM_CPMM_PROGRAM_ID:
		return headInDiscs(data,
			discCpmmSwapIn,
			discCpmmSwapOut,
			instrCpmmInitialize,
			discCpmmDeposit,
			discCpmmWithdraw,
		)
	case RAYDIUM_CLMM_PROGRAM_ID, GrpcRaydiumClmmProgramID:
		return headInDiscs(data,
			instrClmmSwap,
			instrClmmSwapV2,
			instrClmmIncLiqV2,
			instrClmmDecLiqV2,
			instrClmmCreatePool,
			instrClmmCreateCustomizablePool,
			instrClmmOpenPosition,
			instrClmmOpenPositionV2,
			instrClmmOpenPositionWithToken22Nft,
			instrClmmClosePosition,
		)
	case ORCA_WHIRLPOOL_PROGRAM_ID:
		return headInDiscs(data,
			disc8(248, 198, 158, 145, 225, 117, 135, 200),
			disc8(43, 4, 237, 11, 26, 201, 30, 98),
			disc8(46, 156, 243, 118, 13, 205, 251, 178),
			disc8(160, 38, 208, 111, 104, 91, 44, 1),
			disc8(17, 43, 80, 74, 168, 202, 6, 113),
		)
	case METEORA_POOLS_PROGRAM_ID:
		return headInDiscs(data,
			instrMeteoraPoolsSwap,
			instrMeteoraPoolsAddLiquidity,
			instrMeteoraPoolsRemoveLiquidity,
			instrMeteoraPoolsCreatePool,
		)
	default:
		return false
	}
}

// ParseInnerCompiledInstructionIfSupported 尝试按 Rust normal-inner fallback 解析普通编译指令。
func ParseInnerCompiledInstructionIfSupported(
	instructionData []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
	filter EventTypeFilter,
	programID string,
) DexEvent {
	if !normalInstructionDataMayParse(programID, instructionData) {
		return DexEvent{}
	}
	return ParseInstructionUnified(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, programID)
}

func parsePumpFeesInner(disc []byte, data []byte, meta EventMetadata) DexEvent {
	eventDisc, ok := eventCPIDisc8(disc)
	if !ok {
		return DexEvent{}
	}
	switch eventDisc {
	case discPumpFeesCreateFeeSharingConfig:
		return parsePumpFeesCreateFeeSharingConfigFromData(data, meta)
	case discPumpFeesInitializeFeeConfig:
		return parsePumpFeesInitializeFeeConfigFromData(data, meta)
	case discPumpFeesResetFeeSharingConfig:
		return parsePumpFeesResetFeeSharingConfigFromData(data, meta)
	case discPumpFeesRevokeFeeSharingAuthority:
		return parsePumpFeesRevokeFeeSharingAuthorityFromData(data, meta)
	case discPumpFeesTransferFeeSharingAuthority:
		return parsePumpFeesTransferFeeSharingAuthorityFromData(data, meta)
	case discPumpFeesUpdateAdmin:
		return parsePumpFeesUpdateAdminFromData(data, meta)
	case discPumpFeesUpdateFeeConfig:
		return parsePumpFeesUpdateFeeConfigFromData(data, meta)
	case discPumpFeesUpdateFeeShares:
		return parsePumpFeesUpdateFeeSharesFromData(data, meta)
	case discPumpFeesUpsertFeeTiers:
		return parsePumpFeesUpsertFeeTiersFromData(data, meta)
	default:
		return DexEvent{}
	}
}

func dlmmInnerBuffer(disc []byte, inner []byte) []byte {
	buf := make([]byte, 8+len(inner))
	copy(buf[:8], disc[:8])
	copy(buf[8:], inner)
	return buf
}

// ParseInnerInstructionUnified 与 Rust `parse_inner_instruction` 对齐：16 字节 discriminator，data[16..] 为 payload。
func ParseInnerInstructionUnified(
	instructionData []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
	filter EventTypeFilter,
	programID string,
	isCreatedBuy bool,
) DexEvent {
	if len(instructionData) < 16 {
		return DexEvent{}
	}
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	disc := instructionData[:16]
	inner := instructionData[16:]
	disc8Value := binary.LittleEndian.Uint64(disc[:8])

	switch programID {
	case PUMPFUN_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpfun(filter) {
			return DexEvent{}
		}
		switch {
		case bytes.Equal(disc, pumpfunInnerTradeEvent):
			ev := parseTradeFromData(inner, meta, isCreatedBuy)
			if ev.Type != "" {
				if p, ok := ev.Data.(*PumpFunTradeEvent); ok {
					enrichPumpFunTradeFromAccounts(p, accounts)
				}
			}
			return applyActualEventTypeFilter(ev, filter)
		case bytes.Equal(disc, pumpfunInnerCreateToken):
			return applyActualEventTypeFilter(parseCreateFromData(inner, meta), filter)
		case bytes.Equal(disc, pumpfunInnerMigrateComplete):
			return applyActualEventTypeFilter(parseMigrateFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case PUMPSWAP_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpswap(filter) {
			return DexEvent{}
		}
		switch {
		case bytes.Equal(disc, pumpswapInnerBuy):
			ev := parsePSBuyFromData(inner, meta)
			if ev.Type != "" {
				if p, ok := ev.Data.(*PumpSwapBuyEvent); ok {
					enrichPumpSwapBuyFromAccounts(p, accounts)
				}
			}
			return applyActualEventTypeFilter(ev, filter)
		case bytes.Equal(disc, pumpswapInnerSell):
			ev := parsePSSellFromData(inner, meta)
			if ev.Type != "" {
				if p, ok := ev.Data.(*PumpSwapSellEvent); ok {
					enrichPumpSwapSellFromAccounts(p, accounts)
				}
			}
			return applyActualEventTypeFilter(ev, filter)
		case bytes.Equal(disc, pumpswapInnerCreatePool):
			return applyActualEventTypeFilter(parsePSCreatePoolFromData(inner, meta), filter)
		case bytes.Equal(disc, pumpswapInnerAddLiquidity):
			return applyActualEventTypeFilter(parsePSAddLiqFromData(inner, meta), filter)
		case bytes.Equal(disc, pumpswapInnerRemoveLiquidity):
			return applyActualEventTypeFilter(parsePSRemoveLiqFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case PUMP_FEES_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpFees(filter) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(parsePumpFeesInner(disc, inner, meta), filter)
	case RAYDIUM_CLMM_PROGRAM_ID, GrpcRaydiumClmmProgramID:
		if filter != nil && !EventTypeFilterIncludesRaydiumClmm(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		switch disc8Value {
		case discClmmSwap:
			return applyActualEventTypeFilter(parseClmmSwapFromData(inner, meta), filter)
		case discClmmIncLiq:
			return applyActualEventTypeFilter(parseClmmIncFromData(inner, meta), filter)
		case discClmmDecLiq:
			return applyActualEventTypeFilter(parseClmmDecFromData(inner, meta), filter)
		case discClmmLiqChange:
			return applyActualEventTypeFilter(parseClmmLiquidityChangeFromData(inner, meta), filter)
		case discClmmConfigChange:
			return applyActualEventTypeFilter(parseClmmConfigChangeFromData(inner, meta), filter)
		case discClmmCreatePersonalPosition:
			return applyActualEventTypeFilter(parseClmmCreatePersonalPositionFromData(inner, meta), filter)
		case discClmmLiqCalculate:
			return applyActualEventTypeFilter(parseClmmLiquidityCalculateFromData(inner, meta), filter)
		case discClmmOpenLimitOrder:
			return applyActualEventTypeFilter(parseClmmOpenLimitOrderFromData(inner, meta), filter)
		case discClmmIncreaseLimitOrder:
			return applyActualEventTypeFilter(parseClmmIncreaseLimitOrderFromData(inner, meta), filter)
		case discClmmDecreaseLimitOrder:
			return applyActualEventTypeFilter(parseClmmDecreaseLimitOrderFromData(inner, meta), filter)
		case discClmmSettleLimitOrder:
			return applyActualEventTypeFilter(parseClmmSettleLimitOrderFromData(inner, meta), filter)
		case discClmmUpdateRewardInfos:
			return applyActualEventTypeFilter(parseClmmUpdateRewardInfosFromData(inner, meta), filter)
		case discClmmCreate:
			return applyActualEventTypeFilter(parseClmmCreateFromData(inner, meta), filter)
		case discClmmCollectPersonal:
			return applyActualEventTypeFilter(parseClmmCollectPersonalFromData(inner, meta), filter)
		case discClmmCollectProtocol:
			return applyActualEventTypeFilter(parseClmmCollectProtocolFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case RAYDIUM_CPMM_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumCpmm(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		switch disc8Value {
		case discCpmmSwapIn:
			return applyActualEventTypeFilter(parseCpmmSwapInFromData(inner, meta), filter)
		case discCpmmSwapOut:
			return applyActualEventTypeFilter(parseCpmmSwapOutFromData(inner, meta), filter)
		case discCpmmCreatePool:
			return applyActualEventTypeFilter(parseCpmmInitFromData(inner, meta), filter)
		case discCpmmDeposit:
			return applyActualEventTypeFilter(parseCpmmDepositFromData(inner, meta), filter)
		case discCpmmWithdraw:
			return applyActualEventTypeFilter(parseCpmmWithdrawFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case RAYDIUM_AMM_V4_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumAmmV4(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		switch disc8Value {
		case discAmmSwapIn:
			return applyActualEventTypeFilter(parseAmmSwapInFromData(inner, meta), filter)
		case discAmmSwapOut:
			return applyActualEventTypeFilter(parseAmmSwapOutFromData(inner, meta), filter)
		case discAmmDeposit:
			return applyActualEventTypeFilter(parseAmmDepositFromData(inner, meta), filter)
		case discAmmWithdraw:
			return applyActualEventTypeFilter(parseAmmWithdrawFromData(inner, meta), filter)
		case discAmmInit2:
			return applyActualEventTypeFilter(parseAmmInit2FromData(inner, meta), filter)
		case discAmmWithdrawPnl:
			return applyActualEventTypeFilter(parseAmmWithdrawPnlFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case ORCA_WHIRLPOOL_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesOrcaWhirlpool(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		switch disc8Value {
		case discOrcaSwap:
			return applyActualEventTypeFilter(parseOrcaTradedFromData(inner, meta), filter)
		case discOrcaIncLiq:
			return applyActualEventTypeFilter(parseOrcaLiqIncFromData(inner, meta), filter)
		case discOrcaDecLiq:
			return applyActualEventTypeFilter(parseOrcaLiqDecFromData(inner, meta), filter)
		case discOrcaPoolInit:
			return applyActualEventTypeFilter(parseOrcaPoolInitFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case METEORA_POOLS_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraPools(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		switch disc8Value {
		case discMeteoraSwap:
			return applyActualEventTypeFilter(parseMeteoraSwapFromData(inner, meta), filter)
		case discMeteoraAdd:
			return applyActualEventTypeFilter(parseMeteoraAddFromData(inner, meta), filter)
		case discMeteoraRemove:
			return applyActualEventTypeFilter(parseMeteoraRemoveFromData(inner, meta), filter)
		case discMeteoraBootstrap:
			return applyActualEventTypeFilter(parseMeteoraBootstrapFromData(inner, meta), filter)
		case discMeteoraPoolCreated:
			return applyActualEventTypeFilter(parseMeteoraPoolCreatedFromData(inner, meta), filter)
		case discMeteoraSetPoolFees:
			return applyActualEventTypeFilter(parseMeteoraPoolsSetPoolFeesFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	case METEORA_DAMM_V2_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraDammV2(filter) {
			return DexEvent{}
		}
		if !disc16HasPrefix(disc) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(ParseMeteoraDammCpiInstruction(instructionData, meta), filter)
	case METEORA_DLMM_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraDlmm(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		return applyActualEventTypeFilter(parseDlmmFromProgramData(dlmmInnerBuffer(disc, inner), meta), filter)
	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumLaunchlab(filter) {
			return DexEvent{}
		}
		if !disc16HasSuffix(disc) {
			return DexEvent{}
		}
		switch disc8Value {
		case discRaydiumLaunchlabTrade:
			return applyActualEventTypeFilter(parseRaydiumLaunchlabTradeFromData(inner, meta), filter)
		case discRaydiumLaunchlabPoolCreate:
			return applyActualEventTypeFilter(parseRaydiumLaunchlabPoolCreateFromData(inner, meta), filter)
		default:
			return DexEvent{}
		}
	default:
		return DexEvent{}
	}
}

// makeInstrMetadata 构造指令元数据
func makeInstrMetadata(signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64) EventMetadata {
	bt := int64(0)
	if blockTimeUs != nil {
		bt = *blockTimeUs
	}
	return EventMetadata{
		Signature:   signature,
		Slot:        slot,
		TxIndex:     uint64(txIndex),
		BlockTimeUs: bt,
		GrpcRecvUs:  grpcRecvUs,
	}
}

// ParsePumpfunInstruction 与 Rust `pump::parse_instruction` 一致：解析外层 Create/CreateV2、
// legacy/v2 trade 指令，以及内层 CPI Migrate。
func ParsePumpfunInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	outer := binary.LittleEndian.Uint64(data[:8])
	if outer == instrPumpOuterCreateV2 {
		return parsePumpFunCreateV2Instr(data[8:], accounts, meta)
	}
	if outer == instrPumpOuterCreate {
		return parsePumpFunCreateInstr(data, accounts, meta)
	}
	if outer == instrPumpOuterBuy {
		return parsePumpFunLegacyBuyInstr("buy", data[8:], accounts, meta)
	}
	if outer == instrPumpOuterBuyExactSolIn {
		return parsePumpFunLegacyBuyInstr("buy_exact_sol_in", data[8:], accounts, meta)
	}
	if outer == instrPumpOuterSell {
		return parsePumpFunLegacySellInstr(data[8:], accounts, meta)
	}
	if outer == instrPumpOuterBuyV2 {
		return parsePumpFunTradeV2Instr("buy_v2", data[8:], accounts, meta)
	}
	if outer == instrPumpOuterBuyExactQuoteInV2 {
		return parsePumpFunTradeV2Instr("buy_exact_quote_in_v2", data[8:], accounts, meta)
	}
	if outer == instrPumpOuterSellV2 {
		return parsePumpFunTradeV2Instr("sell_v2", data[8:], accounts, meta)
	}
	if len(data) >= 16 {
		cpi := binary.LittleEndian.Uint64(data[8:16])
		if cpi == instrPumpMigrateCPI {
			return parsePumpFunMigrateInstr(data[16:], meta)
		}
	}
	return DexEvent{}
}

func readPumpFunV2Amount(data []byte, offset int) uint64 {
	if offset+8 > len(data) {
		return 0
	}
	return binary.LittleEndian.Uint64(data[offset : offset+8])
}

func mapBoolIndex(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}

func mapBoolString(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func parsePumpFunLegacyBuyInstr(ixName string, data []byte, accounts []string, meta EventMetadata) DexEvent {
	const minAcc = 16
	if len(accounts) < minAcc {
		return DexEvent{}
	}
	first := readPumpFunV2Amount(data, 0)
	second := readPumpFunV2Amount(data, 8)
	exactSolIn := ixName == "buy_exact_sol_in"
	tokenAmount := first
	solAmount := second
	amount := first
	maxSolCost := second
	var spendableSolIn uint64
	var minTokensOut uint64
	if exactSolIn {
		tokenAmount = second
		solAmount = first
		amount = second
		maxSolCost = first
		spendableSolIn = first
		minTokensOut = second
	}
	trackVolume := false
	if b, ok := readU8(data, 16); ok {
		trackVolume = b != 0
	}
	buybackFeeRecipient := getAccountSafe(accounts, 17)
	account := ""
	if buybackFeeRecipient != "" && buybackFeeRecipient != zeroPubkey {
		account = buybackFeeRecipient
	}
	eventType := EventTypePumpFunBuy
	if exactSolIn {
		eventType = EventTypePumpFunBuyExactSolIn
	}
	return DexEvent{
		Type: eventType,
		Data: &PumpFunTradeEvent{
			Metadata:                meta,
			Mint:                    getAccountSafe(accounts, 2),
			IsBuy:                   true,
			Global:                  getAccountSafe(accounts, 0),
			FeeRecipient:            getAccountSafe(accounts, 1),
			BondingCurve:            getAccountSafe(accounts, 3),
			BondingCurveV2:          getAccountSafe(accounts, 16),
			AssociatedBondingCurve:  getAccountSafe(accounts, 4),
			AssociatedUser:          getAccountSafe(accounts, 5),
			User:                    getAccountSafe(accounts, 6),
			SystemProgram:           getAccountSafe(accounts, 7),
			TokenProgram:            getAccountSafe(accounts, 8),
			CreatorVault:            getAccountSafe(accounts, 9),
			EventAuthority:          getAccountSafe(accounts, 10),
			Program:                 getAccountSafe(accounts, 11),
			GlobalVolumeAccumulator: getAccountSafe(accounts, 12),
			UserVolumeAccumulator:   getAccountSafe(accounts, 13),
			FeeConfig:               getAccountSafe(accounts, 14),
			FeeProgram:              getAccountSafe(accounts, 15),
			BuybackFeeRecipient:     buybackFeeRecipient,
			Account:                 account,
			SolAmount:               solAmount,
			TokenAmount:             tokenAmount,
			Amount:                  amount,
			MaxSolCost:              maxSolCost,
			SpendableSolIn:          spendableSolIn,
			MinTokensOut:            minTokensOut,
			TrackVolume:             trackVolume,
			IxName:                  ixName,
		},
	}
}

func parsePumpFunLegacySellInstr(data []byte, accounts []string, meta EventMetadata) DexEvent {
	const minAcc = 14
	if len(accounts) < minAcc {
		return DexEvent{}
	}
	amount := readPumpFunV2Amount(data, 0)
	minSolOutput := readPumpFunV2Amount(data, 8)
	legacyUserVolumeAccumulator := zeroPubkey
	legacyBondingCurveV2 := getAccountSafe(accounts, 14)
	legacyBuybackFeeRecipient := zeroPubkey
	if len(accounts) >= 17 {
		legacyUserVolumeAccumulator = getAccountSafe(accounts, 14)
		legacyBondingCurveV2 = getAccountSafe(accounts, 15)
		legacyBuybackFeeRecipient = getAccountSafe(accounts, 16)
	} else if len(accounts) >= 16 {
		legacyBondingCurveV2 = getAccountSafe(accounts, 14)
		legacyBuybackFeeRecipient = getAccountSafe(accounts, 15)
	}
	account := ""
	if legacyBuybackFeeRecipient != "" && legacyBuybackFeeRecipient != zeroPubkey {
		account = legacyBuybackFeeRecipient
	}
	return DexEvent{
		Type: EventTypePumpFunSell,
		Data: &PumpFunTradeEvent{
			Metadata:                meta,
			Mint:                    getAccountSafe(accounts, 2),
			IsBuy:                   false,
			Global:                  getAccountSafe(accounts, 0),
			FeeRecipient:            getAccountSafe(accounts, 1),
			BondingCurve:            getAccountSafe(accounts, 3),
			BondingCurveV2:          legacyBondingCurveV2,
			AssociatedBondingCurve:  getAccountSafe(accounts, 4),
			AssociatedUser:          getAccountSafe(accounts, 5),
			User:                    getAccountSafe(accounts, 6),
			SystemProgram:           getAccountSafe(accounts, 7),
			CreatorVault:            getAccountSafe(accounts, 8),
			TokenProgram:            getAccountSafe(accounts, 9),
			EventAuthority:          getAccountSafe(accounts, 10),
			Program:                 getAccountSafe(accounts, 11),
			GlobalVolumeAccumulator: zeroPubkey,
			UserVolumeAccumulator:   legacyUserVolumeAccumulator,
			FeeConfig:               getAccountSafe(accounts, 12),
			FeeProgram:              getAccountSafe(accounts, 13),
			BuybackFeeRecipient:     legacyBuybackFeeRecipient,
			Account:                 account,
			SolAmount:               minSolOutput,
			TokenAmount:             amount,
			Amount:                  amount,
			MinSolOutput:            minSolOutput,
			IxName:                  "sell",
		},
	}
}

func parsePumpFunTradeV2Instr(ixName string, data []byte, accounts []string, meta EventMetadata) DexEvent {
	if len(accounts) < 2 || accounts[1] == "" {
		return DexEvent{}
	}
	first := readPumpFunV2Amount(data, 0)
	second := readPumpFunV2Amount(data, 8)
	tokenAmount := first
	solAmount := second
	amount := first
	maxSolCost := second
	var quoteAmount uint64
	var minSolOutput uint64
	var spendableQuoteIn uint64
	var minTokensOut uint64
	if ixName == "buy_exact_quote_in_v2" {
		solAmount = first
		tokenAmount = second
		amount = second
		maxSolCost = 0
		quoteAmount = first
		spendableQuoteIn = first
		minTokensOut = second
	}
	if ixName == "sell_v2" {
		maxSolCost = 0
		minSolOutput = second
	}
	normalizedIxName := normalizePumpfunIxName(ixName)
	eventType := EventTypePumpFunBuy
	if ixName == "sell_v2" {
		eventType = EventTypePumpFunSell
	}
	return DexEvent{
		Type: eventType,
		Data: &PumpFunTradeEvent{
			Metadata:                           meta,
			Mint:                               getAccountSafe(accounts, 1),
			QuoteMint:                          getAccountSafe(accounts, 2),
			Global:                             getAccountSafe(accounts, 0),
			BondingCurve:                       getAccountSafe(accounts, 10),
			User:                               getAccountSafe(accounts, 13),
			SolAmount:                          solAmount,
			TokenAmount:                        tokenAmount,
			Amount:                             amount,
			MaxSolCost:                         maxSolCost,
			QuoteAmount:                        quoteAmount,
			MinSolOutput:                       minSolOutput,
			SpendableQuoteIn:                   spendableQuoteIn,
			MinTokensOut:                       minTokensOut,
			FeeRecipient:                       getAccountSafe(accounts, 6),
			IsBuy:                              ixName != "sell_v2",
			IxName:                             normalizedIxName,
			AssociatedBondingCurve:             getAccountSafe(accounts, 11),
			AssociatedUser:                     getAccountSafe(accounts, 14),
			SystemProgram:                      getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 23, 24)),
			TokenProgram:                       getAccountSafe(accounts, 3),
			QuoteTokenProgram:                  getAccountSafe(accounts, 4),
			AssociatedTokenProgram:             getAccountSafe(accounts, 5),
			CreatorVault:                       getAccountSafe(accounts, 16),
			AssociatedQuoteFeeRecipient:        getAccountSafe(accounts, 7),
			BuybackFeeRecipient:                getAccountSafe(accounts, 8),
			AssociatedQuoteBuybackFeeRecipient: getAccountSafe(accounts, 9),
			AssociatedQuoteBondingCurve:        getAccountSafe(accounts, 12),
			AssociatedQuoteUser:                getAccountSafe(accounts, 15),
			AssociatedCreatorVault:             getAccountSafe(accounts, 17),
			SharingConfig:                      getAccountSafe(accounts, 18),
			EventAuthority:                     getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 24, 25)),
			Program:                            getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 25, 26)),
			GlobalVolumeAccumulator:            mapBoolString(ixName == "sell_v2", "", getAccountSafe(accounts, 19)),
			UserVolumeAccumulator:              getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 19, 20)),
			AssociatedUserVolumeAccumulator:    getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 20, 21)),
			FeeConfig:                          getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 21, 22)),
			FeeProgram:                         getAccountSafe(accounts, mapBoolIndex(ixName == "sell_v2", 22, 23)),
		},
	}
}

func parsePumpFunCreateInstr(data []byte, accounts []string, meta EventMetadata) DexEvent {
	offset := 8 // Skip discriminator

	if offset+4 > len(data) {
		return DexEvent{}
	}
	nameLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+nameLen > len(data) {
		return DexEvent{}
	}
	name := string(data[offset : offset+nameLen])
	offset += nameLen

	if offset+4 > len(data) {
		return DexEvent{}
	}
	symbolLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+symbolLen > len(data) {
		return DexEvent{}
	}
	symbol := string(data[offset : offset+symbolLen])
	offset += symbolLen

	if offset+4 > len(data) {
		return DexEvent{}
	}
	uriLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+uriLen > len(data) {
		return DexEvent{}
	}
	uri := string(data[offset : offset+uriLen])
	offset += uriLen

	creator := zeroPubkey
	if offset+32 <= len(data) {
		creator = ReadPubkey(data, offset)
	}

	return DexEvent{
		Type: EventTypePumpFunCreate,
		Data: &PumpFunCreateEvent{
			Metadata:               meta,
			Name:                   name,
			Symbol:                 symbol,
			Uri:                    uri,
			Creator:                creator,
			Mint:                   getAccountSafe(accounts, 0),
			BondingCurve:           getAccountSafe(accounts, 2),
			User:                   getAccountSafe(accounts, 7),
			QuoteMint:              zeroPubkey,
			IxName:                 "create",
			MintAuthority:          getAccountSafe(accounts, 1),
			AssociatedBondingCurve: getAccountSafe(accounts, 3),
			Global:                 getAccountSafe(accounts, 4),
			SystemProgram:          getAccountSafe(accounts, 8),
			TokenProgram:           getAccountSafe(accounts, 9),
			AssociatedTokenProgram: getAccountSafe(accounts, 10),
			EventAuthority:         getAccountSafe(accounts, 12),
			Program:                getAccountSafe(accounts, 13),
		},
	}
}

func parsePumpFunCreateV2Instr(data []byte, accounts []string, meta EventMetadata) DexEvent {
	const minAcc = 16
	if len(accounts) < minAcc {
		return DexEvent{}
	}
	off := 0
	var ok bool
	var name, symbol, uri string
	name, off, ok = readBorshString(data, off)
	if !ok {
		return DexEvent{}
	}
	symbol, off, ok = readBorshString(data, off)
	if !ok {
		return DexEvent{}
	}
	uri, off, ok = readBorshString(data, off)
	if !ok {
		return DexEvent{}
	}
	if off+33 > len(data) {
		return DexEvent{}
	}
	creator := ReadPubkey(data, off)
	off += 32
	isMayhemMode, ok := readBool(data, off)
	if !ok {
		return DexEvent{}
	}
	off++
	isCashbackEnabled := false
	if v, ok := readBool(data, off); ok {
		isCashbackEnabled = v
	}
	acc := accounts[:minAcc]
	return DexEvent{
		Type: EventTypePumpFunCreateV2,
		Data: &PumpFunCreateV2TokenEvent{
			Metadata:               meta,
			Name:                   name,
			Symbol:                 symbol,
			Uri:                    uri,
			Mint:                   acc[0],
			BondingCurve:           acc[2],
			User:                   acc[5],
			Creator:                creator,
			TokenProgram:           acc[7],
			IsMayhemMode:           isMayhemMode,
			IsCashbackEnabled:      isCashbackEnabled,
			QuoteMint:              getAccountSafe(accounts, 16),
			QuoteVault:             getAccountSafe(accounts, 17),
			QuoteTokenProgram:      getAccountSafe(accounts, 18),
			IxName:                 "create_v2",
			MintAuthority:          acc[1],
			AssociatedBondingCurve: acc[3],
			Global:                 acc[4],
			SystemProgram:          acc[6],
			AssociatedTokenProgram: acc[8],
			MayhemProgramID:        acc[9],
			GlobalParams:           acc[10],
			SolVault:               acc[11],
			MayhemState:            acc[12],
			MayhemTokenVault:       acc[13],
			EventAuthority:         acc[14],
			Program:                acc[15],
		},
	}
}

func parsePumpFunMigrateInstr(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+32+8+32 {
		return DexEvent{}
	}
	o := 0
	user := ReadPubkey(data, o)
	o += 32
	mint := ReadPubkey(data, o)
	o += 32
	ma, ok1 := readU64LE(data, o)
	o += 8
	sa, ok2 := readU64LE(data, o)
	o += 8
	pmf, ok3 := readU64LE(data, o)
	o += 8
	bc := ReadPubkey(data, o)
	o += 32
	ts, ok4 := readU64LE(data, o)
	o += 8
	pool := ReadPubkey(data, o)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypePumpFunMigrate,
		Data: &PumpFunMigrateEvent{
			Metadata:         meta,
			User:             user,
			Mint:             mint,
			MintAmount:       ma,
			SolAmount:        sa,
			PoolMigrationFee: pmf,
			BondingCurve:     bc,
			Timestamp:        int64(ts),
			Pool:             pool,
		},
	}
}

// ParsePumpswapInstruction 与 Rust `pump_amm::parse_instruction` 一致（**指令** discriminator，非 Program log 的 Event disc）。
func ParsePumpswapInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}
	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	switch discriminator {
	case instrPumpSwapBuy:
		return parsePumpSwapBuyInstr(data, accounts, meta, false)
	case instrPumpSwapBuyExactQuote:
		return parsePumpSwapBuyInstr(data, accounts, meta, true)
	case instrPumpSwapSell:
		return parsePumpSwapSellInstr(data, accounts, meta)
	case instrPumpSwapCreatePool:
		return parsePumpSwapCreatePoolInstr(accounts, meta)
	case instrPumpSwapDeposit:
		return parsePumpSwapDepositInstr(accounts, meta)
	case instrPumpSwapWithdraw:
		return parsePumpSwapWithdrawInstr(accounts, meta)
	default:
		return DexEvent{}
	}
}

func fillPumpSwapBuyUpgradeAccounts(ev *PumpSwapBuyEvent, accounts []string) {
	if len(accounts) >= 27 {
		ev.PoolV2 = getAccountSafe(accounts, 24)
		ev.FeeRecipient = getAccountSafe(accounts, 25)
		ev.FeeRecipientQuoteTokenAccount = getAccountSafe(accounts, 26)
	} else if len(accounts) >= 26 {
		ev.PoolV2 = getAccountSafe(accounts, 23)
		ev.FeeRecipient = getAccountSafe(accounts, 24)
		ev.FeeRecipientQuoteTokenAccount = getAccountSafe(accounts, 25)
	} else if len(accounts) >= 24 {
		ev.PoolV2 = getAccountSafe(accounts, 23)
	}
}

func fillPumpSwapSellUpgradeAccounts(ev *PumpSwapSellEvent, accounts []string) {
	if len(accounts) >= 26 {
		ev.PoolV2 = getAccountSafe(accounts, 23)
		ev.FeeRecipient = getAccountSafe(accounts, 24)
		ev.FeeRecipientQuoteTokenAccount = getAccountSafe(accounts, 25)
	} else if len(accounts) >= 24 {
		ev.PoolV2 = getAccountSafe(accounts, 21)
		ev.FeeRecipient = getAccountSafe(accounts, 22)
		ev.FeeRecipientQuoteTokenAccount = getAccountSafe(accounts, 23)
	} else if len(accounts) >= 22 {
		ev.PoolV2 = getAccountSafe(accounts, 21)
	}
}

func parsePumpSwapBuyInstr(data []byte, accounts []string, meta EventMetadata, buyExactQuoteIn bool) DexEvent {
	if len(accounts) < 13 {
		return DexEvent{}
	}
	payload := data[8:]
	var a0, a1 uint64
	if len(payload) >= 16 {
		a0 = binary.LittleEndian.Uint64(payload[0:8])
		a1 = binary.LittleEndian.Uint64(payload[8:16])
	}
	var baseOut, maxQuoteIn uint64
	if buyExactQuoteIn {
		maxQuoteIn, baseOut = a0, a1
	} else {
		baseOut, maxQuoteIn = a0, a1
	}
	ev := &PumpSwapBuyEvent{
		Metadata:                         meta,
		BaseAmountOut:                    baseOut,
		MaxQuoteAmountIn:                 maxQuoteIn,
		Pool:                             getAccountSafe(accounts, 0),
		User:                             getAccountSafe(accounts, 1),
		BaseMint:                         getAccountSafe(accounts, 3),
		QuoteMint:                        getAccountSafe(accounts, 4),
		UserBaseTokenAccount:             getAccountSafe(accounts, 5),
		UserQuoteTokenAccount:            getAccountSafe(accounts, 6),
		PoolBaseTokenAccount:             getAccountSafe(accounts, 7),
		PoolQuoteTokenAccount:            getAccountSafe(accounts, 8),
		ProtocolFeeRecipient:             getAccountSafe(accounts, 9),
		ProtocolFeeRecipientTokenAccount: getAccountSafe(accounts, 10),
		BaseTokenProgram:                 getAccountSafe(accounts, 11),
		QuoteTokenProgram:                getAccountSafe(accounts, 12),
	}
	if buyExactQuoteIn {
		ev.IxName = "buy_exact_quote_in"
	} else {
		ev.IxName = "buy"
	}
	if len(accounts) >= 19 {
		ev.CoinCreatorVaultAta = getAccountSafe(accounts, 17)
		ev.CoinCreatorVaultAuthority = getAccountSafe(accounts, 18)
	}
	fillPumpSwapBuyUpgradeAccounts(ev, accounts)
	return DexEvent{Type: EventTypePumpSwapBuy, Data: ev}
}

func parsePumpSwapSellInstr(data []byte, accounts []string, meta EventMetadata) DexEvent {
	if len(accounts) < 13 {
		return DexEvent{}
	}
	payload := data[8:]
	var baseIn, minQuoteOut uint64
	if len(payload) >= 16 {
		baseIn = binary.LittleEndian.Uint64(payload[0:8])
		minQuoteOut = binary.LittleEndian.Uint64(payload[8:16])
	}
	ev := &PumpSwapSellEvent{
		Metadata:                         meta,
		BaseAmountIn:                     baseIn,
		MinQuoteAmountOut:                minQuoteOut,
		Pool:                             getAccountSafe(accounts, 0),
		User:                             getAccountSafe(accounts, 1),
		BaseMint:                         getAccountSafe(accounts, 3),
		QuoteMint:                        getAccountSafe(accounts, 4),
		UserBaseTokenAccount:             getAccountSafe(accounts, 5),
		UserQuoteTokenAccount:            getAccountSafe(accounts, 6),
		PoolBaseTokenAccount:             getAccountSafe(accounts, 7),
		PoolQuoteTokenAccount:            getAccountSafe(accounts, 8),
		ProtocolFeeRecipient:             getAccountSafe(accounts, 9),
		ProtocolFeeRecipientTokenAccount: getAccountSafe(accounts, 10),
		BaseTokenProgram:                 getAccountSafe(accounts, 11),
		QuoteTokenProgram:                getAccountSafe(accounts, 12),
	}
	if len(accounts) >= 19 {
		ev.CoinCreatorVaultAta = getAccountSafe(accounts, 17)
		ev.CoinCreatorVaultAuthority = getAccountSafe(accounts, 18)
	}
	fillPumpSwapSellUpgradeAccounts(ev, accounts)
	return DexEvent{Type: EventTypePumpSwapSell, Data: ev}
}

func parsePumpSwapCreatePoolInstr(accounts []string, meta EventMetadata) DexEvent {
	if len(accounts) < 5 {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypePumpSwapCreatePool,
		Data: &PumpSwapCreatePoolEvent{
			Metadata:  meta,
			Creator:   getAccountSafe(accounts, 0),
			BaseMint:  getAccountSafe(accounts, 2),
			QuoteMint: getAccountSafe(accounts, 3),
		},
	}
}

func parsePumpSwapDepositInstr(accounts []string, meta EventMetadata) DexEvent {
	if len(accounts) < 8 {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypePumpSwapLiquidityAdded,
		Data: &PumpSwapLiquidityAddedEvent{
			Metadata:              meta,
			Pool:                  getAccountSafe(accounts, 0),
			User:                  getAccountSafe(accounts, 1),
			UserBaseTokenAccount:  getAccountSafe(accounts, 4),
			UserQuoteTokenAccount: getAccountSafe(accounts, 5),
			UserPoolTokenAccount:  getAccountSafe(accounts, 6),
		},
	}
}

func parsePumpSwapWithdrawInstr(accounts []string, meta EventMetadata) DexEvent {
	if len(accounts) < 8 {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypePumpSwapLiquidityRemoved,
		Data: &PumpSwapLiquidityRemovedEvent{
			Metadata:              meta,
			Pool:                  getAccountSafe(accounts, 0),
			User:                  getAccountSafe(accounts, 1),
			UserBaseTokenAccount:  getAccountSafe(accounts, 4),
			UserQuoteTokenAccount: getAccountSafe(accounts, 5),
			UserPoolTokenAccount:  getAccountSafe(accounts, 6),
		},
	}
}

// ParseMeteoraDammInstruction 解析 Meteora DAMM V2 指令（与 Rust `meteora_damm::parse_instruction` 一致：CPI disc 在 [8..16)）。
func ParseMeteoraDammInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) >= 8 && binary.LittleEndian.Uint64(data[:8]) == discDammInit {
		if len(data) < 8+16+16+1 {
			return DexEvent{}
		}
		meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
		payload := data[8:]
		liquidity, _ := readU128LE(payload, 0)
		sqrtPrice, _ := readU128LE(payload, 16)
		var activationPoint uint64
		if len(payload) >= 41 && payload[32] == 1 {
			activationPoint, _ = readU64LE(payload, 33)
		}
		return DexEvent{
			Type: EventTypeMeteoraDammV2InitializePool,
			Data: &MeteoraDammV2InitializePoolEvent{
				Metadata:        meta,
				Creator:         getAccountSafe(accounts, 0),
				PositionNftMint: getAccountSafe(accounts, 1),
				Pool:            getAccountSafe(accounts, 6),
				Position:        getAccountSafe(accounts, 7),
				TokenAMint:      getAccountSafe(accounts, 8),
				TokenBMint:      getAccountSafe(accounts, 9),
				Liquidity:       u128LEDecimalString(liquidity),
				SqrtPrice:       u128LEDecimalString(sqrtPrice),
				ActivationPoint: activationPoint,
			},
		}
	}
	if len(data) < 16 {
		return DexEvent{}
	}
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	return ParseMeteoraDammCpiInstruction(data, meta)
}

// ParseRaydiumClmmInstruction 解析 Raydium CLMM 指令
func ParseRaydiumClmmInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	payload := data[8:]

	switch discriminator {
	case instrClmmSwap, instrClmmSwapV2:
		if len(payload) < 8+8+16+1 {
			return DexEvent{}
		}
		sqrt, _ := readU128LE(payload, 16)
		isBaseInput, ok := readBool(payload, 32)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{
			Type: EventTypeRaydiumClmmSwap,
			Data: &RaydiumClmmSwapEvent{
				Metadata:      meta,
				PoolState:     getAccountSafe(accounts, 2),
				Sender:        getAccountSafe(accounts, 0),
				TokenAccount0: getAccountSafe(accounts, 3),
				TokenAccount1: getAccountSafe(accounts, 4),
				ZeroForOne:    isBaseInput,
				SqrtPriceX64:  u128LEDecimalString(sqrt),
				Liquidity:     "0",
			},
		}
	case instrClmmIncLiqV2:
		if len(payload) < 16+8+8 {
			return DexEvent{}
		}
		liquidity, _ := readU128LE(payload, 0)
		amount0Max, _ := readU64LE(payload, 16)
		amount1Max, _ := readU64LE(payload, 24)
		return DexEvent{
			Type: EventTypeRaydiumClmmIncreaseLiquidity,
			Data: &RaydiumClmmIncreaseLiquidityEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 2),
				PositionNftMint: getAccountSafe(accounts, 1),
				User:            getAccountSafe(accounts, 0),
				Liquidity:       u128LEDecimalString(liquidity),
				Amount0Max:      amount0Max,
				Amount1Max:      amount1Max,
			},
		}
	case instrClmmDecLiqV2:
		if len(payload) < 16+8+8 {
			return DexEvent{}
		}
		liquidity, _ := readU128LE(payload, 0)
		amount0Min, _ := readU64LE(payload, 16)
		amount1Min, _ := readU64LE(payload, 24)
		return DexEvent{
			Type: EventTypeRaydiumClmmDecreaseLiquidity,
			Data: &RaydiumClmmDecreaseLiquidityEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 3),
				PositionNftMint: getAccountSafe(accounts, 1),
				User:            getAccountSafe(accounts, 0),
				Liquidity:       u128LEDecimalString(liquidity),
				Amount0Min:      amount0Min,
				Amount1Min:      amount1Min,
			},
		}
	case instrClmmCreatePool:
		if len(payload) < 16+8 {
			return DexEvent{}
		}
		sqrt, _ := readU128LE(payload, 0)
		openTime, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeRaydiumClmmCreatePool,
			Data: &RaydiumClmmCreatePoolEvent{
				Metadata:     meta,
				Pool:         getAccountSafe(accounts, 2),
				Creator:      getAccountSafe(accounts, 0),
				Token0Mint:   getAccountSafe(accounts, 3),
				Token1Mint:   getAccountSafe(accounts, 4),
				SqrtPriceX64: u128LEDecimalString(sqrt),
				TokenVault0:  getAccountSafe(accounts, 5),
				TokenVault1:  getAccountSafe(accounts, 6),
				OpenTime:     openTime,
			},
		}
	case instrClmmCreateCustomizablePool:
		if len(payload) < 16 {
			return DexEvent{}
		}
		sqrt, _ := readU128LE(payload, 0)
		return DexEvent{
			Type: EventTypeRaydiumClmmCreatePool,
			Data: &RaydiumClmmCreatePoolEvent{
				Metadata:     meta,
				Pool:         getAccountSafe(accounts, 2),
				Creator:      getAccountSafe(accounts, 0),
				Token0Mint:   getAccountSafe(accounts, 3),
				Token1Mint:   getAccountSafe(accounts, 4),
				SqrtPriceX64: u128LEDecimalString(sqrt),
				TokenVault0:  getAccountSafe(accounts, 5),
				TokenVault1:  getAccountSafe(accounts, 6),
			},
		}
	case instrClmmOpenPosition, instrClmmOpenPositionV2, instrClmmOpenPositionWithToken22Nft:
		if len(payload) < 4+4+4+4+16+8+8 {
			return DexEvent{}
		}
		tickLower, _ := readI32LE(payload, 0)
		tickUpper, _ := readI32LE(payload, 4)
		liquidity, _ := readU128LE(payload, 16)
		poolIndex := 5
		if discriminator == instrClmmOpenPositionWithToken22Nft {
			poolIndex = 4
		}
		return DexEvent{
			Type: EventTypeRaydiumClmmOpenPosition,
			Data: &RaydiumClmmOpenPositionEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, poolIndex),
				User:            getAccountSafe(accounts, 1),
				PositionNftMint: getAccountSafe(accounts, 2),
				TickLowerIndex:  tickLower,
				TickUpperIndex:  tickUpper,
				Liquidity:       u128LEDecimalString(liquidity),
			},
		}
	case instrClmmClosePosition:
		return DexEvent{
			Type: EventTypeRaydiumClmmClosePosition,
			Data: &RaydiumClmmClosePositionEvent{
				Metadata:        meta,
				Pool:            zeroPubkey,
				User:            getAccountSafe(accounts, 0),
				PositionNftMint: getAccountSafe(accounts, 1),
			},
		}
	}

	return DexEvent{}
}

// ParseRaydiumCpmmInstruction 解析 Raydium CPMM 指令
func ParseRaydiumCpmmInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch discriminator {
	case discCpmmSwapIn, discCpmmSwapOut:
		if len(data) < 8+16 {
			return DexEvent{}
		}
		input, _ := readU64LE(data, 8)
		output, _ := readU64LE(data, 16)
		return DexEvent{
			Type: EventTypeRaydiumCpmmSwap,
			Data: &RaydiumCpmmSwapEvent{
				Metadata:     meta,
				PoolID:       getAccountSafe(accounts, 0),
				InputAmount:  input,
				OutputAmount: output,
				BaseInput:    discriminator == discCpmmSwapIn,
			},
		}
	case instrCpmmInitialize:
		if len(data) < 8+8+8+8 {
			return DexEvent{}
		}
		initAmount0, _ := readU64LE(data, 8)
		initAmount1, _ := readU64LE(data, 16)
		return DexEvent{
			Type: EventTypeRaydiumCpmmInitialize,
			Data: &RaydiumCpmmInitializeEvent{
				Metadata:    meta,
				Pool:        getAccountSafe(accounts, 0),
				Creator:     getAccountSafe(accounts, 1),
				InitAmount0: initAmount0,
				InitAmount1: initAmount1,
			},
		}
	case discCpmmDeposit:
		if len(data) < 8+8+8+8 {
			return DexEvent{}
		}
		lp, _ := readU64LE(data, 8)
		token0, _ := readU64LE(data, 16)
		token1, _ := readU64LE(data, 24)
		return DexEvent{
			Type: EventTypeRaydiumCpmmDeposit,
			Data: &RaydiumCpmmDepositEvent{
				Metadata:      meta,
				Pool:          getAccountSafe(accounts, 0),
				User:          getAccountSafe(accounts, 1),
				LpTokenAmount: lp,
				Token0Amount:  token0,
				Token1Amount:  token1,
			},
		}
	case discCpmmWithdraw:
		if len(data) < 8+8+8+8 {
			return DexEvent{}
		}
		lp, _ := readU64LE(data, 8)
		token0, _ := readU64LE(data, 16)
		token1, _ := readU64LE(data, 24)
		return DexEvent{
			Type: EventTypeRaydiumCpmmWithdraw,
			Data: &RaydiumCpmmWithdrawEvent{
				Metadata:      meta,
				Pool:          getAccountSafe(accounts, 0),
				User:          getAccountSafe(accounts, 1),
				LpTokenAmount: lp,
				Token0Amount:  token0,
				Token1Amount:  token1,
			},
		}
	}

	return DexEvent{}
}

// ParseRaydiumAmmV4Instruction 解析 Raydium AMM V4 指令
func ParseRaydiumAmmV4Instruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 1 {
		return DexEvent{}
	}

	// Raydium AMM V4 使用单字节 instruction discriminator
	instrType := data[0]
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch instrType {
	case 9, 11: // SwapBaseIn, SwapBaseOut
		if len(data) < 17 {
			return DexEvent{}
		}
		var amountIn, minOut, maxIn, amountOut uint64
		if instrType == 9 {
			amountIn, _ = readU64LE(data, 1)
			minOut, _ = readU64LE(data, 9)
		} else {
			maxIn, _ = readU64LE(data, 1)
			amountOut, _ = readU64LE(data, 9)
		}
		return DexEvent{
			Type: EventTypeRaydiumAmmV4Swap,
			Data: &RaydiumAmmV4SwapEvent{
				Metadata:                    meta,
				Amm:                         getAccountSafe(accounts, 1),
				UserSourceOwner:             getAccountSafe(accounts, 17),
				AmountIn:                    amountIn,
				MinimumAmountOut:            minOut,
				MaxAmountIn:                 maxIn,
				AmountOut:                   amountOut,
				TokenProgram:                getAccountSafe(accounts, 0),
				AmmAuthority:                getAccountSafe(accounts, 2),
				AmmOpenOrders:               getAccountSafe(accounts, 3),
				PoolCoinTokenAccount:        getAccountSafe(accounts, 5),
				PoolPcTokenAccount:          getAccountSafe(accounts, 6),
				SerumProgram:                getAccountSafe(accounts, 7),
				SerumMarket:                 getAccountSafe(accounts, 8),
				SerumBids:                   getAccountSafe(accounts, 9),
				SerumAsks:                   getAccountSafe(accounts, 10),
				SerumEventQueue:             getAccountSafe(accounts, 11),
				SerumCoinVaultAccount:       getAccountSafe(accounts, 12),
				SerumPcVaultAccount:         getAccountSafe(accounts, 13),
				SerumVaultSigner:            getAccountSafe(accounts, 14),
				UserSourceTokenAccount:      getAccountSafe(accounts, 15),
				UserDestinationTokenAccount: getAccountSafe(accounts, 16),
			},
		}
	case 3:
		if len(data) < 25 {
			return DexEvent{}
		}
		maxCoin, _ := readU64LE(data, 1)
		maxPc, _ := readU64LE(data, 9)
		baseSide, _ := readU64LE(data, 17)
		return DexEvent{
			Type: EventTypeRaydiumAmmV4Deposit,
			Data: &RaydiumAmmV4DepositEvent{
				Metadata:             meta,
				Amm:                  getAccountSafe(accounts, 1),
				UserOwner:            getAccountSafe(accounts, 12),
				MaxCoinAmount:        maxCoin,
				MaxPcAmount:          maxPc,
				BaseSide:             baseSide,
				TokenProgram:         getAccountSafe(accounts, 0),
				AmmAuthority:         getAccountSafe(accounts, 2),
				AmmOpenOrders:        getAccountSafe(accounts, 3),
				AmmTargetOrders:      getAccountSafe(accounts, 4),
				LpMintAddress:        getAccountSafe(accounts, 5),
				PoolCoinTokenAccount: getAccountSafe(accounts, 6),
				PoolPcTokenAccount:   getAccountSafe(accounts, 7),
				SerumMarket:          getAccountSafe(accounts, 8),
				UserCoinTokenAccount: getAccountSafe(accounts, 9),
				UserPcTokenAccount:   getAccountSafe(accounts, 10),
				UserLpTokenAccount:   getAccountSafe(accounts, 11),
				SerumEventQueue:      getAccountSafe(accounts, 13),
			},
		}
	case 4:
		if len(data) < 9 {
			return DexEvent{}
		}
		amount, _ := readU64LE(data, 1)
		return DexEvent{
			Type: EventTypeRaydiumAmmV4Withdraw,
			Data: &RaydiumAmmV4WithdrawEvent{
				Metadata:               meta,
				Amm:                    getAccountSafe(accounts, 1),
				UserOwner:              getAccountSafe(accounts, 18),
				Amount:                 amount,
				TokenProgram:           getAccountSafe(accounts, 0),
				AmmAuthority:           getAccountSafe(accounts, 2),
				AmmOpenOrders:          getAccountSafe(accounts, 3),
				AmmTargetOrders:        getAccountSafe(accounts, 4),
				LpMintAddress:          getAccountSafe(accounts, 5),
				PoolCoinTokenAccount:   getAccountSafe(accounts, 6),
				PoolPcTokenAccount:     getAccountSafe(accounts, 7),
				PoolWithdrawQueue:      getAccountSafe(accounts, 8),
				PoolTempLpTokenAccount: getAccountSafe(accounts, 9),
				SerumProgram:           getAccountSafe(accounts, 10),
				SerumMarket:            getAccountSafe(accounts, 11),
				SerumCoinVaultAccount:  getAccountSafe(accounts, 12),
				SerumPcVaultAccount:    getAccountSafe(accounts, 13),
				SerumVaultSigner:       getAccountSafe(accounts, 14),
				UserLpTokenAccount:     getAccountSafe(accounts, 15),
				UserCoinTokenAccount:   getAccountSafe(accounts, 16),
				UserPcTokenAccount:     getAccountSafe(accounts, 17),
				SerumEventQueue:        getAccountSafe(accounts, 19),
				SerumBids:              getAccountSafe(accounts, 20),
				SerumAsks:              getAccountSafe(accounts, 21),
			},
		}
	case 1:
		if len(data) < 26 {
			return DexEvent{}
		}
		openTime, _ := readU64LE(data, 2)
		initPcAmount, _ := readU64LE(data, 10)
		initCoinAmount, _ := readU64LE(data, 18)
		return DexEvent{
			Type: EventTypeRaydiumAmmV4Initialize2,
			Data: &RaydiumAmmV4DepositEvent{
				Metadata:      meta,
				Amm:           getAccountSafe(accounts, 4),
				MaxCoinAmount: initCoinAmount,
				MaxPcAmount:   initPcAmount,
				BaseSide:      openTime,
				TokenProgram:  getAccountSafe(accounts, 0),
			},
		}
	case 7:
		return DexEvent{
			Type: EventTypeRaydiumAmmV4WithdrawPnl,
			Data: &RaydiumAmmV4WithdrawEvent{
				Metadata:     meta,
				Amm:          getAccountSafe(accounts, 1),
				TokenProgram: getAccountSafe(accounts, 0),
			},
		}
	}

	return DexEvent{}
}

// ParseOrcaWhirlpoolInstruction 解析 Orca Whirlpool 指令
func ParseOrcaWhirlpoolInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch discriminator {
	case disc8(248, 198, 158, 145, 225, 117, 135, 200), disc8(43, 4, 237, 11, 26, 201, 30, 98):
		if len(data) < 8+8+8+16+1+1 {
			return DexEvent{}
		}
		amount, _ := readU64LE(data, 8)
		threshold, _ := readU64LE(data, 16)
		sqrt, _ := readU128LE(data, 24)
		inputSpecified, _ := readBool(data, 40)
		aToB, _ := readBool(data, 41)
		inputAmount := uint64(0)
		outputAmount := threshold
		if inputSpecified {
			inputAmount = amount
		} else {
			outputAmount = amount
		}
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolSwap,
			Data: &OrcaWhirlpoolSwapEvent{
				Metadata:     meta,
				Whirlpool:    getAccountSafe(accounts, 1),
				AToB:         aToB,
				PreSqrtPrice: u128LEDecimalString(sqrt),
				InputAmount:  inputAmount,
				OutputAmount: outputAmount,
			},
		}
	case disc8(46, 156, 243, 118, 13, 205, 251, 178):
		if len(data) < 8+16+8+8 {
			return DexEvent{}
		}
		liquidity, _ := readU128LE(data, 8)
		amountA, _ := readU64LE(data, 24)
		amountB, _ := readU64LE(data, 32)
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolLiquidityIncreased,
			Data: &OrcaWhirlpoolLiquidityIncreasedEvent{
				Metadata:     meta,
				Whirlpool:    getAccountSafe(accounts, 1),
				Position:     getAccountSafe(accounts, 3),
				Liquidity:    u128LEDecimalString(liquidity),
				TokenAAmount: amountA,
				TokenBAmount: amountB,
			},
		}
	case disc8(160, 38, 208, 111, 104, 91, 44, 1):
		if len(data) < 8+16+8+8 {
			return DexEvent{}
		}
		liquidity, _ := readU128LE(data, 8)
		amountA, _ := readU64LE(data, 24)
		amountB, _ := readU64LE(data, 32)
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolLiquidityDecreased,
			Data: &OrcaWhirlpoolLiquidityDecreasedEvent{
				Metadata:     meta,
				Whirlpool:    getAccountSafe(accounts, 1),
				Position:     getAccountSafe(accounts, 3),
				Liquidity:    u128LEDecimalString(liquidity),
				TokenAAmount: amountA,
				TokenBAmount: amountB,
			},
		}
	case disc8(17, 43, 80, 74, 168, 202, 6, 113):
		if len(data) < 8+2+16 {
			return DexEvent{}
		}
		tickSpacing, _ := readU16LE(data, 8)
		sqrt, _ := readU128LE(data, 10)
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolPoolInitialized,
			Data: &OrcaWhirlpoolPoolInitializedEvent{
				Metadata:         meta,
				Whirlpool:        getAccountSafe(accounts, 1),
				WhirlpoolsConfig: getAccountSafe(accounts, 2),
				TokenMintA:       getAccountSafe(accounts, 3),
				TokenMintB:       getAccountSafe(accounts, 4),
				TickSpacing:      tickSpacing,
				TokenProgramA:    getAccountSafe(accounts, 8),
				TokenProgramB:    getAccountSafe(accounts, 9),
				InitialSqrtPrice: u128LEDecimalString(sqrt),
			},
		}
	}

	return DexEvent{}
}

// ParseRaydiumLaunchlabInstruction 解析 RaydiumLaunchlab (Raydium LaunchLab) 指令
func ParseRaydiumLaunchlabInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	payload := data[8:]

	switch discriminator {
	case discRaydiumLaunchlabTrade:
		return parseRaydiumLaunchlabTradeFromData(payload, meta)
	case discRaydiumLaunchlabPoolCreate:
		return parseRaydiumLaunchlabPoolCreateFromData(payload, meta)
	case instrRaydiumLaunchlabBuyExactIn, instrRaydiumLaunchlabBuyExactOut,
		instrRaydiumLaunchlabSellExactIn, instrRaydiumLaunchlabSellExactOut:
		first, _ := readU64LE(payload, 0)
		second, _ := readU64LE(payload, 8)
		exactIn := discriminator == instrRaydiumLaunchlabBuyExactIn ||
			discriminator == instrRaydiumLaunchlabSellExactIn
		isBuy := discriminator == instrRaydiumLaunchlabBuyExactIn ||
			discriminator == instrRaydiumLaunchlabBuyExactOut
		amountIn, amountOut := first, second
		if !exactIn {
			amountIn, amountOut = second, first
		}
		dir := "Sell"
		if isBuy {
			dir = "Buy"
		}
		return DexEvent{
			Type: EventTypeRaydiumLaunchlabTrade,
			Data: &RaydiumLaunchlabTradeEvent{
				Metadata:       meta,
				PoolState:      getAccountSafe(accounts, 4),
				User:           getAccountSafe(accounts, 0),
				AmountIn:       amountIn,
				AmountOut:      amountOut,
				IsBuy:          isBuy,
				TradeDirection: dir,
				ExactIn:        exactIn,
			},
		}
	case instrRaydiumLaunchlabInitialize, instrRaydiumLaunchlabInitializeV2,
		instrRaydiumLaunchlabInitializeToken2022:
		if len(payload) < 1 {
			return DexEvent{}
		}
		decimals := payload[0]
		o := 1
		name, next, ok := readBorshString(payload, o)
		if !ok {
			return DexEvent{}
		}
		o = next
		symbol, next, ok := readBorshString(payload, o)
		if !ok {
			return DexEvent{}
		}
		o = next
		uri, _, ok := readBorshString(payload, o)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{
			Type: EventTypeRaydiumLaunchlabPoolCreate,
			Data: &RaydiumLaunchlabPoolCreateEvent{
				Metadata: meta,
				BaseMintParam: RaydiumLaunchlabMintParam{
					Symbol:   symbol,
					Name:     name,
					Uri:      uri,
					Decimals: decimals,
				},
				PoolState: getAccountSafe(accounts, 5),
				Creator:   getAccountSafe(accounts, 1),
			},
		}
	case instrRaydiumLaunchlabMigrateToAmm, instrRaydiumLaunchlabMigrateToCpswap:
		return DexEvent{}
	}

	return DexEvent{}
}

// getAccountSafe 安全获取账户地址
func getAccountSafe(accounts []string, index int) string {
	if index < 0 || index >= len(accounts) {
		return zeroPubkey
	}
	return accounts[index]
}
