package solparser

func parsePSBuyFromData(data []byte, meta EventMetadata) DexEvent {
	const min = 14*8 + 7*32 + 1 + 5*8 + 4
	if len(data) < min {
		return nil
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ts := ri()
	ev := map[string]any{
		"metadata": meta, "timestamp": ts,
		"base_amount_out": rd(), "max_quote_amount_in": rd(),
		"user_base_token_reserves": rd(), "user_quote_token_reserves": rd(),
		"pool_base_token_reserves": rd(), "pool_quote_token_reserves": rd(),
		"quote_amount_in": rd(), "lp_fee_basis_points": rd(), "lp_fee": rd(),
		"protocol_fee_basis_points": rd(), "protocol_fee": rd(),
		"quote_amount_in_with_lp_fee": rd(), "user_quote_amount_in": rd(),
		"pool": rp(), "user": rp(), "user_base_token_account": rp(),
		"user_quote_token_account": rp(), "protocol_fee_recipient": rp(),
		"protocol_fee_recipient_token_account": rp(), "coin_creator": rp(),
		"coin_creator_fee_basis_points": rd(), "coin_creator_fee": rd(),
	}
	tv, _ := readBool(data, o)
	o++
	ev["track_volume"] = tv
	ev["total_unclaimed_tokens"] = rd()
	ev["total_claimed_tokens"] = rd()
	ev["current_sol_volume"] = rd()
	ev["last_update_timestamp"] = ri()
	ev["min_base_amount_out"] = rd()
	ix := ""
	if o+4 <= len(data) {
		l, _ := readU32LE(data, o)
		o += 4
		if o+int(l) <= len(data) {
			ix = string(data[o : o+int(l)])
		}
	}
	ev["ix_name"] = ix
	return DexEvent{"PumpSwapBuy": ev}
}

func parsePSSellFromData(data []byte, meta EventMetadata) DexEvent {
	const req = 13*8 + 7*32
	if len(data) < req {
		return nil
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ev := map[string]any{
		"metadata": meta, "timestamp": ri(),
		"base_amount_in": rd(), "min_quote_amount_out": rd(),
		"user_base_token_reserves": rd(), "user_quote_token_reserves": rd(),
		"pool_base_token_reserves": rd(), "pool_quote_token_reserves": rd(),
		"quote_amount_out": rd(), "lp_fee_basis_points": rd(), "lp_fee": rd(),
		"protocol_fee_basis_points": rd(), "protocol_fee": rd(),
		"quote_amount_out_without_lp_fee": rd(), "user_quote_amount_out": rd(),
		"pool": rp(), "user": rp(), "user_base_token_account": rp(),
		"user_quote_token_account": rp(), "protocol_fee_recipient": rp(),
		"protocol_fee_recipient_token_account": rp(), "coin_creator": rp(),
		"coin_creator_fee_basis_points": rd(), "coin_creator_fee": rd(),
	}
	cashBps, cash := uint64(0), uint64(0)
	if len(data) >= 368 {
		cashBps, _ = readU64LE(data, 352)
		cash, _ = readU64LE(data, 360)
	}
	ev["cashback_fee_basis_points"] = cashBps
	ev["cashback"] = cash
	return DexEvent{"PumpSwapSell": ev}
}

func parsePSCreatePoolFromData(data []byte, meta EventMetadata) DexEvent {
	const req = 8 + 2 + 32*6 + 2 + 8*7 + 1
	if len(data) < req {
		return nil
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
	ev := map[string]any{
		"metadata": meta, "timestamp": ts, "index": idx, "creator": creator,
		"base_mint": bm, "quote_mint": qm, "base_mint_decimals": bd, "quote_mint_decimals": qd,
		"base_amount_in": rd(), "quote_amount_in": rd(), "pool_base_amount": rd(),
		"pool_quote_amount": rd(), "minimum_liquidity": rd(), "initial_liquidity": rd(),
		"lp_token_amount_out": rd(),
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
	ev["pool_bump"] = pb
	ev["pool"] = pool
	ev["lp_mint"] = lp
	ev["user_base_token_account"] = uba
	ev["user_quote_token_account"] = uqa
	ev["coin_creator"] = cc
	if len(data) > 325 {
		mayhemVal, _ := readBool(data, 325)
		ev["is_mayhem_mode"] = mayhemVal
	} else {
		ev["is_mayhem_mode"] = false
	}
	return DexEvent{"PumpSwapCreatePool": ev}
}

func parsePSAddLiqFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 10*8+5*32 {
		return nil
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ev := map[string]any{
		"metadata": meta, "timestamp": ri(), "lp_token_amount_out": rd(),
		"max_base_amount_in": rd(), "max_quote_amount_in": rd(),
		"user_base_token_reserves": rd(), "user_quote_token_reserves": rd(),
		"pool_base_token_reserves": rd(), "pool_quote_token_reserves": rd(),
		"base_amount_in": rd(), "quote_amount_in": rd(), "lp_mint_supply": rd(),
		"pool": rp(), "user": rp(), "user_base_token_account": rp(),
		"user_quote_token_account": rp(), "user_pool_token_account": rp(),
	}
	return DexEvent{"PumpSwapLiquidityAdded": ev}
}

func parsePSRemoveLiqFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 10*8+5*32 {
		return nil
	}
	o := 0
	rd := func() uint64 { v, _ := readU64LE(data, o); o += 8; return v }
	ri := func() int64 { v, _ := readI64LE(data, o); o += 8; return v }
	rp := func() string { s, _ := readPubkey(data, o); o += 32; return s }
	ev := map[string]any{
		"metadata": meta, "timestamp": ri(), "lp_token_amount_in": rd(),
		"min_base_amount_out": rd(), "min_quote_amount_out": rd(),
		"user_base_token_reserves": rd(), "user_quote_token_reserves": rd(),
		"pool_base_token_reserves": rd(), "pool_quote_token_reserves": rd(),
		"base_amount_out": rd(), "quote_amount_out": rd(), "lp_mint_supply": rd(),
		"pool": rp(), "user": rp(), "user_base_token_account": rp(),
		"user_quote_token_account": rp(), "user_pool_token_account": rp(),
	}
	return DexEvent{"PumpSwapLiquidityRemoved": ev}
}
