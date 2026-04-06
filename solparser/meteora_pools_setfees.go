package solparser

func parseMeteoraPoolsSetPoolFeesFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8+8+8+8+32 {
		return DexEvent{}
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
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeMeteoraPoolsSetPoolFees,
		Data: &MeteoraPoolsSetPoolFeesEvent{
			Metadata:                 meta,
			TradeFeeNumerator:        tfn,
			TradeFeeDenominator:      tfd,
			OwnerTradeFeeNumerator:   ofn,
			OwnerTradeFeeDenominator: ofd,
			Pool:                     pool,
		},
	}
}
