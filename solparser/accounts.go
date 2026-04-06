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
const pumpswapProgramID = "pAMMBay6oceH9fJKBRdGP4LmT4saRGfEE7xmrCaGWpZ"

// ParseAccountUnified 统一的账户解析入口
// 对齐 Rust `parse_account_unified`
func ParseAccountUnified(account *AccountData, metadata EventMetadata, filter EventTypeFilter) DexEvent {
	if len(account.Data) == 0 {
		return DexEvent{}
	}

	// Early filtering based on event type filter
	accountTypes := []EventType{
		EventTypeTokenAccount, EventTypeTokenInfo, EventTypeNonceAccount,
		EventTypeAccountPumpSwapGlobalConfig, EventTypeAccountPumpSwapPool,
	}
	shouldParse := false
	for _, t := range accountTypes {
		if filter.ShouldInclude(t) {
			shouldParse = true
			break
		}
	}
	if !shouldParse {
		return DexEvent{}
	}

	// PumpSwap 账户解析
	if account.Owner == pumpswapProgramID {
		if filter.ShouldInclude(EventTypeAccountPumpSwapGlobalConfig) ||
			filter.ShouldInclude(EventTypeAccountPumpSwapPool) {
			event := parsePumpswapAccount(account, metadata)
			if event.Type != "" {
				return event
			}
		}
	}

	// Nonce 账户解析
	if IsNonceAccount(account.Data) {
		if !filter.ShouldInclude(EventTypeNonceAccount) {
			return DexEvent{}
		}
		return ParseNonceAccount(account, metadata)
	}

	// Token 账户解析
	if !filter.ShouldInclude(EventTypeTokenAccount) && !filter.ShouldInclude(EventTypeTokenInfo) {
		return DexEvent{}
	}
	return ParseTokenAccount(account, metadata)
}

// ParseTokenAccount 解析 Token 账户
// 对齐 Rust `parse_token_account`
func ParseTokenAccount(account *AccountData, metadata EventMetadata) DexEvent {
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
				Admin:                         admin,
				LpFeeBasisPoints:              lpFeeBasisPoints,
				ProtocolFeeBasisPoints:        protocolFeeBasisPoints,
				DisableFlags:                  disableFlags,
				ProtocolFeeRecipients:         protocolFeeRecipients,
				CoinCreatorFeeBasisPoints:     coinCreatorFeeBasisPoints,
				AdminSetCoinCreatorAuthority:  adminSetCoinCreatorAuthority,
				WhitelistPda:                  whitelistPda,
				ReservedFeeRecipient:          reservedFeeRecipient,
				MayhemModeEnabled:             mayhemModeEnabled,
				ReservedFeeRecipients:         reservedFeeRecipients,
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
