package solparser

// Bonk discriminators 已在 binary.go 中定义

func parseBonkTradeFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+1+1 {
		return nil
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
	return DexEvent{"BonkTrade": map[string]any{
		"metadata": meta, "pool_state": pool, "user": user,
		"amount_in": ai, "amount_out": ao, "is_buy": isBuy,
		"trade_direction": dir, "exact_in": exIn,
	}}
}

func parseBonkPoolCreateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+32+8+8 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32 + 32 + 32
	creator, _ := readPubkey(data, o)
	return DexEvent{"BonkPoolCreate": map[string]any{
		"metadata": meta,
		"base_mint_param": map[string]any{
			"symbol": "BONK", "name": "Bonk Pool", "uri": "https://bonk.com", "decimals": 5,
		},
		"pool_state": pool, "creator": creator,
	}}
}

func parseBonkMigrateAmmFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8 {
		return nil
	}
	o := 0
	oldP, _ := readPubkey(data, o)
	o += 32
	newP, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU64LE(data, o)
	return DexEvent{"BonkMigrateAmm": map[string]any{
		"metadata": meta, "old_pool": oldP, "new_pool": newP, "user": user, "liquidity_amount": liq,
	}}
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
		return nil
	}
}
