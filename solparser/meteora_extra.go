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
		return DexEvent{}
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
		return DexEvent{}
	}
}

func parseDammSwap(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+1+1+8*8+16+8*4 {
		return DexEvent{}
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
	return DexEvent{
		Type: EventTypeMeteoraDammV2Swap,
		Data: &MeteoraDammV2SwapEvent{
			Metadata: meta, Pool: pool, TradeDirection: td, HasReferral: hr,
			AmountIn: ai, MinimumAmountOut: mo, OutputAmount: oa,
			NextSqrtPrice: u128LEDecimalString(nsp), LpFee: lpf, ProtocolFee: pf,
			PartnerFee: 0, ReferralFee: rf, ActualAmountIn: aai, CurrentTimestamp: ct,
			TokenAVault: zeroPubkey, TokenBVault: zeroPubkey, TokenAMint: zeroPubkey,
			TokenBMint: zeroPubkey, TokenAProgram: zeroPubkey, TokenBProgram: zeroPubkey,
		},
	}
}

func parseDammSwap2(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+1+1+1+8*2+1+8*6+16+8*4+8*3 {
		return DexEvent{}
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
	return DexEvent{
		Type: EventTypeMeteoraDammV2Swap,
		Data: &MeteoraDammV2SwapEvent{
			Metadata: meta, Pool: pool, TradeDirection: td, HasReferral: hr,
			AmountIn: ai, MinimumAmountOut: mo, OutputAmount: oa,
			NextSqrtPrice: u128LEDecimalString(nsp), LpFee: lpf, ProtocolFee: pf,
			PartnerFee: 0, ReferralFee: rf, ActualAmountIn: ifi, CurrentTimestamp: ct,
			TokenAVault: zeroPubkey, TokenBVault: zeroPubkey, TokenAMint: zeroPubkey,
			TokenBMint: zeroPubkey, TokenAProgram: zeroPubkey, TokenBProgram: zeroPubkey,
		},
	}
}

func parseDammCreatePosition(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	nft, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDammV2CreatePosition,
		Data: &MeteoraDammV2CreatePositionEvent{
			Metadata:        meta,
			Pool:            pool,
			Owner:           owner,
			Position:        pos,
			PositionNftMint: nft,
		},
	}
}

func parseDammClosePosition(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*4 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	nft, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDammV2ClosePosition,
		Data: &MeteoraDammV2ClosePositionEvent{
			Metadata:        meta,
			Pool:            pool,
			Owner:           owner,
			Position:        pos,
			PositionNftMint: nft,
		},
	}
}

func parseDammAddLiquidity(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*3+16+8*6 {
		return DexEvent{}
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
	return DexEvent{
		Type: EventTypeMeteoraDammV2AddLiquidity,
		Data: &MeteoraDammV2AddLiquidityEvent{
			Metadata:              meta,
			Pool:                  pool,
			Position:              pos,
			Owner:                 owner,
			LiquidityDelta:        u128LEDecimalString(ld),
			TokenAAmountThreshold: tat,
			TokenBAmountThreshold: tbt,
			TokenAAmount:          ta,
			TokenBAmount:          tb,
			TotalAmountA:          tota,
			TotalAmountB:          totb,
		},
	}
}

func parseDammRemoveLiquidity(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32*3+16+8*4 {
		return DexEvent{}
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
	return DexEvent{
		Type: EventTypeMeteoraDammV2RemoveLiquidity,
		Data: &MeteoraDammV2RemoveLiquidityEvent{
			Metadata:              meta,
			Pool:                  pool,
			Position:              pos,
			Owner:                 owner,
			LiquidityDelta:        u128LEDecimalString(ld),
			TokenAAmountThreshold: tat,
			TokenBAmountThreshold: tbt,
			TokenAAmount:          ta,
			TokenBAmount:          tb,
		},
	}
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
		return DexEvent{}
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
		return DexEvent{}
	}
	o = next
	if o+109 > len(data) {
		return DexEvent{}
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
	return DexEvent{
		Type: EventTypeMeteoraDammV2InitializePool,
		Data: &MeteoraDammV2InitializePoolEvent{
			Metadata:       meta,
			Pool:           pool,
			TokenAMint:     tam,
			TokenBMint:     tbm,
			Creator:        creator,
			Payer:          payer,
			AlphaVault:     av,
			PoolFees:       pf,
			SqrtMinPrice:   u128LEDecimalString(smin),
			SqrtMaxPrice:   u128LEDecimalString(smax),
			ActivationType: act,
			CollectFeeMode: cfm,
			Liquidity:      u128LEDecimalString(liq),
			SqrtPrice:      u128LEDecimalString(sqrt),
			ActivationPoint: ap,
			TokenAFlag:     taf,
			TokenBFlag:     tbf,
			TokenAAmount:   tau,
			TokenBAmount:   tbu,
			TotalAmountA:   tota,
			TotalAmountB:   totb,
			PoolType:       pt,
		},
	}
}

// ParseMeteoraDlmmLog 保留为从日志行解析的入口；与 TS `parseMeteoraDlmmLog` 一致
func ParseMeteoraDlmmLog(log, sig string, slot, tx uint64, blockUs *int64, grpcUs int64) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return DexEvent{}
	}
	meta := makeMetadata(sig, slot, tx, blockUs, grpcUs, "")
	return parseDlmmFromProgramData(buf, meta)
}
