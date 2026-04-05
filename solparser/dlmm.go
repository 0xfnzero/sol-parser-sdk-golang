package solparser

import "encoding/binary"

// 与 TS `logs/meteora_dlmm.ts` 中 DLMM 常量一致
var (
	dlmmSwap        = disc8(143, 190, 90, 218, 196, 30, 51, 222)
	dlmmAddLiq      = disc8(181, 157, 89, 67, 143, 182, 52, 72)
	dlmmRemoveLiq   = disc8(80, 85, 209, 72, 24, 206, 35, 178)
	dlmmInitBin     = disc8(11, 18, 155, 194, 33, 115, 238, 119)
	dlmmInitPool    = disc8(95, 180, 10, 172, 84, 174, 232, 40)
	dlmmCreatePos   = disc8(123, 233, 11, 43, 146, 180, 97, 119)
	dlmmClosePos    = disc8(94, 168, 102, 45, 59, 122, 137, 54)
	dlmmClaimFee    = disc8(152, 70, 208, 111, 104, 91, 44, 1)
)

func parseDlmmFromProgramData(buf []byte, meta EventMetadata) DexEvent {
	if len(buf) < 8 {
		return nil
	}
	d := binary.LittleEndian.Uint64(buf[:8])
	data := buf[8:]

	switch d {
	case dlmmSwap:
		if len(data) < 32+32+4+4+8+8+1+8+8+16+8 {
			return nil
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
			return nil
		}
		o += 16
		hf, _ := readU64LE(data, o)
		return DexEvent{"MeteoraDlmmSwap": map[string]any{
			"metadata": meta, "pool": pool, "from": from,
			"start_bin_id": sb, "end_bin_id": eb,
			"amount_in": ai, "amount_out": ao, "swap_for_y": sy,
			"fee": fee, "protocol_fee": pf, "fee_bps": u128LEDecimalString(fbps), "host_fee": hf,
		}}
	case dlmmAddLiq:
		if len(data) < 32+32+32+8+8+4 {
			return nil
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
		return DexEvent{"MeteoraDlmmAddLiquidity": map[string]any{
			"metadata": meta, "pool": pool, "from": from, "position": pos,
			"amounts": []uint64{a0, a1}, "active_bin_id": ab,
		}}
	case dlmmRemoveLiq:
		if len(data) < 32+32+32+8+8+4 {
			return nil
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
		return DexEvent{"MeteoraDlmmRemoveLiquidity": map[string]any{
			"metadata": meta, "pool": pool, "from": from, "position": pos,
			"amounts": []uint64{a0, a1}, "active_bin_id": ab,
		}}
	case dlmmInitPool:
		if len(data) < 32+32+4+2 {
			return nil
		}
		o := 0
		pool, _ := readPubkey(data, o)
		o += 32
		creator, _ := readPubkey(data, o)
		o += 32
		ab, _ := readI32LE(data, o)
		o += 4
		bs, _ := readU16LE(data, o)
		return DexEvent{"MeteoraDlmmInitializePool": map[string]any{
			"metadata": meta, "pool": pool, "creator": creator,
			"active_bin_id": ab, "bin_step": bs,
		}}
	case dlmmInitBin:
		if len(data) < 32+32+8 {
			return nil
		}
		o := 0
		pool, _ := readPubkey(data, o)
		o += 32
		ba, _ := readPubkey(data, o)
		o += 32
		idx, _ := readU64LE(data, o)
		return DexEvent{"MeteoraDlmmInitializeBinArray": map[string]any{
			"metadata": meta, "pool": pool, "bin_array": ba, "index": idx,
		}}
	case dlmmCreatePos:
		if len(data) < 32+32+32+4+4 {
			return nil
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
		return DexEvent{"MeteoraDlmmCreatePosition": map[string]any{
			"metadata": meta, "pool": pool, "position": pos, "owner": owner,
			"lower_bin_id": lb, "width": w,
		}}
	case dlmmClosePos:
		if len(data) < 32+32+32 {
			return nil
		}
		o := 0
		pool, _ := readPubkey(data, o)
		o += 32
		pos, _ := readPubkey(data, o)
		o += 32
		owner, _ := readPubkey(data, o)
		return DexEvent{"MeteoraDlmmClosePosition": map[string]any{
			"metadata": meta, "pool": pool, "position": pos, "owner": owner,
		}}
	case dlmmClaimFee:
		if len(data) < 32+32+32+8+8 {
			return nil
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
		return DexEvent{"MeteoraDlmmClaimFee": map[string]any{
			"metadata": meta, "pool": pool, "position": pos, "owner": owner,
			"fee_x": fx, "fee_y": fy,
		}}
	default:
		return nil
	}
}
