package solparser

var (
	instrMeteoraPoolsSwap            = disc8(248, 198, 158, 145, 225, 117, 135, 200)
	instrMeteoraPoolsAddLiquidity    = disc8(181, 157, 89, 67, 143, 182, 52, 72)
	instrMeteoraPoolsRemoveLiquidity = disc8(80, 85, 209, 72, 24, 206, 177, 108)
	instrMeteoraPoolsCreatePool      = disc8(95, 180, 10, 172, 84, 174, 232, 40)
)

// ParseMeteoraPoolsInstruction 对齐 Rust `meteora_amm::parse_instruction` 的外层指令路径。
func ParseMeteoraPoolsInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	disc, ok := readDiscU64(data)
	if !ok {
		return DexEvent{}
	}
	payload := data[8:]
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	pool := getAccountSafe(accounts, 0)
	if pool == zeroPubkey {
		return DexEvent{}
	}

	switch disc {
	case instrMeteoraPoolsSwap:
		if len(payload) < 16 {
			return DexEvent{}
		}
		inAmount, _ := readU64LE(payload, 0)
		minOut, _ := readU64LE(payload, 8)
		return DexEvent{
			Type: EventTypeMeteoraPoolsSwap,
			Data: &MeteoraPoolsSwapEvent{
				Metadata:  meta,
				InAmount:  inAmount,
				OutAmount: minOut,
				TradeFee:  0,
				AdminFee:  0,
				HostFee:   0,
			},
		}
	case instrMeteoraPoolsAddLiquidity:
		if len(payload) < 24 {
			return DexEvent{}
		}
		lp, _ := readU64LE(payload, 0)
		tokenA, _ := readU64LE(payload, 8)
		tokenB, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeMeteoraPoolsAddLiquidity,
			Data: &MeteoraPoolsAddLiquidityEvent{
				Metadata:     meta,
				LpMintAmount: lp,
				TokenAAmount: tokenA,
				TokenBAmount: tokenB,
			},
		}
	case instrMeteoraPoolsRemoveLiquidity:
		if len(payload) < 24 {
			return DexEvent{}
		}
		lp, _ := readU64LE(payload, 0)
		tokenA, _ := readU64LE(payload, 8)
		tokenB, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeMeteoraPoolsRemoveLiquidity,
			Data: &MeteoraPoolsRemoveLiquidityEvent{
				Metadata:        meta,
				LpUnmintAmount:  lp,
				TokenAOutAmount: tokenA,
				TokenBOutAmount: tokenB,
			},
		}
	case instrMeteoraPoolsCreatePool:
		if len(payload) < 1+6*8 || len(accounts) <= 9 {
			return DexEvent{}
		}
		return DexEvent{
			Type: EventTypeMeteoraPoolsPoolCreated,
			Data: &MeteoraPoolsPoolCreatedEvent{
				Metadata:   meta,
				LpMint:     getAccountSafe(accounts, 4),
				TokenAMint: getAccountSafe(accounts, 8),
				TokenBMint: getAccountSafe(accounts, 9),
				PoolType:   payload[0],
				Pool:       pool,
			},
		}
	default:
		return DexEvent{}
	}
}

// ParseMeteoraDlmmInstruction 对齐 Rust `meteora_dlmm::parse_instruction` 的外层指令路径。
func ParseMeteoraDlmmInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) == 0 {
		return DexEvent{}
	}
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	pool := getAccountSafe(accounts, 0)
	if pool == zeroPubkey {
		return DexEvent{}
	}
	payload := data[1:]

	switch data[0] {
	case 0:
		if len(payload) < 6 {
			return DexEvent{}
		}
		activeID, _ := readI32LE(payload, 0)
		binStep, _ := readU16LE(payload, 4)
		return DexEvent{
			Type: EventTypeMeteoraDlmmInitializePool,
			Data: &MeteoraDlmmInitializePoolEvent{
				Metadata:    meta,
				Pool:        pool,
				Creator:     getAccountSafe(accounts, 1),
				ActiveBinID: activeID,
				BinStep:     binStep,
			},
		}
	case 1:
		if len(payload) < 8 {
			return DexEvent{}
		}
		index, _ := readU64LE(payload, 0)
		return DexEvent{
			Type: EventTypeMeteoraDlmmInitializeBinArray,
			Data: &MeteoraDlmmInitializeBinArrayEvent{
				Metadata: meta,
				Pool:     pool,
				BinArray: getAccountSafe(accounts, 1),
				Index:    index,
			},
		}
	case 2:
		if len(payload) < 32 {
			return DexEvent{}
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmAddLiquidity,
			Data: &MeteoraDlmmAddLiquidityEvent{
				Metadata:    meta,
				Pool:        pool,
				From:        getAccountSafe(accounts, 1),
				Position:    getAccountSafe(accounts, 2),
				Amounts:     []uint64{0, 0},
				ActiveBinID: 0,
			},
		}
	case 7:
		if len(payload) < 32 {
			return DexEvent{}
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmRemoveLiquidity,
			Data: &MeteoraDlmmRemoveLiquidityEvent{
				Metadata:    meta,
				Pool:        pool,
				From:        getAccountSafe(accounts, 1),
				Position:    getAccountSafe(accounts, 2),
				Amounts:     []uint64{0, 0},
				ActiveBinID: 0,
			},
		}
	case 8:
		if len(payload) < 8 {
			return DexEvent{}
		}
		lowerBinID, _ := readI32LE(payload, 0)
		width, _ := readU32LE(payload, 4)
		return DexEvent{
			Type: EventTypeMeteoraDlmmCreatePosition,
			Data: &MeteoraDlmmCreatePositionEvent{
				Metadata:   meta,
				Pool:       pool,
				Position:   getAccountSafe(accounts, 1),
				Owner:      getAccountSafe(accounts, 2),
				LowerBinID: lowerBinID,
				Width:      width,
			},
		}
	case 11:
		if len(payload) < 16 {
			return DexEvent{}
		}
		amountIn, _ := readU64LE(payload, 0)
		return DexEvent{
			Type: EventTypeMeteoraDlmmSwap,
			Data: &MeteoraDlmmSwapEvent{
				Metadata:    meta,
				Pool:        pool,
				From:        getAccountSafe(accounts, 1),
				StartBinID:  0,
				EndBinID:    0,
				AmountIn:    amountIn,
				AmountOut:   0,
				SwapForY:    false,
				Fee:         0,
				ProtocolFee: 0,
				FeeBps:      "0",
				HostFee:     0,
			},
		}
	case 13:
		return DexEvent{
			Type: EventTypeMeteoraDlmmClaimFee,
			Data: &MeteoraDlmmClaimFeeEvent{
				Metadata: meta,
				Pool:     pool,
				Position: getAccountSafe(accounts, 1),
				Owner:    getAccountSafe(accounts, 2),
				FeeX:     0,
				FeeY:     0,
			},
		}
	case 14:
		return DexEvent{
			Type: EventTypeMeteoraDlmmClosePosition,
			Data: &MeteoraDlmmClosePositionEvent{
				Metadata: meta,
				Pool:     pool,
				Position: getAccountSafe(accounts, 1),
				Owner:    getAccountSafe(accounts, 2),
			},
		}
	default:
		return DexEvent{}
	}
}
