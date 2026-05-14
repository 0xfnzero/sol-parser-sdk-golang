package solparser

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

// 外层**指令** discriminator（Rust `instr/pump.rs` / `pump_amm.rs`）。Program log 里的 Buy/Sell 等 Event disc 仍见 `binary.go` / `matcher.go`，二者不可混用。
var (
	instrPumpOuterCreate            = disc8(24, 30, 200, 40, 5, 28, 7, 119)
	instrPumpOuterCreateV2          = disc8(214, 144, 76, 236, 95, 139, 49, 180)
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
	instrClmmSwap                       = disc8(248, 198, 158, 145, 225, 117, 135, 200)
	instrClmmSwapV2                     = disc8(43, 4, 237, 11, 26, 201, 30, 98)
	instrClmmIncLiqV2                   = disc8(133, 29, 89, 223, 69, 238, 176, 10)
	instrClmmDecLiqV2                   = disc8(58, 127, 188, 62, 79, 82, 196, 96)
	instrClmmCreatePool                 = disc8(233, 146, 209, 142, 207, 104, 64, 188)
	instrClmmOpenPositionV2             = disc8(77, 184, 74, 214, 112, 86, 241, 199)
	instrClmmOpenPositionWithToken22Nft = disc8(77, 255, 174, 82, 125, 29, 201, 46)
	instrClmmClosePosition              = disc8(123, 134, 81, 0, 49, 68, 98, 98)
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
// 覆盖本包已有的外层指令解析器：PumpFun、PumpSwap、Pump Fees、Meteora DAMM V2、Raydium、Orca、Bonk。
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
		return ParsePumpfunInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case PUMPSWAP_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpswap(filter) {
			return DexEvent{}
		}
		return ParsePumpswapInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case METEORA_DAMM_V2_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesMeteoraDammV2(filter) {
			return DexEvent{}
		}
		return ParseMeteoraDammInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case PUMP_FEES_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesPumpFees(filter) {
			return DexEvent{}
		}
		return ParsePumpFeesInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_CLMM_PROGRAM_ID, GrpcRaydiumClmmProgramID:
		if filter != nil && !EventTypeFilterIncludesRaydiumClmm(filter) {
			return DexEvent{}
		}
		return ParseRaydiumClmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_CPMM_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumCpmm(filter) {
			return DexEvent{}
		}
		return ParseRaydiumCpmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_AMM_V4_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesRaydiumAmmV4(filter) {
			return DexEvent{}
		}
		return ParseRaydiumAmmV4Instruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case ORCA_WHIRLPOOL_PROGRAM_ID:
		if filter != nil && !EventTypeFilterIncludesOrcaWhirlpool(filter) {
			return DexEvent{}
		}
		return ParseOrcaWhirlpoolInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case BONK_PROGRAM_ID, GrpcBonkProgramID:
		if filter != nil && !EventTypeFilterIncludesBonk(filter) {
			return DexEvent{}
		}
		return ParseBonkInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)
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

// ParseInnerInstructionUnified 与 Rust `parse_inner_instruction` 对齐：16 字节 discriminator，data[16..] 为 payload。
// 当前实现 PumpFun、PumpSwap（其余 program 可按 Rust `instruction_parser` 扩展）。
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
			return ev
		case bytes.Equal(disc, pumpfunInnerCreateToken):
			return parseCreateFromData(inner, meta)
		case bytes.Equal(disc, pumpfunInnerMigrateComplete):
			return parseMigrateFromData(inner, meta)
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
			return ev
		case bytes.Equal(disc, pumpswapInnerSell):
			ev := parsePSSellFromData(inner, meta)
			if ev.Type != "" {
				if p, ok := ev.Data.(*PumpSwapSellEvent); ok {
					enrichPumpSwapSellFromAccounts(p, accounts)
				}
			}
			return ev
		case bytes.Equal(disc, pumpswapInnerCreatePool):
			return parsePSCreatePoolFromData(inner, meta)
		case bytes.Equal(disc, pumpswapInnerAddLiquidity):
			return parsePSAddLiqFromData(inner, meta)
		case bytes.Equal(disc, pumpswapInnerRemoveLiquidity):
			return parsePSRemoveLiqFromData(inner, meta)
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

// ParsePumpfunInstruction 与 Rust `pump::parse_instruction` 一致：仅解析外层 Create、CreateV2，以及内层 CPI Migrate；不解析 Buy/Sell 外层指令（与 Rust 相同，成交以日志为准）。
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

func parsePumpFunTradeV2Instr(ixName string, data []byte, accounts []string, meta EventMetadata) DexEvent {
	minAcc := 27
	if ixName == "sell_v2" {
		minAcc = 26
	}
	if len(accounts) < minAcc {
		return DexEvent{}
	}
	first := readPumpFunV2Amount(data, 0)
	second := readPumpFunV2Amount(data, 8)
	tokenAmount := first
	solAmount := second
	if ixName == "buy_exact_quote_in_v2" {
		solAmount = first
		tokenAmount = second
	}
	return DexEvent{
		Type: EventTypePumpFunTrade,
		Data: &PumpFunTradeEvent{
			Metadata:               meta,
			Mint:                   getAccountSafe(accounts, 1),
			BondingCurve:           getAccountSafe(accounts, 10),
			User:                   getAccountSafe(accounts, 13),
			SolAmount:              solAmount,
			TokenAmount:            tokenAmount,
			FeeRecipient:           getAccountSafe(accounts, 6),
			IsBuy:                  ixName != "sell_v2",
			IxName:                 ixName,
			AssociatedBondingCurve: getAccountSafe(accounts, 11),
			TokenProgram:           getAccountSafe(accounts, 3),
			CreatorVault:           getAccountSafe(accounts, 16),
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
			Metadata:     meta,
			Name:         name,
			Symbol:       symbol,
			Uri:          uri,
			Creator:      creator,
			Mint:         getAccountSafe(accounts, 0),
			BondingCurve: getAccountSafe(accounts, 2),
			User:         getAccountSafe(accounts, 7),
		},
	}
}

func parsePumpFunCreateV2Instr(data []byte, accounts []string, meta EventMetadata) DexEvent {
	const minAcc = 16
	if len(accounts) < minAcc {
		return DexEvent{}
	}
	off := 0
	readStr := func() string {
		if off+4 > len(data) {
			return ""
		}
		n := int(binary.LittleEndian.Uint32(data[off : off+4]))
		off += 4
		if n < 0 || off+n > len(data) {
			return ""
		}
		s := string(data[off : off+n])
		off += n
		return s
	}
	name := readStr()
	symbol := readStr()
	uri := readStr()
	if off+128 > len(data) {
		return DexEvent{}
	}
	mint := ReadPubkey(data, off)
	off += 32
	bondingCurve := ReadPubkey(data, off)
	off += 32
	user := ReadPubkey(data, off)
	off += 32
	creator := ReadPubkey(data, off)
	acc := accounts[:minAcc]
	return DexEvent{
		Type: EventTypePumpFunCreateV2,
		Data: &PumpFunCreateV2TokenEvent{
			Metadata:               meta,
			Name:                   name,
			Symbol:                 symbol,
			Uri:                    uri,
			Mint:                   mint,
			BondingCurve:           bondingCurve,
			User:                   user,
			Creator:                creator,
			TokenProgram:           acc[7],
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
	_ = accounts
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
		if len(payload) < 8+8+8+1 {
			return DexEvent{}
		}
		sqrt, _ := readU64LE(payload, 16)
		isBaseInput, ok := readBool(payload, 24)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{
			Type: EventTypeRaydiumClmmSwap,
			Data: &RaydiumClmmSwapEvent{
				Metadata:      meta,
				PoolState:     getAccountSafe(accounts, 0),
				Sender:        getAccountSafe(accounts, 1),
				TokenAccount0: zeroPubkey,
				TokenAccount1: zeroPubkey,
				ZeroForOne:    isBaseInput,
				SqrtPriceX64:  strconv.FormatUint(sqrt, 10),
				Liquidity:     "0",
			},
		}
	case instrClmmIncLiqV2:
		if len(payload) < 8+8+8 {
			return DexEvent{}
		}
		liquidity, _ := readU64LE(payload, 0)
		amount0Max, _ := readU64LE(payload, 8)
		amount1Max, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeRaydiumClmmIncreaseLiquidity,
			Data: &RaydiumClmmIncreaseLiquidityEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 0),
				PositionNftMint: getAccountSafe(accounts, 1),
				User:            getAccountSafe(accounts, 2),
				Liquidity:       strconv.FormatUint(liquidity, 10),
				Amount0Max:      amount0Max,
				Amount1Max:      amount1Max,
			},
		}
	case instrClmmDecLiqV2:
		if len(payload) < 8+8+8 {
			return DexEvent{}
		}
		liquidity, _ := readU64LE(payload, 0)
		amount0Min, _ := readU64LE(payload, 8)
		amount1Min, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeRaydiumClmmDecreaseLiquidity,
			Data: &RaydiumClmmDecreaseLiquidityEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 0),
				PositionNftMint: getAccountSafe(accounts, 1),
				User:            getAccountSafe(accounts, 2),
				Liquidity:       strconv.FormatUint(liquidity, 10),
				Amount0Min:      amount0Min,
				Amount1Min:      amount1Min,
			},
		}
	case instrClmmCreatePool:
		if len(payload) < 8+8 {
			return DexEvent{}
		}
		sqrt, _ := readU64LE(payload, 0)
		openTime, _ := readU64LE(payload, 8)
		return DexEvent{
			Type: EventTypeRaydiumClmmCreatePool,
			Data: &RaydiumClmmCreatePoolEvent{
				Metadata:     meta,
				Pool:         getAccountSafe(accounts, 0),
				Creator:      getAccountSafe(accounts, 1),
				Token0Mint:   getAccountSafe(accounts, 2),
				Token1Mint:   getAccountSafe(accounts, 3),
				SqrtPriceX64: strconv.FormatUint(sqrt, 10),
				OpenTime:     openTime,
			},
		}
	case instrClmmOpenPositionV2, instrClmmOpenPositionWithToken22Nft:
		if len(payload) < 4+4+4+4+8+8+8 {
			return DexEvent{}
		}
		tickLower, _ := readI32LE(payload, 0)
		tickUpper, _ := readI32LE(payload, 4)
		liquidity, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeRaydiumClmmOpenPosition,
			Data: &RaydiumClmmOpenPositionEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 0),
				User:            getAccountSafe(accounts, 1),
				PositionNftMint: getAccountSafe(accounts, 2),
				TickLowerIndex:  tickLower,
				TickUpperIndex:  tickUpper,
				Liquidity:       strconv.FormatUint(liquidity, 10),
			},
		}
	case instrClmmClosePosition:
		return DexEvent{
			Type: EventTypeRaydiumClmmClosePosition,
			Data: &RaydiumClmmClosePositionEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 0),
				User:            getAccountSafe(accounts, 1),
				PositionNftMint: getAccountSafe(accounts, 2),
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
		return DexEvent{
			Type: EventTypeRaydiumCpmmSwap,
			Data: &RaydiumCpmmSwapEvent{
				Metadata: meta,
				PoolID:   getAccountSafe(accounts, 2),
			},
		}
	case discCpmmDeposit:
		return DexEvent{
			Type: EventTypeRaydiumCpmmDeposit,
			Data: &RaydiumCpmmDepositEvent{
				Metadata: meta,
				Pool:     getAccountSafe(accounts, 2),
				User:     getAccountSafe(accounts, 0),
			},
		}
	case discCpmmWithdraw:
		return DexEvent{
			Type: EventTypeRaydiumCpmmWithdraw,
			Data: &RaydiumCpmmWithdrawEvent{
				Metadata: meta,
				Pool:     getAccountSafe(accounts, 2),
				User:     getAccountSafe(accounts, 0),
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
		return DexEvent{
			Type: EventTypeRaydiumAmmV4Swap,
			Data: &RaydiumAmmV4SwapEvent{
				Metadata:        meta,
				Amm:             getAccountSafe(accounts, 1),
				UserSourceOwner: getAccountSafe(accounts, 17),
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

	// Orca Whirlpool swap discriminators（与 matcher.go 中对齐）
	switch discriminator {
	case discOrcaSwap:
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolSwap,
			Data: &OrcaWhirlpoolSwapEvent{
				Metadata:  meta,
				Whirlpool: getAccountSafe(accounts, 2),
				AToB:      true,
			},
		}
	case discOrcaIncLiq:
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolLiquidityIncreased,
			Data: &OrcaWhirlpoolLiquidityIncreasedEvent{
				Metadata:  meta,
				Whirlpool: getAccountSafe(accounts, 1),
				Position:  getAccountSafe(accounts, 3),
			},
		}
	case discOrcaDecLiq:
		return DexEvent{
			Type: EventTypeOrcaWhirlpoolLiquidityDecreased,
			Data: &OrcaWhirlpoolLiquidityDecreasedEvent{
				Metadata:  meta,
				Whirlpool: getAccountSafe(accounts, 1),
				Position:  getAccountSafe(accounts, 3),
			},
		}
	}

	return DexEvent{}
}

// ParseBonkInstruction 解析 Bonk (Raydium Launchpad) 指令
func ParseBonkInstruction(
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
	case discBonkTrade:
		return DexEvent{
			Type: EventTypeBonkTrade,
			Data: &BonkTradeEvent{
				Metadata:       meta,
				PoolState:      getAccountSafe(accounts, 1),
				User:           getAccountSafe(accounts, 0),
				IsBuy:          true,
				TradeDirection: "Buy",
				ExactIn:        true,
			},
		}
	case discBonkPoolCreate:
		return DexEvent{
			Type: EventTypeBonkPoolCreate,
			Data: &BonkPoolCreateEvent{
				Metadata: meta,
				BaseMintParam: BonkMintParam{
					Symbol:   "BONK",
					Name:     "Bonk Pool",
					Uri:      "https://bonk.com",
					Decimals: 5,
				},
				PoolState: getAccountSafe(accounts, 1),
				Creator:   getAccountSafe(accounts, 8),
			},
		}
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
