package solparser

import "encoding/binary"

func discLE(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return binary.LittleEndian.Uint64(b[:8])
}

var (
	discPumpCreate  = discLE([]byte{27, 114, 169, 77, 222, 235, 99, 118})
	discPumpTrade   = discLE([]byte{189, 219, 127, 211, 78, 230, 97, 238})
	discPumpMigrate = discLE([]byte{189, 233, 93, 185, 92, 148, 234, 148})
)

func parseTradeFromData(data []byte, meta EventMetadata, isCreatedBuy bool) DexEvent {
	if len(data) < 32+8+8+1+32+8*5+32+8+8+32+8+8 {
		return nil
	}
	o := 0
	mint, ok := readPubkey(data, o)
	if !ok {
		return nil
	}
	o += 32
	solAmount, _ := readU64LE(data, o)
	o += 8
	tokenAmount, _ := readU64LE(data, o)
	o += 8
	isBuy, ok := readBool(data, o)
	if !ok {
		return nil
	}
	o += 1
	user, ok := readPubkey(data, o)
	if !ok {
		return nil
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
		return nil
	}
	o += 32
	feeBps, _ := readU64LE(data, o)
	o += 8
	fee, _ := readU64LE(data, o)
	o += 8
	creator, ok := readPubkey(data, o)
	if !ok {
		return nil
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
	trade := map[string]any{
		"metadata":                   meta,
		"mint":                       mint,
		"sol_amount":                 solAmount,
		"token_amount":               tokenAmount,
		"is_buy":                     isBuy,
		"is_created_buy":             isCreatedBuy,
		"user":                       user,
		"timestamp":                  ts,
		"virtual_sol_reserves":       vsol,
		"virtual_token_reserves":     vtok,
		"real_sol_reserves":          rsol,
		"real_token_reserves":        rtok,
		"fee_recipient":              feeRec,
		"fee_basis_points":           feeBps,
		"fee":                        fee,
		"creator":                    creator,
		"creator_fee_basis_points":   cfbps,
		"creator_fee":                cfee,
		"track_volume":               tv,
		"total_unclaimed_tokens":     tuc,
		"total_claimed_tokens":       tcc,
		"current_sol_volume":         csv,
		"last_update_timestamp":      lut,
		"ix_name":                    ixName,
		"mayhem_mode":                mm,
		"cashback_fee_basis_points":  cbBps,
		"cashback":                   cb,
		"is_cashback_coin":           cbBps > 0,
		"bonding_curve":              zeroPubkey,
		"associated_bonding_curve":   zeroPubkey,
		"token_program":              zeroPubkey,
		"creator_vault":              zeroPubkey,
	}
	switch ixName {
	case "buy":
		return DexEvent{"PumpFunBuy": trade}
	case "sell":
		return DexEvent{"PumpFunSell": trade}
	case "buy_exact_sol_in":
		return DexEvent{"PumpFunBuyExactSolIn": trade}
	default:
		return DexEvent{"PumpFunTrade": trade}
	}
}

func parseCreateFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	name, o, ok := readBorshString(data, o)
	if !ok {
		return nil
	}
	sym, o, ok := readBorshString(data, o)
	if !ok {
		return nil
	}
	uri, o, ok := readBorshString(data, o)
	if !ok {
		return nil
	}
	if len(data) < o+32*4+8*5+32+1 {
		return nil
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
	ev := map[string]any{
		"metadata": meta, "name": name, "symbol": sym, "uri": uri,
		"mint": mint, "bonding_curve": bc, "user": user, "creator": creator,
		"timestamp": ts, "virtual_token_reserves": vtr, "virtual_sol_reserves": vsol,
		"real_token_reserves": rtr, "token_total_supply": tts, "token_program": tp,
		"is_mayhem_mode": mm, "is_cashback_enabled": ice,
	}
	return DexEvent{"PumpFunCreate": ev}
}

func parseMigrateFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) < 32+32+8+8+8+32+8+32 {
		return nil
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
	ev := map[string]any{
		"metadata": meta, "user": user, "mint": mint, "mint_amount": ma,
		"sol_amount": sa, "pool_migration_fee": pmf, "bonding_curve": bc,
		"timestamp": ts, "pool": pool,
	}
	return DexEvent{"PumpFunMigrate": ev}
}
