package solparser

// PumpFun discriminators 已在 binary.go 中定义

func normalizePumpfunIxName(ixName string) string {
	switch ixName {
	case "buy_v2":
		return "buy"
	case "sell_v2":
		return "sell"
	case "buy_exact_quote_in_v2":
		return "buy_exact_quote_in"
	default:
		return ixName
	}
}

func parseTradeFromData(data []byte, meta EventMetadata, isCreatedBuy bool) DexEvent {
	if len(data) < 32+8+8+1+32+8*5+32+8+8+32+8+8 {
		return DexEvent{}
	}
	o := 0
	mint, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	solAmount, _ := readU64LE(data, o)
	o += 8
	tokenAmount, _ := readU64LE(data, o)
	o += 8
	isBuy, ok := readBool(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 1
	user, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	ts, _ := readI64LE(data, o)
	o += 8
	vsol, _ := readU64LE(data, o)
	o += 8
	vtok, _ := readU64LE(data, o)
	o += 8
	rsol, _ := readU64LE(data, o)
	o += 8
	rtok, _ := readU64LE(data, o)
	o += 8
	feeRec, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	feeBps, _ := readU64LE(data, o)
	o += 8
	fee, _ := readU64LE(data, o)
	o += 8
	creator, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	cfbps, _ := readU64LE(data, o)
	o += 8
	cfee, _ := readU64LE(data, o)
	o += 8
	tv := false
	if o < len(data) {
		tv, _ = readBool(data, o)
	}
	o++
	var tuc, tcc, csv uint64
	var lut int64
	if o+8 <= len(data) {
		tuc, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		tcc, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		csv, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		lut, _ = readI64LE(data, o)
		o += 8
	}
	ixName := ""
	if o+4 <= len(data) {
		if s, next, ok2 := readBorshString(data, o); ok2 {
			ixName = s
			o = next
		}
	}
	ixName = normalizePumpfunIxName(ixName)
	mm := false
	if o < len(data) {
		mm, _ = readBool(data, o)
	}
	o++
	var cbBps, cb uint64
	if o+8 <= len(data) {
		cbBps, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		cb, _ = readU64LE(data, o)
		o += 8
	}
	var buybackFeeBps, buybackFee, quoteAmount, virtualQuoteReserves, realQuoteReserves uint64
	if o+8 <= len(data) {
		buybackFeeBps, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		buybackFee, _ = readU64LE(data, o)
		o += 8
	}
	shareholders := []PumpFeesShareholder{}
	if o+4 <= len(data) {
		n32, _ := readU32LE(data, o)
		n := int(n32)
		if n > maxPumpFeesShareholders || o+4+n*34 > len(data) {
			return DexEvent{}
		}
		o += 4
		shareholders = make([]PumpFeesShareholder, 0, n)
		for i := 0; i < n; i++ {
			address, ok := readPubkey(data, o)
			if !ok {
				return DexEvent{}
			}
			o += 32
			shareBps, ok := readU16LE(data, o)
			if !ok {
				return DexEvent{}
			}
			o += 2
			shareholders = append(shareholders, PumpFeesShareholder{Address: address, ShareBps: shareBps})
		}
	}
	quoteMint := zeroPubkey
	if o+32 <= len(data) {
		quoteMint, _ = readPubkey(data, o)
		o += 32
	}
	if o+8 <= len(data) {
		quoteAmount, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		virtualQuoteReserves, _ = readU64LE(data, o)
		o += 8
	}
	if o+8 <= len(data) {
		realQuoteReserves, _ = readU64LE(data, o)
	}

	ev := &PumpFunTradeEvent{
		Metadata:               meta,
		Mint:                   mint,
		SolAmount:              solAmount,
		TokenAmount:            tokenAmount,
		IsBuy:                  isBuy,
		IsCreatedBuy:           isCreatedBuy,
		User:                   user,
		Timestamp:              ts,
		VirtualSolReserves:     vsol,
		VirtualTokenReserves:   vtok,
		RealSolReserves:        rsol,
		RealTokenReserves:      rtok,
		FeeRecipient:           feeRec,
		FeeBasisPoints:         feeBps,
		Fee:                    fee,
		Creator:                creator,
		CreatorFeeBasisPoints:  cfbps,
		CreatorFee:             cfee,
		TrackVolume:            tv,
		TotalUnclaimedTokens:   tuc,
		TotalClaimedTokens:     tcc,
		CurrentSolVolume:       csv,
		LastUpdateTimestamp:    lut,
		IxName:                 ixName,
		MayhemMode:             mm,
		CashbackFeeBasisPoints: cbBps,
		Cashback:               cb,
		BuybackFeeBasisPoints:  buybackFeeBps,
		BuybackFee:             buybackFee,
		Shareholders:           shareholders,
		QuoteMint:              quoteMint,
		QuoteAmount:            quoteAmount,
		VirtualQuoteReserves:   virtualQuoteReserves,
		RealQuoteReserves:      realQuoteReserves,
		IsCashbackCoin:         cbBps > 0,
		BondingCurve:           "",
		AssociatedBondingCurve: "",
		TokenProgram:           "",
		CreatorVault:           "",
	}

	switch ixName {
	case "buy":
		return DexEvent{Type: EventTypePumpFunBuy, Data: ev}
	case "sell":
		return DexEvent{Type: EventTypePumpFunSell, Data: ev}
	case "buy_exact_sol_in":
		return DexEvent{Type: EventTypePumpFunBuyExactSolIn, Data: ev}
	case "buy_exact_quote_in":
		return DexEvent{Type: EventTypePumpFunBuy, Data: ev}
	default:
		return DexEvent{Type: EventTypePumpFunTrade, Data: ev}
	}
}

func parseCreateFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	name, o, ok := readBorshString(data, o)
	if !ok {
		return DexEvent{}
	}
	sym, o, ok := readBorshString(data, o)
	if !ok {
		return DexEvent{}
	}
	uri, o, ok := readBorshString(data, o)
	if !ok {
		return DexEvent{}
	}
	if len(data) < o+32*4+8*5+32+1 {
		return DexEvent{}
	}
	mint, _ := readPubkey(data, o)
	o += 32
	bc, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	creator, _ := readPubkey(data, o)
	o += 32
	ts, _ := readI64LE(data, o)
	o += 8
	vtr, _ := readU64LE(data, o)
	o += 8
	vsol, _ := readU64LE(data, o)
	o += 8
	rtr, _ := readU64LE(data, o)
	o += 8
	tts, _ := readU64LE(data, o)
	o += 8
	tp := zeroPubkey
	if o+32 <= len(data) {
		tp, _ = readPubkey(data, o)
	}
	o += 32
	mm := false
	if o < len(data) {
		mm, _ = readBool(data, o)
	}
	o++
	ice := false
	if o < len(data) {
		ice, _ = readBool(data, o)
	}
	o++
	quoteMint := zeroPubkey
	if o+32 <= len(data) {
		quoteMint, _ = readPubkey(data, o)
	}
	o += 32
	var virtualQuoteReserves uint64
	if o+8 <= len(data) {
		virtualQuoteReserves, _ = readU64LE(data, o)
	}

	return DexEvent{
		Type: EventTypePumpFunCreate,
		Data: &PumpFunCreateEvent{
			Metadata:             meta,
			Name:                 name,
			Symbol:               sym,
			Uri:                  uri,
			Mint:                 mint,
			BondingCurve:         bc,
			User:                 user,
			Creator:              creator,
			Timestamp:            ts,
			VirtualTokenReserves: vtr,
			VirtualSolReserves:   vsol,
			RealTokenReserves:    rtr,
			TokenTotalSupply:     tts,
			TokenProgram:         tp,
			IsMayhemMode:         mm,
			IsCashbackEnabled:    ice,
			QuoteMint:            quoteMint,
			VirtualQuoteReserves: virtualQuoteReserves,
		},
	}
}

func parseMigrateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+32+8+32 {
		return DexEvent{}
	}
	o := 0
	user, _ := readPubkey(data, o)
	o += 32
	mint, _ := readPubkey(data, o)
	o += 32
	ma, _ := readU64LE(data, o)
	o += 8
	sa, _ := readU64LE(data, o)
	o += 8
	pmf, _ := readU64LE(data, o)
	o += 8
	bc, _ := readPubkey(data, o)
	o += 32
	ts, _ := readI64LE(data, o)
	o += 8
	pool, _ := readPubkey(data, o)

	return DexEvent{
		Type: EventTypePumpFunMigrate,
		Data: &PumpFunMigrateEvent{
			Metadata:         meta,
			User:             user,
			Mint:             mint,
			MintAmount:       ma,
			SolAmount:        sa,
			PoolMigrationFee: pmf,
			BondingCurve:     bc,
			Timestamp:        ts,
			Pool:             pool,
		},
	}
}

func parseMigrateBondingCurveCreatorFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8+32*5 {
		return DexEvent{}
	}
	o := 0
	ts, _ := readI64LE(data, o)
	o += 8
	mint, _ := readPubkey(data, o)
	o += 32
	bondingCurve, _ := readPubkey(data, o)
	o += 32
	sharingConfig, _ := readPubkey(data, o)
	o += 32
	oldCreator, _ := readPubkey(data, o)
	o += 32
	newCreator, _ := readPubkey(data, o)

	return DexEvent{
		Type: EventTypePumpFunMigrateBondingCurveCreator,
		Data: &PumpFunMigrateBondingCurveCreatorEvent{
			Metadata:      meta,
			Timestamp:     ts,
			Mint:          mint,
			BondingCurve:  bondingCurve,
			SharingConfig: sharingConfig,
			OldCreator:    oldCreator,
			NewCreator:    newCreator,
		},
	}
}

// enrichPumpFunTradeFromAccounts 按 pump.json IDL 从 **内层 CPI 指令账户** 补全（与 Rust `instr/pump.rs` buy/sell 一致）。
// Program data 日志不含 bonding_curve / token_program 等，仅靠 parseTradeFromData 会得到空字段。
func enrichPumpFunTradeFromAccounts(ev *PumpFunTradeEvent, accounts []string) {
	if ev == nil || len(accounts) < 7 {
		return
	}
	set := func(dst *string, idx int) {
		if *dst != "" && *dst != zeroPubkey {
			return
		}
		s := getAccountSafe(accounts, idx)
		if s != "" && s != zeroPubkey {
			*dst = s
		}
	}
	isV2 := ev.IxName == "buy_v2" ||
		ev.IxName == "sell_v2" ||
		ev.IxName == "buy_exact_quote_in_v2" ||
		(ev.Mint != "" && ev.Mint != zeroPubkey && getAccountSafe(accounts, 1) == ev.Mint)
	if isV2 {
		set(&ev.Global, 0)
		set(&ev.QuoteMint, 2)
		set(&ev.FeeRecipient, 6)
		set(&ev.BondingCurve, 10)
		set(&ev.AssociatedBondingCurve, 11)
		set(&ev.User, 13)
		set(&ev.TokenProgram, 3)
		set(&ev.QuoteTokenProgram, 4)
		set(&ev.AssociatedTokenProgram, 5)
		set(&ev.CreatorVault, 16)
		set(&ev.AssociatedQuoteFeeRecipient, 7)
		set(&ev.BuybackFeeRecipient, 8)
		set(&ev.AssociatedQuoteBuybackFeeRecipient, 9)
		set(&ev.AssociatedQuoteBondingCurve, 12)
		set(&ev.AssociatedUser, 14)
		set(&ev.AssociatedQuoteUser, 15)
		set(&ev.AssociatedCreatorVault, 17)
		set(&ev.SharingConfig, 18)
		if ev.IxName == "sell_v2" || (ev.IxName == "sell" && !ev.IsBuy) {
			set(&ev.UserVolumeAccumulator, 19)
			set(&ev.AssociatedUserVolumeAccumulator, 20)
			set(&ev.FeeConfig, 21)
			set(&ev.FeeProgram, 22)
			set(&ev.SystemProgram, 23)
			set(&ev.EventAuthority, 24)
			set(&ev.Program, 25)
		} else {
			set(&ev.GlobalVolumeAccumulator, 19)
			set(&ev.UserVolumeAccumulator, 20)
			set(&ev.AssociatedUserVolumeAccumulator, 21)
			set(&ev.FeeConfig, 22)
			set(&ev.FeeProgram, 23)
			set(&ev.SystemProgram, 24)
			set(&ev.EventAuthority, 25)
			set(&ev.Program, 26)
		}
		return
	}
	set(&ev.Global, 0)
	set(&ev.FeeRecipient, 1)
	set(&ev.BondingCurve, 3)
	set(&ev.AssociatedBondingCurve, 4)
	set(&ev.AssociatedUser, 5)
	set(&ev.User, 6)
	set(&ev.SystemProgram, 7)
	if ev.IsBuy {
		set(&ev.TokenProgram, 8)
		set(&ev.CreatorVault, 9)
		set(&ev.EventAuthority, 10)
		set(&ev.Program, 11)
		set(&ev.GlobalVolumeAccumulator, 12)
		set(&ev.UserVolumeAccumulator, 13)
		set(&ev.FeeConfig, 14)
		set(&ev.FeeProgram, 15)
		set(&ev.BondingCurveV2, 16)
		set(&ev.BuybackFeeRecipient, 17)
		set(&ev.Account, 17)
	} else {
		set(&ev.CreatorVault, 8)
		set(&ev.TokenProgram, 9)
		set(&ev.EventAuthority, 10)
		set(&ev.Program, 11)
		set(&ev.FeeConfig, 12)
		set(&ev.FeeProgram, 13)
		if len(accounts) >= 17 {
			set(&ev.UserVolumeAccumulator, 14)
			set(&ev.BondingCurveV2, 15)
			set(&ev.BuybackFeeRecipient, 16)
			set(&ev.Account, 16)
		} else if ev.IsCashbackCoin {
			set(&ev.UserVolumeAccumulator, 14)
			set(&ev.BondingCurveV2, 15)
		} else {
			set(&ev.BondingCurveV2, 14)
			set(&ev.BuybackFeeRecipient, 15)
			set(&ev.Account, 15)
		}
	}
}
