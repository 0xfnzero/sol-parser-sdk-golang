package solparser

import (
	"encoding/binary"
	"encoding/hex"
)

// DAMM discriminators 已在 binary.go 中定义

// ParseMeteoraDammLog 与 TS `parseMeteoraDammLog` 对齐（Program data 载荷与 `meteora_damm_ix` CPI 内层一致）
func ParseMeteoraDammLog(log, sig string, slot, tx uint64, blockUs *int64, grpcUs int64) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return nil
	}
	d := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	meta := makeMetadata(sig, slot, tx, blockUs, grpcUs, "")
	switch d {
	case discDammSwap:
		return parseDammSwap(data, meta)
	case discDammSwap2:
		return parseDammSwap2(data, meta)
	case discDammCreatePosition:
		return parseDammCreatePosition(data, meta)
	case discDammClosePosition:
		return parseDammClosePosition(data, meta)
	case discDammAddLiquidity:
		return parseDammAddLiquidity(data, meta)
	case discDammRemoveLiq:
		return parseDammRemoveLiquidity(data, meta)
	case discDammInitPool:
		return parseDammInitializePool(data, meta)
	default:
		return nil
	}
}

func parseDammSwap(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+1+8*8+16+8*4 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	o += 32
	td, _ := readU8(data, o)
	o++
	hr, _ := readBool(data, o)
	o++
	ai, _ := readU64LE(data, o)
	o += 8
	mo, _ := readU64LE(data, o)
	o += 8
	aai, _ := readU64LE(data, o)
	o += 8
	oa, _ := readU64LE(data, o)
	o += 8
	nsp, _ := readU128LE(data, o)
	o += 16
	lpf, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	o += 8
	rf, _ := readU64LE(data, o)
	o += 8
	o += 8
	ct, _ := readU64LE(data, o)
	ev := map[string]any{
		"metadata": meta, "pool": pool, "trade_direction": td, "has_referral": hr,
		"amount_in": ai, "minimum_amount_out": mo, "output_amount": oa,
		"next_sqrt_price": u128LEDecimalString(nsp), "lp_fee": lpf, "protocol_fee": pf,
		"partner_fee": uint64(0), "referral_fee": rf, "actual_amount_in": aai, "current_timestamp": ct,
		"token_a_vault": zeroPubkey, "token_b_vault": zeroPubkey, "token_a_mint": zeroPubkey,
		"token_b_mint": zeroPubkey, "token_a_program": zeroPubkey, "token_b_program": zeroPubkey,
	}
	return DexEvent{"MeteoraDammV2Swap": ev}
}

func parseDammSwap2(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+1+1+1+8*2+1+8*6+16+8*4+8*3 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	td, _ := readU8(data, o)
	o++
	_, _ = readU8(data, o)
	o++
	hr, _ := readBool(data, o)
	o++
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	o += 8
	sm, _ := readU8(data, o)
	o++
	ifi, _ := readU64LE(data, o)
	o += 8
	o += 16
	oa, _ := readU64LE(data, o)
	o += 8
	nsp, _ := readU128LE(data, o)
	o += 16
	lpf, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	o += 8
	rf, _ := readU64LE(data, o)
	o += 8
	o += 8
	o += 8
	o += 8
	ct, _ := readU64LE(data, o)
	ai, mo := a0, a1
	if sm != 0 {
		ai, mo = a1, a0
	}
	ev := map[string]any{
		"metadata": meta, "pool": pool, "trade_direction": td, "has_referral": hr,
		"amount_in": ai, "minimum_amount_out": mo, "output_amount": oa,
		"next_sqrt_price": u128LEDecimalString(nsp), "lp_fee": lpf, "protocol_fee": pf,
		"partner_fee": uint64(0), "referral_fee": rf, "actual_amount_in": ifi, "current_timestamp": ct,
		"token_a_vault": zeroPubkey, "token_b_vault": zeroPubkey, "token_a_mint": zeroPubkey,
		"token_b_mint": zeroPubkey, "token_a_program": zeroPubkey, "token_b_program": zeroPubkey,
	}
	return DexEvent{"MeteoraDammV2Swap": ev}
}

func parseDammCreatePosition(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	nft, _ := readPubkey(data, o)
	return DexEvent{"MeteoraDammV2CreatePosition": map[string]any{
		"metadata": meta, "pool": pool, "owner": owner, "position": pos, "position_nft_mint": nft,
	}}
}

func parseDammClosePosition(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	nft, _ := readPubkey(data, o)
	return DexEvent{"MeteoraDammV2ClosePosition": map[string]any{
		"metadata": meta, "pool": pool, "owner": owner, "position": pos, "position_nft_mint": nft,
	}}
}

func parseDammAddLiquidity(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*3+16+8*6 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	ld, _ := readU128LE(data, o)
	o += 16
	tat, _ := readU64LE(data, o)
	o += 8
	tbt, _ := readU64LE(data, o)
	o += 8
	ta, _ := readU64LE(data, o)
	o += 8
	tb, _ := readU64LE(data, o)
	o += 8
	tota, _ := readU64LE(data, o)
	o += 8
	totb, _ := readU64LE(data, o)
	return DexEvent{"MeteoraDammV2AddLiquidity": map[string]any{
		"metadata": meta, "pool": pool, "position": pos, "owner": owner,
		"liquidity_delta": u128LEDecimalString(ld), "token_a_amount_threshold": tat, "token_b_amount_threshold": tbt,
		"token_a_amount": ta, "token_b_amount": tb, "total_amount_a": tota, "total_amount_b": totb,
	}}
}

func parseDammRemoveLiquidity(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*3+16+8*4 {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	ld, _ := readU128LE(data, o)
	o += 16
	tat, _ := readU64LE(data, o)
	o += 8
	tbt, _ := readU64LE(data, o)
	o += 8
	ta, _ := readU64LE(data, o)
	o += 8
	tb, _ := readU64LE(data, o)
	return DexEvent{"MeteoraDammV2RemoveLiquidity": map[string]any{
		"metadata": meta, "pool": pool, "position": pos, "owner": owner,
		"liquidity_delta": u128LEDecimalString(ld), "token_a_amount_threshold": tat, "token_b_amount_threshold": tbt,
		"token_a_amount": ta, "token_b_amount": tb,
	}}
}

func parseDammDynamicFee(data []byte, o int) map[string]any {
	if o+32 > len(data) {
		return nil
	}
	bs, _ := readU16LE(data, o)
	o += 2
	bu, _ := readU128LE(data, o)
	o += 16
	fp, _ := readU16LE(data, o)
	o += 2
	dp, _ := readU16LE(data, o)
	o += 2
	rf, _ := readU16LE(data, o)
	o += 2
	mva, _ := readU32LE(data, o)
	o += 4
	vfc, _ := readU32LE(data, o)
	return map[string]any{
		"bin_step": bs, "bin_step_u128": u128LEDecimalString(bu),
		"filter_period": fp, "decay_period": dp, "reduction_factor": rf,
		"max_volatility_accumulator": mva, "variable_fee_control": vfc,
	}
}

func parseDammPoolFeeParameters(data []byte, start int) (map[string]any, int, bool) {
	if start+30 > len(data) {
		return nil, start, false
	}
	o := start
	bf := data[o : o+27]
	o += 27
	cfb, ok := readU16LE(data, o)
	if !ok {
		return nil, start, false
	}
	o += 2
	pad, ok := readU8(data, o)
	if !ok {
		return nil, start, false
	}
	o++
	tag, ok := readU8(data, o)
	if !ok {
		return nil, start, false
	}
	o++
	var dyn any
	if tag == 1 {
		d := parseDammDynamicFee(data, o)
		if d == nil {
			return nil, start, false
		}
		dyn = d
		o += 32
	} else if tag != 0 {
		return nil, start, false
	}
	return map[string]any{
		"base_fee_data":           hex.EncodeToString(bf),
		"compounding_fee_bps":     cfb,
		"padding":                 pad,
		"dynamic_fee":             dyn,
	}, o, true
}

func parseDammInitializePool(data []byte, meta EventMetadata) DexEvent {
	const minAfterPub = 31 + 109
	if len(data) < 32*6+minAfterPub {
		return nil
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	tam, _ := readPubkey(data, o)
	o += 32
	tbm, _ := readPubkey(data, o)
	o += 32
	creator, _ := readPubkey(data, o)
	o += 32
	payer, _ := readPubkey(data, o)
	o += 32
	av, _ := readPubkey(data, o)
	o += 32
	pf, next, ok := parseDammPoolFeeParameters(data, o)
	if !ok {
		return nil
	}
	o = next
	if o+109 > len(data) {
		return nil
	}
	smin, _ := readU128LE(data, o)
	o += 16
	smax, _ := readU128LE(data, o)
	o += 16
	act, _ := readU8(data, o)
	o++
	cfm, _ := readU8(data, o)
	o++
	liq, _ := readU128LE(data, o)
	o += 16
	sqrt, _ := readU128LE(data, o)
	o += 16
	ap, _ := readU64LE(data, o)
	o += 8
	taf, _ := readU8(data, o)
	o++
	tbf, _ := readU8(data, o)
	o++
	tau, _ := readU64LE(data, o)
	o += 8
	tbu, _ := readU64LE(data, o)
	o += 8
	tota, _ := readU64LE(data, o)
	o += 8
	totb, _ := readU64LE(data, o)
	o += 8
	pt, _ := readU8(data, o)
	return DexEvent{"MeteoraDammV2InitializePool": map[string]any{
		"metadata": meta, "pool": pool, "token_a_mint": tam, "token_b_mint": tbm,
		"creator": creator, "payer": payer, "alpha_vault": av, "pool_fees": pf,
		"sqrt_min_price": u128LEDecimalString(smin), "sqrt_max_price": u128LEDecimalString(smax),
		"activation_type": act, "collect_fee_mode": cfm,
		"liquidity": u128LEDecimalString(liq), "sqrt_price": u128LEDecimalString(sqrt), "activation_point": ap,
		"token_a_flag": taf, "token_b_flag": tbf,
		"token_a_amount": tau, "token_b_amount": tbu, "total_amount_a": tota, "total_amount_b": totb,
		"pool_type": pt,
	}}
}

// ParseMeteoraDlmmLog 保留为从日志行解析的入口；与 TS `parseMeteoraDlmmLog` 一致
func ParseMeteoraDlmmLog(log, sig string, slot, tx uint64, blockUs *int64, grpcUs int64) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return nil
	}
	meta := makeMetadata(sig, slot, tx, blockUs, grpcUs, "")
	return parseDlmmFromProgramData(buf, meta)
}
