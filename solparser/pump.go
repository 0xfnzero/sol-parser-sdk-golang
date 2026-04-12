package solparser

// PumpFun discriminators 已在 binary.go 中定义

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
	set(&ev.BondingCurve, 3)
	set(&ev.AssociatedBondingCurve, 4)
	if ev.IsBuy {
		set(&ev.TokenProgram, 8)
		set(&ev.CreatorVault, 9)
	} else {
		set(&ev.CreatorVault, 8)
		set(&ev.TokenProgram, 9)
	}
}
