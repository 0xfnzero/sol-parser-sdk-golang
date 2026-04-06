package solparser

// Bonk discriminators 已在 binary.go 中定义

func parseBonkTradeFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+1+1 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	ai, _ := readU64LE(data, o)
	o += 8
	ao, _ := readU64LE(data, o)
	o += 8
	isBuy, _ := readBool(data, o)
	o++
	exIn, _ := readBool(data, o)
	dir := "Sell"
	if isBuy {
		dir = "Buy"
	}
	return DexEvent{
		Type: EventTypeBonkTrade,
		Data: &BonkTradeEvent{
			Metadata:       meta,
			PoolState:      pool,
			User:           user,
			AmountIn:       ai,
			AmountOut:      ao,
			IsBuy:          isBuy,
			TradeDirection: dir,
			ExactIn:        exIn,
		},
	}
}

func parseBonkPoolCreateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+32+8+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32 + 32 + 32
	creator, _ := readPubkey(data, o)
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
			PoolState: pool,
			Creator:   creator,
		},
	}
}

func parseBonkMigrateAmmFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8 {
		return DexEvent{}
	}
	o := 0
	oldP, _ := readPubkey(data, o)
	o += 32
	newP, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeBonkMigrateAmm,
		Data: &BonkMigrateAmmEvent{
			Metadata:        meta,
			OldPool:         oldP,
			NewPool:         newP,
			User:            user,
			LiquidityAmount: liq,
		},
	}
}

// ParseBonkFromDiscriminator 与 TS `parseBonkFromDiscriminator` 对齐
func ParseBonkFromDiscriminator(disc uint64, data []byte, meta EventMetadata) DexEvent {
	switch disc {
	case discBonkTrade:
		return parseBonkTradeFromData(data, meta)
	case discBonkPoolCreate:
		return parseBonkPoolCreateFromData(data, meta)
	case discBonkMigrateAmm:
		return parseBonkMigrateAmmFromData(data, meta)
	default:
		return DexEvent{}
	}
}
