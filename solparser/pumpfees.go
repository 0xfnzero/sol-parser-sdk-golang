package solparser

const (
	maxPumpFeesShareholders = 64
	maxPumpFeesFeeTiers     = 64
)

var (
	instrPumpFeesCreateFeeSharingConfig      = disc8(195, 78, 86, 76, 111, 52, 251, 213)
	instrPumpFeesInitializeFeeConfig         = disc8(62, 162, 20, 133, 121, 65, 145, 27)
	instrPumpFeesResetFeeSharingConfig       = disc8(10, 2, 182, 95, 16, 127, 129, 186)
	instrPumpFeesRevokeFeeSharingAuthority   = disc8(18, 233, 158, 39, 185, 207, 58, 104)
	instrPumpFeesTransferFeeSharingAuthority = disc8(202, 10, 75, 200, 164, 34, 210, 96)
	instrPumpFeesUpdateAdmin                 = disc8(161, 176, 40, 213, 60, 184, 179, 228)
	instrPumpFeesUpdateFeeConfig             = disc8(104, 184, 103, 242, 88, 151, 107, 20)
	instrPumpFeesUpdateFeeShares             = disc8(189, 13, 136, 99, 187, 164, 237, 35)
	instrPumpFeesUpsertFeeTiers              = disc8(227, 23, 150, 12, 77, 86, 94, 4)
)

func readPumpFeesFeesAt(data []byte, o *int) (PumpFeesFees, bool) {
	lp, ok := readU64LE(data, *o)
	if !ok {
		return PumpFeesFees{}, false
	}
	*o += 8
	protocol, ok := readU64LE(data, *o)
	if !ok {
		return PumpFeesFees{}, false
	}
	*o += 8
	creator, ok := readU64LE(data, *o)
	if !ok {
		return PumpFeesFees{}, false
	}
	*o += 8
	return PumpFeesFees{LpFeeBps: lp, ProtocolFeeBps: protocol, CreatorFeeBps: creator}, true
}

func readPumpFeesShareholdersVec(data []byte, o *int) ([]PumpFeesShareholder, bool) {
	n, ok := readU32LE(data, *o)
	if !ok || n > maxPumpFeesShareholders {
		return nil, false
	}
	*o += 4
	out := make([]PumpFeesShareholder, 0, n)
	for i := uint32(0); i < n; i++ {
		address, ok := readPubkey(data, *o)
		if !ok {
			return nil, false
		}
		*o += 32
		share, ok := readU16LE(data, *o)
		if !ok {
			return nil, false
		}
		*o += 2
		out = append(out, PumpFeesShareholder{Address: address, ShareBps: share})
	}
	return out, true
}

func readPumpFeesFeeTiersVec(data []byte, o *int) ([]PumpFeesFeeTier, bool) {
	n, ok := readU32LE(data, *o)
	if !ok || n > maxPumpFeesFeeTiers {
		return nil, false
	}
	*o += 4
	out := make([]PumpFeesFeeTier, 0, n)
	for i := uint32(0); i < n; i++ {
		threshold, ok := readU128LE(data, *o)
		if !ok {
			return nil, false
		}
		*o += 16
		fees, ok := readPumpFeesFeesAt(data, o)
		if !ok {
			return nil, false
		}
		out = append(out, PumpFeesFeeTier{
			MarketCapLamportsThreshold: u128LEDecimalStringInline(threshold),
			Fees:                       fees,
		})
	}
	return out, true
}

func readPumpFeesOptionPubkey(data []byte, o *int) (string, bool) {
	tag, ok := readU8(data, *o)
	if !ok {
		return "", false
	}
	*o += 1
	if tag == 0 {
		return "", true
	}
	if tag != 1 {
		return "", false
	}
	pk, ok := readPubkey(data, *o)
	if !ok {
		return "", false
	}
	*o += 32
	return pk, true
}

func readPumpFeesConfigStatus(data []byte, o *int) (PumpFeesConfigStatus, bool) {
	tag, ok := readU8(data, *o)
	if !ok {
		return "", false
	}
	*o += 1
	switch tag {
	case 0:
		return PumpFeesConfigStatusPaused, true
	case 1:
		return PumpFeesConfigStatusActive, true
	default:
		return "", false
	}
}

func parsePumpFeesCreateFeeSharingConfigFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	ts, ok := readI64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	mint, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	bondingCurve, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	pool, ok := readPumpFeesOptionPubkey(data, &o)
	if !ok {
		return DexEvent{}
	}
	sharingConfig, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	admin, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	shareholders, ok := readPumpFeesShareholdersVec(data, &o)
	if !ok {
		return DexEvent{}
	}
	status, ok := readPumpFeesConfigStatus(data, &o)
	if !ok || o != len(data) {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypePumpFeesCreateFeeSharingConfig,
		Data: &PumpFeesCreateFeeSharingConfigEvent{
			Metadata:            meta,
			Timestamp:           ts,
			Mint:                mint,
			BondingCurve:        bondingCurve,
			Pool:                pool,
			SharingConfig:       sharingConfig,
			Admin:               admin,
			InitialShareholders: shareholders,
			Status:              status,
		},
	}
}

func parsePumpFeesInitializeFeeConfigFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) != 8+32+32 {
		return DexEvent{}
	}
	o := 0
	ts, _ := readI64LE(data, o)
	o += 8
	admin, _ := readPubkey(data, o)
	o += 32
	feeConfig, _ := readPubkey(data, o)
	return DexEvent{Type: EventTypePumpFeesInitializeFeeConfig, Data: &PumpFeesInitializeFeeConfigEvent{
		Metadata: meta, Timestamp: ts, Admin: admin, FeeConfig: feeConfig,
	}}
}

func parsePumpFeesResetFeeSharingConfigFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	ts, ok := readI64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	mint, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	sharingConfig, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	oldAdmin, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	oldShareholders, ok := readPumpFeesShareholdersVec(data, &o)
	if !ok {
		return DexEvent{}
	}
	newAdmin, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	newShareholders, ok := readPumpFeesShareholdersVec(data, &o)
	if !ok || o != len(data) {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypePumpFeesResetFeeSharingConfig, Data: &PumpFeesResetFeeSharingConfigEvent{
		Metadata: meta, Timestamp: ts, Mint: mint, SharingConfig: sharingConfig,
		OldAdmin: oldAdmin, OldShareholders: oldShareholders, NewAdmin: newAdmin,
		NewShareholders: newShareholders,
	}}
}

func parsePumpFeesRevokeFeeSharingAuthorityFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) != 8+32+32+32 {
		return DexEvent{}
	}
	o := 0
	ts, _ := readI64LE(data, o)
	o += 8
	mint, _ := readPubkey(data, o)
	o += 32
	sharingConfig, _ := readPubkey(data, o)
	o += 32
	admin, _ := readPubkey(data, o)
	return DexEvent{Type: EventTypePumpFeesRevokeFeeSharingAuthority, Data: &PumpFeesRevokeFeeSharingAuthorityEvent{
		Metadata: meta, Timestamp: ts, Mint: mint, SharingConfig: sharingConfig, Admin: admin,
	}}
}

func parsePumpFeesTransferFeeSharingAuthorityFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) != 8+32+32+32+32 {
		return DexEvent{}
	}
	o := 0
	ts, _ := readI64LE(data, o)
	o += 8
	mint, _ := readPubkey(data, o)
	o += 32
	sharingConfig, _ := readPubkey(data, o)
	o += 32
	oldAdmin, _ := readPubkey(data, o)
	o += 32
	newAdmin, _ := readPubkey(data, o)
	return DexEvent{Type: EventTypePumpFeesTransferFeeSharingAuthority, Data: &PumpFeesTransferFeeSharingAuthorityEvent{
		Metadata: meta, Timestamp: ts, Mint: mint, SharingConfig: sharingConfig, OldAdmin: oldAdmin, NewAdmin: newAdmin,
	}}
}

func parsePumpFeesUpdateAdminFromData(data []byte, meta EventMetadata) DexEvent {
	if len(data) != 8+32+32 {
		return DexEvent{}
	}
	o := 0
	ts, _ := readI64LE(data, o)
	o += 8
	oldAdmin, _ := readPubkey(data, o)
	o += 32
	newAdmin, _ := readPubkey(data, o)
	return DexEvent{Type: EventTypePumpFeesUpdateAdmin, Data: &PumpFeesUpdateAdminEvent{
		Metadata: meta, Timestamp: ts, OldAdmin: oldAdmin, NewAdmin: newAdmin,
	}}
}

func parsePumpFeesUpdateFeeConfigFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	ts, ok := readI64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	admin, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	feeConfig, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	feeTiers, ok := readPumpFeesFeeTiersVec(data, &o)
	if !ok {
		return DexEvent{}
	}
	flatFees, ok := readPumpFeesFeesAt(data, &o)
	if !ok || o != len(data) {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypePumpFeesUpdateFeeConfig, Data: &PumpFeesUpdateFeeConfigEvent{
		Metadata: meta, Timestamp: ts, Admin: admin, FeeConfig: feeConfig,
		FeeTiers: feeTiers, FlatFees: flatFees,
	}}
}

func parsePumpFeesUpdateFeeSharesFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	ts, ok := readI64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	mint, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	sharingConfig, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	admin, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	shareholders, ok := readPumpFeesShareholdersVec(data, &o)
	if !ok || o != len(data) {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypePumpFeesUpdateFeeShares, Data: &PumpFeesUpdateFeeSharesEvent{
		Metadata: meta, Timestamp: ts, Mint: mint, SharingConfig: sharingConfig, Admin: admin,
		BondingCurve: zeroPubkey, PumpCreatorVault: zeroPubkey, NewShareholders: shareholders,
	}}
}

func parsePumpFeesUpsertFeeTiersFromData(data []byte, meta EventMetadata) DexEvent {
	o := 0
	ts, ok := readI64LE(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 8
	admin, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	feeConfig, ok := readPubkey(data, o)
	if !ok {
		return DexEvent{}
	}
	o += 32
	feeTiers, ok := readPumpFeesFeeTiersVec(data, &o)
	if !ok {
		return DexEvent{}
	}
	offset, ok := readU8(data, o)
	if !ok {
		return DexEvent{}
	}
	o++
	if o != len(data) {
		return DexEvent{}
	}
	return DexEvent{Type: EventTypePumpFeesUpsertFeeTiers, Data: &PumpFeesUpsertFeeTiersEvent{
		Metadata: meta, Timestamp: ts, Admin: admin, FeeConfig: feeConfig, FeeTiers: feeTiers, Offset: offset,
	}}
}

func parsePumpFeesInstruction(data []byte, accounts []string, meta EventMetadata) DexEvent {
	if len(data) < 8 {
		return DexEvent{}
	}
	disc, _ := readDiscU64(data)
	acc := func(i int) (string, bool) {
		if i < 0 || i >= len(accounts) {
			return "", false
		}
		return accounts[i], true
	}
	accOrZero := func(i int) string {
		if i < 0 || i >= len(accounts) {
			return zeroPubkey
		}
		return accounts[i]
	}

	switch disc {
	case instrPumpFeesCreateFeeSharingConfig:
		admin, ok1 := acc(2)
		mint, ok2 := acc(4)
		if !ok1 || !ok2 {
			return DexEvent{}
		}
		pool := ""
		if len(accounts) > 10 {
			pool = accounts[10]
		}
		return DexEvent{Type: EventTypePumpFeesCreateFeeSharingConfig, Data: &PumpFeesCreateFeeSharingConfigEvent{
			Metadata: meta, Timestamp: 0, Mint: mint, BondingCurve: accOrZero(7), Pool: pool,
			SharingConfig: accOrZero(5), Admin: admin, InitialShareholders: []PumpFeesShareholder{},
			Status: PumpFeesConfigStatusActive,
		}}
	case instrPumpFeesUpdateFeeShares:
		admin, ok1 := acc(2)
		mint, ok2 := acc(4)
		sharingConfig, ok3 := acc(5)
		if !ok1 || !ok2 || !ok3 {
			return DexEvent{}
		}
		o := 8
		shareholders, ok := readPumpFeesShareholdersVec(data, &o)
		if !ok || o != len(data) {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesUpdateFeeShares, Data: &PumpFeesUpdateFeeSharesEvent{
			Metadata: meta, Timestamp: 0, Mint: mint, SharingConfig: sharingConfig, Admin: admin,
			BondingCurve: accOrZero(6), PumpCreatorVault: accOrZero(7), NewShareholders: shareholders,
		}}
	case instrPumpFeesInitializeFeeConfig:
		admin, ok1 := acc(0)
		feeConfig, ok2 := acc(1)
		if !ok1 || !ok2 {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesInitializeFeeConfig, Data: &PumpFeesInitializeFeeConfigEvent{
			Metadata: meta, Timestamp: 0, Admin: admin, FeeConfig: feeConfig,
		}}
	case instrPumpFeesResetFeeSharingConfig:
		oldAdmin, ok1 := acc(0)
		newAdmin, ok2 := acc(2)
		mint, ok3 := acc(3)
		sharingConfig, ok4 := acc(4)
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesResetFeeSharingConfig, Data: &PumpFeesResetFeeSharingConfigEvent{
			Metadata: meta, Timestamp: 0, Mint: mint, SharingConfig: sharingConfig,
			OldAdmin: oldAdmin, OldShareholders: []PumpFeesShareholder{},
			NewAdmin: newAdmin, NewShareholders: []PumpFeesShareholder{},
		}}
	case instrPumpFeesRevokeFeeSharingAuthority:
		admin, ok1 := acc(0)
		mint, ok2 := acc(2)
		sharingConfig, ok3 := acc(3)
		if !ok1 || !ok2 || !ok3 {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesRevokeFeeSharingAuthority, Data: &PumpFeesRevokeFeeSharingAuthorityEvent{
			Metadata: meta, Timestamp: 0, Mint: mint, SharingConfig: sharingConfig, Admin: admin,
		}}
	case instrPumpFeesTransferFeeSharingAuthority:
		oldAdmin, ok1 := acc(0)
		mint, ok2 := acc(2)
		sharingConfig, ok3 := acc(3)
		newAdmin, ok4 := acc(4)
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesTransferFeeSharingAuthority, Data: &PumpFeesTransferFeeSharingAuthorityEvent{
			Metadata: meta, Timestamp: 0, Mint: mint, SharingConfig: sharingConfig, OldAdmin: oldAdmin, NewAdmin: newAdmin,
		}}
	case instrPumpFeesUpdateAdmin:
		oldAdmin, ok1 := acc(0)
		newAdmin, ok2 := acc(2)
		if !ok1 || !ok2 {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesUpdateAdmin, Data: &PumpFeesUpdateAdminEvent{
			Metadata: meta, Timestamp: 0, OldAdmin: oldAdmin, NewAdmin: newAdmin,
		}}
	case instrPumpFeesUpdateFeeConfig:
		feeConfig, ok1 := acc(0)
		admin, ok2 := acc(1)
		if !ok1 || !ok2 {
			return DexEvent{}
		}
		o := 8
		feeTiers, ok := readPumpFeesFeeTiersVec(data, &o)
		if !ok {
			return DexEvent{}
		}
		flatFees, ok := readPumpFeesFeesAt(data, &o)
		if !ok || o != len(data) {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesUpdateFeeConfig, Data: &PumpFeesUpdateFeeConfigEvent{
			Metadata: meta, Timestamp: 0, Admin: admin, FeeConfig: feeConfig,
			FeeTiers: feeTiers, FlatFees: flatFees,
		}}
	case instrPumpFeesUpsertFeeTiers:
		feeConfig, ok1 := acc(0)
		admin, ok2 := acc(1)
		if !ok1 || !ok2 {
			return DexEvent{}
		}
		o := 8
		feeTiers, ok := readPumpFeesFeeTiersVec(data, &o)
		if !ok {
			return DexEvent{}
		}
		offset, ok := readU8(data, o)
		if !ok || o+1 != len(data) {
			return DexEvent{}
		}
		return DexEvent{Type: EventTypePumpFeesUpsertFeeTiers, Data: &PumpFeesUpsertFeeTiersEvent{
			Metadata: meta, Timestamp: 0, Admin: admin, FeeConfig: feeConfig, FeeTiers: feeTiers, Offset: offset,
		}}
	default:
		return DexEvent{}
	}
}
