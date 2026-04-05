package solparser

// Raydium CLMM
func parseClmmSwapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+16+1 {
		return nil
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	o += 8
	o += 8
	sqrt, _ := readU128LE(data, o)
	o += 16
	zfo, _ := readBool(data, o)
	ev := map[string]any{
		"metadata": meta, "pool_state": ps, "sender": user,
		"token_account_0": zeroPubkey, "token_account_1": zeroPubkey,
		"amount_0": uint64(0), "amount_1": uint64(0), "zero_for_one": zfo,
		"sqrt_price_x64": u128LEDecimalString(sqrt), "liquidity": "0",
		"transfer_fee_0": uint64(0), "transfer_fee_1": uint64(0), "tick": int32(0),
	}
	return DexEvent{"RaydiumClmmSwap": ev}
}

func parseClmmIncFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+16+8+8 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU128LE(data, o)
	o += 16
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "pool": pool, "position_nft_mint": zeroPubkey,
		"user": user, "liquidity": u128LEDecimalString(liq), "amount0_max": a0, "amount1_max": a1,
	}
	return DexEvent{"RaydiumClmmIncreaseLiquidity": ev}
}

func parseClmmDecFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+16+8+8 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	liq, _ := readU128LE(data, o)
	o += 16
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "pool": pool, "position_nft_mint": zeroPubkey,
		"user": user, "liquidity": u128LEDecimalString(liq), "amount0_min": a0, "amount1_min": a1,
	}
	return DexEvent{"RaydiumClmmDecreaseLiquidity": ev}
}

func parseClmmCreateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+16+8 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	cr, _ := readPubkey(data, o)
	o += 32
	sqrt, _ := readU128LE(data, o)
	o += 16
	ot, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "pool": pool, "creator": cr,
		"token_0_mint": zeroPubkey, "token_1_mint": zeroPubkey,
		"tick_spacing": 0, "fee_rate": 0, "sqrt_price_x64": u128LEDecimalString(sqrt), "open_time": ot,
	}
	return DexEvent{"RaydiumClmmCreatePool": ev}
}

func parseClmmCollectFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8 {
		return nil
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	pn, _ := readPubkey(data, o)
	o += 32
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	ev := map[string]any{"metadata": meta, "pool_state": ps, "position_nft_mint": pn, "amount_0": a0, "amount_1": a1}
	return DexEvent{"RaydiumClmmCollectFee": ev}
}

// Raydium AMM
func parseAmmSwapInFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8 {
		return nil
	}
	o := 0
	amm, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	ai, _ := readU64LE(data, o)
	o += 8
	mo, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "amm": amm, "user_source_owner": user,
		"amount_in": ai, "minimum_amount_out": mo, "max_amount_in": uint64(0), "amount_out": uint64(0),
		"token_program": zeroPubkey, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
		"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
		"serum_program": zeroPubkey, "serum_market": zeroPubkey, "serum_bids": zeroPubkey,
		"serum_asks": zeroPubkey, "serum_event_queue": zeroPubkey,
		"serum_coin_vault_account": zeroPubkey, "serum_pc_vault_account": zeroPubkey,
		"serum_vault_signer": zeroPubkey, "user_source_token_account": zeroPubkey,
		"user_destination_token_account": zeroPubkey,
	}
	return DexEvent{"RaydiumAmmV4Swap": ev}
}

func parseAmmSwapOutFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8 {
		return nil
	}
	o := 0
	amm, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	mai, _ := readU64LE(data, o)
	o += 8
	ao, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "amm": amm, "user_source_owner": user,
		"amount_in": uint64(0), "minimum_amount_out": uint64(0), "max_amount_in": mai, "amount_out": ao,
		"token_program": zeroPubkey, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
		"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
		"serum_program": zeroPubkey, "serum_market": zeroPubkey, "serum_bids": zeroPubkey,
		"serum_asks": zeroPubkey, "serum_event_queue": zeroPubkey,
		"serum_coin_vault_account": zeroPubkey, "serum_pc_vault_account": zeroPubkey,
		"serum_vault_signer": zeroPubkey, "user_source_token_account": zeroPubkey,
		"user_destination_token_account": zeroPubkey,
	}
	return DexEvent{"RaydiumAmmV4Swap": ev}
}

func parseAmmDepositFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8 {
		return nil
	}
	o := 0
	amm, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	mc, _ := readU64LE(data, o)
	o += 8
	mp, _ := readU64LE(data, o)
	o += 8
	bs, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "amm": amm, "user_owner": user,
		"max_coin_amount": mc, "max_pc_amount": mp, "base_side": bs,
		"token_program": zeroPubkey, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
		"amm_target_orders": zeroPubkey, "lp_mint_address": zeroPubkey,
		"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
		"serum_market": zeroPubkey, "user_coin_token_account": zeroPubkey,
		"user_pc_token_account": zeroPubkey, "user_lp_token_account": zeroPubkey,
		"serum_event_queue": zeroPubkey,
	}
	return DexEvent{"RaydiumAmmV4Deposit": ev}
}

func parseAmmWithdrawFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8 {
		return nil
	}
	o := 0
	amm, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	amt, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "amm": amm, "user_owner": user, "amount": amt,
		"token_program": zeroPubkey, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
		"amm_target_orders": zeroPubkey, "lp_mint_address": zeroPubkey,
		"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
		"pool_withdraw_queue": zeroPubkey, "pool_temp_lp_token_account": zeroPubkey,
		"serum_program": zeroPubkey, "serum_market": zeroPubkey,
		"serum_coin_vault_account": zeroPubkey, "serum_pc_vault_account": zeroPubkey,
		"serum_vault_signer": zeroPubkey, "user_lp_token_account": zeroPubkey,
		"user_coin_token_account": zeroPubkey, "user_pc_token_account": zeroPubkey,
		"serum_event_queue": zeroPubkey, "serum_bids": zeroPubkey, "serum_asks": zeroPubkey,
	}
	return DexEvent{"RaydiumAmmV4Withdraw": ev}
}

func parseAmmWithdrawPnlFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32 {
		return nil
	}
	o := 0
	amm, _ := readPubkey(data, o)
	o += 32
	pnlOwner, _ := readPubkey(data, o)
	ev := map[string]any{
		"metadata": meta, "token_program": zeroPubkey, "amm": amm, "amm_config": zeroPubkey,
		"amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
		"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
		"coin_pnl_token_account": zeroPubkey, "pc_pnl_token_account": zeroPubkey,
		"pnl_owner": pnlOwner, "amm_target_orders": zeroPubkey,
		"serum_program": zeroPubkey, "serum_market": zeroPubkey, "serum_event_queue": zeroPubkey,
		"serum_coin_vault_account": zeroPubkey, "serum_pc_vault_account": zeroPubkey,
		"serum_vault_signer": zeroPubkey,
	}
	return DexEvent{"RaydiumAmmV4WithdrawPnl": ev}
}

func parseAmmInit2FromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+8+8+8 {
		return nil
	}
	o := 0
	amm, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	nonce, _ := readU8(data, o)
	o++
	ot, _ := readU64LE(data, o)
	o += 8
	ipc, _ := readU64LE(data, o)
	o += 8
	ic, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "nonce": nonce, "open_time": ot,
		"init_pc_amount": ipc, "init_coin_amount": ic,
		"token_program": zeroPubkey, "spl_associated_token_account": zeroPubkey,
		"system_program": zeroPubkey, "rent": zeroPubkey,
		"amm": amm, "amm_authority": zeroPubkey, "amm_open_orders": zeroPubkey,
		"lp_mint": zeroPubkey, "coin_mint": zeroPubkey, "pc_mint": zeroPubkey,
		"pool_coin_token_account": zeroPubkey, "pool_pc_token_account": zeroPubkey,
		"pool_withdraw_queue": zeroPubkey, "amm_target_orders": zeroPubkey, "pool_temp_lp": zeroPubkey,
		"serum_program": zeroPubkey, "serum_market": zeroPubkey,
		"user_wallet": user, "user_token_coin": zeroPubkey, "user_token_pc": zeroPubkey,
		"user_lp_token_account": zeroPubkey,
	}
	return DexEvent{"RaydiumAmmV4Initialize2": ev}
}

// Raydium CPMM
func parseCpmmSwapInFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+1 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 64
	ai, _ := readU64LE(data, o)
	o += 16
	ao, _ := readU64LE(data, o)
	o += 8
	bi, _ := readBool(data, o)
	ev := map[string]any{
		"metadata": meta, "pool_id": pool, "input_amount": ai, "output_amount": ao,
		"input_vault_before": uint64(0), "output_vault_before": uint64(0),
		"input_transfer_fee": uint64(0), "output_transfer_fee": uint64(0), "base_input": bi,
	}
	return DexEvent{"RaydiumCpmmSwap": ev}
}

func parseCpmmSwapOutFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+1 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 64
	o += 8
	ao, _ := readU64LE(data, o)
	o += 8
	ai, _ := readU64LE(data, o)
	o += 8
	bo, _ := readBool(data, o)
	ev := map[string]any{
		"metadata": meta, "pool_id": pool, "input_amount": ai, "output_amount": ao,
		"base_input": !bo,
		"input_vault_before": uint64(0), "output_vault_before": uint64(0),
		"input_transfer_fee": uint64(0), "output_transfer_fee": uint64(0),
	}
	return DexEvent{"RaydiumCpmmSwap": ev}
}

func parseCpmmDepositFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	lp, _ := readU64LE(data, o)
	o += 8
	t0, _ := readU64LE(data, o)
	o += 8
	t1, _ := readU64LE(data, o)
	ev := map[string]any{"metadata": meta, "pool": pool, "user": user, "lp_token_amount": lp, "token0_amount": t0, "token1_amount": t1}
	return DexEvent{"RaydiumCpmmDeposit": ev}
}

func parseCpmmWithdrawFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	user, _ := readPubkey(data, o)
	o += 32
	lp, _ := readU64LE(data, o)
	o += 8
	t0, _ := readU64LE(data, o)
	o += 8
	t1, _ := readU64LE(data, o)
	ev := map[string]any{"metadata": meta, "pool": pool, "user": user, "lp_token_amount": lp, "token0_amount": t0, "token1_amount": t1}
	return DexEvent{"RaydiumCpmmWithdraw": ev}
}

func parseCpmmInitFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+32+8+8 {
		return nil
	}
	o := 0
	ps, _ := readPubkey(data, o)
	o += 32
	o += 32
	o += 32
	cr, _ := readPubkey(data, o)
	o += 32
	i0, _ := readU64LE(data, o)
	o += 8
	i1, _ := readU64LE(data, o)
	ev := map[string]any{"metadata": meta, "pool": ps, "creator": cr, "init_amount0": i0, "init_amount1": i1}
	return DexEvent{"RaydiumCpmmInitialize": ev}
}

// Orca
func parseOrcaTradedFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+1+16+16+8*6 {
		return nil
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	atb, _ := readBool(data, o)
	o++
	pre, _ := readU128LE(data, o)
	o += 16
	post, _ := readU128LE(data, o)
	o += 16
	ia, _ := readU64LE(data, o)
	o += 8
	oa, _ := readU64LE(data, o)
	o += 8
	itf, _ := readU64LE(data, o)
	o += 8
	otf, _ := readU64LE(data, o)
	o += 8
	lpf, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "whirlpool": w, "a_to_b": atb,
		"pre_sqrt_price": u128LEDecimalString(pre), "post_sqrt_price": u128LEDecimalString(post),
		"input_amount": ia, "output_amount": oa, "input_transfer_fee": itf, "output_transfer_fee": otf,
		"lp_fee": lpf, "protocol_fee": pf,
	}
	return DexEvent{"OrcaWhirlpoolSwap": ev}
}

func parseOrcaLiqIncFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+4+16+8*4 {
		return nil
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	p, _ := readPubkey(data, o)
	o += 32
	tl, _ := readI32LE(data, o)
	o += 4
	tu, _ := readI32LE(data, o)
	o += 4
	liq, _ := readU128LE(data, o)
	o += 16
	ta, _ := readU64LE(data, o)
	o += 8
	tb, _ := readU64LE(data, o)
	o += 8
	taf, _ := readU64LE(data, o)
	o += 8
	tbf, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "whirlpool": w, "position": p,
		"tick_lower_index": tl, "tick_upper_index": tu, "liquidity": u128LEDecimalString(liq),
		"token_a_amount": ta, "token_b_amount": tb, "token_a_transfer_fee": taf, "token_b_transfer_fee": tbf,
	}
	return DexEvent{"OrcaWhirlpoolLiquidityIncreased": ev}
}

func parseOrcaLiqDecFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+4+16+8*4 {
		return nil
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	p, _ := readPubkey(data, o)
	o += 32
	tl, _ := readI32LE(data, o)
	o += 4
	tu, _ := readI32LE(data, o)
	o += 4
	liq, _ := readU128LE(data, o)
	o += 16
	ta, _ := readU64LE(data, o)
	o += 8
	tb, _ := readU64LE(data, o)
	o += 8
	taf, _ := readU64LE(data, o)
	o += 8
	tbf, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "whirlpool": w, "position": p,
		"tick_lower_index": tl, "tick_upper_index": tu, "liquidity": u128LEDecimalString(liq),
		"token_a_amount": ta, "token_b_amount": tb, "token_a_transfer_fee": taf, "token_b_transfer_fee": tbf,
	}
	return DexEvent{"OrcaWhirlpoolLiquidityDecreased": ev}
}

func parseOrcaPoolInitFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*5+2+1+1+16 {
		return nil
	}
	o := 0
	w, _ := readPubkey(data, o)
	o += 32
	cfg, _ := readPubkey(data, o)
	o += 32
	ma, _ := readPubkey(data, o)
	o += 32
	mb, _ := readPubkey(data, o)
	o += 32
	ts, _ := readU16LE(data, o)
	o += 2
	tpa, _ := readPubkey(data, o)
	o += 32
	tpb, _ := readPubkey(data, o)
	o += 32
	da, _ := readU8(data, o)
	o++
	db, _ := readU8(data, o)
	o++
	isp, _ := readU128LE(data, o)
	ev := map[string]any{
		"metadata": meta, "whirlpool": w, "whirlpools_config": cfg,
		"token_mint_a": ma, "token_mint_b": mb, "tick_spacing": ts,
		"token_program_a": tpa, "token_program_b": tpb,
		"decimals_a": da, "decimals_b": db, "initial_sqrt_price": u128LEDecimalString(isp),
	}
	return DexEvent{"OrcaWhirlpoolPoolInitialized": ev}
}

// Meteora Pools
func parseMeteoraSwapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 8*5 {
		return nil
	}
	o := 0
	ev := map[string]any{
		"metadata": meta,
		"in_amount": readU64At(data, &o), "out_amount": readU64At(data, &o),
		"trade_fee": readU64At(data, &o), "admin_fee": readU64At(data, &o), "host_fee": readU64At(data, &o),
	}
	return DexEvent{"MeteoraPoolsSwap": ev}
}

func readU64At(b []byte, o *int) uint64 {
	v, _ := readU64LE(b, *o)
	*o += 8
	return v
}

func parseMeteoraAddFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 24 {
		return nil
	}
	o := 0
	ev := map[string]any{"metadata": meta, "lp_mint_amount": readU64At(data, &o), "token_a_amount": readU64At(data, &o), "token_b_amount": readU64At(data, &o)}
	return DexEvent{"MeteoraPoolsAddLiquidity": ev}
}

func parseMeteoraRemoveFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 24 {
		return nil
	}
	o := 0
	ev := map[string]any{"metadata": meta, "lp_unmint_amount": readU64At(data, &o), "token_a_out_amount": readU64At(data, &o), "token_b_out_amount": readU64At(data, &o)}
	return DexEvent{"MeteoraPoolsRemoveLiquidity": ev}
}

func parseMeteoraBootstrapFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 24+32 {
		return nil
	}
	o := 0
	lp := readU64At(data, &o)
	ta := readU64At(data, &o)
	tb := readU64At(data, &o)
	pl, _ := readPubkey(data, o)
	ev := map[string]any{"metadata": meta, "lp_mint_amount": lp, "token_a_amount": ta, "token_b_amount": tb, "pool": pl}
	return DexEvent{"MeteoraPoolsBootstrapLiquidity": ev}
}

func parseMeteoraPoolCreatedFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4+1 {
		return nil
	}
	o := 0
	lm, _ := readPubkey(data, o)
	o += 32
	ta, _ := readPubkey(data, o)
	o += 32
	tb, _ := readPubkey(data, o)
	o += 32
	pt, _ := readU8(data, o)
	o++
	pl, _ := readPubkey(data, o)
	ev := map[string]any{"metadata": meta, "lp_mint": lm, "token_a_mint": ta, "token_b_mint": tb, "pool_type": pt, "pool": pl}
	return DexEvent{"MeteoraPoolsPoolCreated": ev}
}
