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

// ParseLogOptimized 超低延迟日志解析（与 Rust `parse_log_optimized` 等价）
// 使用预定义的 discriminator 常量，避免运行时计算
func ParseLogOptimized(log, signature string, slot, txIndex uint64, blockTimeUs *int64, grpcRecvUs int64, _ any, isCreatedBuy bool, recentB58 string) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return DexEvent{}
	}
	disc := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	meta := makeMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs, recentB58)

	// 热路径：PumpFun Trade（最频繁的事件）
	if disc == discPumpTrade {
		return parseTradeFromData(data, meta, isCreatedBuy)
	}

	// 热路径：Raydium CLMM Swap
	if disc == discClmmSwap {
		return parseClmmSwapFromData(data, meta)
	}

	// 热路径：Raydium AMM Swap In
	if disc == discAmmSwapIn {
		return parseAmmSwapInFromData(data, meta)
	}

	// 热路径：PumpSwap Buy/Sell
	if disc == discPSBuy {
		return parsePSBuyFromData(data, meta)
	}
	if disc == discPSSell {
		return parsePSSellFromData(data, meta)
	}

	// 其他事件类型使用 switch
	switch disc {
	// PumpFun
	case discPumpCreate:
		return parseCreateFromData(data, meta)
	case discPumpMigrate:
		return parseMigrateFromData(data, meta)

	// PumpSwap
	case discPSCreatePool:
		return parsePSCreatePoolFromData(data, meta)
	case discPSAddLiq:
		return parsePSAddLiqFromData(data, meta)
	case discPSRemLiq:
		return parsePSRemoveLiqFromData(data, meta)

	// Raydium CLMM
	case discClmmIncLiq:
		return parseClmmIncFromData(data, meta)
	case discClmmDecLiq:
		return parseClmmDecFromData(data, meta)
	case discClmmCreate:
		return parseClmmCreateFromData(data, meta)
	case discClmmCollect:
		return parseClmmCollectFromData(data, meta)

	// Raydium CPMM
	case discCpmmSwapIn:
		return parseCpmmSwapInFromData(data, meta)
	case discCpmmSwapOut:
		return parseCpmmSwapOutFromData(data, meta)
	case discCpmmDeposit:
		return parseCpmmDepositFromData(data, meta)
	case discCpmmWithdraw:
		return parseCpmmWithdrawFromData(data, meta)

	// Raydium AMM V4
	case discAmmSwapOut:
		return parseAmmSwapOutFromData(data, meta)
	case discAmmDeposit:
		return parseAmmDepositFromData(data, meta)
	case discAmmWithdraw:
		return parseAmmWithdrawFromData(data, meta)
	case discAmmWithdrawPnl:
		return parseAmmWithdrawPnlFromData(data, meta)
	case discAmmInit2:
		return parseAmmInit2FromData(data, meta)

	// Orca
	case discOrcaSwap:
		return parseOrcaTradedFromData(data, meta)
	case discOrcaIncLiq:
		return parseOrcaLiqIncFromData(data, meta)
	case discOrcaDecLiq:
		return parseOrcaLiqDecFromData(data, meta)
	case discOrcaPoolInit:
		return parseOrcaPoolInitFromData(data, meta)

	// Meteora Pools
	case discMeteoraSwap:
		return parseMeteoraSwapFromData(data, meta)
	case discMeteoraAdd:
		return parseMeteoraAddFromData(data, meta)
	case discMeteoraRemove:
		return parseMeteoraRemoveFromData(data, meta)
	case discMeteoraBootstrap:
		return parseMeteoraBootstrapFromData(data, meta)
	case discMeteoraPoolCreated:
		return parseMeteoraPoolCreatedFromData(data, meta)
	case discMeteoraSetPoolFees:
		return parseMeteoraPoolsSetPoolFeesFromData(data, meta)

	// Meteora DAMM v2
	case discDammSwap, discDammSwap2, discDammAdd, discDammRem, discDammInit, discDammCreate, discDammClose:
		return ParseMeteoraDammLog(log, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	default:
		// Bonk 事件
		if disc == discBonkTrade || disc == discBonkPoolCreate || disc == discBonkMigrateAmm {
			return ParseBonkFromDiscriminator(disc, data, meta)
		}
		// Meteora DLMM 事件
		return parseDlmmFromProgramData(buf, meta)
	}
}
