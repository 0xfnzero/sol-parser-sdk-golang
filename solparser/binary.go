package solparser

import (
	"encoding/binary"
	"math/big"
	"sync"
)

// ============================================================================
// 性能优化：sync.Pool 用于复用 big.Int 对象
// ============================================================================

var bigIntPool = sync.Pool{
	New: func() any {
		return new(big.Int)
	},
}

func getBigInt() *big.Int {
	return bigIntPool.Get().(*big.Int)
}

func putBigInt(v *big.Int) {
	v.SetInt64(0)
	bigIntPool.Put(v)
}

// ============================================================================
// 超低延迟辅助函数
// ============================================================================

// disc8 将最多 8 字节转换为小端 uint64（内联友好）
func disc8(b ...byte) uint64 {
	var a [8]byte
	copy(a[:], b)
	return binary.LittleEndian.Uint64(a[:])
}

// readU8 快速读取 uint8（内联）
func readU8(b []byte, o int) (byte, bool) {
	if o >= len(b) {
		return 0, false
	}
	return b[o], true
}

// readU16LE 快速读取小端 uint16（内联）
func readU16LE(b []byte, o int) (uint16, bool) {
	if o+2 > len(b) {
		return 0, false
	}
	return binary.LittleEndian.Uint16(b[o:]), true
}

// readU32LE 快速读取小端 uint32（内联）
func readU32LE(b []byte, o int) (uint32, bool) {
	if o+4 > len(b) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b[o:]), true
}

// readI32LE 快速读取小端 int32
func readI32LE(b []byte, o int) (int32, bool) {
	if o+4 > len(b) {
		return 0, false
	}
	v := binary.LittleEndian.Uint32(b[o:])
	return int32(v), true
}

// readU64LE 快速读取小端 uint64（内联）
func readU64LE(b []byte, o int) (uint64, bool) {
	if o+8 > len(b) {
		return 0, false
	}
	return binary.LittleEndian.Uint64(b[o:]), true
}

// readI64LE 快速读取小端 int64
func readI64LE(b []byte, o int) (int64, bool) {
	u, ok := readU64LE(b, o)
	if !ok {
		return 0, false
	}
	return int64(u), true
}

// readU128LE 快速读取 16 字节到数组
func readU128LE(b []byte, o int) ([16]byte, bool) {
	var out [16]byte
	if o+16 > len(b) {
		return out, false
	}
	copy(out[:], b[o:o+16])
	return out, true
}

// readBool 快速读取 bool
func readBool(b []byte, o int) (bool, bool) {
	v, ok := readU8(b, o)
	if !ok {
		return false, false
	}
	return v == 1, true
}

// readDiscU64 快速读取 discriminator
func readDiscU64(b []byte) (uint64, bool) {
	return readU64LE(b, 0)
}

// readBorshString 读取 Borsh 编码字符串
func readBorshString(b []byte, o int) (s string, next int, ok bool) {
	l, ok := readU32LE(b, o)
	if !ok || int(o)+4+int(l) > len(b) {
		return "", o, false
	}
	return string(b[o+4 : o+4+int(l)]), o + 4 + int(l), true
}

// u128LEDecimalString 将 16 字节小端 u128 转为十进制字符串（优化版）
func u128LEDecimalString(u [16]byte) string {
	// 从池中获取 big.Int
	bi := getBigInt()
	defer putBigInt(bi)

	// 转换为大端字节序
	be := make([]byte, 16)
	for i := 0; i < 16; i++ {
		be[15-i] = u[i]
	}
	bi.SetBytes(be)
	return bi.String()
}

// u128LEDecimalStringInline 内联版本（无池，用于热路径）
func u128LEDecimalStringInline(u [16]byte) string {
	be := make([]byte, 16)
	for i := 0; i < 16; i++ {
		be[15-i] = u[i]
	}
	return new(big.Int).SetBytes(be).String()
}

// ============================================================================
// 预定义 discriminator 常量（避免运行时计算）
// ============================================================================

var (
	// PumpFun
	discPumpCreate  = disc8(27, 114, 169, 77, 222, 235, 99, 118)
	discPumpTrade   = disc8(189, 219, 127, 211, 78, 230, 97, 238)
	discPumpMigrate = disc8(189, 233, 93, 185, 92, 148, 234, 148)

	// PumpSwap
	discPSBuy        = disc8(103, 244, 82, 31, 44, 245, 119, 119)
	discPSSell       = disc8(62, 47, 55, 10, 165, 3, 220, 42)
	discPSCreatePool = disc8(177, 49, 12, 210, 160, 118, 167, 116)
	discPSAddLiq     = disc8(120, 248, 61, 83, 31, 142, 107, 144)
	discPSRemLiq     = disc8(22, 9, 133, 26, 160, 44, 71, 192)

	// Raydium CLMM
	discClmmSwap    = disc8(248, 198, 158, 145, 225, 117, 135, 200)
	discClmmIncLiq  = disc8(133, 29, 89, 223, 69, 238, 176, 10)
	discClmmDecLiq  = disc8(160, 38, 208, 111, 104, 91, 44, 1)
	discClmmCreate  = disc8(233, 146, 209, 142, 207, 104, 64, 188)
	discClmmCollect = disc8(164, 152, 207, 99, 187, 104, 171, 119)

	// Raydium CPMM
	discCpmmSwapIn  = disc8(143, 190, 90, 218, 196, 30, 51, 222)
	discCpmmSwapOut = disc8(55, 217, 98, 86, 163, 74, 180, 173)
	discCpmmDeposit = disc8(242, 35, 198, 137, 82, 225, 242, 182)
	discCpmmWithdraw = disc8(183, 18, 70, 156, 148, 109, 161, 34)

	// Raydium AMM V4 (单字节 discriminator 作为 u64)
	discAmmSwapIn     = disc8(0, 0, 0, 0, 0, 0, 0, 9)
	discAmmSwapOut    = disc8(0, 0, 0, 0, 0, 0, 0, 11)
	discAmmDeposit    = disc8(0, 0, 0, 0, 0, 0, 0, 3)
	discAmmWithdraw   = disc8(0, 0, 0, 0, 0, 0, 0, 4)
	discAmmWithdrawPnl = disc8(0, 0, 0, 0, 0, 0, 0, 7)
	discAmmInit2      = disc8(0, 0, 0, 0, 0, 0, 0, 1)

	// Orca
	discOrcaSwap    = disc8(225, 202, 73, 175, 147, 43, 160, 150)
	discOrcaIncLiq  = disc8(30, 7, 144, 181, 102, 254, 155, 161)
	discOrcaDecLiq  = disc8(166, 1, 36, 71, 112, 202, 181, 171)
	discOrcaPoolInit = disc8(100, 118, 173, 87, 12, 198, 254, 229)

	// Meteora Pools
	discMeteoraSwap      = disc8(81, 108, 227, 190, 205, 208, 10, 196)
	discMeteoraAdd       = disc8(31, 94, 125, 90, 227, 52, 61, 186)
	discMeteoraRemove    = disc8(116, 244, 97, 232, 103, 31, 152, 58)
	discMeteoraBootstrap = disc8(121, 127, 38, 136, 92, 55, 14, 247)
	discMeteoraPoolCreated = disc8(202, 44, 41, 88, 104, 220, 157, 82)
	discMeteoraSetPoolFees = disc8(245, 26, 198, 164, 88, 18, 75, 9)

	// Meteora DAMM v2
	discDammSwap   = disc8(27, 60, 21, 213, 138, 170, 187, 147)
	discDammSwap2  = disc8(189, 66, 51, 168, 38, 80, 117, 153)
	discDammAdd    = disc8(175, 242, 8, 157, 30, 247, 185, 169)
	discDammRem    = disc8(87, 46, 88, 98, 175, 96, 34, 91)
	discDammInit   = disc8(228, 50, 246, 85, 203, 66, 134, 37)
	discDammCreate = disc8(156, 15, 119, 198, 29, 181, 221, 55)
	discDammClose  = disc8(20, 145, 144, 68, 143, 142, 214, 178)

	// 别名（用于 meteora_extra.go）
	discDammCreatePosition = discDammCreate
	discDammClosePosition  = discDammClose
	discDammAddLiquidity   = discDammAdd
	discDammRemoveLiq      = discDammRem
	discDammInitPool       = discDammInit

	// Bonk (Raydium Launchpad)
	discBonkTrade       = disc8(2, 3, 4, 5, 6, 7, 8, 9)
	discBonkPoolCreate  = disc8(1, 2, 3, 4, 5, 6, 7, 8)
	discBonkMigrateAmm  = disc8(3, 4, 5, 6, 7, 8, 9, 10)

	// Meteora DLMM
	dlmmSwap      = disc8(142, 35, 199, 193, 77, 169, 172, 85)
	dlmmAddLiq    = disc8(83, 14, 120, 139, 140, 137, 206, 127)
	dlmmRemoveLiq = disc8(21, 163, 85, 22, 151, 198, 162, 20)
	dlmmInitPool  = disc8(147, 185, 18, 29, 117, 51, 73, 11)
	dlmmInitBin   = disc8(234, 167, 189, 121, 145, 110, 218, 57)
	dlmmCreatePos = disc8(83, 240, 75, 157, 184, 163, 193, 19)
	dlmmClosePos  = disc8(240, 142, 81, 161, 159, 188, 231, 188)
	dlmmClaimFee  = disc8(233, 53, 144, 233, 46, 160, 53, 114)
)
