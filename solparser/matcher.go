package solparser

import (
	"encoding/base64"
	"encoding/binary"
	"strings"
)

// 与 Rust `optimized_matcher::detect_pumpfun_create` 一致：日志中出现 PumpFun Create 的 Program data 前缀。
const pumpfunCreateLogPrefix = "Program data: G3KpTd7rY3Y"

// DetectPumpfunCreateFromLogs 若任一日志行包含 PumpFun Create 的 base64 前缀则返回 true（用于 inner trade 的 is_created_buy，与 Rust `parse_instructions_enhanced` 一致）。
func DetectPumpfunCreateFromLogs(logs []string) bool {
	for _, log := range logs {
		if strings.Contains(log, pumpfunCreateLogPrefix) {
			return true
		}
	}
	return false
}

func ParseLogUnified(log, signature string, slot uint64, blockTimeUs *int64) DexEvent {
	return ParseLogOptimized(log, signature, slot, 0, blockTimeUs, NowUs(), nil, false, "")
}

func logDiscriminatorEventType(disc uint64) (EventType, bool) {
	switch disc {
	case discPumpCreate:
		return EventTypePumpFunCreate, true
	case discPumpTrade:
		return EventTypePumpFunTrade, true
	case discPumpMigrate:
		return EventTypePumpFunMigrate, true
	case discPumpMigrateBondingCurveCreator:
		return EventTypePumpFunMigrateBondingCurveCreator, true
	case discPumpFeesCreateFeeSharingConfig:
		return EventTypePumpFeesCreateFeeSharingConfig, true
	case discPumpFeesInitializeFeeConfig:
		return EventTypePumpFeesInitializeFeeConfig, true
	case discPumpFeesResetFeeSharingConfig:
		return EventTypePumpFeesResetFeeSharingConfig, true
	case discPumpFeesRevokeFeeSharingAuthority:
		return EventTypePumpFeesRevokeFeeSharingAuthority, true
	case discPumpFeesTransferFeeSharingAuthority:
		return EventTypePumpFeesTransferFeeSharingAuthority, true
	case discPumpFeesUpdateAdmin:
		return EventTypePumpFeesUpdateAdmin, true
	case discPumpFeesUpdateFeeConfig:
		return EventTypePumpFeesUpdateFeeConfig, true
	case discPumpFeesUpdateFeeShares:
		return EventTypePumpFeesUpdateFeeShares, true
	case discPumpFeesUpsertFeeTiers:
		return EventTypePumpFeesUpsertFeeTiers, true
	case discPSBuy:
		return EventTypePumpSwapBuy, true
	case discPSSell:
		return EventTypePumpSwapSell, true
	case discPSCreatePool:
		return EventTypePumpSwapCreatePool, true
	case discPSAddLiq:
		return EventTypePumpSwapLiquidityAdded, true
	case discPSRemLiq:
		return EventTypePumpSwapLiquidityRemoved, true
	case discClmmSwap:
		return EventTypeRaydiumClmmSwap, true
	case discClmmIncLiq:
		return EventTypeRaydiumClmmIncreaseLiquidity, true
	case discClmmDecLiq:
		return EventTypeRaydiumClmmDecreaseLiquidity, true
	case discClmmLiqChange:
		return EventTypeRaydiumClmmLiquidityChange, true
	case discClmmConfigChange:
		return EventTypeRaydiumClmmConfigChange, true
	case discClmmCreatePersonalPosition:
		return EventTypeRaydiumClmmCreatePersonalPosition, true
	case discClmmLiqCalculate:
		return EventTypeRaydiumClmmLiquidityCalculate, true
	case discClmmOpenLimitOrder:
		return EventTypeRaydiumClmmOpenLimitOrder, true
	case discClmmIncreaseLimitOrder:
		return EventTypeRaydiumClmmIncreaseLimitOrder, true
	case discClmmDecreaseLimitOrder:
		return EventTypeRaydiumClmmDecreaseLimitOrder, true
	case discClmmSettleLimitOrder:
		return EventTypeRaydiumClmmSettleLimitOrder, true
	case discClmmUpdateRewardInfos:
		return EventTypeRaydiumClmmUpdateRewardInfos, true
	case discClmmCreate:
		return EventTypeRaydiumClmmCreatePool, true
	case discClmmCollectPersonal, discClmmCollectProtocol:
		return EventTypeRaydiumClmmCollectFee, true
	case discCpmmSwapIn, discCpmmSwapOut:
		return EventTypeRaydiumCpmmSwap, true
	case discCpmmCreatePool:
		return EventTypeRaydiumCpmmInitialize, true
	case discCpmmDeposit:
		return EventTypeRaydiumCpmmDeposit, true
	case discCpmmWithdraw:
		return EventTypeRaydiumCpmmWithdraw, true
	case discAmmSwapIn, discAmmSwapOut:
		return EventTypeRaydiumAmmV4Swap, true
	case discAmmDeposit:
		return EventTypeRaydiumAmmV4Deposit, true
	case discAmmWithdraw:
		return EventTypeRaydiumAmmV4Withdraw, true
	case discAmmWithdrawPnl:
		return EventTypeRaydiumAmmV4WithdrawPnl, true
	case discAmmInit2:
		return EventTypeRaydiumAmmV4Initialize2, true
	case discOrcaSwap:
		return EventTypeOrcaWhirlpoolSwap, true
	case discOrcaIncLiq:
		return EventTypeOrcaWhirlpoolLiquidityIncreased, true
	case discOrcaDecLiq:
		return EventTypeOrcaWhirlpoolLiquidityDecreased, true
	case discOrcaPoolInit:
		return EventTypeOrcaWhirlpoolPoolInitialized, true
	case discMeteoraSwap:
		return EventTypeMeteoraPoolsSwap, true
	case discMeteoraAdd:
		return EventTypeMeteoraPoolsAddLiquidity, true
	case discMeteoraRemove:
		return EventTypeMeteoraPoolsRemoveLiquidity, true
	case discMeteoraBootstrap:
		return EventTypeMeteoraPoolsBootstrapLiquidity, true
	case discMeteoraPoolCreated:
		return EventTypeMeteoraPoolsPoolCreated, true
	case discMeteoraSetPoolFees:
		return EventTypeMeteoraPoolsSetPoolFees, true
	case discDammSwap, discDammSwap2:
		return EventTypeMeteoraDammV2Swap, true
	case discDammAdd:
		return EventTypeMeteoraDammV2AddLiquidity, true
	case discDammRem:
		return EventTypeMeteoraDammV2RemoveLiquidity, true
	case discDammInit:
		return EventTypeMeteoraDammV2InitializePool, true
	case discDammCreate:
		return EventTypeMeteoraDammV2CreatePosition, true
	case discDammClose:
		return EventTypeMeteoraDammV2ClosePosition, true
	case discRaydiumLaunchlabPoolCreate:
		return EventTypeRaydiumLaunchlabPoolCreate, true
	case dlmmSwap, dlmmSwap2, dlmmLegacySwap:
		return EventTypeMeteoraDlmmSwap, true
	case dlmmAddLiq, dlmmLegacyAddLiq:
		return EventTypeMeteoraDlmmAddLiquidity, true
	case dlmmRemoveLiq, dlmmLegacyRemoveLiq:
		return EventTypeMeteoraDlmmRemoveLiquidity, true
	case dlmmInitPool, dlmmLegacyInitPool:
		return EventTypeMeteoraDlmmInitializePool, true
	case dlmmInitBin:
		return EventTypeMeteoraDlmmInitializeBinArray, true
	case dlmmCreatePos, dlmmLegacyCreatePos:
		return EventTypeMeteoraDlmmCreatePosition, true
	case dlmmClosePos, dlmmLegacyClosePos:
		return EventTypeMeteoraDlmmClosePosition, true
	case dlmmClaimFee, dlmmClaimFee2, dlmmLegacyClaimFee:
		return EventTypeMeteoraDlmmClaimFee, true
	default:
		return "", false
	}
}

func programScopedLogDiscriminatorEventType(programID string, disc uint64) (EventType, bool) {
	switch programID {
	case PUMPFUN_PROGRAM_ID:
		switch disc {
		case discPumpCreate:
			return EventTypePumpFunCreate, true
		case discPumpTrade:
			return EventTypePumpFunTrade, true
		case discPumpMigrate:
			return EventTypePumpFunMigrate, true
		case discPumpMigrateBondingCurveCreator:
			return EventTypePumpFunMigrateBondingCurveCreator, true
		default:
			return "", false
		}
	case PUMP_FEES_PROGRAM_ID:
		switch disc {
		case discPumpFeesCreateFeeSharingConfig:
			return EventTypePumpFeesCreateFeeSharingConfig, true
		case discPumpFeesInitializeFeeConfig:
			return EventTypePumpFeesInitializeFeeConfig, true
		case discPumpFeesResetFeeSharingConfig:
			return EventTypePumpFeesResetFeeSharingConfig, true
		case discPumpFeesRevokeFeeSharingAuthority:
			return EventTypePumpFeesRevokeFeeSharingAuthority, true
		case discPumpFeesTransferFeeSharingAuthority:
			return EventTypePumpFeesTransferFeeSharingAuthority, true
		case discPumpFeesUpdateAdmin:
			return EventTypePumpFeesUpdateAdmin, true
		case discPumpFeesUpdateFeeConfig:
			return EventTypePumpFeesUpdateFeeConfig, true
		case discPumpFeesUpdateFeeShares:
			return EventTypePumpFeesUpdateFeeShares, true
		case discPumpFeesUpsertFeeTiers:
			return EventTypePumpFeesUpsertFeeTiers, true
		default:
			return "", false
		}
	case PUMPSWAP_PROGRAM_ID:
		switch disc {
		case discPSBuy:
			return EventTypePumpSwapBuy, true
		case discPSSell:
			return EventTypePumpSwapSell, true
		case discPSCreatePool:
			return EventTypePumpSwapCreatePool, true
		case discPSAddLiq:
			return EventTypePumpSwapLiquidityAdded, true
		case discPSRemLiq:
			return EventTypePumpSwapLiquidityRemoved, true
		default:
			return "", false
		}
	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		switch disc {
		case discRaydiumLaunchlabTrade:
			return EventTypeRaydiumLaunchlabTrade, true
		case discRaydiumLaunchlabPoolCreate:
			return EventTypeRaydiumLaunchlabPoolCreate, true
		default:
			return "", false
		}
	case RAYDIUM_CLMM_PROGRAM_ID:
		switch disc {
		case discClmmSwap:
			return EventTypeRaydiumClmmSwap, true
		case discClmmIncLiq:
			return EventTypeRaydiumClmmIncreaseLiquidity, true
		case discClmmDecLiq:
			return EventTypeRaydiumClmmDecreaseLiquidity, true
		case discClmmLiqChange:
			return EventTypeRaydiumClmmLiquidityChange, true
		case discClmmConfigChange:
			return EventTypeRaydiumClmmConfigChange, true
		case discClmmCreatePersonalPosition:
			return EventTypeRaydiumClmmCreatePersonalPosition, true
		case discClmmLiqCalculate:
			return EventTypeRaydiumClmmLiquidityCalculate, true
		case discClmmOpenLimitOrder:
			return EventTypeRaydiumClmmOpenLimitOrder, true
		case discClmmIncreaseLimitOrder:
			return EventTypeRaydiumClmmIncreaseLimitOrder, true
		case discClmmDecreaseLimitOrder:
			return EventTypeRaydiumClmmDecreaseLimitOrder, true
		case discClmmSettleLimitOrder:
			return EventTypeRaydiumClmmSettleLimitOrder, true
		case discClmmUpdateRewardInfos:
			return EventTypeRaydiumClmmUpdateRewardInfos, true
		case discClmmCreate:
			return EventTypeRaydiumClmmCreatePool, true
		case discClmmCollectPersonal, discClmmCollectProtocol:
			return EventTypeRaydiumClmmCollectFee, true
		default:
			return "", false
		}
	case RAYDIUM_CPMM_PROGRAM_ID:
		switch disc {
		case discCpmmSwapEvent, discCpmmSwapIn, discCpmmSwapOut:
			return EventTypeRaydiumCpmmSwap, true
		case discCpmmCreatePool:
			return EventTypeRaydiumCpmmInitialize, true
		case discCpmmDeposit:
			return EventTypeRaydiumCpmmDeposit, true
		case discCpmmWithdraw:
			return EventTypeRaydiumCpmmWithdraw, true
		default:
			return "", false
		}
	case RAYDIUM_AMM_V4_PROGRAM_ID:
		switch disc {
		case discAmmSwapIn, discAmmSwapOut:
			return EventTypeRaydiumAmmV4Swap, true
		case discAmmDeposit:
			return EventTypeRaydiumAmmV4Deposit, true
		case discAmmWithdraw:
			return EventTypeRaydiumAmmV4Withdraw, true
		case discAmmInit2:
			return EventTypeRaydiumAmmV4Initialize2, true
		case discAmmWithdrawPnl:
			return EventTypeRaydiumAmmV4WithdrawPnl, true
		default:
			return "", false
		}
	case ORCA_WHIRLPOOL_PROGRAM_ID:
		switch disc {
		case discOrcaSwap:
			return EventTypeOrcaWhirlpoolSwap, true
		case discOrcaIncLiq:
			return EventTypeOrcaWhirlpoolLiquidityIncreased, true
		case discOrcaDecLiq:
			return EventTypeOrcaWhirlpoolLiquidityDecreased, true
		case discOrcaPoolInit:
			return EventTypeOrcaWhirlpoolPoolInitialized, true
		default:
			return "", false
		}
	case METEORA_POOLS_PROGRAM_ID:
		switch disc {
		case discMeteoraSwap:
			return EventTypeMeteoraPoolsSwap, true
		case discMeteoraAdd:
			return EventTypeMeteoraPoolsAddLiquidity, true
		case discMeteoraRemove:
			return EventTypeMeteoraPoolsRemoveLiquidity, true
		case discMeteoraBootstrap:
			return EventTypeMeteoraPoolsBootstrapLiquidity, true
		case discMeteoraPoolCreated:
			return EventTypeMeteoraPoolsPoolCreated, true
		case discMeteoraSetPoolFees:
			return EventTypeMeteoraPoolsSetPoolFees, true
		default:
			return "", false
		}
	case METEORA_DAMM_V2_PROGRAM_ID:
		switch disc {
		case discDammSwap, discDammSwap2:
			return EventTypeMeteoraDammV2Swap, true
		case discDammAdd:
			return EventTypeMeteoraDammV2AddLiquidity, true
		case discDammRem:
			return EventTypeMeteoraDammV2RemoveLiquidity, true
		case discDammInit:
			return EventTypeMeteoraDammV2InitializePool, true
		case discDammCreate:
			return EventTypeMeteoraDammV2CreatePosition, true
		case discDammClose:
			return EventTypeMeteoraDammV2ClosePosition, true
		default:
			return "", false
		}
	case METEORA_DBC_PROGRAM_ID:
		switch disc {
		case discDbcSwap:
			return EventTypeMeteoraDbcSwap, true
		case discDbcInit:
			return EventTypeMeteoraDbcInitializePool, true
		case discDbcCurve:
			return EventTypeMeteoraDbcCurveComplete, true
		default:
			return "", false
		}
	case METEORA_DLMM_PROGRAM_ID:
		switch disc {
		case dlmmSwap, dlmmSwap2, dlmmLegacySwap:
			return EventTypeMeteoraDlmmSwap, true
		case dlmmAddLiq, dlmmLegacyAddLiq:
			return EventTypeMeteoraDlmmAddLiquidity, true
		case dlmmRemoveLiq, dlmmLegacyRemoveLiq:
			return EventTypeMeteoraDlmmRemoveLiquidity, true
		case dlmmInitPool, dlmmLegacyInitPool:
			return EventTypeMeteoraDlmmInitializePool, true
		case dlmmInitBin:
			return EventTypeMeteoraDlmmInitializeBinArray, true
		case dlmmCreatePos, dlmmLegacyCreatePos:
			return EventTypeMeteoraDlmmCreatePosition, true
		case dlmmClosePos, dlmmLegacyClosePos:
			return EventTypeMeteoraDlmmClosePosition, true
		case dlmmClaimFee, dlmmClaimFee2, dlmmLegacyClaimFee:
			return EventTypeMeteoraDlmmClaimFee, true
		default:
			return "", false
		}
	default:
		return logDiscriminatorEventType(disc)
	}
}

func filterIncludesProgram(filter EventTypeFilter, programID string) bool {
	if filter == nil {
		return true
	}
	switch programID {
	case PUMPFUN_PROGRAM_ID:
		return EventTypeFilterIncludesPumpfun(filter)
	case PUMP_FEES_PROGRAM_ID:
		return EventTypeFilterIncludesPumpFees(filter)
	case PUMPSWAP_PROGRAM_ID:
		return EventTypeFilterIncludesPumpswap(filter)
	case RAYDIUM_LAUNCHLAB_PROGRAM_ID:
		return EventTypeFilterIncludesRaydiumLaunchlab(filter)
	case RAYDIUM_CLMM_PROGRAM_ID:
		return EventTypeFilterIncludesRaydiumClmm(filter)
	case RAYDIUM_CPMM_PROGRAM_ID:
		return EventTypeFilterIncludesRaydiumCpmm(filter)
	case RAYDIUM_AMM_V4_PROGRAM_ID:
		return EventTypeFilterIncludesRaydiumAmmV4(filter)
	case ORCA_WHIRLPOOL_PROGRAM_ID:
		return EventTypeFilterIncludesOrcaWhirlpool(filter)
	case METEORA_POOLS_PROGRAM_ID:
		return EventTypeFilterIncludesMeteoraPools(filter)
	case METEORA_DAMM_V2_PROGRAM_ID:
		return EventTypeFilterIncludesMeteoraDammV2(filter)
	case METEORA_DBC_PROGRAM_ID:
		return EventTypeFilterIncludesMeteoraDbc(filter)
	case METEORA_DLMM_PROGRAM_ID:
		return EventTypeFilterIncludesMeteoraDlmm(filter)
	default:
		return logFilterAllowsUnknown(filter)
	}
}

func logFilterAllowsUnknown(filter EventTypeFilter) bool {
	if filter == nil {
		return true
	}
	_, ok := filter.(*IncludeOnlyFilter)
	return !ok
}

func filterWantsSupportedLogs(filter EventTypeFilter) bool {
	return filterIncludesProgram(filter, PUMPFUN_PROGRAM_ID) ||
		filterIncludesProgram(filter, PUMP_FEES_PROGRAM_ID) ||
		filterIncludesProgram(filter, PUMPSWAP_PROGRAM_ID) ||
		filterIncludesProgram(filter, RAYDIUM_LAUNCHLAB_PROGRAM_ID) ||
		filterIncludesProgram(filter, RAYDIUM_CLMM_PROGRAM_ID) ||
		filterIncludesProgram(filter, RAYDIUM_CPMM_PROGRAM_ID) ||
		filterIncludesProgram(filter, RAYDIUM_AMM_V4_PROGRAM_ID) ||
		filterIncludesProgram(filter, ORCA_WHIRLPOOL_PROGRAM_ID) ||
		filterIncludesProgram(filter, METEORA_POOLS_PROGRAM_ID) ||
		filterIncludesProgram(filter, METEORA_DAMM_V2_PROGRAM_ID) ||
		filterIncludesProgram(filter, METEORA_DLMM_PROGRAM_ID) ||
		filterIncludesProgram(filter, METEORA_DBC_PROGRAM_ID)
}

func applyActualEventTypeFilter(ev DexEvent, filter EventTypeFilter) DexEvent {
	if ev.Type == "" || filter == nil {
		return ev
	}
	if !filter.ShouldInclude(ev.Type) {
		return DexEvent{}
	}
	return ev
}

func pumpfunTradeMatchesIncludeOnly(ev DexEvent, includeOnly []EventType) bool {
	switch ev.Type {
	case EventTypePumpFunBuy:
		return eventTypeSliceContains(includeOnly, EventTypePumpFunBuy) ||
			eventTypeSliceContains(includeOnly, EventTypePumpFunBuyExactSolIn)
	case EventTypePumpFunSell:
		return eventTypeSliceContains(includeOnly, EventTypePumpFunSell)
	case EventTypePumpFunBuyExactSolIn:
		return eventTypeSliceContains(includeOnly, EventTypePumpFunBuy) ||
			eventTypeSliceContains(includeOnly, EventTypePumpFunBuyExactSolIn)
	case EventTypePumpFunTrade:
		return eventTypeSliceContains(includeOnly, EventTypePumpFunTrade)
	case EventTypePumpFunCreate, EventTypePumpFunCreateV2:
		return eventTypeSliceContains(includeOnly, EventTypePumpFunCreate) ||
			eventTypeSliceContains(includeOnly, EventTypePumpFunCreateV2)
	default:
		return false
	}
}

func applyPumpfunSecondaryFilter(ev DexEvent, filter EventTypeFilter) DexEvent {
	if ev.Type == "" || filter == nil {
		return ev
	}
	includeOnly, ok := filter.(*IncludeOnlyFilter)
	if ok {
		hasSpecific := false
		for _, t := range includeOnly.IncludeOnly {
			switch t {
			case EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn,
				EventTypePumpFunCreate, EventTypePumpFunCreateV2:
				hasSpecific = true
			}
		}
		if hasSpecific && !pumpfunTradeMatchesIncludeOnly(ev, includeOnly.IncludeOnly) {
			return DexEvent{}
		}
	}
	return applyActualEventTypeFilter(ev, filter)
}

func filterWantsPumpfunTrade(filter EventTypeFilter) bool {
	if filter == nil {
		return true
	}
	return filter.ShouldInclude(EventTypePumpFunTrade) ||
		filter.ShouldInclude(EventTypePumpFunBuy) ||
		filter.ShouldInclude(EventTypePumpFunSell) ||
		filter.ShouldInclude(EventTypePumpFunBuyExactSolIn)
}

func filterWantsRaydiumLaunchlabTrade(filter EventTypeFilter) bool {
	return filter == nil || filter.ShouldInclude(EventTypeRaydiumLaunchlabTrade)
}

func filterAllowsUnscopedDiscriminator(filter EventTypeFilter, disc uint64) bool {
	if filter == nil {
		return true
	}
	switch disc {
	case discPumpTrade:
		return filterWantsPumpfunTrade(filter) || filterWantsRaydiumLaunchlabTrade(filter)
	case discCpmmSwapIn:
		return filter.ShouldInclude(EventTypeRaydiumCpmmSwap) ||
			filter.ShouldInclude(EventTypeMeteoraDlmmSwap)
	default:
		if eventType, ok := logDiscriminatorEventType(disc); ok {
			return filter.ShouldInclude(eventType)
		}
		return filterWantsSupportedLogs(filter)
	}
}

func parseUnscopedPumpfunLaunchlabTrade(data []byte, meta EventMetadata, filter EventTypeFilter, isCreatedBuy bool) DexEvent {
	if filterWantsPumpfunTrade(filter) {
		if ev := applyPumpfunSecondaryFilter(parseTradeFromData(data, meta, isCreatedBuy), filter); ev.Type != "" {
			return ev
		}
	}
	if filterWantsRaydiumLaunchlabTrade(filter) {
		return applyActualEventTypeFilter(ParseRaydiumLaunchlabFromDiscriminator(discPumpTrade, data, meta), filter)
	}
	return DexEvent{}
}

func parseScopedPumpfunData(disc uint64, data []byte, meta EventMetadata, filter EventTypeFilter, isCreatedBuy bool) DexEvent {
	switch disc {
	case discPumpTrade:
		return applyPumpfunSecondaryFilter(parseTradeFromData(data, meta, isCreatedBuy), filter)
	case discPumpCreate:
		return applyActualEventTypeFilter(parseCreateFromData(data, meta), filter)
	case discPumpMigrate:
		return applyActualEventTypeFilter(parseMigrateFromData(data, meta), filter)
	case discPumpMigrateBondingCurveCreator:
		return applyActualEventTypeFilter(parseMigrateBondingCurveCreatorFromData(data, meta), filter)
	default:
		return DexEvent{}
	}
}

func parseScopedPumpFeesData(disc uint64, data []byte, meta EventMetadata, filter EventTypeFilter) DexEvent {
	switch disc {
	case discPumpFeesCreateFeeSharingConfig:
		return applyActualEventTypeFilter(parsePumpFeesCreateFeeSharingConfigFromData(data, meta), filter)
	case discPumpFeesInitializeFeeConfig:
		return applyActualEventTypeFilter(parsePumpFeesInitializeFeeConfigFromData(data, meta), filter)
	case discPumpFeesResetFeeSharingConfig:
		return applyActualEventTypeFilter(parsePumpFeesResetFeeSharingConfigFromData(data, meta), filter)
	case discPumpFeesRevokeFeeSharingAuthority:
		return applyActualEventTypeFilter(parsePumpFeesRevokeFeeSharingAuthorityFromData(data, meta), filter)
	case discPumpFeesTransferFeeSharingAuthority:
		return applyActualEventTypeFilter(parsePumpFeesTransferFeeSharingAuthorityFromData(data, meta), filter)
	case discPumpFeesUpdateAdmin:
		return applyActualEventTypeFilter(parsePumpFeesUpdateAdminFromData(data, meta), filter)
	case discPumpFeesUpdateFeeConfig:
		return applyActualEventTypeFilter(parsePumpFeesUpdateFeeConfigFromData(data, meta), filter)
	case discPumpFeesUpdateFeeShares:
		return applyActualEventTypeFilter(parsePumpFeesUpdateFeeSharesFromData(data, meta), filter)
	case discPumpFeesUpsertFeeTiers:
		return applyActualEventTypeFilter(parsePumpFeesUpsertFeeTiersFromData(data, meta), filter)
	default:
		return DexEvent{}
	}
}

func parseScopedPumpswapData(disc uint64, data []byte, meta EventMetadata, filter EventTypeFilter) DexEvent {
	switch disc {
	case discPSBuy:
		return applyActualEventTypeFilter(parsePSBuyFromData(data, meta), filter)
	case discPSSell:
		return applyActualEventTypeFilter(parsePSSellFromData(data, meta), filter)
	case discPSCreatePool:
		return applyActualEventTypeFilter(parsePSCreatePoolFromData(data, meta), filter)
	case discPSAddLiq:
		return applyActualEventTypeFilter(parsePSAddLiqFromData(data, meta), filter)
	case discPSRemLiq:
		return applyActualEventTypeFilter(parsePSRemoveLiqFromData(data, meta), filter)
	default:
		return DexEvent{}
	}
}

// ParseLogOptimized 超低延迟日志解析（与 Rust `parse_log_optimized` 等价）
// 使用预定义的 discriminator 常量，避免运行时计算
func ParseLogOptimized(log, signature string, slot, txIndex uint64, blockTimeUs *int64, grpcRecvUs int64, filter any, isCreatedBuy bool, recentB58 string) DexEvent {
	return ParseLogOptimizedWithProgramID(log, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, isCreatedBuy, recentB58, "")
}

func ParseLogOptimizedWithProgramID(log, signature string, slot, txIndex uint64, blockTimeUs *int64, grpcRecvUs int64, filter any, isCreatedBuy bool, recentB58 string, programID string) DexEvent {
	eventFilter, _ := filter.(EventTypeFilter)
	if programID == RAYDIUM_AMM_V4_PROGRAM_ID {
		if start := strings.Index(log, "ray_log: "); start >= 0 {
			if eventFilter != nil && !eventFilter.ShouldInclude(EventTypeRaydiumAmmV4Swap) {
				return DexEvent{}
			}
			encoded := strings.TrimSpace(log[start+len("ray_log: "):])
			data, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return DexEvent{}
			}
			meta := makeMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs, recentB58)
			return parseAmmRayLogSwap(data, meta)
		}
	}
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return DexEvent{}
	}
	disc := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	meta := makeMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs, recentB58)
	if eventFilter != nil {
		unscopedShared := programID == "" && (disc == discPumpTrade || disc == discCpmmSwapIn)
		eventType, ok := logDiscriminatorEventType(disc)
		if programID != "" {
			eventType, ok = programScopedLogDiscriminatorEventType(programID, disc)
		}
		if unscopedShared {
			if !filterAllowsUnscopedDiscriminator(eventFilter, disc) {
				return DexEvent{}
			}
		} else if programID == PUMPFUN_PROGRAM_ID && disc == discPumpTrade {
			if !filterWantsPumpfunTrade(eventFilter) {
				return DexEvent{}
			}
		} else if ok {
			if !eventFilter.ShouldInclude(eventType) {
				return DexEvent{}
			}
		} else if programID != "" && !filterIncludesProgram(eventFilter, programID) {
			return DexEvent{}
		} else if !filterAllowsUnscopedDiscriminator(eventFilter, disc) {
			return DexEvent{}
		}
	}

	if programID == RAYDIUM_LAUNCHLAB_PROGRAM_ID {
		return applyActualEventTypeFilter(ParseRaydiumLaunchlabFromDiscriminator(disc, data, meta), eventFilter)
	}
	if programID == RAYDIUM_CLMM_PROGRAM_ID {
		switch disc {
		case discClmmSwap:
			return applyActualEventTypeFilter(parseClmmSwapFromData(data, meta), eventFilter)
		case discClmmIncLiq:
			return applyActualEventTypeFilter(parseClmmIncFromData(data, meta), eventFilter)
		case discClmmDecLiq:
			return applyActualEventTypeFilter(parseClmmDecFromData(data, meta), eventFilter)
		case discClmmLiqChange:
			return applyActualEventTypeFilter(parseClmmLiquidityChangeFromData(data, meta), eventFilter)
		case discClmmConfigChange:
			return applyActualEventTypeFilter(parseClmmConfigChangeFromData(data, meta), eventFilter)
		case discClmmCreatePersonalPosition:
			return applyActualEventTypeFilter(parseClmmCreatePersonalPositionFromData(data, meta), eventFilter)
		case discClmmLiqCalculate:
			return applyActualEventTypeFilter(parseClmmLiquidityCalculateFromData(data, meta), eventFilter)
		case discClmmOpenLimitOrder:
			return applyActualEventTypeFilter(parseClmmOpenLimitOrderFromData(data, meta), eventFilter)
		case discClmmIncreaseLimitOrder:
			return applyActualEventTypeFilter(parseClmmIncreaseLimitOrderFromData(data, meta), eventFilter)
		case discClmmDecreaseLimitOrder:
			return applyActualEventTypeFilter(parseClmmDecreaseLimitOrderFromData(data, meta), eventFilter)
		case discClmmSettleLimitOrder:
			return applyActualEventTypeFilter(parseClmmSettleLimitOrderFromData(data, meta), eventFilter)
		case discClmmUpdateRewardInfos:
			return applyActualEventTypeFilter(parseClmmUpdateRewardInfosFromData(data, meta), eventFilter)
		case discClmmCreate:
			return applyActualEventTypeFilter(parseClmmCreateFromData(data, meta), eventFilter)
		case discClmmCollectPersonal:
			return applyActualEventTypeFilter(parseClmmCollectPersonalFromData(data, meta), eventFilter)
		case discClmmCollectProtocol:
			return applyActualEventTypeFilter(parseClmmCollectProtocolFromData(data, meta), eventFilter)
		default:
			return DexEvent{}
		}
	}
	if programID == RAYDIUM_CPMM_PROGRAM_ID {
		switch disc {
		case discCpmmSwapEvent:
			return applyActualEventTypeFilter(parseCpmmSwapEventFromData(data, meta), eventFilter)
		case discCpmmSwapIn:
			return applyActualEventTypeFilter(parseCpmmSwapInFromData(data, meta), eventFilter)
		case discCpmmSwapOut:
			return applyActualEventTypeFilter(parseCpmmSwapOutFromData(data, meta), eventFilter)
		case discCpmmCreatePool:
			return applyActualEventTypeFilter(parseCpmmInitFromData(data, meta), eventFilter)
		case discCpmmDeposit:
			return applyActualEventTypeFilter(parseCpmmDepositFromData(data, meta), eventFilter)
		case discCpmmWithdraw:
			return applyActualEventTypeFilter(parseCpmmWithdrawFromData(data, meta), eventFilter)
		default:
			return DexEvent{}
		}
	}
	if programID == RAYDIUM_AMM_V4_PROGRAM_ID {
		switch disc {
		case discAmmSwapIn:
			return applyActualEventTypeFilter(parseAmmSwapInFromData(data, meta), eventFilter)
		case discAmmSwapOut:
			return applyActualEventTypeFilter(parseAmmSwapOutFromData(data, meta), eventFilter)
		case discAmmDeposit:
			return applyActualEventTypeFilter(parseAmmDepositFromData(data, meta), eventFilter)
		case discAmmWithdraw:
			return applyActualEventTypeFilter(parseAmmWithdrawFromData(data, meta), eventFilter)
		case discAmmInit2:
			return applyActualEventTypeFilter(parseAmmInit2FromData(data, meta), eventFilter)
		case discAmmWithdrawPnl:
			return applyActualEventTypeFilter(parseAmmWithdrawPnlFromData(data, meta), eventFilter)
		default:
			return DexEvent{}
		}
	}
	if programID == ORCA_WHIRLPOOL_PROGRAM_ID {
		switch disc {
		case discOrcaSwap:
			return applyActualEventTypeFilter(parseOrcaTradedFromData(data, meta), eventFilter)
		case discOrcaIncLiq:
			return applyActualEventTypeFilter(parseOrcaLiqIncFromData(data, meta), eventFilter)
		case discOrcaDecLiq:
			return applyActualEventTypeFilter(parseOrcaLiqDecFromData(data, meta), eventFilter)
		case discOrcaPoolInit:
			return applyActualEventTypeFilter(parseOrcaPoolInitFromData(data, meta), eventFilter)
		default:
			return DexEvent{}
		}
	}
	if programID == METEORA_POOLS_PROGRAM_ID {
		switch disc {
		case discMeteoraSwap:
			return applyActualEventTypeFilter(parseMeteoraSwapFromData(data, meta), eventFilter)
		case discMeteoraAdd:
			return applyActualEventTypeFilter(parseMeteoraAddFromData(data, meta), eventFilter)
		case discMeteoraRemove:
			return applyActualEventTypeFilter(parseMeteoraRemoveFromData(data, meta), eventFilter)
		case discMeteoraBootstrap:
			return applyActualEventTypeFilter(parseMeteoraBootstrapFromData(data, meta), eventFilter)
		case discMeteoraPoolCreated:
			return applyActualEventTypeFilter(parseMeteoraPoolCreatedFromData(data, meta), eventFilter)
		case discMeteoraSetPoolFees:
			return applyActualEventTypeFilter(parseMeteoraPoolsSetPoolFeesFromData(data, meta), eventFilter)
		default:
			return DexEvent{}
		}
	}
	if programID == METEORA_DAMM_V2_PROGRAM_ID {
		return applyActualEventTypeFilter(ParseMeteoraDammLog(log, signature, slot, txIndex, blockTimeUs, grpcRecvUs), eventFilter)
	}
	if programID == METEORA_DBC_PROGRAM_ID {
		return applyActualEventTypeFilter(parseMeteoraDbcFromDiscriminator(disc, data, meta), eventFilter)
	}
	if programID == METEORA_DLMM_PROGRAM_ID {
		return applyActualEventTypeFilter(parseDlmmFromProgramData(buf, meta), eventFilter)
	}
	if programID == PUMPFUN_PROGRAM_ID {
		return parseScopedPumpfunData(disc, data, meta, eventFilter, isCreatedBuy)
	}
	if programID == PUMP_FEES_PROGRAM_ID {
		return parseScopedPumpFeesData(disc, data, meta, eventFilter)
	}
	if programID == PUMPSWAP_PROGRAM_ID {
		return parseScopedPumpswapData(disc, data, meta, eventFilter)
	}

	// 热路径：PumpFun Trade（最频繁的事件）
	if disc == discPumpTrade {
		return parseUnscopedPumpfunLaunchlabTrade(data, meta, eventFilter, isCreatedBuy)
	}

	// 热路径：Raydium CLMM Swap
	if disc == discClmmSwap {
		return applyActualEventTypeFilter(parseClmmSwapFromData(data, meta), eventFilter)
	}

	// 热路径：Raydium AMM Swap In
	if disc == discAmmSwapIn {
		return applyActualEventTypeFilter(parseAmmSwapInFromData(data, meta), eventFilter)
	}

	// 热路径：PumpSwap Buy/Sell
	if disc == discPSBuy {
		return applyActualEventTypeFilter(parsePSBuyFromData(data, meta), eventFilter)
	}
	if disc == discPSSell {
		return applyActualEventTypeFilter(parsePSSellFromData(data, meta), eventFilter)
	}

	// 其他事件类型使用 switch
	switch disc {
	// PumpFun
	case discPumpCreate:
		return applyActualEventTypeFilter(parseCreateFromData(data, meta), eventFilter)
	case discPumpMigrate:
		return applyActualEventTypeFilter(parseMigrateFromData(data, meta), eventFilter)
	case discPumpMigrateBondingCurveCreator:
		return applyActualEventTypeFilter(parseMigrateBondingCurveCreatorFromData(data, meta), eventFilter)
	case discPumpFeesCreateFeeSharingConfig:
		return applyActualEventTypeFilter(parsePumpFeesCreateFeeSharingConfigFromData(data, meta), eventFilter)
	case discPumpFeesInitializeFeeConfig:
		return applyActualEventTypeFilter(parsePumpFeesInitializeFeeConfigFromData(data, meta), eventFilter)
	case discPumpFeesResetFeeSharingConfig:
		return applyActualEventTypeFilter(parsePumpFeesResetFeeSharingConfigFromData(data, meta), eventFilter)
	case discPumpFeesRevokeFeeSharingAuthority:
		return applyActualEventTypeFilter(parsePumpFeesRevokeFeeSharingAuthorityFromData(data, meta), eventFilter)
	case discPumpFeesTransferFeeSharingAuthority:
		return applyActualEventTypeFilter(parsePumpFeesTransferFeeSharingAuthorityFromData(data, meta), eventFilter)
	case discPumpFeesUpdateAdmin:
		return applyActualEventTypeFilter(parsePumpFeesUpdateAdminFromData(data, meta), eventFilter)
	case discPumpFeesUpdateFeeConfig:
		return applyActualEventTypeFilter(parsePumpFeesUpdateFeeConfigFromData(data, meta), eventFilter)
	case discPumpFeesUpdateFeeShares:
		return applyActualEventTypeFilter(parsePumpFeesUpdateFeeSharesFromData(data, meta), eventFilter)
	case discPumpFeesUpsertFeeTiers:
		return applyActualEventTypeFilter(parsePumpFeesUpsertFeeTiersFromData(data, meta), eventFilter)

	// PumpSwap
	case discPSCreatePool:
		return applyActualEventTypeFilter(parsePSCreatePoolFromData(data, meta), eventFilter)
	case discPSAddLiq:
		return applyActualEventTypeFilter(parsePSAddLiqFromData(data, meta), eventFilter)
	case discPSRemLiq:
		return applyActualEventTypeFilter(parsePSRemoveLiqFromData(data, meta), eventFilter)

	// Raydium CLMM
	case discClmmIncLiq:
		return applyActualEventTypeFilter(parseClmmIncFromData(data, meta), eventFilter)
	case discClmmDecLiq:
		return applyActualEventTypeFilter(parseClmmDecFromData(data, meta), eventFilter)
	case discClmmLiqChange:
		return applyActualEventTypeFilter(parseClmmLiquidityChangeFromData(data, meta), eventFilter)
	case discClmmConfigChange:
		return applyActualEventTypeFilter(parseClmmConfigChangeFromData(data, meta), eventFilter)
	case discClmmCreatePersonalPosition:
		return applyActualEventTypeFilter(parseClmmCreatePersonalPositionFromData(data, meta), eventFilter)
	case discClmmLiqCalculate:
		return applyActualEventTypeFilter(parseClmmLiquidityCalculateFromData(data, meta), eventFilter)
	case discClmmOpenLimitOrder:
		return applyActualEventTypeFilter(parseClmmOpenLimitOrderFromData(data, meta), eventFilter)
	case discClmmIncreaseLimitOrder:
		return applyActualEventTypeFilter(parseClmmIncreaseLimitOrderFromData(data, meta), eventFilter)
	case discClmmDecreaseLimitOrder:
		return applyActualEventTypeFilter(parseClmmDecreaseLimitOrderFromData(data, meta), eventFilter)
	case discClmmSettleLimitOrder:
		return applyActualEventTypeFilter(parseClmmSettleLimitOrderFromData(data, meta), eventFilter)
	case discClmmUpdateRewardInfos:
		return applyActualEventTypeFilter(parseClmmUpdateRewardInfosFromData(data, meta), eventFilter)
	case discClmmCreate:
		return applyActualEventTypeFilter(parseClmmCreateFromData(data, meta), eventFilter)
	case discClmmCollectPersonal:
		return applyActualEventTypeFilter(parseClmmCollectPersonalFromData(data, meta), eventFilter)
	case discClmmCollectProtocol:
		return applyActualEventTypeFilter(parseClmmCollectProtocolFromData(data, meta), eventFilter)

	// Raydium CPMM
	case discCpmmSwapIn:
		return applyActualEventTypeFilter(parseCpmmSwapInFromData(data, meta), eventFilter)
	case discCpmmSwapOut:
		return applyActualEventTypeFilter(parseCpmmSwapOutFromData(data, meta), eventFilter)
	case discCpmmDeposit:
		return applyActualEventTypeFilter(parseCpmmDepositFromData(data, meta), eventFilter)
	case discCpmmWithdraw:
		return applyActualEventTypeFilter(parseCpmmWithdrawFromData(data, meta), eventFilter)

	// Raydium AMM V4
	case discAmmSwapOut:
		return applyActualEventTypeFilter(parseAmmSwapOutFromData(data, meta), eventFilter)
	case discAmmDeposit:
		return applyActualEventTypeFilter(parseAmmDepositFromData(data, meta), eventFilter)
	case discAmmWithdraw:
		return applyActualEventTypeFilter(parseAmmWithdrawFromData(data, meta), eventFilter)
	case discAmmWithdrawPnl:
		return applyActualEventTypeFilter(parseAmmWithdrawPnlFromData(data, meta), eventFilter)
	case discAmmInit2:
		return applyActualEventTypeFilter(parseAmmInit2FromData(data, meta), eventFilter)

	// Orca
	case discOrcaSwap:
		return applyActualEventTypeFilter(parseOrcaTradedFromData(data, meta), eventFilter)
	case discOrcaIncLiq:
		return applyActualEventTypeFilter(parseOrcaLiqIncFromData(data, meta), eventFilter)
	case discOrcaDecLiq:
		return applyActualEventTypeFilter(parseOrcaLiqDecFromData(data, meta), eventFilter)
	case discOrcaPoolInit:
		return applyActualEventTypeFilter(parseOrcaPoolInitFromData(data, meta), eventFilter)

	// Meteora Pools
	case discMeteoraSwap:
		return applyActualEventTypeFilter(parseMeteoraSwapFromData(data, meta), eventFilter)
	case discMeteoraAdd:
		return applyActualEventTypeFilter(parseMeteoraAddFromData(data, meta), eventFilter)
	case discMeteoraRemove:
		return applyActualEventTypeFilter(parseMeteoraRemoveFromData(data, meta), eventFilter)
	case discMeteoraBootstrap:
		return applyActualEventTypeFilter(parseMeteoraBootstrapFromData(data, meta), eventFilter)
	case discMeteoraPoolCreated:
		return applyActualEventTypeFilter(parseMeteoraPoolCreatedFromData(data, meta), eventFilter)
	case discMeteoraSetPoolFees:
		return applyActualEventTypeFilter(parseMeteoraPoolsSetPoolFeesFromData(data, meta), eventFilter)

	// Meteora DAMM v2
	case discDammSwap, discDammSwap2, discDammAdd, discDammRem, discDammInit, discDammCreate, discDammClose:
		return applyActualEventTypeFilter(ParseMeteoraDammLog(log, signature, slot, txIndex, blockTimeUs, grpcRecvUs), eventFilter)

	default:
		// RaydiumLaunchlab 事件；Trade discriminator 与 PumpFunTrade 相同，需依赖上游 program context。
		if disc == discRaydiumLaunchlabPoolCreate {
			return applyActualEventTypeFilter(ParseRaydiumLaunchlabFromDiscriminator(disc, data, meta), eventFilter)
		}
		// Meteora DLMM 事件
		return applyActualEventTypeFilter(parseDlmmFromProgramData(buf, meta), eventFilter)
	}
}

func ParseInvokeInfo(log string) (string, int, bool) {
	const prefix = "Program "
	start := strings.Index(log, prefix)
	if start < 0 {
		return "", 0, false
	}
	const marker = " invoke ["
	mid := strings.Index(log[start+len(prefix):], marker)
	if mid < 0 {
		return "", 0, false
	}
	mid += start + len(prefix)
	programID := log[start+len(prefix) : mid]
	depthStart := mid + len(marker)
	depthEndRel := strings.Index(log[depthStart:], "]")
	if depthEndRel < 0 {
		return "", 0, false
	}
	depth := 0
	for _, ch := range log[depthStart : depthStart+depthEndRel] {
		if ch < '0' || ch > '9' {
			return "", 0, false
		}
		depth = depth*10 + int(ch-'0')
	}
	if programID == "" || depth <= 0 {
		return "", 0, false
	}
	return programID, depth, true
}

func ParseProgramCompleteInfo(log string) (string, bool) {
	const prefix = "Program "
	start := strings.Index(log, prefix)
	if start < 0 {
		return "", false
	}
	restStart := start + len(prefix)
	if idx := strings.Index(log[restStart:], " success"); idx >= 0 {
		return log[restStart : restStart+idx], true
	}
	if idx := strings.Index(log[restStart:], " failed:"); idx >= 0 {
		return log[restStart : restStart+idx], true
	}
	return "", false
}
