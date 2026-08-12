package solparser

import "encoding/binary"

func parseDlmmFromProgramData(buf []byte, meta EventMetadata) DexEvent {
	if len(buf) < 8 {
		return DexEvent{}
	}
	d := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]
	return parseDlmmEventData(d, data, meta)
}

func parseDlmmEventData(d uint64, data []byte, meta EventMetadata) DexEvent {
	switch d {
	case dlmmSwap, dlmmLegacySwap:
		return parseDlmmSwapData(data, meta)
	case dlmmSwap2:
		return parseDlmmSwap2Data(data, meta)
	case dlmmAddLiq, dlmmLegacyAddLiq:
		return parseDlmmAddLiquidityData(data, meta)
	case dlmmRemoveLiq, dlmmLegacyRemoveLiq:
		return parseDlmmRemoveLiquidityData(data, meta)
	case dlmmInitPool:
		return parseDlmmLbPairCreateData(data, meta)
	case dlmmLegacyInitPool:
		return parseDlmmLegacyInitPoolData(data, meta)
	case dlmmInitBin:
		return parseDlmmInitBinData(data, meta)
	case dlmmCreatePos:
		return parseDlmmPositionCreateData(data, meta)
	case dlmmLegacyCreatePos:
		return parseDlmmLegacyCreatePositionData(data, meta)
	case dlmmClosePos:
		return parseDlmmPositionCloseData(data, meta)
	case dlmmLegacyClosePos:
		return parseDlmmLegacyClosePositionData(data, meta)
	case dlmmClaimFee, dlmmLegacyClaimFee:
		return parseDlmmClaimFeeData(data, meta)
	case dlmmClaimFee2:
		return parseDlmmClaimFee2Data(data, meta)
	default:
		return DexEvent{}
	}
}

func parseDlmmSwapData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+4+8+8+1+8+8+16+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	from, _ := readPubkey(data, o)
	o += 32
	sb, _ := readI32LE(data, o)
	o += 4
	eb, _ := readI32LE(data, o)
	o += 4
	ai, _ := readU64LE(data, o)
	o += 8
	ao, _ := readU64LE(data, o)
	o += 8
	sy, _ := readBool(data, o)
	o++
	fee, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	o += 8
	fbps, ok := readU128LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 16
	hf, _ := readU64LE(data, o)
	return dlmmSwapEvent(meta, pool, from, sb, eb, ai, ao, sy, fee, pf, fbps, hf)
}

func parseDlmmSwap2Data(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+4+1+16+8+8+8+8+8+8+8+1+1 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	from, _ := readPubkey(data, o)
	o += 32
	sb, _ := readI32LE(data, o)
	o += 4
	eb, _ := readI32LE(data, o)
	o += 4
	sy, _ := readBool(data, o)
	o++
	fbps, ok := readU128LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 16
	ai, _ := readU64LE(data, o)
	o += 8
	o += 8
	ao, _ := readU64LE(data, o)
	o += 8
	fee, _ := readU64LE(data, o)
	o += 8
	pf, _ := readU64LE(data, o)
	o += 8
	o += 8
	hf, _ := readU64LE(data, o)
	return dlmmSwapEvent(meta, pool, from, sb, eb, ai, ao, sy, fee, pf, fbps, hf)
}

func dlmmSwapEvent(meta EventMetadata, pool, from string, sb, eb int32, ai, ao uint64, sy bool, fee, pf uint64, fbps [16]byte, hf uint64) DexEvent {
	return DexEvent{
		Type: EventTypeMeteoraDlmmSwap,
		Data: &MeteoraDlmmSwapEvent{
			Metadata:    meta,
			Pool:        pool,
			From:        from,
			StartBinID:  sb,
			EndBinID:    eb,
			AmountIn:    ai,
			AmountOut:   ao,
			SwapForY:    sy,
			Fee:         fee,
			ProtocolFee: pf,
			FeeBps:      u128LEDecimalString(fbps),
			HostFee:     hf,
		},
	}
}

func parseDlmmAddLiquidityData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8+8+4 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	from, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	o += 8
	ab, _ := readI32LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmAddLiquidity,
		Data: &MeteoraDlmmAddLiquidityEvent{Metadata: meta, Pool: pool, From: from, Position: pos, Amounts: []uint64{a0, a1}, ActiveBinID: ab},
	}
}

func parseDlmmRemoveLiquidityData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8+8+4 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	from, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	a0, _ := readU64LE(data, o)
	o += 8
	a1, _ := readU64LE(data, o)
	o += 8
	ab, _ := readI32LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmRemoveLiquidity,
		Data: &MeteoraDlmmRemoveLiquidityEvent{Metadata: meta, Pool: pool, From: from, Position: pos, Amounts: []uint64{a0, a1}, ActiveBinID: ab},
	}
}

func parseDlmmLbPairCreateData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+2+32+32 {
		return DexEvent{}
	}
	pool, _ := readPubkey(data, 0)
	bs, _ := readU16LE(data, 32)
	return DexEvent{
		Type: EventTypeMeteoraDlmmInitializePool,
		Data: &MeteoraDlmmInitializePoolEvent{Metadata: meta, Pool: pool, Creator: zeroPubkey, ActiveBinID: 0, BinStep: bs},
	}
}

func parseDlmmLegacyInitPoolData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+4+2 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	creator, _ := readPubkey(data, o)
	o += 32
	ab, _ := readI32LE(data, o)
	o += 4
	bs, _ := readU16LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmInitializePool,
		Data: &MeteoraDlmmInitializePoolEvent{Metadata: meta, Pool: pool, Creator: creator, ActiveBinID: ab, BinStep: bs},
	}
}

func parseDlmmInitBinData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	ba, _ := readPubkey(data, o)
	o += 32
	idx, _ := readI64LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmInitializeBinArray,
		Data: &MeteoraDlmmInitializeBinArrayEvent{Metadata: meta, Pool: pool, BinArray: ba, Index: idx},
	}
}

func parseDlmmPositionCreateData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmCreatePosition,
		Data: &MeteoraDlmmCreatePositionEvent{Metadata: meta, Pool: pool, Position: pos, Owner: owner, LowerBinID: 0, Width: 0},
	}
}

func parseDlmmLegacyCreatePositionData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+4+4 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	lb, _ := readI32LE(data, o)
	o += 4
	w, _ := readU32LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmCreatePosition,
		Data: &MeteoraDlmmCreatePositionEvent{Metadata: meta, Pool: pool, Position: pos, Owner: owner, LowerBinID: lb, Width: w},
	}
}

func parseDlmmPositionCloseData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32 {
		return DexEvent{}
	}
	pos, _ := readPubkey(data, 0)
	owner, _ := readPubkey(data, 32)
	return DexEvent{
		Type: EventTypeMeteoraDlmmClosePosition,
		Data: &MeteoraDlmmClosePositionEvent{Metadata: meta, Pool: zeroPubkey, Position: pos, Owner: owner},
	}
}

func parseDlmmLegacyClosePositionData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmClosePosition,
		Data: &MeteoraDlmmClosePositionEvent{Metadata: meta, Pool: pool, Position: pos, Owner: owner},
	}
}

func parseDlmmClaimFeeData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8+8 {
		return DexEvent{}
	}
	o := 0
	pool, _ := readPubkey(data, o)
	o += 32
	pos, _ := readPubkey(data, o)
	o += 32
	owner, _ := readPubkey(data, o)
	o += 32
	fx, _ := readU64LE(data, o)
	o += 8
	fy, _ := readU64LE(data, o)
	return DexEvent{
		Type: EventTypeMeteoraDlmmClaimFee,
		Data: &MeteoraDlmmClaimFeeEvent{Metadata: meta, Pool: pool, Position: pos, Owner: owner, FeeX: fx, FeeY: fy},
	}
}

func parseDlmmClaimFee2Data(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+32+8+8+4 {
		return DexEvent{}
	}
	return parseDlmmClaimFeeData(data[:32+32+32+8+8], meta)
}
