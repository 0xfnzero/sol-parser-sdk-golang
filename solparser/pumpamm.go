package solparser

import "unicode/utf8"

type pumpSwapTradeTail struct {
	CashbackFeeBasisPoints uint64
	Cashback               uint64
	BuybackFeeBasisPoints  uint64
	BuybackFee             uint64
	VirtualQuoteReserves   string
	CanBoost               bool
	BaseSupply             uint64
}

func parsePumpSwapTradeTail(data []byte) (pumpSwapTradeTail, bool) {
	tail := pumpSwapTradeTail{VirtualQuoteReserves: "0"}
	if len(data) == 0 {
		return tail, true
	}
	if len(data) < 16 {
		return tail, false
	}

	tail.CashbackFeeBasisPoints, _ = readU64LE(data, 0)
	tail.Cashback, _ = readU64LE(data, 8)
	if len(data) == 16 {
		return tail, true
	}
	if len(data) < 32 {
		return tail, false
	}

	tail.BuybackFeeBasisPoints, _ = readU64LE(data, 16)
	tail.BuybackFee, _ = readU64LE(data, 24)
	if len(data) == 32 {
		return tail, true
	}
	if len(data) < 57 {
		return tail, false
	}

	raw, _ := readU128LE(data, 32)
	tail.VirtualQuoteReserves = i128LEDecimalString(raw[:])
	switch data[48] {
	case 0:
		tail.CanBoost = false
	case 1:
		tail.CanBoost = true
	default:
		return tail, false
	}
	tail.BaseSupply, _ = readU64LE(data, 49)
	return tail, true
}

func parsePSBuyFromData(data []byte, meta EventMetadata) DexEvent {
	const legacyLen = 16*8 + 7*32 + 1 + 4*8
	const minRequiredLen = legacyLen + 8 + 4
	if len(data) != legacyLen && len(data) < minRequiredLen {
		return DexEvent{}
	}
	if data[352] != 0 && data[352] != 1 {
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
	tv := data[o] == 1
	o++
	ev.TrackVolume = tv
	ev.TotalUnclaimedTokens = rd()
	ev.TotalClaimedTokens = rd()
	ev.CurrentSolVolume = rd()
	ev.LastUpdateTimestamp = ri()
	tail := pumpSwapTradeTail{VirtualQuoteReserves: "0"}
	if len(data) != legacyLen {
		ev.MinBaseAmountOut = rd()
		ix, next, ok := readBorshString(data, o)
		if !ok || !utf8.ValidString(ix) {
			return DexEvent{}
		}
		o = next
		tail, ok = parsePumpSwapTradeTail(data[o:])
		if !ok {
			return DexEvent{}
		}
		ev.IxName = ix
	}
	ev.CashbackFeeBasisPoints = tail.CashbackFeeBasisPoints
	ev.Cashback = tail.Cashback
	ev.BuybackFeeBasisPoints = tail.BuybackFeeBasisPoints
	ev.BuybackFee = tail.BuybackFee
	ev.VirtualQuoteReserves = tail.VirtualQuoteReserves
	ev.CanBoost = tail.CanBoost
	ev.BaseSupply = tail.BaseSupply
	ev.IsCashbackCoin = tail.CashbackFeeBasisPoints > 0
	return DexEvent{Type: EventTypePumpSwapBuy, Data: ev}
}

func parsePSSellFromData(data []byte, meta EventMetadata) DexEvent {
	const req = 14*8 + 7*32 + 2*8
	if len(data) < req {
		return DexEvent{}
	}
	tail, ok := parsePumpSwapTradeTail(data[req:])
	if !ok {
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
	ev.CashbackFeeBasisPoints = tail.CashbackFeeBasisPoints
	ev.Cashback = tail.Cashback
	ev.BuybackFeeBasisPoints = tail.BuybackFeeBasisPoints
	ev.BuybackFee = tail.BuybackFee
	ev.VirtualQuoteReserves = tail.VirtualQuoteReserves
	ev.CanBoost = tail.CanBoost
	ev.BaseSupply = tail.BaseSupply
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

// enrichPumpSwapBuyFromAccounts 将内层 CPI 指令账户写入事件（与 parsePumpSwapBuyInstr 下标一致）。
// parsePSBuyFromData 仅有 Program data 数值与部分 pubkey，mint/池子 ATA 等需从指令账户补全。
func enrichPumpSwapBuyFromAccounts(ev *PumpSwapBuyEvent, accounts []string) {
	if len(accounts) < 13 {
		return
	}
	set := func(dst *string, idx int) {
		if *dst != "" {
			return
		}
		s := getAccountSafe(accounts, idx)
		if s != "" && s != zeroPubkey {
			*dst = s
		}
	}
	set(&ev.BaseMint, 3)
	set(&ev.QuoteMint, 4)
	set(&ev.PoolBaseTokenAccount, 7)
	set(&ev.PoolQuoteTokenAccount, 8)
	set(&ev.BaseTokenProgram, 11)
	set(&ev.QuoteTokenProgram, 12)
	if len(accounts) >= 19 {
		set(&ev.CoinCreatorVaultAta, 17)
		set(&ev.CoinCreatorVaultAuthority, 18)
	}
	if len(accounts) >= 27 {
		set(&ev.PoolV2, 24)
		set(&ev.FeeRecipient, 25)
		set(&ev.FeeRecipientQuoteTokenAccount, 26)
	} else if len(accounts) >= 26 {
		set(&ev.PoolV2, 23)
		set(&ev.FeeRecipient, 24)
		set(&ev.FeeRecipientQuoteTokenAccount, 25)
	} else if len(accounts) >= 24 {
		set(&ev.PoolV2, 23)
	}
}

func enrichPumpSwapSellFromAccounts(ev *PumpSwapSellEvent, accounts []string) {
	if len(accounts) < 13 {
		return
	}
	set := func(dst *string, idx int) {
		if *dst != "" {
			return
		}
		s := getAccountSafe(accounts, idx)
		if s != "" && s != zeroPubkey {
			*dst = s
		}
	}
	set(&ev.BaseMint, 3)
	set(&ev.QuoteMint, 4)
	set(&ev.PoolBaseTokenAccount, 7)
	set(&ev.PoolQuoteTokenAccount, 8)
	set(&ev.BaseTokenProgram, 11)
	set(&ev.QuoteTokenProgram, 12)
	if len(accounts) >= 19 {
		set(&ev.CoinCreatorVaultAta, 17)
		set(&ev.CoinCreatorVaultAuthority, 18)
	}
	if len(accounts) >= 26 {
		set(&ev.PoolV2, 23)
		set(&ev.FeeRecipient, 24)
		set(&ev.FeeRecipientQuoteTokenAccount, 25)
	} else if len(accounts) >= 24 {
		set(&ev.PoolV2, 21)
		set(&ev.FeeRecipient, 22)
		set(&ev.FeeRecipientQuoteTokenAccount, 23)
	} else if len(accounts) >= 22 {
		set(&ev.PoolV2, 21)
	}
}
