package solparser

// RaydiumLaunchlab discriminators 已在 binary.go 中定义

func parseRaydiumLaunchlabTradeFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) != 139 || data[136] > 1 || data[137] > 2 || data[138] > 1 {
		return DexEvent{}
	}
	pool, _ := readPubkey(data, 0)
	ai, _ := readU64LE(data, 88)
	ao, _ := readU64LE(data, 96)
	dirByte, _ := readU8(data, 136)
	exIn, _ := readBool(data, 138)
	isBuy := dirByte == 0
	dir := "Sell"
	if isBuy {
		dir = "Buy"
	}
	number := func(o int) uint64 { v, _ := readU64LE(data, o); return v }

	return DexEvent{
		Type: EventTypeRaydiumLaunchlabTrade,
		Data: &RaydiumLaunchlabTradeEvent{
			Metadata:        meta,
			PoolState:       pool,
			User:            zeroPubkey,
			AmountIn:        ai,
			AmountOut:       ao,
			IsBuy:           isBuy,
			TradeDirection:  dir,
			TotalBaseSell:   number(32),
			VirtualBase:     number(40),
			VirtualQuote:    number(48),
			RealBaseBefore:  number(56),
			RealQuoteBefore: number(64),
			RealBaseAfter:   number(72),
			RealQuoteAfter:  number(80),
			ProtocolFee:     number(104),
			PlatformFee:     number(112),
			CreatorFee:      number(120),
			ShareFee:        number(128),
			PoolStatus:      []string{"Fund", "Migrate", "Trade"}[data[137]],
			ExactIn:         exIn,
		},
	}
}

func parseRaydiumLaunchlabPoolCreateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 97 {
		return DexEvent{}
	}
	pool, _ := readPubkey(data, 0)
	creator, _ := readPubkey(data, 32)
	decimals, _ := readU8(data, 96)
	o := 97
	name, next, ok := readBorshString(data, o)
	if !ok {
		return DexEvent{}
	}
	o = next
	symbol, next, ok := readBorshString(data, o)
	if !ok {
		return DexEvent{}
	}
	o = next
	uri, _, ok := readBorshString(data, o)
	if !ok {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeRaydiumLaunchlabPoolCreate,
		Data: &RaydiumLaunchlabPoolCreateEvent{
			Metadata: meta,
			BaseMintParam: RaydiumLaunchlabMintParam{
				Symbol:   symbol,
				Name:     name,
				Uri:      uri,
				Decimals: decimals,
			},
			PoolState: pool,
			Creator:   creator,
		},
	}
}

// ParseRaydiumLaunchlabFromDiscriminator 与 TS `parseRaydiumLaunchlabFromDiscriminator` 对齐
func ParseRaydiumLaunchlabFromDiscriminator(disc uint64, data []byte, meta EventMetadata) DexEvent {
	switch disc {
	case discRaydiumLaunchlabTrade:
		return parseRaydiumLaunchlabTradeFromData(data, meta)
	case discRaydiumLaunchlabPoolCreate:
		return parseRaydiumLaunchlabPoolCreateFromData(data, meta)
	default:
		return DexEvent{}
	}
}
