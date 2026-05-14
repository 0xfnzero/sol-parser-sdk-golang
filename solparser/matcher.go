package solparser

import (
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
	case discClmmCreate:
		return EventTypeRaydiumClmmCreatePool, true
	case discClmmCollect:
		return EventTypeRaydiumClmmCollectFee, true
	case discCpmmSwapIn, discCpmmSwapOut:
		return EventTypeRaydiumCpmmSwap, true
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
	case discBonkTrade:
		return EventTypeBonkTrade, true
	case discBonkPoolCreate:
		return EventTypeBonkPoolCreate, true
	case discBonkMigrateAmm:
		return EventTypeBonkMigrateAmm, true
	case dlmmSwap:
		return EventTypeMeteoraDlmmSwap, true
	case dlmmAddLiq:
		return EventTypeMeteoraDlmmAddLiquidity, true
	case dlmmRemoveLiq:
		return EventTypeMeteoraDlmmRemoveLiquidity, true
	case dlmmInitPool:
		return EventTypeMeteoraDlmmInitializePool, true
	case dlmmInitBin:
		return EventTypeMeteoraDlmmInitializeBinArray, true
	case dlmmCreatePos:
		return EventTypeMeteoraDlmmCreatePosition, true
	case dlmmClosePos:
		return EventTypeMeteoraDlmmClosePosition, true
	case dlmmClaimFee:
		return EventTypeMeteoraDlmmClaimFee, true
	default:
		return "", false
	}
}

func logFilterAllowsUnknown(filter EventTypeFilter) bool {
	if filter == nil {
		return true
	}
	_, ok := filter.(*IncludeOnlyFilter)
	return !ok
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

// ParseLogOptimized 超低延迟日志解析（与 Rust `parse_log_optimized` 等价）
// 使用预定义的 discriminator 常量，避免运行时计算
func ParseLogOptimized(log, signature string, slot, txIndex uint64, blockTimeUs *int64, grpcRecvUs int64, filter any, isCreatedBuy bool, recentB58 string) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return DexEvent{}
	}
	disc := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	meta := makeMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs, recentB58)
	eventFilter, _ := filter.(EventTypeFilter)
	if eventFilter != nil {
		if eventType, ok := logDiscriminatorEventType(disc); ok {
			if !eventFilter.ShouldInclude(eventType) {
				return DexEvent{}
			}
		} else if !logFilterAllowsUnknown(eventFilter) {
			return DexEvent{}
		}
	}

	// 热路径：PumpFun Trade（最频繁的事件）
	if disc == discPumpTrade {
		return applyActualEventTypeFilter(parseTradeFromData(data, meta, isCreatedBuy), eventFilter)
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
	case discClmmCreate:
		return applyActualEventTypeFilter(parseClmmCreateFromData(data, meta), eventFilter)
	case discClmmCollect:
		return applyActualEventTypeFilter(parseClmmCollectFromData(data, meta), eventFilter)

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
		// Bonk 事件
		if disc == discBonkTrade || disc == discBonkPoolCreate || disc == discBonkMigrateAmm {
			return applyActualEventTypeFilter(ParseBonkFromDiscriminator(disc, data, meta), eventFilter)
		}
		// Meteora DLMM 事件
		return applyActualEventTypeFilter(parseDlmmFromProgramData(buf, meta), eventFilter)
	}
}
