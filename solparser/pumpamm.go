package solparser

func parsePSBuyFromData(data []byte, meta EventMetadata) DexEvent {
	const min = 14*8 + 7*32 + 1 + 5*8 + 4
	if len(data) < min {
		return DexEvent{}
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ts := ri()
	ev := &PumpSwapBuyEvent{
		Metadata:                         meta,
		Timestamp:                        ts,
		BaseAmountOut:                    rd(),
		MaxQuoteAmountIn:                 rd(),
		UserBaseTokenReserves:            rd(),
		UserQuoteTokenReserves:           rd(),
		PoolBaseTokenReserves:            rd(),
		PoolQuoteTokenReserves:           rd(),
		QuoteAmountIn:                    rd(),
		LpFeeBasisPoints:                 rd(),
		LpFee:                            rd(),
		ProtocolFeeBasisPoints:           rd(),
		ProtocolFee:                      rd(),
		QuoteAmountInWithLpFee:           rd(),
		UserQuoteAmountIn:                rd(),
		Pool:                             rp(),
		User:                             rp(),
		UserBaseTokenAccount:             rp(),
		UserQuoteTokenAccount:            rp(),
		ProtocolFeeRecipient:             rp(),
		ProtocolFeeRecipientTokenAccount: rp(),
		CoinCreator:                      rp(),
		CoinCreatorFeeBasisPoints:        rd(),
		CoinCreatorFee:                   rd(),
	}
	tv, _ := readBool(data, o)
	o++
	ev.TrackVolume = tv
	ev.TotalUnclaimedTokens = rd()
	ev.TotalClaimedTokens = rd()
	ev.CurrentSolVolume = rd()
	ev.LastUpdateTimestamp = ri()
	ev.MinBaseAmountOut = rd()
	ix := ""
	if o+4 <= len(data) {
		l, _ := readU32LE(data, o)
		o += 4
		if o+int(l) <= len(data) {
			ix = string(data[o : o+int(l)])
			o += int(l)
		}
	}
	ev.IxName = ix
	mm := false
	if o < len(data) {
		mm, _ = readBool(data, o)
		o++
	}
	cbBps := uint64(0)
	cb := uint64(0)
	if o+16 <= len(data) {
		cbBps, _ = readU64LE(data, o)
		o += 8
		cb, _ = readU64LE(data, o)
	}
	ev.MayhemMode = mm
	ev.CashbackFeeBasisPoints = cbBps
	ev.Cashback = cb
	ev.IsCashbackCoin = cbBps > 0
	return DexEvent{Type: EventTypePumpSwapBuy, Data: ev}
}

func parsePSSellFromData(data []byte, meta EventMetadata) DexEvent {
	const req = 13*8 + 7*32
	if len(data) < req {
		return DexEvent{}
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ev := &PumpSwapSellEvent{
		Metadata:                         meta,
		Timestamp:                        ri(),
		BaseAmountIn:                     rd(),
		MinQuoteAmountOut:                rd(),
		UserBaseTokenReserves:            rd(),
		UserQuoteTokenReserves:           rd(),
		PoolBaseTokenReserves:            rd(),
		PoolQuoteTokenReserves:           rd(),
		QuoteAmountOut:                   rd(),
		LpFeeBasisPoints:                 rd(),
		LpFee:                            rd(),
		ProtocolFeeBasisPoints:           rd(),
		ProtocolFee:                      rd(),
		QuoteAmountOutWithoutLpFee:       rd(),
		UserQuoteAmountOut:               rd(),
		Pool:                             rp(),
		User:                             rp(),
		UserBaseTokenAccount:             rp(),
		UserQuoteTokenAccount:            rp(),
		ProtocolFeeRecipient:             rp(),
		ProtocolFeeRecipientTokenAccount: rp(),
		CoinCreator:                      rp(),
		CoinCreatorFeeBasisPoints:        rd(),
		CoinCreatorFee:                   rd(),
	}
	cashBps, cash := uint64(0), uint64(0)
	if len(data) >= 368 {
		cashBps, _ = readU64LE(data, 352)
		cash, _ = readU64LE(data, 360)
	}
	ev.CashbackFeeBasisPoints = cashBps
	ev.Cashback = cash
	return DexEvent{Type: EventTypePumpSwapSell, Data: ev}
}

func parsePSCreatePoolFromData(data []byte, meta EventMetadata) DexEvent {
	const req = 8 + 2 + 32*6 + 2 + 8*7 + 1
	if len(data) < req {
		return DexEvent{}
	}
	o := 0
	ts, _ := readI64LE(data, o)
	o += 8
	idx, _ := readU16LE(data, o)
	o += 2
	creator, _ := readPubkey(data, o)
	o += 32
	bm, _ := readPubkey(data, o)
	o += 32
	qm, _ := readPubkey(data, o)
	o += 32
	bd, _ := readU8(data, o)
	o++
	qd, _ := readU8(data, o)
	o++
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ev := &PumpSwapCreatePoolEvent{
		Metadata:          meta,
		Timestamp:         ts,
		Index:             idx,
		Creator:           creator,
		BaseMint:          bm,
		QuoteMint:         qm,
		BaseMintDecimals:  bd,
		QuoteMintDecimals: qd,
		BaseAmountIn:      rd(),
		QuoteAmountIn:     rd(),
		PoolBaseAmount:    rd(),
		PoolQuoteAmount:   rd(),
		MinimumLiquidity:  rd(),
		InitialLiquidity:  rd(),
		LpTokenAmountOut:  rd(),
	}
	pb, _ := readU8(data, o)
	o++
	pool, _ := readPubkey(data, o)
	o += 32
	lp, _ := readPubkey(data, o)
	o += 32
	uba, _ := readPubkey(data, o)
	o += 32
	uqa, _ := readPubkey(data, o)
	o += 32
	cc, _ := readPubkey(data, o)
	ev.PoolBump = pb
	ev.Pool = pool
	ev.LpMint = lp
	ev.UserBaseTokenAccount = uba
	ev.UserQuoteTokenAccount = uqa
	ev.CoinCreator = cc
	if len(data) > 325 {
		mayhemVal, _ := readBool(data, 325)
		ev.IsMayhemMode = mayhemVal
	}
	return DexEvent{Type: EventTypePumpSwapCreatePool, Data: ev}
}

func parsePSAddLiqFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 10*8+5*32 {
		return DexEvent{}
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ev := &PumpSwapLiquidityAddedEvent{
		Metadata:               meta,
		Timestamp:              ri(),
		LpTokenAmountOut:       rd(),
		MaxBaseAmountIn:        rd(),
		MaxQuoteAmountIn:       rd(),
		UserBaseTokenReserves:  rd(),
		UserQuoteTokenReserves: rd(),
		PoolBaseTokenReserves:  rd(),
		PoolQuoteTokenReserves: rd(),
		BaseAmountIn:           rd(),
		QuoteAmountIn:          rd(),
		LpMintSupply:           rd(),
		Pool:                   rp(),
		User:                   rp(),
		UserBaseTokenAccount:   rp(),
		UserQuoteTokenAccount:  rp(),
		UserPoolTokenAccount:   rp(),
	}
	return DexEvent{Type: EventTypePumpSwapLiquidityAdded, Data: ev}
}

func parsePSRemoveLiqFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 10*8+5*32 {
		return DexEvent{}
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ev := &PumpSwapLiquidityRemovedEvent{
		Metadata:               meta,
		Timestamp:              ri(),
		LpTokenAmountIn:        rd(),
		MinBaseAmountOut:       rd(),
		MinQuoteAmountOut:      rd(),
		UserBaseTokenReserves:  rd(),
		UserQuoteTokenReserves: rd(),
		PoolBaseTokenReserves:  rd(),
		PoolQuoteTokenReserves: rd(),
		BaseAmountOut:          rd(),
		QuoteAmountOut:         rd(),
		LpMintSupply:           rd(),
		Pool:                   rp(),
		User:                   rp(),
		UserBaseTokenAccount:   rp(),
		UserQuoteTokenAccount:  rp(),
		UserPoolTokenAccount:   rp(),
	}
	return DexEvent{Type: EventTypePumpSwapLiquidityRemoved, Data: ev}
}
