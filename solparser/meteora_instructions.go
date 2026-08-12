package solparser

// ParseMeteoraPoolsInstruction mirrors the Rust outer-instruction parser.
func ParseMeteoraPoolsInstruction(
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
	disc, _ := readDiscU64(data)
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
		outAmount, _ := readU64LE(payload, 8)
		return DexEvent{
			Type: EventTypeMeteoraPoolsSwap,
			Data: &MeteoraPoolsSwapEvent{
				Metadata:  meta,
				InAmount:  inAmount,
				OutAmount: outAmount,
				TradeFee:  0,
				AdminFee:  0,
				HostFee:   0,
			},
		}
	case instrMeteoraPoolsAddLiquidity:
		if len(payload) < 24 {
			return DexEvent{}
		}
		lpAmount, _ := readU64LE(payload, 0)
		tokenA, _ := readU64LE(payload, 8)
		tokenB, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeMeteoraPoolsAddLiquidity,
			Data: &MeteoraPoolsAddLiquidityEvent{
				Metadata:     meta,
				LpMintAmount: lpAmount,
				TokenAAmount: tokenA,
				TokenBAmount: tokenB,
			},
		}
	case instrMeteoraPoolsRemoveLiquidity:
		if len(payload) < 24 {
			return DexEvent{}
		}
		lpAmount, _ := readU64LE(payload, 0)
		tokenA, _ := readU64LE(payload, 8)
		tokenB, _ := readU64LE(payload, 16)
		return DexEvent{
			Type: EventTypeMeteoraPoolsRemoveLiquidity,
			Data: &MeteoraPoolsRemoveLiquidityEvent{
				Metadata:        meta,
				LpUnmintAmount:  lpAmount,
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

// ParseMeteoraDlmmInstruction follows the current DLMM Anchor IDL layout.
func ParseMeteoraDlmmInstruction(
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
	disc, _ := readDiscU64(data)
	payload := data[8:]
	meta := makeInstrMetadata(signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	switch disc {
	case instrDlmmInitializeLbPair:
		if len(payload) < 6 {
			return DexEvent{}
		}
		activeID, _ := readI32LE(payload, 0)
		binStep, _ := readU16LE(payload, 4)
		pool := getAccountSafe(accounts, 0)
		return DexEvent{
			Type: EventTypeMeteoraDlmmInitializePool,
			Data: &MeteoraDlmmInitializePoolEvent{
				Metadata:    meta,
				Pool:        pool,
				Creator:     getAccountSafe(accounts, 8),
				ActiveBinID: activeID,
				BinStep:     binStep,
			},
		}
	case instrDlmmInitializeLbPair2:
		if len(payload) < 4 {
			return DexEvent{}
		}
		activeID, _ := readI32LE(payload, 0)
		pool := getAccountSafe(accounts, 0)
		return DexEvent{
			Type: EventTypeMeteoraDlmmInitializePool,
			Data: &MeteoraDlmmInitializePoolEvent{
				Metadata:    meta,
				Pool:        pool,
				Creator:     getAccountSafe(accounts, 8),
				ActiveBinID: activeID,
				BinStep:     0,
			},
		}
	case instrDlmmInitializeBinArray:
		if len(payload) < 8 {
			return DexEvent{}
		}
		index, _ := readI64LE(payload, 0)
		return DexEvent{
			Type: EventTypeMeteoraDlmmInitializeBinArray,
			Data: &MeteoraDlmmInitializeBinArrayEvent{
				Metadata: meta,
				Pool:     getAccountSafe(accounts, 0),
				BinArray: getAccountSafe(accounts, 1),
				Index:    index,
			},
		}
	case instrDlmmAddLiquidity, instrDlmmAddLiquidity2:
		senderIndex := 11
		if disc == instrDlmmAddLiquidity2 {
			senderIndex = 9
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmAddLiquidity,
			Data: &MeteoraDlmmAddLiquidityEvent{
				Metadata:    meta,
				Pool:        getAccountSafe(accounts, 1),
				From:        getAccountSafe(accounts, senderIndex),
				Position:    getAccountSafe(accounts, 0),
				Amounts:     []uint64{0, 0},
				ActiveBinID: 0,
			},
		}
	case instrDlmmRemoveLiquidity, instrDlmmRemoveLiquidity2:
		senderIndex := 11
		if disc == instrDlmmRemoveLiquidity2 {
			senderIndex = 9
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmRemoveLiquidity,
			Data: &MeteoraDlmmRemoveLiquidityEvent{
				Metadata:    meta,
				Pool:        getAccountSafe(accounts, 1),
				From:        getAccountSafe(accounts, senderIndex),
				Position:    getAccountSafe(accounts, 0),
				Amounts:     []uint64{0, 0},
				ActiveBinID: 0,
			},
		}
	case instrDlmmInitializePosition, instrDlmmInitializePosition2, instrDlmmInitializePositionPda:
		if len(payload) < 8 {
			return DexEvent{}
		}
		lowerBinID, _ := readI32LE(payload, 0)
		widthSigned, _ := readI32LE(payload, 4)
		if widthSigned < 0 {
			return DexEvent{}
		}
		positionIndex, poolIndex, ownerIndex := 1, 2, 3
		if disc == instrDlmmInitializePositionPda {
			positionIndex, poolIndex, ownerIndex = 2, 3, 4
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmCreatePosition,
			Data: &MeteoraDlmmCreatePositionEvent{
				Metadata:   meta,
				Pool:       getAccountSafe(accounts, poolIndex),
				Position:   getAccountSafe(accounts, positionIndex),
				Owner:      getAccountSafe(accounts, ownerIndex),
				LowerBinID: lowerBinID,
				Width:      uint32(widthSigned),
			},
		}
	case instrDlmmSwap, instrDlmmSwap2:
		if len(payload) < 8 {
			return DexEvent{}
		}
		amountIn, _ := readU64LE(payload, 0)
		return dlmmSwapInstructionEvent(meta, accounts, amountIn, 0)
	case instrDlmmSwapExactOut, instrDlmmSwapExactOut2:
		if len(payload) < 16 {
			return DexEvent{}
		}
		amountIn, _ := readU64LE(payload, 0)
		amountOut, _ := readU64LE(payload, 8)
		return dlmmSwapInstructionEvent(meta, accounts, amountIn, amountOut)
	case instrDlmmSwapWithPriceImpact, instrDlmmSwapWithPriceImpact2:
		if len(payload) < 8 {
			return DexEvent{}
		}
		amountIn, _ := readU64LE(payload, 0)
		return dlmmSwapInstructionEvent(meta, accounts, amountIn, 0)
	case instrDlmmClaimFee, instrDlmmClaimFee2:
		ownerIndex := 4
		if disc == instrDlmmClaimFee2 {
			ownerIndex = 2
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmClaimFee,
			Data: &MeteoraDlmmClaimFeeEvent{
				Metadata: meta,
				Pool:     getAccountSafe(accounts, 0),
				Position: getAccountSafe(accounts, 1),
				Owner:    getAccountSafe(accounts, ownerIndex),
				FeeX:     0,
				FeeY:     0,
			},
		}
	case instrDlmmClosePosition, instrDlmmClosePosition2:
		pool := getAccountSafe(accounts, 1)
		ownerIndex := 4
		if disc == instrDlmmClosePosition2 {
			pool = zeroPubkey
			ownerIndex = 1
		}
		return DexEvent{
			Type: EventTypeMeteoraDlmmClosePosition,
			Data: &MeteoraDlmmClosePositionEvent{
				Metadata: meta,
				Pool:     pool,
				Position: getAccountSafe(accounts, 0),
				Owner:    getAccountSafe(accounts, ownerIndex),
			},
		}
	default:
		return DexEvent{}
	}
}

func dlmmSwapInstructionEvent(meta EventMetadata, accounts []string, amountIn, amountOut uint64) DexEvent {
	return DexEvent{
		Type: EventTypeMeteoraDlmmSwap,
		Data: &MeteoraDlmmSwapEvent{
			Metadata:    meta,
			Pool:        getAccountSafe(accounts, 0),
			From:        getAccountSafe(accounts, 10),
			StartBinID:  0,
			EndBinID:    0,
			AmountIn:    amountIn,
			AmountOut:   amountOut,
			SwapForY:    false,
			Fee:         0,
			ProtocolFee: 0,
			FeeBps:      "0",
			HostFee:     0,
		},
	}
}
