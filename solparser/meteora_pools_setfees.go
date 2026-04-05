package solparser

func parseMeteoraPoolsSetPoolFeesFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8+8+8+8+32 {
		return nil
	}
	o := 0
	tfn, _ := readU64LE(data, o)
	o += 8
	tfd, _ := readU64LE(data, o)
	o += 8
	ofn, _ := readU64LE(data, o)
	o += 8
	ofd, _ := readU64LE(data, o)
	o += 8
	pool, ok := readPubkey(data, o)
	if !ok {
		return nil
	}
	return DexEvent{"MeteoraPoolsSetPoolFees": map[string]any{
		"metadata": meta, "trade_fee_numerator": tfn, "trade_fee_denominator": tfd,
		"owner_trade_fee_numerator": ofn, "owner_trade_fee_denominator": ofd, "pool": pool,
	}}
}
