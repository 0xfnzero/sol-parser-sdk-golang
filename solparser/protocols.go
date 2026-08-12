package solparser

// Raydium CLMM
func parseClmmSwapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+32+8+8+8+8+1+16+16+4 {
		return DexEvent{}
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	sender, _ := readPubkey(data, o)
	o += 32
	tokenAccount0, _ := readPubkey(data, o)
	o += 32
	tokenAccount1, _ := readPubkey(data, o)
	o += 32
	amount0, _ := readU64LE(data, o)
	o += 8
	transferFee0, _ := readU64LE(data, o)
	o += 8
	amount1, _ := readU64LE(data, o)
	o += 8
	transferFee1, _ := readU64LE(data, o)
	o += 8
	zfo, _ := readBool(data, o)
	o++
	sqrt, _ := readU128LE(data, o)
	o += 16
	liquidity, _ := readU128LE(data, o)
	o += 16
	tick, _ := readI32LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmSwap,
		Data: &RaydiumClmmSwapEvent{
			Metadata:      meta,
			PoolState:     ps,
			Sender:        sender,
			TokenAccount0: tokenAccount0,
			TokenAccount1: tokenAccount1,
			Amount0:       amount0,
			Amount1:       amount1,
			ZeroForOne:    zfo,
			SqrtPriceX64:  u128LEDecimalString(sqrt),
			Liquidity:     u128LEDecimalString(liquidity),
			TransferFee0:  transferFee0,
			TransferFee1:  transferFee1,
			Tick:          tick,
		},
	}
}

func parseClmmIncFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+16+8+8+8+8 {
		return DexEvent{}
	}
	o := 0
	positionNftMint, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU128LE(data, o)
	o += 16
	amount0, _ := readU64LE(data, o)
	o += 8
	amount1, _ := readU64LE(data, o)
	o += 8
	amount0TransferFee, _ := readU64LE(data, o)
	o += 8
	amount1TransferFee, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmIncreaseLiquidity,
		Data: &RaydiumClmmIncreaseLiquidityEvent{
			Metadata:           meta,
			Pool:               zeroPubkey,
			PositionNftMint:    positionNftMint,
			User:               zeroPubkey,
			Liquidity:          u128LEDecimalString(liq),
			Amount0:            amount0,
			Amount1:            amount1,
			Amount0TransferFee: amount0TransferFee,
			Amount1TransferFee: amount1TransferFee,
		},
	}
}

func parseClmmDecFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+16+8+8+8+8+8+8+8+8+8 {
		return DexEvent{}
	}
	o := 0
	positionNftMint, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU128LE(data, o)
	o += 16
	decreaseAmount0, _ := readU64LE(data, o)
	o += 8
	decreaseAmount1, _ := readU64LE(data, o)
	o += 8
	feeAmount0, _ := readU64LE(data, o)
	o += 8
	feeAmount1, _ := readU64LE(data, o)
	o += 8
	reward0, _ := readU64LE(data, o)
	o += 8
	reward1, _ := readU64LE(data, o)
	o += 8
	reward2, _ := readU64LE(data, o)
	o += 8
	transferFee0, _ := readU64LE(data, o)
	o += 8
	transferFee1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmDecreaseLiquidity,
		Data: &RaydiumClmmDecreaseLiquidityEvent{
			Metadata:        meta,
			Pool:            zeroPubkey,
			PositionNftMint: positionNftMint,
			User:            zeroPubkey,
			Liquidity:       u128LEDecimalString(liq),
			DecreaseAmount0: decreaseAmount0,
			DecreaseAmount1: decreaseAmount1,
			FeeAmount0:      feeAmount0,
			FeeAmount1:      feeAmount1,
			RewardAmounts:   [3]uint64{reward0, reward1, reward2},
			TransferFee0:    transferFee0,
			TransferFee1:    transferFee1,
		},
	}
}

func parseClmmCreateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+2+32+16+4+32+32 {
		return DexEvent{}
	}
	o := 0
	token0Mint, _ := readPubkey(data, o)
	o += 32
	token1Mint, _ := readPubkey(data, o)
	o += 32
	tickSpacing, _ := readU16LE(data, o)
	o += 2
	pool, _ := readPubkey(data, o)
	o += 32
	sqrt, _ := readU128LE(data, o)
	o += 16
	tick, _ := readI32LE(data, o)
	o += 4
	tokenVault0, _ := readPubkey(data, o)
	o += 32
	tokenVault1, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmCreatePool,
		Data: &RaydiumClmmCreatePoolEvent{
			Metadata:     meta,
			Pool:         pool,
			Creator:      zeroPubkey,
			Token0Mint:   token0Mint,
			Token1Mint:   token1Mint,
			TickSpacing:  int(tickSpacing),
			FeeRate:      0,
			SqrtPriceX64: u128LEDecimalString(sqrt),
			Tick:         tick,
			TokenVault0:  tokenVault0,
			TokenVault1:  tokenVault1,
		},
	}
}

func parseClmmCollectPersonalFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8+8 {
		return DexEvent{}
	}
	o := 0
	pn, _ := readPubkey(data, o)
	o += 32
	recipient0, _ := readPubkey(data, o)
	o += 32
	recipient1, _ := readPubkey(data, o)
	o += 32
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmCollectFee,
		Data: &RaydiumClmmCollectFeeEvent{
			Metadata:               meta,
			PoolState:              zeroPubkey,
			PositionNftMint:        pn,
			RecipientTokenAccount0: recipient0,
			RecipientTokenAccount1: recipient1,
			Amount0:                a0,
			Amount1:                a1,
		},
	}
}

func parseClmmCollectProtocolFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8+8 {
		return DexEvent{}
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	recipient0, _ := readPubkey(data, o)
	o += 32
	recipient1, _ := readPubkey(data, o)
	o += 32
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmCollectFee,
		Data: &RaydiumClmmCollectFeeEvent{
			Metadata:               meta,
			PoolState:              ps,
			PositionNftMint:        zeroPubkey,
			RecipientTokenAccount0: recipient0,
			RecipientTokenAccount1: recipient1,
			Amount0:                a0,
			Amount1:                a1,
		},
	}
}

func parseClmmLiquidityChangeFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+4+4+4+16+16 {
		return DexEvent{}
	}
	o := 0
	poolState, _ := readPubkey(data, o)
	o += 32
	tick, _ := readI32LE(data, o)
	o += 4
	tickLower, _ := readI32LE(data, o)
	o += 4
	tickUpper, _ := readI32LE(data, o)
	o += 4
	before, _ := readU128LE(data, o)
	o += 16
	after, _ := readU128LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmLiquidityChange, Data: &RaydiumClmmLiquidityChangeEvent{
		Metadata:        meta,
		PoolState:       poolState,
		Tick:            tick,
		TickLower:       tickLower,
		TickUpper:       tickUpper,
		LiquidityBefore: u128LEDecimalString(before),
		LiquidityAfter:  u128LEDecimalString(after),
	}}
}

func parseClmmConfigChangeFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 2+32+4+4+2+4+32 {
		return DexEvent{}
	}
	o := 0
	index, _ := readU16LE(data, o)
	o += 2
	owner, _ := readPubkey(data, o)
	o += 32
	protocolFeeRate, _ := readU32LE(data, o)
	o += 4
	tradeFeeRate, _ := readU32LE(data, o)
	o += 4
	tickSpacing, _ := readU16LE(data, o)
	o += 2
	fundFeeRate, _ := readU32LE(data, o)
	o += 4
	fundOwner, _ := readPubkey(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmConfigChange, Data: &RaydiumClmmConfigChangeEvent{
		Metadata:        meta,
		Index:           index,
		Owner:           owner,
		ProtocolFeeRate: protocolFeeRate,
		TradeFeeRate:    tradeFeeRate,
		TickSpacing:     tickSpacing,
		FundFeeRate:     fundFeeRate,
		FundOwner:       fundOwner,
	}}
}

func parseClmmCreatePersonalPositionFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+4+4+16+8+8+8+8 {
		return DexEvent{}
	}
	o := 0
	poolState, _ := readPubkey(data, o)
	o += 32
	minter, _ := readPubkey(data, o)
	o += 32
	nftOwner, _ := readPubkey(data, o)
	o += 32
	tickLowerIndex, _ := readI32LE(data, o)
	o += 4
	tickUpperIndex, _ := readI32LE(data, o)
	o += 4
	liquidity, _ := readU128LE(data, o)
	o += 16
	depositAmount0, _ := readU64LE(data, o)
	o += 8
	depositAmount1, _ := readU64LE(data, o)
	o += 8
	depositAmount0TransferFee, _ := readU64LE(data, o)
	o += 8
	depositAmount1TransferFee, _ := readU64LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmCreatePersonalPosition, Data: &RaydiumClmmCreatePersonalPositionEvent{
		Metadata:                  meta,
		PoolState:                 poolState,
		Minter:                    minter,
		NftOwner:                  nftOwner,
		TickLowerIndex:            tickLowerIndex,
		TickUpperIndex:            tickUpperIndex,
		Liquidity:                 u128LEDecimalString(liquidity),
		DepositAmount0:            depositAmount0,
		DepositAmount1:            depositAmount1,
		DepositAmount0TransferFee: depositAmount0TransferFee,
		DepositAmount1TransferFee: depositAmount1TransferFee,
	}}
}

func parseClmmLiquidityCalculateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 16+16+4+8+8+8+8+8+8 {
		return DexEvent{}
	}
	o := 0
	poolLiquidity, _ := readU128LE(data, o)
	o += 16
	poolSqrtPriceX64, _ := readU128LE(data, o)
	o += 16
	poolTick, _ := readI32LE(data, o)
	o += 4
	calcAmount0, _ := readU64LE(data, o)
	o += 8
	calcAmount1, _ := readU64LE(data, o)
	o += 8
	tradeFeeOwed0, _ := readU64LE(data, o)
	o += 8
	tradeFeeOwed1, _ := readU64LE(data, o)
	o += 8
	transferFee0, _ := readU64LE(data, o)
	o += 8
	transferFee1, _ := readU64LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmLiquidityCalculate, Data: &RaydiumClmmLiquidityCalculateEvent{
		Metadata:         meta,
		PoolLiquidity:    u128LEDecimalString(poolLiquidity),
		PoolSqrtPriceX64: u128LEDecimalString(poolSqrtPriceX64),
		PoolTick:         poolTick,
		CalcAmount0:      calcAmount0,
		CalcAmount1:      calcAmount1,
		TradeFeeOwed0:    tradeFeeOwed0,
		TradeFeeOwed1:    tradeFeeOwed1,
		TransferFee0:     transferFee0,
		TransferFee1:     transferFee1,
	}}
}

func parseClmmOpenLimitOrderFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+4+8+8 {
		return DexEvent{}
	}
	o := 0
	poolID, _ := readPubkey(data, o)
	o += 32
	limitOrder, _ := readPubkey(data, o)
	o += 32
	zeroForOne, _ := readBool(data, o)
	o += 1
	tickIndex, _ := readI32LE(data, o)
	o += 4
	totalAmount, _ := readU64LE(data, o)
	o += 8
	transferFee, _ := readU64LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmOpenLimitOrder, Data: &RaydiumClmmOpenLimitOrderEvent{
		Metadata:    meta,
		PoolID:      poolID,
		LimitOrder:  limitOrder,
		ZeroForOne:  zeroForOne,
		TickIndex:   tickIndex,
		TotalAmount: totalAmount,
		TransferFee: transferFee,
	}}
}

func parseClmmIncreaseLimitOrderFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+4+8+8+8 {
		return DexEvent{}
	}
	o := 0
	poolID, _ := readPubkey(data, o)
	o += 32
	limitOrder, _ := readPubkey(data, o)
	o += 32
	zeroForOne, _ := readBool(data, o)
	o += 1
	tickIndex, _ := readI32LE(data, o)
	o += 4
	totalAmount, _ := readU64LE(data, o)
	o += 8
	increasedAmount, _ := readU64LE(data, o)
	o += 8
	transferFee, _ := readU64LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmIncreaseLimitOrder, Data: &RaydiumClmmIncreaseLimitOrderEvent{
		Metadata:        meta,
		PoolID:          poolID,
		LimitOrder:      limitOrder,
		ZeroForOne:      zeroForOne,
		TickIndex:       tickIndex,
		TotalAmount:     totalAmount,
		IncreasedAmount: increasedAmount,
		TransferFee:     transferFee,
	}}
}

func parseClmmDecreaseLimitOrderFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+4+8+8+8+8 {
		return DexEvent{}
	}
	o := 0
	poolID, _ := readPubkey(data, o)
	o += 32
	limitOrder, _ := readPubkey(data, o)
	o += 32
	zeroForOne, _ := readBool(data, o)
	o += 1
	tickIndex, _ := readI32LE(data, o)
	o += 4
	totalAmount, _ := readU64LE(data, o)
	o += 8
	filledAmount, _ := readU64LE(data, o)
	o += 8
	settledOutputAmount, _ := readU64LE(data, o)
	o += 8
	decreasedAmount, _ := readU64LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmDecreaseLimitOrder, Data: &RaydiumClmmDecreaseLimitOrderEvent{
		Metadata:            meta,
		PoolID:              poolID,
		LimitOrder:          limitOrder,
		ZeroForOne:          zeroForOne,
		TickIndex:           tickIndex,
		TotalAmount:         totalAmount,
		FilledAmount:        filledAmount,
		SettledOutputAmount: settledOutputAmount,
		DecreasedAmount:     decreasedAmount,
	}}
}

func parseClmmSettleLimitOrderFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+4+8+8+8 {
		return DexEvent{}
	}
	o := 0
	poolID, _ := readPubkey(data, o)
	o += 32
	limitOrder, _ := readPubkey(data, o)
	o += 32
	zeroForOne, _ := readBool(data, o)
	o += 1
	tickIndex, _ := readI32LE(data, o)
	o += 4
	totalAmount, _ := readU64LE(data, o)
	o += 8
	filledAmount, _ := readU64LE(data, o)
	o += 8
	settledAmountOut, _ := readU64LE(data, o)
	return DexEvent{Type: EventTypeRaydiumClmmSettleLimitOrder, Data: &RaydiumClmmSettleLimitOrderEvent{
		Metadata:         meta,
		PoolID:           poolID,
		LimitOrder:       limitOrder,
		ZeroForOne:       zeroForOne,
		TickIndex:        tickIndex,
		TotalAmount:      totalAmount,
		FilledAmount:     filledAmount,
		SettledAmountOut: settledAmountOut,
	}}
}

func parseClmmUpdateRewardInfosFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 16*3 {
		return DexEvent{}
	}
	rewards := make([]string, 3)
	for i := 0; i < 3; i++ {
		raw, _ := readU128LE(data, i*16)
		rewards[i] = u128LEDecimalString(raw)
	}
	return DexEvent{Type: EventTypeRaydiumClmmUpdateRewardInfos, Data: &RaydiumClmmUpdateRewardInfosEvent{
		Metadata:              meta,
		RewardGrowthGlobalX64: rewards,
	}}
}

// Raydium CPMM
func parseCpmmSwapEventFromData(data []byte, meta EventMetadata) DexEvent {
	const payloadLen = 32 + 6*8 + 1
	if len(data) < payloadLen {
		return DexEvent{}
	}
	pool, ok := readPubkey(data, 0)
	if !ok {
		return DexEvent{}
	}
	inputVaultBefore, _ := readU64LE(data, 32)
	outputVaultBefore, _ := readU64LE(data, 40)
	inputAmount, _ := readU64LE(data, 48)
	outputAmount, _ := readU64LE(data, 56)
	inputTransferFee, _ := readU64LE(data, 64)
	outputTransferFee, _ := readU64LE(data, 72)
	baseInput, _ := readBool(data, 80)
	return DexEvent{Type: EventTypeRaydiumCpmmSwap, Data: &RaydiumCpmmSwapEvent{
		Metadata:          meta,
		PoolID:            pool,
		InputAmount:       inputAmount,
		OutputAmount:      outputAmount,
		InputVaultBefore:  inputVaultBefore,
		OutputVaultBefore: outputVaultBefore,
		InputTransferFee:  inputTransferFee,
		OutputTransferFee: outputTransferFee,
		BaseInput:         baseInput,
	}}
}

func parseCpmmSwapInFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+1 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 64
	ai, _ := readU64LE(data, o)
	o += 16
	ao, _ := readU64LE(data, o)
	o += 8
	bi, _ := readBool(data, o)
	return DexEvent{
		Type: EventTypeRaydiumCpmmSwap,
		Data: &RaydiumCpmmSwapEvent{
			Metadata:          meta,
			PoolID:            pool,
			InputAmount:       ai,
			OutputAmount:      ao,
			InputVaultBefore:  0,
			OutputVaultBefore: 0,
			InputTransferFee:  0,
			OutputTransferFee: 0,
			BaseInput:         bi,
		},
	}
}

func parseCpmmSwapOutFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+1 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 64
	o += 8
	ao, _ := readU64LE(data, o)
	o += 8
	ai, _ := readU64LE(data, o)
	o += 8
	bo, _ := readBool(data, o)
	return DexEvent{
		Type: EventTypeRaydiumCpmmSwap,
		Data: &RaydiumCpmmSwapEvent{
			Metadata:          meta,
			PoolID:            pool,
			InputAmount:       ai,
			OutputAmount:      ao,
			InputVaultBefore:  0,
			OutputVaultBefore: 0,
			InputTransferFee:  0,
			OutputTransferFee: 0,
			BaseInput:         !bo,
		},
	}
}

func parseCpmmDepositFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	lp, _ := readU64LE(data, o)
	o += 8
	t0, _ := readU64LE(data, o)
	o += 8
	t1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumCpmmDeposit,
		Data: &RaydiumCpmmDepositEvent{
			Metadata:      meta,
			Pool:          pool,
			User:          user,
			LpTokenAmount: lp,
			Token0Amount:  t0,
			Token1Amount:  t1,
		},
	}
}

func parseCpmmWithdrawFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	lp, _ := readU64LE(data, o)
	o += 8
	t0, _ := readU64LE(data, o)
	o += 8
	t1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumCpmmWithdraw,
		Data: &RaydiumCpmmWithdrawEvent{
			Metadata:      meta,
			Pool:          pool,
			User:          user,
			LpTokenAmount: lp,
			Token0Amount:  t0,
			Token1Amount:  t1,
		},
	}
}

func parseCpmmInitFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+32+8+8 {
		return DexEvent{}
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	o += 32
	o += 32
	cr, _ := readPubkey(data, o)
	o += 32
	i0, _ := readU64LE(data, o)
	o += 8
	i1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumCpmmInitialize,
		Data: &RaydiumCpmmInitializeEvent{
			Metadata:    meta,
			Pool:        ps,
			Creator:     cr,
			InitAmount0: i0,
			InitAmount1: i1,
		},
	}
}

// Orca Whirlpool
func parseOrcaTradedFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+1+16+16+8*6 {
		return DexEvent{}
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	atb, _ := readBool(data, o)
	o++
	pre, _ := readU128LE(data, o)
	o += 16
	post, _ := readU128LE(data, o)
	o += 16
	ia, _ := readU64LE(data, o)
	o += 8
	oa, _ := readU64LE(data, o)
	o += 8
	itf, _ := readU64LE(data, o)
	o += 8
	otf, _ := readU64LE(data, o)
	o += 8
	lpf, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeOrcaWhirlpoolSwap,
		Data: &OrcaWhirlpoolSwapEvent{
			Metadata:          meta,
			Whirlpool:         w,
			AToB:              atb,
			PreSqrtPrice:      u128LEDecimalString(pre),
			PostSqrtPrice:     u128LEDecimalString(post),
			InputAmount:       ia,
			OutputAmount:      oa,
			InputTransferFee:  itf,
			OutputTransferFee: otf,
			LpFee:             lpf,
			ProtocolFee:       pf,
		},
	}
}

func parseOrcaLiqIncFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+4+16+8*4 {
		return DexEvent{}
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	p, _ := readPubkey(data, o)
	o += 32
	tl, _ := readI32LE(data, o)
	o += 4
	tu, _ := readI32LE(data, o)
	o += 4
	liq, _ := readU128LE(data, o)
	o += 16
	ta, _ := readU64LE(data, o)
	o += 8
	tb, _ := readU64LE(data, o)
	o += 8
	taf, _ := readU64LE(data, o)
	o += 8
	tbf, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeOrcaWhirlpoolLiquidityIncreased,
		Data: &OrcaWhirlpoolLiquidityIncreasedEvent{
			Metadata:          meta,
			Whirlpool:         w,
			Position:          p,
			TickLowerIndex:    tl,
			TickUpperIndex:    tu,
			Liquidity:         u128LEDecimalString(liq),
			TokenAAmount:      ta,
			TokenBAmount:      tb,
			TokenATransferFee: taf,
			TokenBTransferFee: tbf,
		},
	}
}

func parseOrcaLiqDecFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+4+16+8*4 {
		return DexEvent{}
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	p, _ := readPubkey(data, o)
	o += 32
	tl, _ := readI32LE(data, o)
	o += 4
	tu, _ := readI32LE(data, o)
	o += 4
	liq, _ := readU128LE(data, o)
	o += 16
	ta, _ := readU64LE(data, o)
	o += 8
	tb, _ := readU64LE(data, o)
	o += 8
	taf, _ := readU64LE(data, o)
	o += 8
	tbf, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeOrcaWhirlpoolLiquidityDecreased,
		Data: &OrcaWhirlpoolLiquidityDecreasedEvent{
			Metadata:          meta,
			Whirlpool:         w,
			Position:          p,
			TickLowerIndex:    tl,
			TickUpperIndex:    tu,
			Liquidity:         u128LEDecimalString(liq),
			TokenAAmount:      ta,
			TokenBAmount:      tb,
			TokenATransferFee: taf,
			TokenBTransferFee: tbf,
		},
	}
}

func parseOrcaPoolInitFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*5+2+1+1+16 {
		return DexEvent{}
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	cfg, _ := readPubkey(data, o)
	o += 32
	ma, _ := readPubkey(data, o)
	o += 32
	mb, _ := readPubkey(data, o)
	o += 32
	ts, _ := readU16LE(data, o)
	o += 2
	tpa, _ := readPubkey(data, o)
	o += 32
	tpb, _ := readPubkey(data, o)
	o += 32
	da, _ := readU8(data, o)
	o++
	db, _ := readU8(data, o)
	o++
	isp, _ := readU128LE(data, o)
	return DexEvent{
		Type: EventTypeOrcaWhirlpoolPoolInitialized,
		Data: &OrcaWhirlpoolPoolInitializedEvent{
			Metadata:         meta,
			Whirlpool:        w,
			WhirlpoolsConfig: cfg,
			TokenMintA:       ma,
			TokenMintB:       mb,
			TickSpacing:      ts,
			TokenProgramA:    tpa,
			TokenProgramB:    tpb,
			DecimalsA:        da,
			DecimalsB:        db,
			InitialSqrtPrice: u128LEDecimalString(isp),
		},
	}
}

// Meteora Pools
func parseMeteoraSwapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8*5 {
		return DexEvent{}
	}
	o := 0
	return DexEvent{
		Type: EventTypeMeteoraPoolsSwap,
		Data: &MeteoraPoolsSwapEvent{
			Metadata:  meta,
			InAmount:  readU64At(data, &o),
			OutAmount: readU64At(data, &o),
			TradeFee:  readU64At(data, &o),
			AdminFee:  readU64At(data, &o),
			HostFee:   readU64At(data, &o),
		},
	}
}

func readU64At(b []byte, o *int) uint64 {
	v, _ := readU64LE(b, *o)
	*o += 8
	return v
}

func parseMeteoraAddFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 24 {
		return DexEvent{}
	}
	o := 0
	return DexEvent{
		Type: EventTypeMeteoraPoolsAddLiquidity,
		Data: &MeteoraPoolsAddLiquidityEvent{
			Metadata:     meta,
			LpMintAmount: readU64At(data, &o),
			TokenAAmount: readU64At(data, &o),
			TokenBAmount: readU64At(data, &o),
		},
	}
}

func parseMeteoraRemoveFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 24 {
		return DexEvent{}
	}
	o := 0
	return DexEvent{
		Type: EventTypeMeteoraPoolsRemoveLiquidity,
		Data: &MeteoraPoolsRemoveLiquidityEvent{
			Metadata:        meta,
			LpUnmintAmount:  readU64At(data, &o),
			TokenAOutAmount: readU64At(data, &o),
			TokenBOutAmount: readU64At(data, &o),
		},
	}
}

func parseMeteoraBootstrapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 24+32 {
		return DexEvent{}
	}
	o := 0
	lp := readU64At(data, &o)
	ta := readU64At(data, &o)
	tb := readU64At(data, &o)
	pl, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeMeteoraPoolsBootstrapLiquidity,
		Data: &MeteoraPoolsBootstrapLiquidityEvent{
			Metadata:     meta,
			LpMintAmount: lp,
			TokenAAmount: ta,
			TokenBAmount: tb,
			Pool:         pl,
		},
	}
}

func parseMeteoraPoolCreatedFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4+1 {
		return DexEvent{}
	}
	o := 0
	lm, _ := readPubkey(data, o)
	o += 32
	ta, _ := readPubkey(data, o)
	o += 32
	tb, _ := readPubkey(data, o)
	o += 32
	pt, _ := readU8(data, o)
	o++
	pl, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeMeteoraPoolsPoolCreated,
		Data: &MeteoraPoolsPoolCreatedEvent{
			Metadata:   meta,
			LpMint:     lm,
			TokenAMint: ta,
			TokenBMint: tb,
			PoolType:   pt,
			Pool:       pl,
		},
	}
}

// Raydium AMM V4
func parseAmmRayLogSwap(data []byte, meta EventMetadata) DexEvent {
	// Official bincode format: one u8 log type followed by seven u64 values.
	if len(data) != 57 || (data[0] != 3 && data[0] != 4) {
		return DexEvent{}
	}
	first, _ := readU64LE(data, 1)
	second, _ := readU64LE(data, 9)
	actual, _ := readU64LE(data, 49)
	event := &RaydiumAmmV4SwapEvent{Metadata: meta, Amm: zeroPubkey, UserSourceOwner: zeroPubkey}
	if data[0] == 3 {
		event.AmountIn = first
		event.MinimumAmountOut = second
		event.AmountOut = actual
	} else {
		event.MaxAmountIn = first
		event.AmountOut = second
		event.AmountIn = actual
	}
	return DexEvent{Type: EventTypeRaydiumAmmV4Swap, Data: event}
}

func parseAmmSwapInFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8+8+8 {
		return DexEvent{}
	}
	o := 0
	ai, _ := readU64LE(data, o)
	o += 8
	mao, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumAmmV4Swap,
		Data: &RaydiumAmmV4SwapEvent{
			Metadata:         meta,
			AmountIn:         ai,
			MinimumAmountOut: mao,
		},
	}
}

func parseAmmSwapOutFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8+8+8 {
		return DexEvent{}
	}
	o := 0
	mai, _ := readU64LE(data, o)
	o += 8
	ao, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumAmmV4Swap,
		Data: &RaydiumAmmV4SwapEvent{
			Metadata:    meta,
			MaxAmountIn: mai,
			AmountOut:   ao,
		},
	}
}

func parseAmmDepositFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8+8+8 {
		return DexEvent{}
	}
	o := 0
	mca, _ := readU64LE(data, o)
	o += 8
	mpa, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumAmmV4Deposit,
		Data: &RaydiumAmmV4DepositEvent{
			Metadata:      meta,
			MaxCoinAmount: mca,
			MaxPcAmount:   mpa,
		},
	}
}

func parseAmmWithdrawFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}
	o := 0
	amt, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumAmmV4Withdraw,
		Data: &RaydiumAmmV4WithdrawEvent{
			Metadata: meta,
			Amount:   amt,
		},
	}
}

func parseAmmWithdrawPnlFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}
	o := 0
	amt, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumAmmV4WithdrawPnl,
		Data: &RaydiumAmmV4WithdrawEvent{
			Metadata: meta,
			Amount:   amt,
		},
	}
}

func parseAmmInit2FromData(data []byte, meta EventMetadata) DexEvent {
	return DexEvent{
		Type: EventTypeRaydiumAmmV4Initialize2,
		Data: &RaydiumAmmV4DepositEvent{Metadata: meta},
	}
}
