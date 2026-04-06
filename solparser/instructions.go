package solparser

import (
	"encoding/binary"
)

// InstructionData 指令数据
type InstructionData struct {
	ProgramIDIndex uint32
	Accounts       []uint32
	Data           []byte
}

// 程序 ID 常量
const (
	PUMPFUN_PROGRAM_ID          = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
	PUMPSWAP_PROGRAM_ID         = "pAMMBay6oceH9fJKBRdGP4LmT4saRGfEE7xmrCaGWpZ"
	METEORA_DAMM_V2_PROGRAM_ID  = "cpamdpZCGKUy5JxQXB2MWgCm3hcnGjEJbYTJgfm4E8a"
	RAYDIUM_CLMM_PROGRAM_ID     = "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK"
	RAYDIUM_CPMM_PROGRAM_ID     = "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C"
	RAYDIUM_AMM_V4_PROGRAM_ID   = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	ORCA_WHIRLPOOL_PROGRAM_ID   = "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"
	BONK_LAUNCHPAD_PROGRAM_ID   = "LanCh3hDdY7M6x8urBSLJhsQBgPNGKHNqJqGwzAEmBm"
)

// ParseInstructionUnified 统一的指令解析入口
// 对齐 Rust `parse_instruction_unified`
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
	// 快速检查指令数据长度
	if len(instructionData) == 0 {
		return DexEvent{}
	}

	// 根据程序 ID 路由到相应的解析器
	switch programID {
	case PUMPFUN_PROGRAM_ID:
		if !EventTypeFilterIncludesPumpfun(filter) {
			return DexEvent{}
		}
		return ParsePumpfunInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case PUMPSWAP_PROGRAM_ID:
		if !EventTypeFilterIncludesPumpswap(filter) {
			return DexEvent{}
		}
		return ParsePumpswapInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case METEORA_DAMM_V2_PROGRAM_ID:
		if !EventTypeFilterIncludesMeteoraDammV2(filter) {
			return DexEvent{}
		}
		return ParseMeteoraDammInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_CLMM_PROGRAM_ID:
		if !EventTypeFilterIncludesRaydiumClmm(filter) {
			return DexEvent{}
		}
		return ParseRaydiumClmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_CPMM_PROGRAM_ID:
		if !EventTypeFilterIncludesRaydiumCpmm(filter) {
			return DexEvent{}
		}
		return ParseRaydiumCpmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_AMM_V4_PROGRAM_ID:
		if !EventTypeFilterIncludesRaydiumAmmV4(filter) {
			return DexEvent{}
		}
		return ParseRaydiumAmmV4Instruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case ORCA_WHIRLPOOL_PROGRAM_ID:
		if !EventTypeFilterIncludesOrcaWhirlpool(filter) {
			return DexEvent{}
		}
		return ParseOrcaWhirlpoolInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case BONK_LAUNCHPAD_PROGRAM_ID:
		if !EventTypeFilterIncludesBonk(filter) {
			return DexEvent{}
		}
		return ParseBonkInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	}

	return DexEvent{}
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

// ParsePumpfunInstruction 解析 PumpFun 指令
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

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	// PumpFun Create: 8576854823835016728
	if discriminator == 8576854823835016728 {
		return parsePumpFunCreateInstr(data, accounts, meta)
	}
	// PumpFun Buy: 16927863322537900544
	if discriminator == 16927863322537900544 {
		return DexEvent{
			Type: EventTypePumpFunBuy,
			Data: &PumpFunTradeEvent{
				Metadata: meta,
				Mint:     getAccountSafe(accounts, 2),
				User:     getAccountSafe(accounts, 7),
			},
		}
	}
	// PumpFun Sell: 12502976635542175488
	if discriminator == 12502976635542175488 {
		return DexEvent{
			Type: EventTypePumpFunSell,
			Data: &PumpFunTradeEvent{
				Metadata: meta,
				Mint:     getAccountSafe(accounts, 2),
				User:     getAccountSafe(accounts, 7),
			},
		}
	}

	return DexEvent{}
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
			Metadata: meta,
			Name:     name,
			Symbol:   symbol,
			Uri:      uri,
			Creator:  creator,
			Mint:     getAccountSafe(accounts, 0),
		},
	}
}

// ParsePumpswapInstruction 解析 PumpSwap 指令
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

	// PumpSwap Buy: disc8(103, 244, 82, 31, 44, 245, 119, 119)
	if discriminator == disc8(103, 244, 82, 31, 44, 245, 119, 119) {
		return DexEvent{
			Type: EventTypePumpSwapBuy,
			Data: &PumpSwapBuyEvent{Metadata: meta},
		}
	}
	// PumpSwap Sell: disc8(62, 47, 55, 10, 165, 3, 220, 42)
	if discriminator == disc8(62, 47, 55, 10, 165, 3, 220, 42) {
		return DexEvent{
			Type: EventTypePumpSwapSell,
			Data: &PumpSwapSellEvent{Metadata: meta},
		}
	}

	return DexEvent{}
}

// ParseMeteoraDammInstruction 解析 Meteora DAMM 指令
func ParseMeteoraDammInstruction(
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

	// 复用 Log 解析器中的 DAMM discriminators
	switch discriminator {
	case discDammSwap, discDammSwap2:
		return DexEvent{
			Type: EventTypeMeteoraDammV2Swap,
			Data: &MeteoraDammV2SwapEvent{Metadata: meta},
		}
	case discDammAdd:
		return DexEvent{
			Type: EventTypeMeteoraDammV2AddLiquidity,
			Data: &MeteoraDammV2AddLiquidityEvent{Metadata: meta},
		}
	case discDammRem:
		return DexEvent{
			Type: EventTypeMeteoraDammV2RemoveLiquidity,
			Data: &MeteoraDammV2RemoveLiquidityEvent{Metadata: meta},
		}
	case discDammCreate:
		return DexEvent{
			Type: EventTypeMeteoraDammV2CreatePosition,
			Data: &MeteoraDammV2CreatePositionEvent{Metadata: meta},
		}
	case discDammClose:
		return DexEvent{
			Type: EventTypeMeteoraDammV2ClosePosition,
			Data: &MeteoraDammV2ClosePositionEvent{Metadata: meta},
		}
	case discDammInit:
		return DexEvent{
			Type: EventTypeMeteoraDammV2InitializePool,
			Data: &MeteoraDammV2InitializePoolEvent{Metadata: meta},
		}
	}

	return DexEvent{}
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

	switch discriminator {
	case discClmmSwap:
		return DexEvent{
			Type: EventTypeRaydiumClmmSwap,
			Data: &RaydiumClmmSwapEvent{
				Metadata:      meta,
				PoolState:     getAccountSafe(accounts, 2),
				Sender:        getAccountSafe(accounts, 0),
				TokenAccount0: zeroPubkey,
				TokenAccount1: zeroPubkey,
			},
		}
	case discClmmIncLiq:
		return DexEvent{
			Type: EventTypeRaydiumClmmIncreaseLiquidity,
			Data: &RaydiumClmmIncreaseLiquidityEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 3),
				PositionNftMint: zeroPubkey,
				User:            getAccountSafe(accounts, 0),
			},
		}
	case discClmmDecLiq:
		return DexEvent{
			Type: EventTypeRaydiumClmmDecreaseLiquidity,
			Data: &RaydiumClmmDecreaseLiquidityEvent{
				Metadata:        meta,
				Pool:            getAccountSafe(accounts, 3),
				PositionNftMint: zeroPubkey,
				User:            getAccountSafe(accounts, 0),
			},
		}
	case discClmmCreate:
		return DexEvent{
			Type: EventTypeRaydiumClmmCreatePool,
			Data: &RaydiumClmmCreatePoolEvent{
				Metadata: meta,
				Pool:     getAccountSafe(accounts, 4),
				Creator:  getAccountSafe(accounts, 0),
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
