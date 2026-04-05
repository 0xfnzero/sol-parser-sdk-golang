package solparser

import "encoding/binary"

func ParseLogUnified(log, signature string, slot uint64, blockTimeUs *int64) DexEvent {
	return ParseLogOptimized(log, signature, slot, 0, blockTimeUs, NowUs(), nil, false, "")
}

// ParseLogOptimized 与 Rust `parse_log_optimized` 等价（过滤器参数留作后续扩展）
func ParseLogOptimized(log, signature string, slot, txIndex uint64, blockTimeUs *int64, grpcRecvUs int64, _ any, isCreatedBuy bool, recentB58 string) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return nil
	}
	disc := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	meta := makeMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs, recentB58)

	switch disc {
	case discPumpTrade:
		return parseTradeFromData(data, meta, isCreatedBuy)
	case disc8(248, 198, 158, 145, 225, 117, 135, 200):
		return parseClmmSwapFromData(data, meta)
	case disc8(0, 0, 0, 0, 0, 0, 0, 9):
		return parseAmmSwapInFromData(data, meta)
	case disc8(103, 244, 82, 31, 44, 245, 119, 119):
		return parsePSBuyFromData(data, meta)
	case disc8(62, 47, 55, 10, 165, 3, 220, 42):
		return parsePSSellFromData(data, meta)
	case discPumpCreate:
		return parseCreateFromData(data, meta)
	case discPumpMigrate:
		return parseMigrateFromData(data, meta)
	case disc8(177, 49, 12, 210, 160, 118, 167, 116):
		return parsePSCreatePoolFromData(data, meta)
	case disc8(120, 248, 61, 83, 31, 142, 107, 144):
		return parsePSAddLiqFromData(data, meta)
	case disc8(22, 9, 133, 26, 160, 44, 71, 192):
		return parsePSRemoveLiqFromData(data, meta)
	case disc8(133, 29, 89, 223, 69, 238, 176, 10):
		return parseClmmIncFromData(data, meta)
	case disc8(160, 38, 208, 111, 104, 91, 44, 1):
		return parseClmmDecFromData(data, meta)
	case disc8(233, 146, 209, 142, 207, 104, 64, 188):
		return parseClmmCreateFromData(data, meta)
	case disc8(164, 152, 207, 99, 187, 104, 171, 119):
		return parseClmmCollectFromData(data, meta)
	case disc8(143, 190, 90, 218, 196, 30, 51, 222):
		return parseCpmmSwapInFromData(data, meta)
	case disc8(55, 217, 98, 86, 163, 74, 180, 173):
		return parseCpmmSwapOutFromData(data, meta)
	case disc8(242, 35, 198, 137, 82, 225, 242, 182):
		return parseCpmmDepositFromData(data, meta)
	case disc8(183, 18, 70, 156, 148, 109, 161, 34):
		return parseCpmmWithdrawFromData(data, meta)
	case disc8(0, 0, 0, 0, 0, 0, 0, 11):
		return parseAmmSwapOutFromData(data, meta)
	case disc8(0, 0, 0, 0, 0, 0, 0, 3):
		return parseAmmDepositFromData(data, meta)
	case disc8(0, 0, 0, 0, 0, 0, 0, 4):
		return parseAmmWithdrawFromData(data, meta)
	case disc8(0, 0, 0, 0, 0, 0, 0, 7):
		return parseAmmWithdrawPnlFromData(data, meta)
	case disc8(0, 0, 0, 0, 0, 0, 0, 1):
		return parseAmmInit2FromData(data, meta)
	case disc8(225, 202, 73, 175, 147, 43, 160, 150):
		return parseOrcaTradedFromData(data, meta)
	case disc8(30, 7, 144, 181, 102, 254, 155, 161):
		return parseOrcaLiqIncFromData(data, meta)
	case disc8(166, 1, 36, 71, 112, 202, 181, 171):
		return parseOrcaLiqDecFromData(data, meta)
	case disc8(100, 118, 173, 87, 12, 198, 254, 229):
		return parseOrcaPoolInitFromData(data, meta)
	case disc8(81, 108, 227, 190, 205, 208, 10, 196):
		return parseMeteoraSwapFromData(data, meta)
	case disc8(31, 94, 125, 90, 227, 52, 61, 186):
		return parseMeteoraAddFromData(data, meta)
	case disc8(116, 244, 97, 232, 103, 31, 152, 58):
		return parseMeteoraRemoveFromData(data, meta)
	case disc8(121, 127, 38, 136, 92, 55, 14, 247):
		return parseMeteoraBootstrapFromData(data, meta)
	case disc8(202, 44, 41, 88, 104, 220, 157, 82):
		return parseMeteoraPoolCreatedFromData(data, meta)
	case disc8(245, 26, 198, 164, 88, 18, 75, 9):
		return parseMeteoraPoolsSetPoolFeesFromData(data, meta)
	case discDammSwap, discDammSwap2,
		disc8(175, 242, 8, 157, 30, 247, 185, 169),
		disc8(87, 46, 88, 98, 175, 96, 34, 91),
		disc8(228, 50, 246, 85, 203, 66, 134, 37),
		disc8(156, 15, 119, 198, 29, 181, 221, 55),
		disc8(20, 145, 144, 68, 143, 142, 214, 178):
		return ParseMeteoraDammLog(log, signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	default:
		if ev := ParseBonkFromDiscriminator(disc, data, meta); ev != nil {
			return ev
		}
		return parseDlmmFromProgramData(buf, meta)
	}
}
