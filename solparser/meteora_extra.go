package solparser

import (
	"encoding/binary"
	"encoding/hex"
)

// DAMM discriminators 已在 binary.go 中定义

func usesCompoundingFeeLayout(slot uint64) bool {
	return slot == 0 || slot >= compoundingFeeLayoutActivationSlot
}

// ParseMeteoraDammLog 与 TS `parseMeteoraDammLog` 对齐（Program data 载荷与 `meteora_damm_ix` CPI 内层一致）
func ParseMeteoraDammLog(log, sig string, slot, tx uint64, blockUs *int64, grpcUs int64) DexEvent {
	buf := decodeProgramDataLine(log)
	if len(buf) < 8 {
		return DexEvent{}
	}
	d := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	meta := makeMetadata(sig, slot, tx, blockUs, grpcUs, "")
	return parseMeteoraDammFromDiscriminator(d, data, meta)
}

func parseMeteoraDammFromDiscriminator(d uint64, data []byte, meta EventMetadata) DexEvent {
	switch d {
	case discDammSwap:
		return parseDammSwap(data, meta)
	case discDammSwap2:
		return parseDammSwap2(data, meta)
	case discDammClaimPositionFee:
		return parseDammClaimPositionFee(data, meta)
	case discDammCreatePosition:
		return parseDammCreatePosition(data, meta)
	case discDammClosePosition:
		return parseDammClosePosition(data, meta)
	case discDammAddLiquidity:
		return parseDammAddLiquidity(data, meta)
	case discDammRemoveLiq:
		return parseDammRemoveLiquidity(data, meta)
	case discDammLiquidityChange:
		return parseDammLiquidityChange(data, meta)
	case discDammInitPool:
		return parseDammInitializePool(data, meta)
	case discDammUpdateDelegatePermission:
		return parseDammUpdateDelegatePermission(data, meta)
	case discDammWithdrawDeadLiquidityReward:
		return parseDammWithdrawDeadLiquidityReward(data, meta)
	case discDammCreateConfig:
		return parseDammCreateConfig(data, meta)
	case discDammCreateDynamicConfig:
		return parseDammCreateDynamicConfig(data, meta)
	default:
		return DexEvent{}
	}
}

// ParseMeteoraDammCpiInstruction 与 Rust `meteora_damm::parse_instruction` 一致：CPI 事件 discriminator 位于指令数据 [8..16)，载荷从 [16..) 开始。
func ParseMeteoraDammCpiInstruction(ixData []byte, meta EventMetadata) DexEvent {
	if len(ixData) < 16 {
		return DexEvent{}
	}
	cpi := binary.LittleEndian.Uint64(ixData[8:16])
	payload := ixData[16:]
	return parseMeteoraDammFromDiscriminator(cpi, payload, meta)
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

// parseDammSwap2 aligns with Rust/Node parse_swap2_from_data (full 180-byte EvtSwap2).
func parseDammSwap2(data []byte, meta EventMetadata) DexEvent {
	const swap2Len = 180
	if len(data) < swap2Len {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	td, _ := readU8(data, o)
	o++
	cfm, _ := readU8(data, o)
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
	efi, _ := readU64LE(data, o)
	o += 8
	left, _ := readU64LE(data, o)
	o += 8
	oa, _ := readU64LE(data, o)
	o += 8
	nsp, _ := readU128LE(data, o)
	o += 16
	claimOrTrade, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	o += 8
	compOrPartner, _ := readU64LE(data, o)
	o += 8
	rf, _ := readU64LE(data, o)
	o += 8
	incIn, _ := readU64LE(data, o)
	o += 8
	incOut, _ := readU64LE(data, o)
	o += 8
	excOut, _ := readU64LE(data, o)
	o += 8
	ct, _ := readU64LE(data, o)
	o += 8
	ra, _ := readU64LE(data, o)
	o += 8
	rb, _ := readU64LE(data, o)

	var ai, mo uint64
	switch sm {
	case 0, 1:
		ai, mo = a0, a1
	case 2:
		ai, mo = a1, a0
	default:
		return DexEvent{}
	}

	var lpFee, partnerFee, claimingFee, compoundingFee uint64
	if usesCompoundingFeeLayout(meta.Slot) {
		lpFee = claimOrTrade + compOrPartner
		partnerFee = compOrPartner
		claimingFee = claimOrTrade
		compoundingFee = compOrPartner
	} else {
		lpFee = claimOrTrade
		partnerFee = compOrPartner
		claimingFee = 0
		compoundingFee = 0
	}

	return DexEvent{
		Type: EventTypeMeteoraDammV2Swap,
		Data: &MeteoraDammV2SwapEvent{
			Metadata: meta, Pool: pool, TradeDirection: td, CollectFeeMode: cfm, HasReferral: hr,
			Amount0: a0, Amount1: a1, SwapMode: sm,
			AmountIn: ai, MinimumAmountOut: mo, OutputAmount: oa,
			NextSqrtPrice: u128LEDecimalString(nsp), LpFee: lpFee, ProtocolFee: pf,
			PartnerFee: partnerFee, ReferralFee: rf, ActualAmountIn: ifi,
			ExcludedFeeInputAmount: efi, AmountLeft: left,
			ClaimingFee: claimingFee, CompoundingFee: compoundingFee,
			IncludedTransferFeeAmountIn: incIn, IncludedTransferFeeAmountOut: incOut,
			ExcludedTransferFeeAmountOut: excOut, CurrentTimestamp: ct,
			ReserveAAmount: ra, ReserveBAmount: rb,
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

// parseDammLiquidityChange: EvtLiquidityChange; change_type 0=add, 1=remove.
func parseDammLiquidityChange(data []byte, meta EventMetadata) DexEvent {
	const lenChange = 177
	if len(data) < lenChange {
		return DexEvent{}
	}
	pool, ok := readPubkey(data, 0)
	if !ok {
		return DexEvent{}
	}
	pos, ok := readPubkey(data, 32)
	if !ok {
		return DexEvent{}
	}
	owner, ok := readPubkey(data, 64)
	if !ok {
		return DexEvent{}
	}
	ta, _ := readU64LE(data, 96)
	tb, _ := readU64LE(data, 104)
	tota, _ := readU64LE(data, 112)
	totb, _ := readU64LE(data, 120)
	ra, _ := readU64LE(data, 128)
	rb, _ := readU64LE(data, 136)
	ld, ok := readU128LE(data, 144)
	if !ok {
		return DexEvent{}
	}
	tat, _ := readU64LE(data, 160)
	tbt, _ := readU64LE(data, 168)
	ct, _ := readU8(data, 176)
	switch ct {
	case 0:
		return DexEvent{
			Type: EventTypeMeteoraDammV2AddLiquidity,
			Data: &MeteoraDammV2AddLiquidityEvent{
				Metadata: meta, Pool: pool, Position: pos, Owner: owner,
				TokenAAmount: ta, TokenBAmount: tb, LiquidityDelta: u128LEDecimalString(ld),
				TokenAAmountThreshold: tat, TokenBAmountThreshold: tbt,
				TotalAmountA: tota, TotalAmountB: totb, ReserveAAmount: ra, ReserveBAmount: rb,
			},
		}
	case 1:
		return DexEvent{
			Type: EventTypeMeteoraDammV2RemoveLiquidity,
			Data: &MeteoraDammV2RemoveLiquidityEvent{
				Metadata: meta, Pool: pool, Position: pos, Owner: owner,
				TokenAAmount: ta, TokenBAmount: tb, LiquidityDelta: u128LEDecimalString(ld),
				TokenAAmountThreshold: tat, TokenBAmountThreshold: tbt,
				TotalAmountA: tota, TotalAmountB: totb, ReserveAAmount: ra, ReserveBAmount: rb,
			},
		}
	default:
		return DexEvent{}
	}
}

func parseDammUpdateDelegatePermission(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+1 {
		return DexEvent{}
	}
	o := 0
	pos, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	owner, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	perm, ok := readU32LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 4
	hasDel, ok := readBool(data, o)
	if !ok {
		return DexEvent{}
	}
	o++
	var delegate *string
	if hasDel {
		d, ok := readPubkey(data, o)
		if !ok {
			return DexEvent{}
		}
		delegate = &d
	}
	return DexEvent{
		Type: EventTypeMeteoraDammV2UpdateDelegatePermission,
		Data: &MeteoraDammV2UpdateDelegatePermissionEvent{
			Metadata: meta, Position: pos, Owner: owner, Permission: perm, Delegate: delegate,
		},
	}
}

func parseDammWithdrawDeadLiquidityReward(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8 {
		return DexEvent{}
	}
	o := 0
	pool, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	mint, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	amt, ok := readU64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeMeteoraDammV2WithdrawDeadLiquidityReward,
		Data: &MeteoraDammV2WithdrawDeadLiquidityRewardEvent{
			Metadata: meta, Pool: pool, RewardMint: mint, Amount: amt,
		},
	}
}

func parseDammDynamicFeeParams(data []byte, o int) (*MeteoraDammV2DynamicFeeParameters, int, bool) {
	if o+2+16+2+2+2+4+4 > len(data) {
		return nil, o, false
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
	o += 4
	return &MeteoraDammV2DynamicFeeParameters{
		BinStep: bs, BinStepU128: u128LEDecimalString(bu),
		FilterPeriod: fp, DecayPeriod: dp, ReductionFactor: rf,
		MaxVolatilityAccumulator: mva, VariableFeeControl: vfc,
	}, o, true
}

func parseDammCreateConfig(data []byte, meta EventMetadata) DexEvent {
	o := 0
	if len(data) < o+27 {
		return DexEvent{}
	}
	var baseFee [27]byte
	copy(baseFee[:], data[o:o+27])
	o += 27
	cfb, ok := readU16LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 2
	pad, ok := readU8(data, o)
	if !ok {
		return DexEvent{}
	}
	o++
	hasDyn, ok := readBool(data, o)
	if !ok {
		return DexEvent{}
	}
	o++
	var dyn *MeteoraDammV2DynamicFeeParameters
	if hasDyn {
		params, next, ok := parseDammDynamicFeeParams(data, o)
		if !ok {
			return DexEvent{}
		}
		dyn = params
		o = next
	}
	vault, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	auth, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	act, ok := readU8(data, o)
	if !ok {
		return DexEvent{}
	}
	o++
	smin, ok := readU128LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 16
	smax, ok := readU128LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 16
	cfm, ok := readU8(data, o)
	if !ok {
		return DexEvent{}
	}
	o++
	idx, ok := readU64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	cfg, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	perm, ok := readU128LE(data, o)
	if !ok {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeMeteoraDammV2CreateConfig,
		Data: &MeteoraDammV2CreateConfigEvent{
			Metadata: meta, BaseFeeData: baseFee, CompoundingFeeBps: cfb, Padding: pad,
			DynamicFee: dyn, VaultConfigKey: vault, PoolCreatorAuthority: auth,
			ActivationType: act, SqrtMinPrice: u128LEDecimalString(smin),
			SqrtMaxPrice: u128LEDecimalString(smax), CollectFeeMode: cfm,
			Index: idx, Config: cfg, Permission: u128LEDecimalString(perm),
		},
	}
}

func parseDammCreateDynamicConfig(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+16 {
		return DexEvent{}
	}
	o := 0
	cfg, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	auth, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	idx, ok := readU64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	perm, ok := readU128LE(data, o)
	if !ok {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeMeteoraDammV2CreateDynamicConfig,
		Data: &MeteoraDammV2CreateDynamicConfigEvent{
			Metadata: meta, Config: cfg, PoolCreatorAuthority: auth,
			Index: idx, Permission: u128LEDecimalString(perm),
		},
	}
}

func parseDammDynamicFee(data []byte, o int) map[string]any {
	params, _, ok := parseDammDynamicFeeParams(data, o)
	if !ok {
		return nil
	}
	return map[string]any{
		"bin_step": params.BinStep, "bin_step_u128": params.BinStepU128,
		"filter_period": params.FilterPeriod, "decay_period": params.DecayPeriod,
		"reduction_factor":           params.ReductionFactor,
		"max_volatility_accumulator": params.MaxVolatilityAccumulator,
		"variable_fee_control":       params.VariableFeeControl,
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
		"base_fee_data":       hex.EncodeToString(bf),
		"compounding_fee_bps": cfb,
		"padding":             pad,
		"dynamic_fee":         dyn,
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
			Metadata:        meta,
			Pool:            pool,
			TokenAMint:      tam,
			TokenBMint:      tbm,
			Creator:         creator,
			Payer:           payer,
			AlphaVault:      av,
			PoolFees:        pf,
			SqrtMinPrice:    u128LEDecimalString(smin),
			SqrtMaxPrice:    u128LEDecimalString(smax),
			ActivationType:  act,
			CollectFeeMode:  cfm,
			Liquidity:       u128LEDecimalString(liq),
			SqrtPrice:       u128LEDecimalString(sqrt),
			ActivationPoint: ap,
			TokenAFlag:      taf,
			TokenBFlag:      tbf,
			TokenAAmount:    tau,
			TokenBAmount:    tbu,
			TotalAmountA:    tota,
			TotalAmountB:    totb,
			PoolType:        pt,
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

func parseDammClaimPositionFee(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 112 {
		return DexEvent{}
	}
	pool, _ := readPubkey(data, 0)
	position, _ := readPubkey(data, 32)
	owner, _ := readPubkey(data, 64)
	return DexEvent{Type: EventTypeMeteoraDammV2ClaimPositionFee, Data: &MeteoraDammV2ClaimPositionFeeEvent{
		Metadata: meta, Pool: pool, Position: position, Owner: owner, FeeAClaimed: binary.LittleEndian.Uint64(data[96:104]), FeeBClaimed: binary.LittleEndian.Uint64(data[104:112])}}
}
