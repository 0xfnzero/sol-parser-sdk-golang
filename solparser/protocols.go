package solparser

// Raydium CLMM
func parseClmmSwapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+16+1 {
		return DexEvent{}
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	o += 8
	o += 8
	sqrt, _ := readU128LE(data, o)
	o += 16
	zfo, _ := readBool(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmSwap,
		Data: &RaydiumClmmSwapEvent{
			Metadata:      meta,
			PoolState:     ps,
			Sender:        user,
			TokenAccount0: zeroPubkey,
			TokenAccount1: zeroPubkey,
			Amount0:       0,
			Amount1:       0,
			ZeroForOne:    zfo,
			SqrtPriceX64:  u128LEDecimalString(sqrt),
			Liquidity:     "0",
			TransferFee0:  0,
			TransferFee1:  0,
			Tick:          0,
		},
	}
}

func parseClmmIncFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+16+8+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU128LE(data, o)
	o += 16
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmIncreaseLiquidity,
		Data: &RaydiumClmmIncreaseLiquidityEvent{
			Metadata:        meta,
			Pool:            pool,
			PositionNftMint: zeroPubkey,
			User:            user,
			Liquidity:       u128LEDecimalString(liq),
			Amount0Max:      a0,
			Amount1Max:      a1,
		},
	}
}

func parseClmmDecFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+16+8+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU128LE(data, o)
	o += 16
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmDecreaseLiquidity,
		Data: &RaydiumClmmDecreaseLiquidityEvent{
			Metadata:        meta,
			Pool:            pool,
			PositionNftMint: zeroPubkey,
			User:            user,
			Liquidity:       u128LEDecimalString(liq),
			Amount0Min:      a0,
			Amount1Min:      a1,
		},
	}
}

func parseClmmCreateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+16+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	cr, _ := readPubkey(data, o)
	o += 32
	sqrt, _ := readU128LE(data, o)
	o += 16
	ot, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmCreatePool,
		Data: &RaydiumClmmCreatePoolEvent{
			Metadata:     meta,
			Pool:         pool,
			Creator:      cr,
			Token0Mint:   zeroPubkey,
			Token1Mint:   zeroPubkey,
			TickSpacing:  0,
			FeeRate:      0,
			SqrtPriceX64: u128LEDecimalString(sqrt),
			OpenTime:     ot,
		},
	}
}

func parseClmmCollectFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8 {
		return DexEvent{}
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	pn, _ := readPubkey(data, o)
	o += 32
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeRaydiumClmmCollectFee,
		Data: &RaydiumClmmCollectFeeEvent{
			Metadata:        meta,
			PoolState:       ps,
			PositionNftMint: pn,
			Amount0:         a0,
			Amount1:         a1,
		},
	}
}

// Raydium CPMM
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
			Metadata:         meta,
			Whirlpool:        w,
			AToB:             atb,
			PreSqrtPrice:     u128LEDecimalString(pre),
			PostSqrtPrice:    u128LEDecimalString(post),
			InputAmount:      ia,
			OutputAmount:     oa,
			InputTransferFee: itf,
			OutputTransferFee: otf,
			LpFee:            lpf,
			ProtocolFee:      pf,
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
			Metadata:          meta,
			Whirlpool:         w,
			WhirlpoolsConfig:  cfg,
			TokenMintA:        ma,
			TokenMintB:        mb,
			TickSpacing:       ts,
			TokenProgramA:     tpa,
			TokenProgramB:     tpb,
			DecimalsA:         da,
			DecimalsB:         db,
			InitialSqrtPrice:  u128LEDecimalString(isp),
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
			Metadata:         meta,
			LpUnmintAmount:   readU64At(data, &o),
			TokenAOutAmount:  readU64At(data, &o),
			TokenBOutAmount:  readU64At(data, &o),
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
			Metadata:         meta,
			MaxAmountIn:      mai,
			AmountOut:        ao,
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
