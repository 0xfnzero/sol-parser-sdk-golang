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
		return nil
	}

	// 根据程序 ID 路由到相应的解析器
	switch programID {
	case PUMPFUN_PROGRAM_ID:
		if !EventTypeFilterIncludesPumpfun(filter) {
			return nil
		}
		return ParsePumpfunInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case PUMPSWAP_PROGRAM_ID:
		if !EventTypeFilterIncludesPumpswap(filter) {
			return nil
		}
		return ParsePumpswapInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case METEORA_DAMM_V2_PROGRAM_ID:
		if !EventTypeFilterIncludesMeteoraDammV2(filter) {
			return nil
		}
		return ParseMeteoraDammInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_CLMM_PROGRAM_ID:
		if !EventTypeFilterIncludesRaydiumClmm(filter) {
			return nil
		}
		return ParseRaydiumClmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_CPMM_PROGRAM_ID:
		if !EventTypeFilterIncludesRaydiumCpmm(filter) {
			return nil
		}
		return ParseRaydiumCpmmInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case RAYDIUM_AMM_V4_PROGRAM_ID:
		if !EventTypeFilterIncludesRaydiumAmmV4(filter) {
			return nil
		}
		return ParseRaydiumAmmV4Instruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case ORCA_WHIRLPOOL_PROGRAM_ID:
		if !EventTypeFilterIncludesOrcaWhirlpool(filter) {
			return nil
		}
		return ParseOrcaWhirlpoolInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case BONK_LAUNCHPAD_PROGRAM_ID:
		if !EventTypeFilterIncludesBonk(filter) {
			return nil
		}
		return ParseBonkInstruction(instructionData, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	}

	return nil
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	// PumpFun Create: 8576854823835016728
	if discriminator == 8576854823835016728 {
		return parsePumpFunCreateInstr(data, accounts, meta)
	}
	// PumpFun Buy: 16927863322537900544
	if discriminator == 16927863322537900544 {
		return DexEvent{"PumpFunBuy": map[string]any{
			"metadata": meta,
			"mint":     getAccountSafe(accounts, 2),
			"user":     getAccountSafe(accounts, 7),
		}}
	}
	// PumpFun Sell: 12502976635542175488
	if discriminator == 12502976635542175488 {
		return DexEvent{"PumpFunSell": map[string]any{
			"metadata": meta,
			"mint":     getAccountSafe(accounts, 2),
			"user":     getAccountSafe(accounts, 7),
		}}
	}

	return nil
}

func parsePumpFunCreateInstr(data []byte, accounts []string, meta EventMetadata) DexEvent {
	offset := 8 // Skip discriminator

	if offset+4 > len(data) {
		return nil
	}
	nameLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+nameLen > len(data) {
		return nil
	}
	name := string(data[offset : offset+nameLen])
	offset += nameLen

	if offset+4 > len(data) {
		return nil
	}
	symbolLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+symbolLen > len(data) {
		return nil
	}
	symbol := string(data[offset : offset+symbolLen])
	offset += symbolLen

	if offset+4 > len(data) {
		return nil
	}
	uriLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+uriLen > len(data) {
		return nil
	}
	uri := string(data[offset : offset+uriLen])
	offset += uriLen

	creator := zeroPubkey
	if offset+32 <= len(data) {
		creator = ReadPubkey(data, offset)
	}

	return DexEvent{"PumpFunCreate": map[string]any{
		"metadata": meta,
		"name":     name,
		"symbol":   symbol,
		"uri":      uri,
		"creator":  creator,
		"mint":     getAccountSafe(accounts, 0),
	}}
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	// PumpSwap Buy: disc8(103, 244, 82, 31, 44, 245, 119, 119)
	if discriminator == disc8(103, 244, 82, 31, 44, 245, 119, 119) {
		return DexEvent{"PumpSwapBuy": map[string]any{"metadata": meta}}
	}
	// PumpSwap Sell: disc8(62, 47, 55, 10, 165, 3, 220, 42)
	if discriminator == disc8(62, 47, 55, 10, 165, 3, 220, 42) {
		return DexEvent{"PumpSwapSell": map[string]any{"metadata": meta}}
	}

	return nil
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	// 复用 Log 解析器中的 DAMM discriminators
	switch discriminator {
	case discDammSwap, discDammSwap2:
		return DexEvent{"MeteoraDammV2Swap": map[string]any{"metadata": meta}}
	case disc8(175, 242, 8, 157, 30, 247, 185, 169):
		return DexEvent{"MeteoraDammV2AddLiquidity": map[string]any{"metadata": meta}}
	case disc8(87, 46, 88, 98, 175, 96, 34, 91):
		return DexEvent{"MeteoraDammV2RemoveLiquidity": map[string]any{"metadata": meta}}
	case disc8(156, 15, 119, 198, 29, 181, 221, 55):
		return DexEvent{"MeteoraDammV2CreatePosition": map[string]any{"metadata": meta}}
	case disc8(20, 145, 144, 68, 143, 142, 214, 178):
		return DexEvent{"MeteoraDammV2ClosePosition": map[string]any{"metadata": meta}}
	case disc8(228, 50, 246, 85, 203, 66, 134, 37):
		return DexEvent{"MeteoraDammV2InitializePool": map[string]any{"metadata": meta}}
	}

	return nil
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch discriminator {
	case disc8(248, 198, 158, 145, 225, 117, 135, 200):
		return DexEvent{"RaydiumClmmSwap": map[string]any{
			"metadata":        meta,
			"pool_state":      getAccountSafe(accounts, 2),
			"sender":          getAccountSafe(accounts, 0),
			"token_account_0": zeroPubkey,
			"token_account_1": zeroPubkey,
			"amount_0": uint64(0), "amount_1": uint64(0),
			"zero_for_one": false, "sqrt_price_x64": "0",
			"liquidity": "0", "transfer_fee_0": uint64(0), "transfer_fee_1": uint64(0), "tick": int32(0),
		}}
	case disc8(133, 29, 89, 223, 69, 238, 176, 10):
		return DexEvent{"RaydiumClmmIncreaseLiquidity": map[string]any{
			"metadata": meta, "pool": getAccountSafe(accounts, 3),
			"position_nft_mint": zeroPubkey, "user": getAccountSafe(accounts, 0),
			"liquidity": "0", "amount0_max": uint64(0), "amount1_max": uint64(0),
		}}
	case disc8(160, 38, 208, 111, 104, 91, 44, 1):
		return DexEvent{"RaydiumClmmDecreaseLiquidity": map[string]any{
			"metadata": meta, "pool": getAccountSafe(accounts, 3),
			"position_nft_mint": zeroPubkey, "user": getAccountSafe(accounts, 0),
			"liquidity": "0", "amount0_min": uint64(0), "amount1_min": uint64(0),
		}}
	case disc8(233, 146, 209, 142, 207, 104, 64, 188):
		return DexEvent{"RaydiumClmmCreatePool": map[string]any{
			"metadata": meta, "pool": getAccountSafe(accounts, 4),
			"creator": getAccountSafe(accounts, 0),
			"token_0_mint": zeroPubkey, "token_1_mint": zeroPubkey,
			"tick_spacing": 0, "fee_rate": 0, "sqrt_price_x64": "0", "open_time": uint64(0),
		}}
	}

	return nil
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch discriminator {
	case disc8(143, 190, 90, 218, 196, 30, 51, 222):
		return DexEvent{"RaydiumCpmmSwap": map[string]any{
			"metadata": meta, "pool_id": getAccountSafe(accounts, 2),
			"input_amount": uint64(0), "output_amount": uint64(0),
			"input_vault_before": uint64(0), "output_vault_before": uint64(0),
			"input_transfer_fee": uint64(0), "output_transfer_fee": uint64(0), "base_input": true,
		}}
	case disc8(242, 35, 198, 137, 82, 225, 242, 182):
		return DexEvent{"RaydiumCpmmDeposit": map[string]any{
			"metadata": meta, "pool": getAccountSafe(accounts, 2),
			"user": getAccountSafe(accounts, 0),
			"lp_token_amount": uint64(0), "token0_amount": uint64(0), "token1_amount": uint64(0),
		}}
	case disc8(183, 18, 70, 156, 148, 109, 161, 34):
		return DexEvent{"RaydiumCpmmWithdraw": map[string]any{
			"metadata": meta, "pool": getAccountSafe(accounts, 2),
			"user": getAccountSafe(accounts, 0),
			"lp_token_amount": uint64(0), "token0_amount": uint64(0), "token1_amount": uint64(0),
		}}
	}

	return nil
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
		return nil
	}

	// Raydium AMM V4 使用单字节 instruction discriminator
	instrType := data[0]
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch instrType {
	case 9: // SwapBaseIn
		return DexEvent{"RaydiumAmmV4Swap": map[string]any{
			"metadata": meta, "amm": getAccountSafe(accounts, 1),
			"user_source_owner": getAccountSafe(accounts, 17),
			"amount_in": uint64(0), "minimum_amount_out": uint64(0),
			"max_amount_in": uint64(0), "amount_out": uint64(0),
			"token_program": zeroPubkey, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
			"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
			"serum_program": zeroPubkey, "serum_market": zeroPubkey, "serum_bids": zeroPubkey,
			"serum_asks": zeroPubkey, "serum_event_queue": zeroPubkey,
			"serum_coin_vault_account": zeroPubkey, "serum_pc_vault_account": zeroPubkey,
			"serum_vault_signer": zeroPubkey, "user_source_token_account": zeroPubkey,
			"user_destination_token_account": zeroPubkey,
		}}
	case 11: // SwapBaseOut
		return DexEvent{"RaydiumAmmV4Swap": map[string]any{
			"metadata": meta, "amm": getAccountSafe(accounts, 1),
			"user_source_owner": getAccountSafe(accounts, 17),
			"amount_in": uint64(0), "minimum_amount_out": uint64(0),
			"max_amount_in": uint64(0), "amount_out": uint64(0),
			"token_program": zeroPubkey, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
			"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
			"serum_program": zeroPubkey, "serum_market": zeroPubkey, "serum_bids": zeroPubkey,
			"serum_asks": zeroPubkey, "serum_event_queue": zeroPubkey,
			"serum_coin_vault_account": zeroPubkey, "serum_pc_vault_account": zeroPubkey,
			"serum_vault_signer": zeroPubkey, "user_source_token_account": zeroPubkey,
			"user_destination_token_account": zeroPubkey,
		}}
	}

	return nil
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	// Orca Whirlpool swap discriminators（与 matcher.go 中对齐）
	switch discriminator {
	case disc8(225, 202, 73, 175, 147, 43, 160, 150):
		return DexEvent{"OrcaWhirlpoolSwap": map[string]any{
			"metadata":               meta,
			"whirlpool":              getAccountSafe(accounts, 2),
			"a_to_b":                 true,
			"pre_sqrt_price":         "0",
			"post_sqrt_price":        "0",
			"input_amount":           uint64(0),
			"output_amount":          uint64(0),
			"input_transfer_fee":     uint64(0),
			"output_transfer_fee":    uint64(0),
			"lp_fee":                 uint64(0),
			"protocol_fee":           uint64(0),
		}}
	case disc8(30, 7, 144, 181, 102, 254, 155, 161):
		return DexEvent{"OrcaWhirlpoolLiquidityIncreased": map[string]any{
			"metadata": meta, "whirlpool": getAccountSafe(accounts, 1),
			"position": getAccountSafe(accounts, 3),
			"tick_lower_index": int32(0), "tick_upper_index": int32(0),
			"liquidity": "0", "token_a_amount": uint64(0), "token_b_amount": uint64(0),
			"token_a_transfer_fee": uint64(0), "token_b_transfer_fee": uint64(0),
		}}
	case disc8(166, 1, 36, 71, 112, 202, 181, 171):
		return DexEvent{"OrcaWhirlpoolLiquidityDecreased": map[string]any{
			"metadata": meta, "whirlpool": getAccountSafe(accounts, 1),
			"position": getAccountSafe(accounts, 3),
			"tick_lower_index": int32(0), "tick_upper_index": int32(0),
			"liquidity": "0", "token_a_amount": uint64(0), "token_b_amount": uint64(0),
			"token_a_transfer_fee": uint64(0), "token_b_transfer_fee": uint64(0),
		}}
	}

	return nil
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
		return nil
	}

	discriminator := binary.LittleEndian.Uint64(data[:8])
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch discriminator {
	case discBonkTrade:
		return DexEvent{"BonkTrade": map[string]any{
			"metadata":        meta,
			"pool_state":      getAccountSafe(accounts, 1),
			"user":            getAccountSafe(accounts, 0),
			"amount_in":       uint64(0),
			"amount_out":      uint64(0),
			"is_buy":          true,
			"trade_direction": "Buy",
			"exact_in":        true,
		}}
	case discBonkPoolCreate:
		return DexEvent{"BonkPoolCreate": map[string]any{
			"metadata": meta,
			"base_mint_param": map[string]any{
				"symbol": "BONK", "name": "Bonk Pool", "uri": "https://bonk.com", "decimals": 5,
			},
			"pool_state": getAccountSafe(accounts, 1),
			"creator":    getAccountSafe(accounts, 8),
		}}
	}

	return nil
}

// getAccountSafe 安全获取账户地址
func getAccountSafe(accounts []string, index int) string {
	if index < 0 || index >= len(accounts) {
		return zeroPubkey
	}
	return accounts[index]
}
