package solparser

import "encoding/binary"

// DLMM discriminators 已在 binary.go 中定义

func parseDlmmFromProgramData(buf []byte, meta EventMetadata) DexEvent {
	if len(buf) < 8 {
		return DexEvent{}
	}
	d := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]

	switch d {
	case dlmmSwap:
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
	case dlmmAddLiq:
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
			Data: &MeteoraDlmmAddLiquidityEvent{
				Metadata:    meta,
				Pool:        pool,
				From:        from,
				Position:    pos,
				Amounts:     []uint64{a0, a1},
				ActiveBinID: ab,
			},
		}
	case dlmmRemoveLiq:
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
			Data: &MeteoraDlmmRemoveLiquidityEvent{
				Metadata:    meta,
				Pool:        pool,
				From:        from,
				Position:    pos,
				Amounts:     []uint64{a0, a1},
				ActiveBinID: ab,
			},
		}
	case dlmmInitPool:
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
			Data: &MeteoraDlmmInitializePoolEvent{
				Metadata:    meta,
				Pool:        pool,
				Creator:     creator,
				ActiveBinID: ab,
				BinStep:     bs,
			},
		}
	case dlmmInitBin:
		if len(data) < 32+32+8 {
			return DexEvent{}
		}
		o := 0
		pool, _ := readPubkey(data, o)
		o += 32
		ba, _ := readPubkey(data, o)
		o += 32
		idx, _ := readU64LE(data, o)
		return DexEvent{
			Type: EventTypeMeteoraDlmmInitializeBinArray,
			Data: &MeteoraDlmmInitializeBinArrayEvent{
				Metadata: meta,
				Pool:     pool,
				BinArray: ba,
				Index:    idx,
			},
		}
	case dlmmCreatePos:
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
			Data: &MeteoraDlmmCreatePositionEvent{
				Metadata:   meta,
				Pool:       pool,
				Position:   pos,
				Owner:      owner,
				LowerBinID: lb,
				Width:      w,
			},
		}
	case dlmmClosePos:
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
			Data: &MeteoraDlmmClosePositionEvent{
				Metadata: meta,
				Pool:     pool,
				Position: pos,
				Owner:    owner,
			},
		}
	case dlmmClaimFee:
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
			Data: &MeteoraDlmmClaimFeeEvent{
				Metadata: meta,
				Pool:     pool,
				Position: pos,
				Owner:    owner,
				FeeX:     fx,
				FeeY:     fy,
			},
		}
	default:
		return DexEvent{}
	}
}
