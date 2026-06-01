package solparser

func parseMeteoraDbcFromDiscriminator(disc uint64, data []byte, meta EventMetadata) DexEvent {
	switch disc {
	case discDbcSwap:
		return parseDbcSwap(data, meta)
	case discDbcInit:
		return parseDbcInitializePool(data, meta)
	case discDbcCurve:
		return parseDbcCurveComplete(data, meta)
	default:
		return DexEvent{}
	}
}

func parseDbcSwap(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*2+2+8*9+16 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	config, _ := readPubkey(data, o)
	o += 32
	tradeDirection, _ := readU8(data, o)
	o++
	hasReferral, _ := readBool(data, o)
	o++
	paramsAmountIn, _ := readU64LE(data, o)
	o += 8
	minimumAmountOut, _ := readU64LE(data, o)
	o += 8
	actualInputAmount, _ := readU64LE(data, o)
	o += 8
	outputAmount, _ := readU64LE(data, o)
	o += 8
	nextSqrtPrice, _ := readU128LE(data, o)
	o += 16
	tradingFee, _ := readU64LE(data, o)
	o += 8
	protocolFee, _ := readU64LE(data, o)
	o += 8
	referralFee, _ := readU64LE(data, o)
	o += 8
	amountIn := paramsAmountIn
	if v, ok := readU64LE(data, o); ok {
		amountIn = v
	}
	o += 8
	currentTimestamp, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDbcSwap,
		Data: &MeteoraDbcSwapEvent{
			Metadata:          meta,
			Pool:              pool,
			Config:            config,
			TradeDirection:    tradeDirection,
			HasReferral:       hasReferral,
			AmountIn:          amountIn,
			MinimumAmountOut:  minimumAmountOut,
			ActualInputAmount: actualInputAmount,
			OutputAmount:      outputAmount,
			NextSqrtPrice:     u128LEDecimalString(nextSqrtPrice),
			TradingFee:        tradingFee,
			ProtocolFee:       protocolFee,
			ReferralFee:       referralFee,
			CurrentTimestamp:  currentTimestamp,
		},
	}
}

func parseDbcInitializePool(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4+1+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	config, _ := readPubkey(data, o)
	o += 32
	creator, _ := readPubkey(data, o)
	o += 32
	baseMint, _ := readPubkey(data, o)
	o += 32
	poolType, _ := readU8(data, o)
	o++
	activationPoint, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDbcInitializePool,
		Data: &MeteoraDbcInitializePoolEvent{
			Metadata:        meta,
			Pool:            pool,
			Config:          config,
			Creator:         creator,
			BaseMint:        baseMint,
			PoolType:        poolType,
			ActivationPoint: activationPoint,
		},
	}
}

func parseDbcCurveComplete(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*2+8*2 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	config, _ := readPubkey(data, o)
	o += 32
	baseReserve, _ := readU64LE(data, o)
	o += 8
	quoteReserve, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDbcCurveComplete,
		Data: &MeteoraDbcCurveCompleteEvent{
			Metadata:     meta,
			Pool:         pool,
			Config:       config,
			BaseReserve:  baseReserve,
			QuoteReserve: quoteReserve,
		},
	}
}
