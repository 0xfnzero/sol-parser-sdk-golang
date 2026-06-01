package solparser

import (
	"encoding/binary"

	base58lib "github.com/mr-tron/base58/base58"
)

// AccountData 账户数据结构
type AccountData struct {
	Pubkey     string
	Executable bool
	Lamports   uint64
	Owner      string
	RentEpoch  uint64
	Data       []byte
}

// 程序 ID 常量（accounts 包内部使用）
const pumpswapProgramID = PUMPSWAP_PROGRAM_ID
const pumpfunProgramID = PUMPFUN_PROGRAM_ID
const splTokenProgramID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
const splToken2022ProgramID = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"

var accountEventTypes = []EventType{
	EventTypeTokenAccount, EventTypeTokenInfo, EventTypeNonceAccount,
	EventTypeAccountPumpFunGlobal, EventTypeAccountPumpFunBondingCurve,
	EventTypeAccountPumpFunFeeConfig, EventTypeAccountPumpFunSharingConfig,
	EventTypeAccountPumpFunGlobalVolumeAccumulator, EventTypeAccountPumpFunUserVolumeAccumulator,
	EventTypeAccountPumpSwapGlobalConfig, EventTypeAccountPumpSwapPool,
}

// ParseAccountUnified 统一的账户解析入口
// 对齐 Rust `parse_account_unified`
func ParseAccountUnified(account *AccountData, metadata EventMetadata, filter EventTypeFilter) DexEvent {
	if len(account.Data) == 0 {
		return DexEvent{}
	}

	// Early filtering based on event type filter
	if filter != nil {
		shouldParse := false
		for _, t := range accountEventTypes {
			if filter.ShouldInclude(t) {
				shouldParse = true
				break
			}
		}
		if !shouldParse {
			return DexEvent{}
		}
	}

	// PumpSwap 账户解析
	if account.Owner == pumpswapProgramID {
		if filter == nil ||
			filter.ShouldInclude(EventTypeAccountPumpSwapGlobalConfig) ||
			filter.ShouldInclude(EventTypeAccountPumpSwapPool) {
			event := parsePumpswapAccount(account, metadata)
			if event.Type != "" {
				return applyActualEventTypeFilter(event, filter)
			}
		}
		return DexEvent{}
	}

	// PumpFun 账户解析
	if account.Owner == pumpfunProgramID || account.Owner == PUMP_FEES_PROGRAM_ID {
		if filter == nil ||
			filter.ShouldInclude(EventTypeAccountPumpFunGlobal) ||
			filter.ShouldInclude(EventTypeAccountPumpFunBondingCurve) ||
			filter.ShouldInclude(EventTypeAccountPumpFunFeeConfig) ||
			filter.ShouldInclude(EventTypeAccountPumpFunSharingConfig) ||
			filter.ShouldInclude(EventTypeAccountPumpFunGlobalVolumeAccumulator) ||
			filter.ShouldInclude(EventTypeAccountPumpFunUserVolumeAccumulator) {
			event := parsePumpfunAccount(account, metadata)
			if event.Type != "" {
				return applyActualEventTypeFilter(event, filter)
			}
		}
		return DexEvent{}
	}

	// Nonce 账户解析
	if IsNonceAccount(account.Data) {
		if filter != nil && !filter.ShouldInclude(EventTypeNonceAccount) {
			return DexEvent{}
		}
		return ParseNonceAccount(account, metadata)
	}

	// Token 账户解析
	if filter != nil && !filter.ShouldInclude(EventTypeTokenAccount) && !filter.ShouldInclude(EventTypeTokenInfo) {
		return DexEvent{}
	}
	return applyActualEventTypeFilter(ParseTokenAccount(account, metadata), filter)
}

// ParseTokenAccount 解析 Token 账户
// 对齐 Rust `parse_token_account`
func ParseTokenAccount(account *AccountData, metadata EventMetadata) DexEvent {
	if !isTokenProgramOwner(account.Owner) {
		return DexEvent{}
	}

	// 快速路径：尝试解析 Mint 账户
	if len(account.Data) <= 100 {
		event := parseMintFast(account, metadata)
		if event.Type != "" {
			return event
		}
	}

	// 快速路径：尝试解析 Token Account
	event := parseTokenFast(account, metadata)
	if event.Type != "" {
		return event
	}

	return DexEvent{}
}

func isTokenProgramOwner(owner string) bool {
	return owner == splTokenProgramID || owner == splToken2022ProgramID
}

// parseMintFast 快速解析 Mint 账户（零拷贝）
func parseMintFast(account *AccountData, metadata EventMetadata) DexEvent {
	const mintSize = 82
	const supplyOffset = 36
	const decimalsOffset = 44

	if len(account.Data) < mintSize {
		return DexEvent{}
	}

	supply := binary.LittleEndian.Uint64(account.Data[supplyOffset : supplyOffset+8])
	decimals := account.Data[decimalsOffset]

	return DexEvent{
		Type: EventTypeTokenInfo,
		Data: &TokenInfoEvent{
			Metadata:   metadata,
			Pubkey:     account.Pubkey,
			Executable: account.Executable,
			Lamports:   account.Lamports,
			Owner:      account.Owner,
			RentEpoch:  account.RentEpoch,
			Supply:     supply,
			Decimals:   decimals,
		},
	}
}

// parseTokenFast 快速解析 Token Account（零拷贝）
func parseTokenFast(account *AccountData, metadata EventMetadata) DexEvent {
	const tokenAccountSize = 165
	const amountOffset = 64

	if len(account.Data) < tokenAccountSize {
		return DexEvent{}
	}

	amount := binary.LittleEndian.Uint64(account.Data[amountOffset : amountOffset+8])

	return DexEvent{
		Type: EventTypeTokenAccount,
		Data: &TokenAccountEvent{
			Metadata:   metadata,
			Pubkey:     account.Pubkey,
			Executable: account.Executable,
			Lamports:   account.Lamports,
			Owner:      account.Owner,
			RentEpoch:  account.RentEpoch,
			Amount:     amount,
		},
	}
}

// ParseNonceAccount 解析 Nonce 账户
// 对齐 Rust `parse_nonce_account`
func ParseNonceAccount(account *AccountData, metadata EventMetadata) DexEvent {
	const nonceAccountSize = 80
	const authorityOffset = 8
	const nonceOffset = 40

	if len(account.Data) != nonceAccountSize {
		return DexEvent{}
	}

	// Extract authority (32 bytes at offset 8)
	authority := Base58Encode(account.Data[authorityOffset : authorityOffset+32])

	// Extract nonce/blockhash (32 bytes at offset 40)
	nonce := Base58Encode(account.Data[nonceOffset : nonceOffset+32])

	return DexEvent{
		Type: EventTypeNonceAccount,
		Data: &NonceAccountEvent{
			Metadata:   metadata,
			Pubkey:     account.Pubkey,
			Executable: account.Executable,
			Lamports:   account.Lamports,
			Owner:      account.Owner,
			RentEpoch:  account.RentEpoch,
			Nonce:      nonce,
			Authority:  authority,
		},
	}
}

// IsNonceAccount 检测是否为 Nonce 账户
// 对齐 Rust `is_nonce_account`
func IsNonceAccount(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	discriminator := []byte{1, 0, 0, 0, 1, 0, 0, 0}
	for i, b := range discriminator {
		if data[i] != b {
			return false
		}
	}
	return true
}

// ParsePumpfunGlobal 解析 PumpFun Global 账户
// 对齐 Rust `parse_pumpfun_account` 中的 Global 分支。
func ParsePumpfunGlobal(account *AccountData, metadata EventMetadata) DexEvent {
	const globalBody = 1037
	if len(account.Data) < 8+globalBody {
		return DexEvent{}
	}
	globalDisc := []byte{167, 232, 232, 177, 200, 108, 114, 127}
	if !HasDiscriminator(account.Data, globalDisc) {
		return DexEvent{}
	}

	data := account.Data[8:]
	offset := 0

	initialized := data[offset] != 0
	offset++
	authority := ReadPubkey(data, offset)
	offset += 32
	feeRecipient := ReadPubkey(data, offset)
	offset += 32
	initialVirtualTokenReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	initialVirtualSolReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	initialRealTokenReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	tokenTotalSupply := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	feeBasisPoints := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	withdrawAuthority := ReadPubkey(data, offset)
	offset += 32
	enableMigrate := data[offset] != 0
	offset++
	poolMigrationFee := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	creatorFeeBasisPoints := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	feeRecipients := make([]string, 7)
	for i := 0; i < 7; i++ {
		feeRecipients[i] = ReadPubkey(data, offset)
		offset += 32
	}
	setCreatorAuthority := ReadPubkey(data, offset)
	offset += 32
	adminSetCreatorAuthority := ReadPubkey(data, offset)
	offset += 32
	createV2Enabled := data[offset] != 0
	offset++
	whitelistPda := ReadPubkey(data, offset)
	offset += 32
	reservedFeeRecipient := ReadPubkey(data, offset)
	offset += 32
	mayhemModeEnabled := data[offset] != 0
	offset++
	reservedFeeRecipients := make([]string, 7)
	for i := 0; i < 7; i++ {
		reservedFeeRecipients[i] = ReadPubkey(data, offset)
		offset += 32
	}
	isCashbackEnabled := data[offset] != 0
	offset++
	buybackFeeRecipients := make([]string, 8)
	for i := 0; i < 8; i++ {
		buybackFeeRecipients[i] = ReadPubkey(data, offset)
		offset += 32
	}
	buybackBasisPoints := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	initialVirtualQuoteReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	whitelistedQuoteMints := make([]string, 1)
	whitelistedQuoteMints[0] = ReadPubkey(data, offset)
	offset += 32

	return DexEvent{
		Type: EventTypeAccountPumpFunGlobal,
		Data: &PumpFunGlobalAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			Global: PumpFunGlobal{
				Initialized:                 initialized,
				Authority:                   authority,
				FeeRecipient:                feeRecipient,
				InitialVirtualTokenReserves: initialVirtualTokenReserves,
				InitialVirtualSolReserves:   initialVirtualSolReserves,
				InitialRealTokenReserves:    initialRealTokenReserves,
				TokenTotalSupply:            tokenTotalSupply,
				FeeBasisPoints:              feeBasisPoints,
				WithdrawAuthority:           withdrawAuthority,
				EnableMigrate:               enableMigrate,
				PoolMigrationFee:            poolMigrationFee,
				CreatorFeeBasisPoints:       creatorFeeBasisPoints,
				FeeRecipients:               feeRecipients,
				SetCreatorAuthority:         setCreatorAuthority,
				AdminSetCreatorAuthority:    adminSetCreatorAuthority,
				CreateV2Enabled:             createV2Enabled,
				WhitelistPda:                whitelistPda,
				ReservedFeeRecipient:        reservedFeeRecipient,
				MayhemModeEnabled:           mayhemModeEnabled,
				ReservedFeeRecipients:       reservedFeeRecipients,
				IsCashbackEnabled:           isCashbackEnabled,
				BuybackFeeRecipients:        buybackFeeRecipients,
				BuybackBasisPoints:          buybackBasisPoints,
				InitialVirtualQuoteReserves: initialVirtualQuoteReserves,
				WhitelistedQuoteMints:       whitelistedQuoteMints,
			},
		},
	}
}

// ParsePumpfunBondingCurve 解析 PumpFun BondingCurve 账户。
func ParsePumpfunBondingCurve(account *AccountData, metadata EventMetadata) DexEvent {
	const bondingCurveBody = 107
	if len(account.Data) < 8+bondingCurveBody {
		return DexEvent{}
	}
	bondingCurveDisc := []byte{23, 183, 248, 55, 96, 216, 172, 96}
	if !HasDiscriminator(account.Data, bondingCurveDisc) {
		return DexEvent{}
	}

	data := account.Data[8:]
	offset := 0
	virtualTokenReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	virtualQuoteReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	realTokenReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	realQuoteReserves := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	tokenTotalSupply := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	complete := data[offset] != 0
	offset++
	creator := ReadPubkey(data, offset)
	offset += 32
	isMayhemMode := data[offset] != 0
	offset++
	isCashbackCoin := data[offset] != 0
	offset++
	quoteMint := ReadPubkey(data, offset)

	return DexEvent{
		Type: EventTypeAccountPumpFunBondingCurve,
		Data: &PumpFunBondingCurveAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			BondingCurve: PumpFunBondingCurve{
				VirtualTokenReserves: virtualTokenReserves,
				VirtualQuoteReserves: virtualQuoteReserves,
				RealTokenReserves:    realTokenReserves,
				RealQuoteReserves:    realQuoteReserves,
				TokenTotalSupply:     tokenTotalSupply,
				Complete:             complete,
				Creator:              creator,
				IsMayhemMode:         isMayhemMode,
				IsCashbackCoin:       isCashbackCoin,
				QuoteMint:            quoteMint,
			},
		},
	}
}

const maxPumpfunFeeTiers = 64
const maxPumpfunShareholders = 64

func readPumpfunFees(data []byte, offset *int) (PumpFeesFees, bool) {
	if *offset+24 > len(data) {
		return PumpFeesFees{}, false
	}
	fees := PumpFeesFees{
		LpFeeBps:       binary.LittleEndian.Uint64(data[*offset : *offset+8]),
		ProtocolFeeBps: binary.LittleEndian.Uint64(data[*offset+8 : *offset+16]),
		CreatorFeeBps:  binary.LittleEndian.Uint64(data[*offset+16 : *offset+24]),
	}
	*offset += 24
	return fees, true
}

func readPumpfunFeeTiers(data []byte, offset *int) ([]PumpFeesFeeTier, bool) {
	if *offset+4 > len(data) {
		return nil, false
	}
	n := int(binary.LittleEndian.Uint32(data[*offset : *offset+4]))
	*offset += 4
	if n > maxPumpfunFeeTiers || *offset+n*40 > len(data) {
		return nil, false
	}
	out := make([]PumpFeesFeeTier, 0, n)
	for i := 0; i < n; i++ {
		var raw [16]byte
		copy(raw[:], data[*offset:*offset+16])
		*offset += 16
		fees, ok := readPumpfunFees(data, offset)
		if !ok {
			return nil, false
		}
		out = append(out, PumpFeesFeeTier{
			MarketCapLamportsThreshold: u128LEDecimalString(raw),
			Fees:                       fees,
		})
	}
	return out, true
}

func readPumpfunShareholders(data []byte, offset *int) ([]PumpFeesShareholder, bool) {
	if *offset+4 > len(data) {
		return nil, false
	}
	n := int(binary.LittleEndian.Uint32(data[*offset : *offset+4]))
	*offset += 4
	if n > maxPumpfunShareholders || *offset+n*34 > len(data) {
		return nil, false
	}
	out := make([]PumpFeesShareholder, 0, n)
	for i := 0; i < n; i++ {
		address := ReadPubkey(data, *offset)
		*offset += 32
		shareBps := binary.LittleEndian.Uint16(data[*offset : *offset+2])
		*offset += 2
		out = append(out, PumpFeesShareholder{Address: address, ShareBps: shareBps})
	}
	return out, true
}

func ParsePumpfunFeeConfig(account *AccountData, metadata EventMetadata) DexEvent {
	disc := []byte{143, 52, 146, 187, 219, 123, 76, 155}
	if len(account.Data) < 8+1+32+24+4+4 || !HasDiscriminator(account.Data, disc) {
		return DexEvent{}
	}
	data := account.Data[8:]
	offset := 0
	bump := data[offset]
	offset++
	admin := ReadPubkey(data, offset)
	offset += 32
	flatFees, ok := readPumpfunFees(data, &offset)
	if !ok {
		return DexEvent{}
	}
	feeTiers, ok := readPumpfunFeeTiers(data, &offset)
	if !ok {
		return DexEvent{}
	}
	stableFeeTiers, ok := readPumpfunFeeTiers(data, &offset)
	if !ok {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeAccountPumpFunFeeConfig,
		Data: &PumpFunFeeConfigAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			FeeConfig: PumpFunFeeConfig{
				Bump:           bump,
				Admin:          admin,
				FlatFees:       flatFees,
				FeeTiers:       feeTiers,
				StableFeeTiers: stableFeeTiers,
			},
		},
	}
}

func ParsePumpfunSharingConfig(account *AccountData, metadata EventMetadata) DexEvent {
	disc := []byte{216, 74, 9, 0, 56, 140, 93, 75}
	if len(account.Data) < 8+1+1+1+32+32+1+4 || !HasDiscriminator(account.Data, disc) {
		return DexEvent{}
	}
	data := account.Data[8:]
	offset := 0
	bump := data[offset]
	offset++
	version := data[offset]
	offset++
	statusRaw := data[offset]
	offset++
	var status PumpFeesConfigStatus
	switch statusRaw {
	case 0:
		status = PumpFeesConfigStatusPaused
	case 1:
		status = PumpFeesConfigStatusActive
	default:
		return DexEvent{}
	}
	mint := ReadPubkey(data, offset)
	offset += 32
	admin := ReadPubkey(data, offset)
	offset += 32
	adminRevoked := data[offset] != 0
	offset++
	shareholders, ok := readPumpfunShareholders(data, &offset)
	if !ok {
		return DexEvent{}
	}
	return DexEvent{
		Type: EventTypeAccountPumpFunSharingConfig,
		Data: &PumpFunSharingConfigAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			SharingConfig: PumpFunSharingConfig{
				Bump:         bump,
				Version:      version,
				Status:       status,
				Mint:         mint,
				Admin:        admin,
				AdminRevoked: adminRevoked,
				Shareholders: shareholders,
			},
		},
	}
}

func ParsePumpfunGlobalVolumeAccumulator(account *AccountData, metadata EventMetadata) DexEvent {
	disc := []byte{202, 42, 246, 43, 142, 190, 30, 255}
	const body = 536
	if len(account.Data) < 8+body || !HasDiscriminator(account.Data, disc) {
		return DexEvent{}
	}
	data := account.Data[8:]
	offset := 0
	startTime := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	endTime := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	secondsInADay := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	mint := ReadPubkey(data, offset)
	offset += 32
	totalTokenSupply := make([]uint64, 30)
	for i := range totalTokenSupply {
		totalTokenSupply[i] = binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	solVolumes := make([]uint64, 30)
	for i := range solVolumes {
		solVolumes[i] = binary.LittleEndian.Uint64(data[offset : offset+8])
		offset += 8
	}
	return DexEvent{
		Type: EventTypeAccountPumpFunGlobalVolumeAccumulator,
		Data: &PumpFunGlobalVolumeAccumulatorAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			GlobalVolumeAccumulator: PumpFunGlobalVolumeAccumulator{
				StartTime:        startTime,
				EndTime:          endTime,
				SecondsInADay:    secondsInADay,
				Mint:             mint,
				TotalTokenSupply: totalTokenSupply,
				SolVolumes:       solVolumes,
			},
		},
	}
}

func ParsePumpfunUserVolumeAccumulator(account *AccountData, metadata EventMetadata) DexEvent {
	disc := []byte{86, 255, 112, 14, 102, 53, 154, 250}
	const body = 98
	if len(account.Data) < 8+body || !HasDiscriminator(account.Data, disc) {
		return DexEvent{}
	}
	data := account.Data[8:]
	offset := 0
	user := ReadPubkey(data, offset)
	offset += 32
	needsClaim := data[offset] != 0
	offset++
	totalUnclaimedTokens := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	totalClaimedTokens := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	currentSolVolume := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	lastUpdateTimestamp := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	hasTotalClaimedTokens := data[offset] != 0
	offset++
	cashbackEarned := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	totalCashbackClaimed := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	stableCashbackEarned := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8
	totalStableCashbackClaimed := binary.LittleEndian.Uint64(data[offset : offset+8])
	return DexEvent{
		Type: EventTypeAccountPumpFunUserVolumeAccumulator,
		Data: &PumpFunUserVolumeAccumulatorAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			UserVolumeAccumulator: PumpFunUserVolumeAccumulator{
				User:                       user,
				NeedsClaim:                 needsClaim,
				TotalUnclaimedTokens:       totalUnclaimedTokens,
				TotalClaimedTokens:         totalClaimedTokens,
				CurrentSolVolume:           currentSolVolume,
				LastUpdateTimestamp:        lastUpdateTimestamp,
				HasTotalClaimedTokens:      hasTotalClaimedTokens,
				CashbackEarned:             cashbackEarned,
				TotalCashbackClaimed:       totalCashbackClaimed,
				StableCashbackEarned:       stableCashbackEarned,
				TotalStableCashbackClaimed: totalStableCashbackClaimed,
			},
		},
	}
}

// ParsePumpswapGlobalConfig 解析 PumpSwap Global Config 账户
// 对齐 Rust `parse_pumpswap_global_config`
func ParsePumpswapGlobalConfig(account *AccountData, metadata EventMetadata) DexEvent {
	const globalConfigSize = 32 + 8 + 8 + 1 + 32*8 + 8 + 32

	if len(account.Data) < globalConfigSize+8 {
		return DexEvent{}
	}

	// Check discriminator
	globalConfigDisc := []byte{149, 8, 156, 202, 160, 252, 176, 217}
	if !HasDiscriminator(account.Data, globalConfigDisc) {
		return DexEvent{}
	}

	data := account.Data[8:]
	offset := 0

	admin := ReadPubkey(data, offset)
	offset += 32

	lpFeeBasisPoints := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	protocolFeeBasisPoints := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	disableFlags := data[offset]
	offset++

	// Read 8 protocol_fee_recipients
	protocolFeeRecipients := make([]string, 8)
	for i := 0; i < 8; i++ {
		protocolFeeRecipients[i] = ReadPubkey(data, offset)
		offset += 32
	}

	coinCreatorFeeBasisPoints := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	adminSetCoinCreatorAuthority := ReadPubkey(data, offset)
	offset += 32

	whitelistPda := ReadPubkey(data, offset)
	offset += 32

	reservedFeeRecipient := ReadPubkey(data, offset)
	offset += 32

	mayhemModeEnabled := data[offset] != 0
	offset++

	// Read 7 reserved_fee_recipients
	reservedFeeRecipients := make([]string, 7)
	for i := 0; i < 7; i++ {
		reservedFeeRecipients[i] = ReadPubkey(data, offset)
		offset += 32
	}

	return DexEvent{
		Type: EventTypeAccountPumpSwapGlobalConfig,
		Data: &PumpSwapGlobalConfigAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			Config: PumpSwapGlobalConfigAccountData{
				Admin:                        admin,
				LpFeeBasisPoints:             lpFeeBasisPoints,
				ProtocolFeeBasisPoints:       protocolFeeBasisPoints,
				DisableFlags:                 disableFlags,
				ProtocolFeeRecipients:        protocolFeeRecipients,
				CoinCreatorFeeBasisPoints:    coinCreatorFeeBasisPoints,
				AdminSetCoinCreatorAuthority: adminSetCoinCreatorAuthority,
				WhitelistPda:                 whitelistPda,
				ReservedFeeRecipient:         reservedFeeRecipient,
				MayhemModeEnabled:            mayhemModeEnabled,
				ReservedFeeRecipients:        reservedFeeRecipients,
			},
		},
	}
}

// ParsePumpswapPool 解析 PumpSwap Pool 账户
// 对齐 Rust `parse_pumpswap_pool`
// 结构体布局（按顺序）：
// - pool_bump: u8 (1 byte)
// - index: u16 (2 bytes)
// - creator: pubkey (32 bytes)
// - base_mint: pubkey (32 bytes)
// - quote_mint: pubkey (32 bytes)
// - lp_mint: pubkey (32 bytes)
// - pool_base_token_account: pubkey (32 bytes)
// - pool_quote_token_account: pubkey (32 bytes)
// - lp_supply: u64 (8 bytes)
// - coin_creator: pubkey (32 bytes)
// - is_mayhem_mode: bool (1 byte)
// - is_cashback_coin: bool (1 byte)
func ParsePumpswapPool(account *AccountData, metadata EventMetadata) DexEvent {
	const poolBody = 244

	if len(account.Data) < 8+poolBody {
		return DexEvent{}
	}

	// Check discriminator
	poolDisc := []byte{241, 154, 109, 4, 17, 177, 109, 188}
	if !HasDiscriminator(account.Data, poolDisc) {
		return DexEvent{}
	}

	data := account.Data[8:]
	offset := 0

	poolBump := data[offset]
	offset++

	index := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	// creator field (was missing in original implementation)
	creator := ReadPubkey(data, offset)
	offset += 32

	baseMint := ReadPubkey(data, offset)
	offset += 32
	quoteMint := ReadPubkey(data, offset)
	offset += 32
	lpMint := ReadPubkey(data, offset)
	offset += 32
	poolBaseTokenAccount := ReadPubkey(data, offset)
	offset += 32
	poolQuoteTokenAccount := ReadPubkey(data, offset)
	offset += 32

	lpSupply := binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	coinCreator := ReadPubkey(data, offset)
	offset += 32

	isMayhemMode := data[offset] != 0
	offset++

	isCashbackCoin := data[offset] != 0

	return DexEvent{
		Type: EventTypeAccountPumpSwapPool,
		Data: &PumpSwapPoolAccountEvent{
			Metadata: metadata,
			Pubkey:   account.Pubkey,
			Pool: PumpSwapPoolAccountData{
				PoolBump:              poolBump,
				Index:                 index,
				Creator:               creator,
				BaseMint:              baseMint,
				QuoteMint:             quoteMint,
				LpMint:                lpMint,
				PoolBaseTokenAccount:  poolBaseTokenAccount,
				PoolQuoteTokenAccount: poolQuoteTokenAccount,
				LpSupply:              lpSupply,
				CoinCreator:           coinCreator,
				IsMayhemMode:          isMayhemMode,
				IsCashbackCoin:        isCashbackCoin,
			},
		},
	}
}

// parsePumpswapAccount 解析 PumpSwap 账户（内部函数）
func parsePumpswapAccount(account *AccountData, metadata EventMetadata) DexEvent {
	// Check Global Config discriminator
	globalConfigDisc := []byte{149, 8, 156, 202, 160, 252, 176, 217}
	if HasDiscriminator(account.Data, globalConfigDisc) {
		return ParsePumpswapGlobalConfig(account, metadata)
	}

	// Check Pool discriminator
	poolDisc := []byte{241, 154, 109, 4, 17, 177, 109, 188}
	if HasDiscriminator(account.Data, poolDisc) {
		return ParsePumpswapPool(account, metadata)
	}

	return DexEvent{}
}

// parsePumpfunAccount 解析 PumpFun 账户（内部函数）
func parsePumpfunAccount(account *AccountData, metadata EventMetadata) DexEvent {
	feeConfigDisc := []byte{143, 52, 146, 187, 219, 123, 76, 155}
	if HasDiscriminator(account.Data, feeConfigDisc) {
		return ParsePumpfunFeeConfig(account, metadata)
	}
	sharingConfigDisc := []byte{216, 74, 9, 0, 56, 140, 93, 75}
	if HasDiscriminator(account.Data, sharingConfigDisc) {
		return ParsePumpfunSharingConfig(account, metadata)
	}
	globalVolumeAccumulatorDisc := []byte{202, 42, 246, 43, 142, 190, 30, 255}
	if HasDiscriminator(account.Data, globalVolumeAccumulatorDisc) {
		return ParsePumpfunGlobalVolumeAccumulator(account, metadata)
	}
	userVolumeAccumulatorDisc := []byte{86, 255, 112, 14, 102, 53, 154, 250}
	if HasDiscriminator(account.Data, userVolumeAccumulatorDisc) {
		return ParsePumpfunUserVolumeAccumulator(account, metadata)
	}
	if IsPumpfunBondingCurveAccount(account.Data) {
		return ParsePumpfunBondingCurve(account, metadata)
	}
	if IsPumpfunGlobalAccount(account.Data) {
		return ParsePumpfunGlobal(account, metadata)
	}
	return DexEvent{}
}

// IsPumpfunGlobalAccount 检查是否为 PumpFun Global 账户
func IsPumpfunGlobalAccount(data []byte) bool {
	globalDisc := []byte{167, 232, 232, 177, 200, 108, 114, 127}
	return HasDiscriminator(data, globalDisc)
}

// IsPumpfunBondingCurveAccount 检查是否为 PumpFun BondingCurve 账户
func IsPumpfunBondingCurveAccount(data []byte) bool {
	bondingCurveDisc := []byte{23, 183, 248, 55, 96, 216, 172, 96}
	return HasDiscriminator(data, bondingCurveDisc)
}

// IsGlobalConfigAccount 检查是否为 Global Config 账户
func IsGlobalConfigAccount(data []byte) bool {
	globalConfigDisc := []byte{149, 8, 156, 202, 160, 252, 176, 217}
	return HasDiscriminator(data, globalConfigDisc)
}

// IsPoolAccount 检查是否为 Pool 账户
func IsPoolAccount(data []byte) bool {
	poolDisc := []byte{241, 154, 109, 4, 17, 177, 109, 188}
	return HasDiscriminator(data, poolDisc)
}

// HasDiscriminator 检查是否有指定的 discriminator
func HasDiscriminator(data []byte, discriminator []byte) bool {
	if len(data) < len(discriminator) {
		return false
	}
	for i, b := range discriminator {
		if data[i] != b {
			return false
		}
	}
	return true
}

// Base58Encode 将字节编码为 Base58 字符串
func Base58Encode(data []byte) string {
	return base58lib.Encode(data)
}

// ReadPubkey 从字节数组读取公钥（32字节），返回 Base58 编码字符串
func ReadPubkey(data []byte, offset int) string {
	if offset+32 > len(data) {
		return zeroPubkey
	}
	return Base58Encode(data[offset : offset+32])
}

// ReadU64Le 读取小端序 uint64
func ReadU64Le(data []byte, offset int) uint64 {
	if offset+8 > len(data) {
		return 0
	}
	return binary.LittleEndian.Uint64(data[offset : offset+8])
}

// ReadU8 读取 uint8
func ReadU8(data []byte, offset int) uint8 {
	if offset >= len(data) {
		return 0
	}
	return data[offset]
}
