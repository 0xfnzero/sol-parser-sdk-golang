package solparser

// ============================================================
// 基础类型定义
// ============================================================

// 注意: EventMetadata 已在 metadata.go 中定义

// DexEventInterface 事件接口 - 所有事件类型都需要实现此接口
type DexEventInterface interface {
	EventType() EventType
	GetMetadata() EventMetadata
}

// ============================================================
// PumpFun 事件结构体
// ============================================================

// PumpFunTradeEvent PumpFun 交易事件
type PumpFunTradeEvent struct {
	Metadata                           EventMetadata         `json:"metadata"`
	Mint                               string                `json:"mint"`
	SolAmount                          uint64                `json:"sol_amount"`
	TokenAmount                        uint64                `json:"token_amount"`
	IsBuy                              bool                  `json:"is_buy"`
	IsCreatedBuy                       bool                  `json:"is_created_buy"`
	User                               string                `json:"user"`
	Timestamp                          int64                 `json:"timestamp"`
	VirtualSolReserves                 uint64                `json:"virtual_sol_reserves"`
	VirtualTokenReserves               uint64                `json:"virtual_token_reserves"`
	RealSolReserves                    uint64                `json:"real_sol_reserves"`
	RealTokenReserves                  uint64                `json:"real_token_reserves"`
	FeeRecipient                       string                `json:"fee_recipient"`
	FeeBasisPoints                     uint64                `json:"fee_basis_points"`
	Fee                                uint64                `json:"fee"`
	Creator                            string                `json:"creator"`
	CreatorFeeBasisPoints              uint64                `json:"creator_fee_basis_points"`
	CreatorFee                         uint64                `json:"creator_fee"`
	TrackVolume                        bool                  `json:"track_volume"`
	TotalUnclaimedTokens               uint64                `json:"total_unclaimed_tokens"`
	TotalClaimedTokens                 uint64                `json:"total_claimed_tokens"`
	CurrentSolVolume                   uint64                `json:"current_sol_volume"`
	LastUpdateTimestamp                int64                 `json:"last_update_timestamp"`
	IxName                             string                `json:"ix_name"`
	MayhemMode                         bool                  `json:"mayhem_mode"`
	CashbackFeeBasisPoints             uint64                `json:"cashback_fee_basis_points"`
	Cashback                           uint64                `json:"cashback"`
	BuybackFeeBasisPoints              uint64                `json:"buyback_fee_basis_points"`
	BuybackFee                         uint64                `json:"buyback_fee"`
	Shareholders                       []PumpFeesShareholder `json:"shareholders"`
	QuoteMint                          string                `json:"quote_mint"`
	QuoteAmount                        uint64                `json:"quote_amount"`
	VirtualQuoteReserves               uint64                `json:"virtual_quote_reserves"`
	RealQuoteReserves                  uint64                `json:"real_quote_reserves"`
	IsCashbackCoin                     bool                  `json:"is_cashback_coin"`
	Amount                             uint64                `json:"amount"`
	MaxSolCost                         uint64                `json:"max_sol_cost"`
	MinSolOutput                       uint64                `json:"min_sol_output"`
	SpendableSolIn                     uint64                `json:"spendable_sol_in"`
	SpendableQuoteIn                   uint64                `json:"spendable_quote_in"`
	MinTokensOut                       uint64                `json:"min_tokens_out"`
	Global                             string                `json:"global"`
	BondingCurve                       string                `json:"bonding_curve"`
	BondingCurveV2                     string                `json:"bonding_curve_v2"`
	AssociatedBondingCurve             string                `json:"associated_bonding_curve"`
	AssociatedUser                     string                `json:"associated_user"`
	SystemProgram                      string                `json:"system_program"`
	TokenProgram                       string                `json:"token_program"`
	QuoteTokenProgram                  string                `json:"quote_token_program"`
	AssociatedTokenProgram             string                `json:"associated_token_program"`
	CreatorVault                       string                `json:"creator_vault"`
	AssociatedQuoteFeeRecipient        string                `json:"associated_quote_fee_recipient"`
	BuybackFeeRecipient                string                `json:"buyback_fee_recipient"`
	AssociatedQuoteBuybackFeeRecipient string                `json:"associated_quote_buyback_fee_recipient"`
	AssociatedQuoteBondingCurve        string                `json:"associated_quote_bonding_curve"`
	AssociatedQuoteUser                string                `json:"associated_quote_user"`
	AssociatedCreatorVault             string                `json:"associated_creator_vault"`
	SharingConfig                      string                `json:"sharing_config"`
	EventAuthority                     string                `json:"event_authority"`
	Program                            string                `json:"program"`
	GlobalVolumeAccumulator            string                `json:"global_volume_accumulator"`
	UserVolumeAccumulator              string                `json:"user_volume_accumulator"`
	AssociatedUserVolumeAccumulator    string                `json:"associated_user_volume_accumulator"`
	FeeConfig                          string                `json:"fee_config"`
	FeeProgram                         string                `json:"fee_program"`
	Account                            string                `json:"account,omitempty"`
}

func (e *PumpFunTradeEvent) EventType() EventType       { return EventTypePumpFunTrade }
func (e *PumpFunTradeEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpFunCreateEvent PumpFun 创建代币事件
type PumpFunCreateEvent struct {
	Metadata               EventMetadata `json:"metadata"`
	Name                   string        `json:"name"`
	Symbol                 string        `json:"symbol"`
	Uri                    string        `json:"uri"`
	Mint                   string        `json:"mint"`
	BondingCurve           string        `json:"bonding_curve"`
	User                   string        `json:"user"`
	Creator                string        `json:"creator"`
	Timestamp              int64         `json:"timestamp"`
	VirtualTokenReserves   uint64        `json:"virtual_token_reserves"`
	VirtualSolReserves     uint64        `json:"virtual_sol_reserves"`
	RealTokenReserves      uint64        `json:"real_token_reserves"`
	TokenTotalSupply       uint64        `json:"token_total_supply"`
	TokenProgram           string        `json:"token_program"`
	IsMayhemMode           bool          `json:"is_mayhem_mode"`
	IsCashbackEnabled      bool          `json:"is_cashback_enabled"`
	QuoteMint              string        `json:"quote_mint"`
	QuoteVault             string        `json:"quote_vault"`
	QuoteTokenProgram      string        `json:"quote_token_program"`
	VirtualQuoteReserves   uint64        `json:"virtual_quote_reserves"`
	IxName                 string        `json:"ix_name"`
	MintAuthority          string        `json:"mint_authority"`
	AssociatedBondingCurve string        `json:"associated_bonding_curve"`
	Global                 string        `json:"global"`
	SystemProgram          string        `json:"system_program"`
	AssociatedTokenProgram string        `json:"associated_token_program"`
	MayhemProgramID        string        `json:"mayhem_program_id"`
	GlobalParams           string        `json:"global_params"`
	SolVault               string        `json:"sol_vault"`
	MayhemState            string        `json:"mayhem_state"`
	MayhemTokenVault       string        `json:"mayhem_token_vault"`
	EventAuthority         string        `json:"event_authority"`
	Program                string        `json:"program"`
	ObservedFeeRecipient   string        `json:"observed_fee_recipient"`
}

func (e *PumpFunCreateEvent) EventType() EventType       { return EventTypePumpFunCreate }
func (e *PumpFunCreateEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpFunCreateV2TokenEvent PumpFun create_v2（SPL-22 / Mayhem），与 Rust `PumpFunCreateV2TokenEvent` 对齐
type PumpFunCreateV2TokenEvent struct {
	Metadata               EventMetadata `json:"metadata"`
	Name                   string        `json:"name"`
	Symbol                 string        `json:"symbol"`
	Uri                    string        `json:"uri"`
	Mint                   string        `json:"mint"`
	BondingCurve           string        `json:"bonding_curve"`
	User                   string        `json:"user"`
	Creator                string        `json:"creator"`
	Timestamp              int64         `json:"timestamp"`
	VirtualTokenReserves   uint64        `json:"virtual_token_reserves"`
	VirtualSolReserves     uint64        `json:"virtual_sol_reserves"`
	RealTokenReserves      uint64        `json:"real_token_reserves"`
	TokenTotalSupply       uint64        `json:"token_total_supply"`
	TokenProgram           string        `json:"token_program"`
	IsMayhemMode           bool          `json:"is_mayhem_mode"`
	IsCashbackEnabled      bool          `json:"is_cashback_enabled"`
	QuoteMint              string        `json:"quote_mint"`
	QuoteVault             string        `json:"quote_vault"`
	QuoteTokenProgram      string        `json:"quote_token_program"`
	VirtualQuoteReserves   uint64        `json:"virtual_quote_reserves"`
	IxName                 string        `json:"ix_name"`
	MintAuthority          string        `json:"mint_authority"`
	AssociatedBondingCurve string        `json:"associated_bonding_curve"`
	Global                 string        `json:"global"`
	SystemProgram          string        `json:"system_program"`
	AssociatedTokenProgram string        `json:"associated_token_program"`
	MayhemProgramID        string        `json:"mayhem_program_id"`
	GlobalParams           string        `json:"global_params"`
	SolVault               string        `json:"sol_vault"`
	MayhemState            string        `json:"mayhem_state"`
	MayhemTokenVault       string        `json:"mayhem_token_vault"`
	EventAuthority         string        `json:"event_authority"`
	Program                string        `json:"program"`
	ObservedFeeRecipient   string        `json:"observed_fee_recipient"`
}

func (e *PumpFunCreateV2TokenEvent) EventType() EventType       { return EventTypePumpFunCreateV2 }
func (e *PumpFunCreateV2TokenEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpFunMigrateEvent PumpFun 迁移事件
type PumpFunMigrateEvent struct {
	Metadata         EventMetadata `json:"metadata"`
	User             string        `json:"user"`
	Mint             string        `json:"mint"`
	MintAmount       uint64        `json:"mint_amount"`
	SolAmount        uint64        `json:"sol_amount"`
	PoolMigrationFee uint64        `json:"pool_migration_fee"`
	BondingCurve     string        `json:"bonding_curve"`
	Timestamp        int64         `json:"timestamp"`
	Pool             string        `json:"pool"`
}

func (e *PumpFunMigrateEvent) EventType() EventType       { return EventTypePumpFunMigrate }
func (e *PumpFunMigrateEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunGlobalAccountEvent struct {
	Metadata EventMetadata `json:"metadata"`
	Pubkey   string        `json:"pubkey"`
	Global   PumpFunGlobal `json:"global"`
}

func (e *PumpFunGlobalAccountEvent) EventType() EventType {
	return EventTypeAccountPumpFunGlobal
}
func (e *PumpFunGlobalAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunBondingCurveAccountEvent struct {
	Metadata     EventMetadata       `json:"metadata"`
	Pubkey       string              `json:"pubkey"`
	BondingCurve PumpFunBondingCurve `json:"bonding_curve"`
}

func (e *PumpFunBondingCurveAccountEvent) EventType() EventType {
	return EventTypeAccountPumpFunBondingCurve
}
func (e *PumpFunBondingCurveAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunBondingCurve struct {
	VirtualTokenReserves uint64 `json:"virtual_token_reserves"`
	VirtualQuoteReserves uint64 `json:"virtual_quote_reserves"`
	RealTokenReserves    uint64 `json:"real_token_reserves"`
	RealQuoteReserves    uint64 `json:"real_quote_reserves"`
	TokenTotalSupply     uint64 `json:"token_total_supply"`
	Complete             bool   `json:"complete"`
	Creator              string `json:"creator"`
	IsMayhemMode         bool   `json:"is_mayhem_mode"`
	IsCashbackCoin       bool   `json:"is_cashback_coin"`
	QuoteMint            string `json:"quote_mint"`
}

type PumpFunGlobal struct {
	Initialized                 bool     `json:"initialized"`
	Authority                   string   `json:"authority"`
	FeeRecipient                string   `json:"fee_recipient"`
	InitialVirtualTokenReserves uint64   `json:"initial_virtual_token_reserves"`
	InitialVirtualSolReserves   uint64   `json:"initial_virtual_sol_reserves"`
	InitialRealTokenReserves    uint64   `json:"initial_real_token_reserves"`
	TokenTotalSupply            uint64   `json:"token_total_supply"`
	FeeBasisPoints              uint64   `json:"fee_basis_points"`
	WithdrawAuthority           string   `json:"withdraw_authority"`
	EnableMigrate               bool     `json:"enable_migrate"`
	PoolMigrationFee            uint64   `json:"pool_migration_fee"`
	CreatorFeeBasisPoints       uint64   `json:"creator_fee_basis_points"`
	FeeRecipients               []string `json:"fee_recipients"`
	SetCreatorAuthority         string   `json:"set_creator_authority"`
	AdminSetCreatorAuthority    string   `json:"admin_set_creator_authority"`
	CreateV2Enabled             bool     `json:"create_v2_enabled"`
	WhitelistPda                string   `json:"whitelist_pda"`
	ReservedFeeRecipient        string   `json:"reserved_fee_recipient"`
	MayhemModeEnabled           bool     `json:"mayhem_mode_enabled"`
	ReservedFeeRecipients       []string `json:"reserved_fee_recipients"`
	IsCashbackEnabled           bool     `json:"is_cashback_enabled"`
	BuybackFeeRecipients        []string `json:"buyback_fee_recipients"`
	BuybackBasisPoints          uint64   `json:"buyback_basis_points"`
	InitialVirtualQuoteReserves uint64   `json:"initial_virtual_quote_reserves"`
	WhitelistedQuoteMints       []string `json:"whitelisted_quote_mints"`
}

type PumpFunFeeConfigAccountEvent struct {
	Metadata  EventMetadata    `json:"metadata"`
	Pubkey    string           `json:"pubkey"`
	FeeConfig PumpFunFeeConfig `json:"fee_config"`
}

func (e *PumpFunFeeConfigAccountEvent) EventType() EventType {
	return EventTypeAccountPumpFunFeeConfig
}
func (e *PumpFunFeeConfigAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunFeeConfig struct {
	Bump           uint8             `json:"bump"`
	Admin          string            `json:"admin"`
	FlatFees       PumpFeesFees      `json:"flat_fees"`
	FeeTiers       []PumpFeesFeeTier `json:"fee_tiers"`
	StableFeeTiers []PumpFeesFeeTier `json:"stable_fee_tiers"`
}

type PumpFunSharingConfigAccountEvent struct {
	Metadata      EventMetadata        `json:"metadata"`
	Pubkey        string               `json:"pubkey"`
	SharingConfig PumpFunSharingConfig `json:"sharing_config"`
}

func (e *PumpFunSharingConfigAccountEvent) EventType() EventType {
	return EventTypeAccountPumpFunSharingConfig
}
func (e *PumpFunSharingConfigAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunSharingConfig struct {
	Bump         uint8                 `json:"bump"`
	Version      uint8                 `json:"version"`
	Status       PumpFeesConfigStatus  `json:"status"`
	Mint         string                `json:"mint"`
	Admin        string                `json:"admin"`
	AdminRevoked bool                  `json:"admin_revoked"`
	Shareholders []PumpFeesShareholder `json:"shareholders"`
}

type PumpFunGlobalVolumeAccumulatorAccountEvent struct {
	Metadata                EventMetadata                  `json:"metadata"`
	Pubkey                  string                         `json:"pubkey"`
	GlobalVolumeAccumulator PumpFunGlobalVolumeAccumulator `json:"global_volume_accumulator"`
}

func (e *PumpFunGlobalVolumeAccumulatorAccountEvent) EventType() EventType {
	return EventTypeAccountPumpFunGlobalVolumeAccumulator
}
func (e *PumpFunGlobalVolumeAccumulatorAccountEvent) GetMetadata() EventMetadata {
	return e.Metadata
}

type PumpFunGlobalVolumeAccumulator struct {
	StartTime        int64    `json:"start_time"`
	EndTime          int64    `json:"end_time"`
	SecondsInADay    int64    `json:"seconds_in_a_day"`
	Mint             string   `json:"mint"`
	TotalTokenSupply []uint64 `json:"total_token_supply"`
	SolVolumes       []uint64 `json:"sol_volumes"`
}

type PumpFunUserVolumeAccumulatorAccountEvent struct {
	Metadata              EventMetadata                `json:"metadata"`
	Pubkey                string                       `json:"pubkey"`
	UserVolumeAccumulator PumpFunUserVolumeAccumulator `json:"user_volume_accumulator"`
}

func (e *PumpFunUserVolumeAccumulatorAccountEvent) EventType() EventType {
	return EventTypeAccountPumpFunUserVolumeAccumulator
}
func (e *PumpFunUserVolumeAccumulatorAccountEvent) GetMetadata() EventMetadata {
	return e.Metadata
}

type PumpFunUserVolumeAccumulator struct {
	User                       string `json:"user"`
	NeedsClaim                 bool   `json:"needs_claim"`
	TotalUnclaimedTokens       uint64 `json:"total_unclaimed_tokens"`
	TotalClaimedTokens         uint64 `json:"total_claimed_tokens"`
	CurrentSolVolume           uint64 `json:"current_sol_volume"`
	LastUpdateTimestamp        int64  `json:"last_update_timestamp"`
	HasTotalClaimedTokens      bool   `json:"has_total_claimed_tokens"`
	CashbackEarned             uint64 `json:"cashback_earned"`
	TotalCashbackClaimed       uint64 `json:"total_cashback_claimed"`
	StableCashbackEarned       uint64 `json:"stable_cashback_earned"`
	TotalStableCashbackClaimed uint64 `json:"total_stable_cashback_claimed"`
}

type PumpFeesShareholder struct {
	Address  string `json:"address"`
	ShareBps uint16 `json:"share_bps"`
}

type PumpFeesConfigStatus string

const (
	PumpFeesConfigStatusPaused PumpFeesConfigStatus = "Paused"
	PumpFeesConfigStatusActive PumpFeesConfigStatus = "Active"
)

type PumpFeesFees struct {
	LpFeeBps       uint64 `json:"lp_fee_bps"`
	ProtocolFeeBps uint64 `json:"protocol_fee_bps"`
	CreatorFeeBps  uint64 `json:"creator_fee_bps"`
}

type PumpFeesFeeTier struct {
	MarketCapLamportsThreshold string       `json:"market_cap_lamports_threshold"`
	Fees                       PumpFeesFees `json:"fees"`
}

type PumpFeesCreateFeeSharingConfigEvent struct {
	Metadata            EventMetadata         `json:"metadata"`
	Timestamp           int64                 `json:"timestamp"`
	Mint                string                `json:"mint"`
	BondingCurve        string                `json:"bonding_curve"`
	Pool                string                `json:"pool,omitempty"`
	SharingConfig       string                `json:"sharing_config"`
	Admin               string                `json:"admin"`
	InitialShareholders []PumpFeesShareholder `json:"initial_shareholders"`
	Status              PumpFeesConfigStatus  `json:"status"`
}

func (e *PumpFeesCreateFeeSharingConfigEvent) EventType() EventType {
	return EventTypePumpFeesCreateFeeSharingConfig
}
func (e *PumpFeesCreateFeeSharingConfigEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesInitializeFeeConfigEvent struct {
	Metadata  EventMetadata `json:"metadata"`
	Timestamp int64         `json:"timestamp"`
	Admin     string        `json:"admin"`
	FeeConfig string        `json:"fee_config"`
}

func (e *PumpFeesInitializeFeeConfigEvent) EventType() EventType {
	return EventTypePumpFeesInitializeFeeConfig
}
func (e *PumpFeesInitializeFeeConfigEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesResetFeeSharingConfigEvent struct {
	Metadata        EventMetadata         `json:"metadata"`
	Timestamp       int64                 `json:"timestamp"`
	Mint            string                `json:"mint"`
	SharingConfig   string                `json:"sharing_config"`
	OldAdmin        string                `json:"old_admin"`
	OldShareholders []PumpFeesShareholder `json:"old_shareholders"`
	NewAdmin        string                `json:"new_admin"`
	NewShareholders []PumpFeesShareholder `json:"new_shareholders"`
}

func (e *PumpFeesResetFeeSharingConfigEvent) EventType() EventType {
	return EventTypePumpFeesResetFeeSharingConfig
}
func (e *PumpFeesResetFeeSharingConfigEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesRevokeFeeSharingAuthorityEvent struct {
	Metadata      EventMetadata `json:"metadata"`
	Timestamp     int64         `json:"timestamp"`
	Mint          string        `json:"mint"`
	SharingConfig string        `json:"sharing_config"`
	Admin         string        `json:"admin"`
}

func (e *PumpFeesRevokeFeeSharingAuthorityEvent) EventType() EventType {
	return EventTypePumpFeesRevokeFeeSharingAuthority
}
func (e *PumpFeesRevokeFeeSharingAuthorityEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesTransferFeeSharingAuthorityEvent struct {
	Metadata      EventMetadata `json:"metadata"`
	Timestamp     int64         `json:"timestamp"`
	Mint          string        `json:"mint"`
	SharingConfig string        `json:"sharing_config"`
	OldAdmin      string        `json:"old_admin"`
	NewAdmin      string        `json:"new_admin"`
}

func (e *PumpFeesTransferFeeSharingAuthorityEvent) EventType() EventType {
	return EventTypePumpFeesTransferFeeSharingAuthority
}
func (e *PumpFeesTransferFeeSharingAuthorityEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesUpdateAdminEvent struct {
	Metadata  EventMetadata `json:"metadata"`
	Timestamp int64         `json:"timestamp"`
	OldAdmin  string        `json:"old_admin"`
	NewAdmin  string        `json:"new_admin"`
}

func (e *PumpFeesUpdateAdminEvent) EventType() EventType       { return EventTypePumpFeesUpdateAdmin }
func (e *PumpFeesUpdateAdminEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesUpdateFeeConfigEvent struct {
	Metadata  EventMetadata     `json:"metadata"`
	Timestamp int64             `json:"timestamp"`
	Admin     string            `json:"admin"`
	FeeConfig string            `json:"fee_config"`
	FeeTiers  []PumpFeesFeeTier `json:"fee_tiers"`
	FlatFees  PumpFeesFees      `json:"flat_fees"`
}

func (e *PumpFeesUpdateFeeConfigEvent) EventType() EventType {
	return EventTypePumpFeesUpdateFeeConfig
}
func (e *PumpFeesUpdateFeeConfigEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesUpdateFeeSharesEvent struct {
	Metadata         EventMetadata         `json:"metadata"`
	Timestamp        int64                 `json:"timestamp"`
	Mint             string                `json:"mint"`
	SharingConfig    string                `json:"sharing_config"`
	Admin            string                `json:"admin"`
	BondingCurve     string                `json:"bonding_curve"`
	PumpCreatorVault string                `json:"pump_creator_vault"`
	NewShareholders  []PumpFeesShareholder `json:"new_shareholders"`
}

func (e *PumpFeesUpdateFeeSharesEvent) EventType() EventType {
	return EventTypePumpFeesUpdateFeeShares
}
func (e *PumpFeesUpdateFeeSharesEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFeesUpsertFeeTiersEvent struct {
	Metadata  EventMetadata     `json:"metadata"`
	Timestamp int64             `json:"timestamp"`
	Admin     string            `json:"admin"`
	FeeConfig string            `json:"fee_config"`
	FeeTiers  []PumpFeesFeeTier `json:"fee_tiers"`
	Offset    uint8             `json:"offset"`
}

func (e *PumpFeesUpsertFeeTiersEvent) EventType() EventType {
	return EventTypePumpFeesUpsertFeeTiers
}
func (e *PumpFeesUpsertFeeTiersEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunMigrateBondingCurveCreatorEvent struct {
	Metadata      EventMetadata `json:"metadata"`
	Timestamp     int64         `json:"timestamp"`
	Mint          string        `json:"mint"`
	BondingCurve  string        `json:"bonding_curve"`
	SharingConfig string        `json:"sharing_config"`
	OldCreator    string        `json:"old_creator"`
	NewCreator    string        `json:"new_creator"`
}

func (e *PumpFunMigrateBondingCurveCreatorEvent) EventType() EventType {
	return EventTypePumpFunMigrateBondingCurveCreator
}
func (e *PumpFunMigrateBondingCurveCreatorEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// PumpSwap 事件结构体
// ============================================================

// PumpSwapBuyEvent PumpSwap 买入事件
type PumpSwapBuyEvent struct {
	Metadata                         EventMetadata `json:"metadata"`
	Timestamp                        int64         `json:"timestamp"`
	BaseAmountOut                    uint64        `json:"base_amount_out"`
	MaxQuoteAmountIn                 uint64        `json:"max_quote_amount_in"`
	UserBaseTokenReserves            uint64        `json:"user_base_token_reserves"`
	UserQuoteTokenReserves           uint64        `json:"user_quote_token_reserves"`
	PoolBaseTokenReserves            uint64        `json:"pool_base_token_reserves"`
	PoolQuoteTokenReserves           uint64        `json:"pool_quote_token_reserves"`
	QuoteAmountIn                    uint64        `json:"quote_amount_in"`
	LpFeeBasisPoints                 uint64        `json:"lp_fee_basis_points"`
	LpFee                            uint64        `json:"lp_fee"`
	ProtocolFeeBasisPoints           uint64        `json:"protocol_fee_basis_points"`
	ProtocolFee                      uint64        `json:"protocol_fee"`
	QuoteAmountInWithLpFee           uint64        `json:"quote_amount_in_with_lp_fee"`
	UserQuoteAmountIn                uint64        `json:"user_quote_amount_in"`
	Pool                             string        `json:"pool"`
	User                             string        `json:"user"`
	UserBaseTokenAccount             string        `json:"user_base_token_account"`
	UserQuoteTokenAccount            string        `json:"user_quote_token_account"`
	ProtocolFeeRecipient             string        `json:"protocol_fee_recipient"`
	ProtocolFeeRecipientTokenAccount string        `json:"protocol_fee_recipient_token_account"`
	CoinCreator                      string        `json:"coin_creator"`
	CoinCreatorFeeBasisPoints        uint64        `json:"coin_creator_fee_basis_points"`
	CoinCreatorFee                   uint64        `json:"coin_creator_fee"`
	TrackVolume                      bool          `json:"track_volume"`
	TotalUnclaimedTokens             uint64        `json:"total_unclaimed_tokens"`
	TotalClaimedTokens               uint64        `json:"total_claimed_tokens"`
	CurrentSolVolume                 uint64        `json:"current_sol_volume"`
	LastUpdateTimestamp              int64         `json:"last_update_timestamp"`
	MinBaseAmountOut                 uint64        `json:"min_base_amount_out"`
	IxName                           string        `json:"ix_name"`
	MayhemMode                       bool          `json:"mayhem_mode"`
	CashbackFeeBasisPoints           uint64        `json:"cashback_fee_basis_points"`
	Cashback                         uint64        `json:"cashback"`
	BuybackFeeBasisPoints            uint64        `json:"buyback_fee_basis_points"`
	BuybackFee                       uint64        `json:"buyback_fee"`
	VirtualQuoteReserves             string        `json:"virtual_quote_reserves"`
	CanBoost                         bool          `json:"can_boost"`
	BaseSupply                       uint64        `json:"base_supply"`
	IsCashbackCoin                   bool          `json:"is_cashback_coin"`
	BaseMint                         string        `json:"base_mint"`
	QuoteMint                        string        `json:"quote_mint"`
	PoolBaseTokenAccount             string        `json:"pool_base_token_account"`
	PoolQuoteTokenAccount            string        `json:"pool_quote_token_account"`
	CoinCreatorVaultAta              string        `json:"coin_creator_vault_ata"`
	CoinCreatorVaultAuthority        string        `json:"coin_creator_vault_authority"`
	BaseTokenProgram                 string        `json:"base_token_program"`
	QuoteTokenProgram                string        `json:"quote_token_program"`
	PoolV2                           string        `json:"pool_v2"`
	FeeRecipient                     string        `json:"fee_recipient"`
	FeeRecipientQuoteTokenAccount    string        `json:"fee_recipient_quote_token_account"`
	IsPumpPool                       bool          `json:"is_pump_pool"`
}

func (e *PumpSwapBuyEvent) EventType() EventType       { return EventTypePumpSwapBuy }
func (e *PumpSwapBuyEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpSwapSellEvent PumpSwap 卖出事件
type PumpSwapSellEvent struct {
	Metadata                         EventMetadata `json:"metadata"`
	Timestamp                        int64         `json:"timestamp"`
	BaseAmountIn                     uint64        `json:"base_amount_in"`
	MinQuoteAmountOut                uint64        `json:"min_quote_amount_out"`
	UserBaseTokenReserves            uint64        `json:"user_base_token_reserves"`
	UserQuoteTokenReserves           uint64        `json:"user_quote_token_reserves"`
	PoolBaseTokenReserves            uint64        `json:"pool_base_token_reserves"`
	PoolQuoteTokenReserves           uint64        `json:"pool_quote_token_reserves"`
	QuoteAmountOut                   uint64        `json:"quote_amount_out"`
	LpFeeBasisPoints                 uint64        `json:"lp_fee_basis_points"`
	LpFee                            uint64        `json:"lp_fee"`
	ProtocolFeeBasisPoints           uint64        `json:"protocol_fee_basis_points"`
	ProtocolFee                      uint64        `json:"protocol_fee"`
	QuoteAmountOutWithoutLpFee       uint64        `json:"quote_amount_out_without_lp_fee"`
	UserQuoteAmountOut               uint64        `json:"user_quote_amount_out"`
	Pool                             string        `json:"pool"`
	User                             string        `json:"user"`
	UserBaseTokenAccount             string        `json:"user_base_token_account"`
	UserQuoteTokenAccount            string        `json:"user_quote_token_account"`
	ProtocolFeeRecipient             string        `json:"protocol_fee_recipient"`
	ProtocolFeeRecipientTokenAccount string        `json:"protocol_fee_recipient_token_account"`
	CoinCreator                      string        `json:"coin_creator"`
	CoinCreatorFeeBasisPoints        uint64        `json:"coin_creator_fee_basis_points"`
	CoinCreatorFee                   uint64        `json:"coin_creator_fee"`
	CashbackFeeBasisPoints           uint64        `json:"cashback_fee_basis_points"`
	Cashback                         uint64        `json:"cashback"`
	BuybackFeeBasisPoints            uint64        `json:"buyback_fee_basis_points"`
	BuybackFee                       uint64        `json:"buyback_fee"`
	VirtualQuoteReserves             string        `json:"virtual_quote_reserves"`
	CanBoost                         bool          `json:"can_boost"`
	BaseSupply                       uint64        `json:"base_supply"`
	BaseMint                         string        `json:"base_mint"`
	QuoteMint                        string        `json:"quote_mint"`
	PoolBaseTokenAccount             string        `json:"pool_base_token_account"`
	PoolQuoteTokenAccount            string        `json:"pool_quote_token_account"`
	CoinCreatorVaultAta              string        `json:"coin_creator_vault_ata"`
	CoinCreatorVaultAuthority        string        `json:"coin_creator_vault_authority"`
	BaseTokenProgram                 string        `json:"base_token_program"`
	QuoteTokenProgram                string        `json:"quote_token_program"`
	PoolV2                           string        `json:"pool_v2"`
	FeeRecipient                     string        `json:"fee_recipient"`
	FeeRecipientQuoteTokenAccount    string        `json:"fee_recipient_quote_token_account"`
	IsPumpPool                       bool          `json:"is_pump_pool"`
}

func (e *PumpSwapSellEvent) EventType() EventType       { return EventTypePumpSwapSell }
func (e *PumpSwapSellEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpSwapCreatePoolEvent PumpSwap 创建池子事件
type PumpSwapCreatePoolEvent struct {
	Metadata              EventMetadata `json:"metadata"`
	Timestamp             int64         `json:"timestamp"`
	Index                 uint16        `json:"index"`
	Creator               string        `json:"creator"`
	BaseMint              string        `json:"base_mint"`
	QuoteMint             string        `json:"quote_mint"`
	BaseMintDecimals      uint8         `json:"base_mint_decimals"`
	QuoteMintDecimals     uint8         `json:"quote_mint_decimals"`
	BaseAmountIn          uint64        `json:"base_amount_in"`
	QuoteAmountIn         uint64        `json:"quote_amount_in"`
	PoolBaseAmount        uint64        `json:"pool_base_amount"`
	PoolQuoteAmount       uint64        `json:"pool_quote_amount"`
	MinimumLiquidity      uint64        `json:"minimum_liquidity"`
	InitialLiquidity      uint64        `json:"initial_liquidity"`
	LpTokenAmountOut      uint64        `json:"lp_token_amount_out"`
	PoolBump              uint8         `json:"pool_bump"`
	Pool                  string        `json:"pool"`
	LpMint                string        `json:"lp_mint"`
	UserBaseTokenAccount  string        `json:"user_base_token_account"`
	UserQuoteTokenAccount string        `json:"user_quote_token_account"`
	CoinCreator           string        `json:"coin_creator"`
	IsMayhemMode          bool          `json:"is_mayhem_mode"`
}

func (e *PumpSwapCreatePoolEvent) EventType() EventType       { return EventTypePumpSwapCreatePool }
func (e *PumpSwapCreatePoolEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpSwapLiquidityAddedEvent PumpSwap 添加流动性事件
type PumpSwapLiquidityAddedEvent struct {
	Metadata               EventMetadata `json:"metadata"`
	Timestamp              int64         `json:"timestamp"`
	LpTokenAmountOut       uint64        `json:"lp_token_amount_out"`
	MaxBaseAmountIn        uint64        `json:"max_base_amount_in"`
	MaxQuoteAmountIn       uint64        `json:"max_quote_amount_in"`
	UserBaseTokenReserves  uint64        `json:"user_base_token_reserves"`
	UserQuoteTokenReserves uint64        `json:"user_quote_token_reserves"`
	PoolBaseTokenReserves  uint64        `json:"pool_base_token_reserves"`
	PoolQuoteTokenReserves uint64        `json:"pool_quote_token_reserves"`
	BaseAmountIn           uint64        `json:"base_amount_in"`
	QuoteAmountIn          uint64        `json:"quote_amount_in"`
	LpMintSupply           uint64        `json:"lp_mint_supply"`
	Pool                   string        `json:"pool"`
	User                   string        `json:"user"`
	UserBaseTokenAccount   string        `json:"user_base_token_account"`
	UserQuoteTokenAccount  string        `json:"user_quote_token_account"`
	UserPoolTokenAccount   string        `json:"user_pool_token_account"`
}

func (e *PumpSwapLiquidityAddedEvent) EventType() EventType       { return EventTypePumpSwapLiquidityAdded }
func (e *PumpSwapLiquidityAddedEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpSwapLiquidityRemovedEvent PumpSwap 移除流动性事件
type PumpSwapLiquidityRemovedEvent struct {
	Metadata               EventMetadata `json:"metadata"`
	Timestamp              int64         `json:"timestamp"`
	LpTokenAmountIn        uint64        `json:"lp_token_amount_in"`
	MinBaseAmountOut       uint64        `json:"min_base_amount_out"`
	MinQuoteAmountOut      uint64        `json:"min_quote_amount_out"`
	UserBaseTokenReserves  uint64        `json:"user_base_token_reserves"`
	UserQuoteTokenReserves uint64        `json:"user_quote_token_reserves"`
	PoolBaseTokenReserves  uint64        `json:"pool_base_token_reserves"`
	PoolQuoteTokenReserves uint64        `json:"pool_quote_token_reserves"`
	BaseAmountOut          uint64        `json:"base_amount_out"`
	QuoteAmountOut         uint64        `json:"quote_amount_out"`
	LpMintSupply           uint64        `json:"lp_mint_supply"`
	Pool                   string        `json:"pool"`
	User                   string        `json:"user"`
	UserBaseTokenAccount   string        `json:"user_base_token_account"`
	UserQuoteTokenAccount  string        `json:"user_quote_token_account"`
	UserPoolTokenAccount   string        `json:"user_pool_token_account"`
}

func (e *PumpSwapLiquidityRemovedEvent) EventType() EventType {
	return EventTypePumpSwapLiquidityRemoved
}
func (e *PumpSwapLiquidityRemovedEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Raydium CLMM 事件结构体
// ============================================================

// RaydiumClmmSwapEvent Raydium CLMM 交换事件
type RaydiumClmmSwapEvent struct {
	Metadata      EventMetadata `json:"metadata"`
	PoolState     string        `json:"pool_state"`
	Sender        string        `json:"sender"`
	TokenAccount0 string        `json:"token_account_0"`
	TokenAccount1 string        `json:"token_account_1"`
	Amount0       uint64        `json:"amount_0"`
	Amount1       uint64        `json:"amount_1"`
	ZeroForOne    bool          `json:"zero_for_one"`
	SqrtPriceX64  string        `json:"sqrt_price_x64"`
	Liquidity     string        `json:"liquidity"`
	TransferFee0  uint64        `json:"transfer_fee_0"`
	TransferFee1  uint64        `json:"transfer_fee_1"`
	Tick          int32         `json:"tick"`
}

func (e *RaydiumClmmSwapEvent) EventType() EventType       { return EventTypeRaydiumClmmSwap }
func (e *RaydiumClmmSwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumClmmIncreaseLiquidityEvent Raydium CLMM 增加流动性事件
type RaydiumClmmIncreaseLiquidityEvent struct {
	Metadata           EventMetadata `json:"metadata"`
	Pool               string        `json:"pool"`
	PositionNftMint    string        `json:"position_nft_mint"`
	User               string        `json:"user"`
	Liquidity          string        `json:"liquidity"`
	Amount0            uint64        `json:"amount_0"`
	Amount1            uint64        `json:"amount_1"`
	Amount0TransferFee uint64        `json:"amount_0_transfer_fee"`
	Amount1TransferFee uint64        `json:"amount_1_transfer_fee"`
	Amount0Max         uint64        `json:"amount0_max"`
	Amount1Max         uint64        `json:"amount1_max"`
}

func (e *RaydiumClmmIncreaseLiquidityEvent) EventType() EventType {
	return EventTypeRaydiumClmmIncreaseLiquidity
}
func (e *RaydiumClmmIncreaseLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumClmmDecreaseLiquidityEvent Raydium CLMM 减少流动性事件
type RaydiumClmmDecreaseLiquidityEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	PositionNftMint string        `json:"position_nft_mint"`
	User            string        `json:"user"`
	Liquidity       string        `json:"liquidity"`
	DecreaseAmount0 uint64        `json:"decrease_amount_0"`
	DecreaseAmount1 uint64        `json:"decrease_amount_1"`
	FeeAmount0      uint64        `json:"fee_amount_0"`
	FeeAmount1      uint64        `json:"fee_amount_1"`
	RewardAmounts   [3]uint64     `json:"reward_amounts"`
	TransferFee0    uint64        `json:"transfer_fee_0"`
	TransferFee1    uint64        `json:"transfer_fee_1"`
	Amount0Min      uint64        `json:"amount0_min"`
	Amount1Min      uint64        `json:"amount1_min"`
}

func (e *RaydiumClmmDecreaseLiquidityEvent) EventType() EventType {
	return EventTypeRaydiumClmmDecreaseLiquidity
}
func (e *RaydiumClmmDecreaseLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumClmmOpenPositionEvent Raydium CLMM 开启仓位事件
type RaydiumClmmOpenPositionEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	User            string        `json:"user"`
	PositionNftMint string        `json:"position_nft_mint"`
	TickLowerIndex  int32         `json:"tick_lower_index"`
	TickUpperIndex  int32         `json:"tick_upper_index"`
	Liquidity       string        `json:"liquidity"`
}

func (e *RaydiumClmmOpenPositionEvent) EventType() EventType {
	return EventTypeRaydiumClmmOpenPosition
}
func (e *RaydiumClmmOpenPositionEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumClmmOpenPositionWithTokenExtNftEvent Raydium CLMM Token-2022 NFT 开启仓位事件
type RaydiumClmmOpenPositionWithTokenExtNftEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	User            string        `json:"user"`
	PositionNftMint string        `json:"position_nft_mint"`
	TickLowerIndex  int32         `json:"tick_lower_index"`
	TickUpperIndex  int32         `json:"tick_upper_index"`
	Liquidity       string        `json:"liquidity"`
}

func (e *RaydiumClmmOpenPositionWithTokenExtNftEvent) EventType() EventType {
	return EventTypeRaydiumClmmOpenPositionWithTokenExtNft
}
func (e *RaydiumClmmOpenPositionWithTokenExtNftEvent) GetMetadata() EventMetadata {
	return e.Metadata
}

// RaydiumClmmClosePositionEvent Raydium CLMM 关闭仓位事件
type RaydiumClmmClosePositionEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	User            string        `json:"user"`
	PositionNftMint string        `json:"position_nft_mint"`
}

func (e *RaydiumClmmClosePositionEvent) EventType() EventType {
	return EventTypeRaydiumClmmClosePosition
}
func (e *RaydiumClmmClosePositionEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumClmmCreatePoolEvent Raydium CLMM 创建池子事件
type RaydiumClmmCreatePoolEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	Pool         string        `json:"pool"`
	Creator      string        `json:"creator"`
	Token0Mint   string        `json:"token_0_mint"`
	Token1Mint   string        `json:"token_1_mint"`
	TickSpacing  int           `json:"tick_spacing"`
	FeeRate      int           `json:"fee_rate"`
	SqrtPriceX64 string        `json:"sqrt_price_x64"`
	Tick         int32         `json:"tick"`
	TokenVault0  string        `json:"token_vault_0"`
	TokenVault1  string        `json:"token_vault_1"`
	OpenTime     uint64        `json:"open_time"`
}

func (e *RaydiumClmmCreatePoolEvent) EventType() EventType       { return EventTypeRaydiumClmmCreatePool }
func (e *RaydiumClmmCreatePoolEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumClmmCollectFeeEvent Raydium CLMM 收取费用事件
type RaydiumClmmCollectFeeEvent struct {
	Metadata               EventMetadata `json:"metadata"`
	PoolState              string        `json:"pool_state"`
	PositionNftMint        string        `json:"position_nft_mint"`
	RecipientTokenAccount0 string        `json:"recipient_token_account_0"`
	RecipientTokenAccount1 string        `json:"recipient_token_account_1"`
	Amount0                uint64        `json:"amount_0"`
	Amount1                uint64        `json:"amount_1"`
}

func (e *RaydiumClmmCollectFeeEvent) EventType() EventType       { return EventTypeRaydiumClmmCollectFee }
func (e *RaydiumClmmCollectFeeEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmLiquidityChangeEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	PoolState       string        `json:"pool_state"`
	Tick            int32         `json:"tick"`
	TickLower       int32         `json:"tick_lower"`
	TickUpper       int32         `json:"tick_upper"`
	LiquidityBefore string        `json:"liquidity_before"`
	LiquidityAfter  string        `json:"liquidity_after"`
}

func (e *RaydiumClmmLiquidityChangeEvent) EventType() EventType {
	return EventTypeRaydiumClmmLiquidityChange
}
func (e *RaydiumClmmLiquidityChangeEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmConfigChangeEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Index           uint16        `json:"index"`
	Owner           string        `json:"owner"`
	ProtocolFeeRate uint32        `json:"protocol_fee_rate"`
	TradeFeeRate    uint32        `json:"trade_fee_rate"`
	TickSpacing     uint16        `json:"tick_spacing"`
	FundFeeRate     uint32        `json:"fund_fee_rate"`
	FundOwner       string        `json:"fund_owner"`
}

func (e *RaydiumClmmConfigChangeEvent) EventType() EventType {
	return EventTypeRaydiumClmmConfigChange
}
func (e *RaydiumClmmConfigChangeEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmCreatePersonalPositionEvent struct {
	Metadata                  EventMetadata `json:"metadata"`
	PoolState                 string        `json:"pool_state"`
	Minter                    string        `json:"minter"`
	NftOwner                  string        `json:"nft_owner"`
	TickLowerIndex            int32         `json:"tick_lower_index"`
	TickUpperIndex            int32         `json:"tick_upper_index"`
	Liquidity                 string        `json:"liquidity"`
	DepositAmount0            uint64        `json:"deposit_amount_0"`
	DepositAmount1            uint64        `json:"deposit_amount_1"`
	DepositAmount0TransferFee uint64        `json:"deposit_amount_0_transfer_fee"`
	DepositAmount1TransferFee uint64        `json:"deposit_amount_1_transfer_fee"`
}

func (e *RaydiumClmmCreatePersonalPositionEvent) EventType() EventType {
	return EventTypeRaydiumClmmCreatePersonalPosition
}
func (e *RaydiumClmmCreatePersonalPositionEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmLiquidityCalculateEvent struct {
	Metadata         EventMetadata `json:"metadata"`
	PoolLiquidity    string        `json:"pool_liquidity"`
	PoolSqrtPriceX64 string        `json:"pool_sqrt_price_x64"`
	PoolTick         int32         `json:"pool_tick"`
	CalcAmount0      uint64        `json:"calc_amount_0"`
	CalcAmount1      uint64        `json:"calc_amount_1"`
	TradeFeeOwed0    uint64        `json:"trade_fee_owed_0"`
	TradeFeeOwed1    uint64        `json:"trade_fee_owed_1"`
	TransferFee0     uint64        `json:"transfer_fee_0"`
	TransferFee1     uint64        `json:"transfer_fee_1"`
}

func (e *RaydiumClmmLiquidityCalculateEvent) EventType() EventType {
	return EventTypeRaydiumClmmLiquidityCalculate
}
func (e *RaydiumClmmLiquidityCalculateEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmOpenLimitOrderEvent struct {
	Metadata    EventMetadata `json:"metadata"`
	PoolID      string        `json:"pool_id"`
	LimitOrder  string        `json:"limit_order"`
	ZeroForOne  bool          `json:"zero_for_one"`
	TickIndex   int32         `json:"tick_index"`
	TotalAmount uint64        `json:"total_amount"`
	TransferFee uint64        `json:"transfer_fee"`
}

func (e *RaydiumClmmOpenLimitOrderEvent) EventType() EventType {
	return EventTypeRaydiumClmmOpenLimitOrder
}
func (e *RaydiumClmmOpenLimitOrderEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmIncreaseLimitOrderEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	PoolID          string        `json:"pool_id"`
	LimitOrder      string        `json:"limit_order"`
	ZeroForOne      bool          `json:"zero_for_one"`
	TickIndex       int32         `json:"tick_index"`
	TotalAmount     uint64        `json:"total_amount"`
	IncreasedAmount uint64        `json:"increased_amount"`
	TransferFee     uint64        `json:"transfer_fee"`
}

func (e *RaydiumClmmIncreaseLimitOrderEvent) EventType() EventType {
	return EventTypeRaydiumClmmIncreaseLimitOrder
}
func (e *RaydiumClmmIncreaseLimitOrderEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmDecreaseLimitOrderEvent struct {
	Metadata            EventMetadata `json:"metadata"`
	PoolID              string        `json:"pool_id"`
	LimitOrder          string        `json:"limit_order"`
	ZeroForOne          bool          `json:"zero_for_one"`
	TickIndex           int32         `json:"tick_index"`
	TotalAmount         uint64        `json:"total_amount"`
	FilledAmount        uint64        `json:"filled_amount"`
	SettledOutputAmount uint64        `json:"settled_output_amount"`
	DecreasedAmount     uint64        `json:"decreased_amount"`
}

func (e *RaydiumClmmDecreaseLimitOrderEvent) EventType() EventType {
	return EventTypeRaydiumClmmDecreaseLimitOrder
}
func (e *RaydiumClmmDecreaseLimitOrderEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmSettleLimitOrderEvent struct {
	Metadata         EventMetadata `json:"metadata"`
	PoolID           string        `json:"pool_id"`
	LimitOrder       string        `json:"limit_order"`
	ZeroForOne       bool          `json:"zero_for_one"`
	TickIndex        int32         `json:"tick_index"`
	TotalAmount      uint64        `json:"total_amount"`
	FilledAmount     uint64        `json:"filled_amount"`
	SettledAmountOut uint64        `json:"settled_amount_out"`
}

func (e *RaydiumClmmSettleLimitOrderEvent) EventType() EventType {
	return EventTypeRaydiumClmmSettleLimitOrder
}
func (e *RaydiumClmmSettleLimitOrderEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmUpdateRewardInfosEvent struct {
	Metadata              EventMetadata `json:"metadata"`
	RewardGrowthGlobalX64 []string      `json:"reward_growth_global_x64"`
}

func (e *RaydiumClmmUpdateRewardInfosEvent) EventType() EventType {
	return EventTypeRaydiumClmmUpdateRewardInfos
}
func (e *RaydiumClmmUpdateRewardInfosEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Raydium AMM v4 事件结构体
// ============================================================

// RaydiumAmmV4SwapEvent Raydium AMM v4 交换事件
type RaydiumAmmV4SwapEvent struct {
	Metadata                    EventMetadata `json:"metadata"`
	Amm                         string        `json:"amm"`
	UserSourceOwner             string        `json:"user_source_owner"`
	AmountIn                    uint64        `json:"amount_in"`
	MinimumAmountOut            uint64        `json:"minimum_amount_out"`
	MaxAmountIn                 uint64        `json:"max_amount_in"`
	AmountOut                   uint64        `json:"amount_out"`
	TokenProgram                string        `json:"token_program"`
	AmmAuthority                string        `json:"amm_authority"`
	AmmOpenOrders               string        `json:"amm_open_orders"`
	PoolCoinTokenAccount        string        `json:"pool_coin_token_account"`
	PoolPcTokenAccount          string        `json:"pool_pc_token_account"`
	SerumProgram                string        `json:"serum_program"`
	SerumMarket                 string        `json:"serum_market"`
	SerumBids                   string        `json:"serum_bids"`
	SerumAsks                   string        `json:"serum_asks"`
	SerumEventQueue             string        `json:"serum_event_queue"`
	SerumCoinVaultAccount       string        `json:"serum_coin_vault_account"`
	SerumPcVaultAccount         string        `json:"serum_pc_vault_account"`
	SerumVaultSigner            string        `json:"serum_vault_signer"`
	UserSourceTokenAccount      string        `json:"user_source_token_account"`
	UserDestinationTokenAccount string        `json:"user_destination_token_account"`
}

func (e *RaydiumAmmV4SwapEvent) EventType() EventType       { return EventTypeRaydiumAmmV4Swap }
func (e *RaydiumAmmV4SwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumAmmV4DepositEvent Raydium AMM v4 存款事件
type RaydiumAmmV4DepositEvent struct {
	Metadata             EventMetadata `json:"metadata"`
	Amm                  string        `json:"amm"`
	UserOwner            string        `json:"user_owner"`
	MaxCoinAmount        uint64        `json:"max_coin_amount"`
	MaxPcAmount          uint64        `json:"max_pc_amount"`
	BaseSide             uint64        `json:"base_side"`
	TokenProgram         string        `json:"token_program"`
	AmmAuthority         string        `json:"amm_authority"`
	AmmOpenOrders        string        `json:"amm_open_orders"`
	AmmTargetOrders      string        `json:"amm_target_orders"`
	LpMintAddress        string        `json:"lp_mint_address"`
	PoolCoinTokenAccount string        `json:"pool_coin_token_account"`
	PoolPcTokenAccount   string        `json:"pool_pc_token_account"`
	SerumMarket          string        `json:"serum_market"`
	UserCoinTokenAccount string        `json:"user_coin_token_account"`
	UserPcTokenAccount   string        `json:"user_pc_token_account"`
	UserLpTokenAccount   string        `json:"user_lp_token_account"`
	SerumEventQueue      string        `json:"serum_event_queue"`
}

func (e *RaydiumAmmV4DepositEvent) EventType() EventType       { return EventTypeRaydiumAmmV4Deposit }
func (e *RaydiumAmmV4DepositEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumAmmV4WithdrawEvent Raydium AMM v4 取款事件
type RaydiumAmmV4WithdrawEvent struct {
	Metadata               EventMetadata `json:"metadata"`
	Amm                    string        `json:"amm"`
	UserOwner              string        `json:"user_owner"`
	Amount                 uint64        `json:"amount"`
	TokenProgram           string        `json:"token_program"`
	AmmAuthority           string        `json:"amm_authority"`
	AmmOpenOrders          string        `json:"amm_open_orders"`
	AmmTargetOrders        string        `json:"amm_target_orders"`
	LpMintAddress          string        `json:"lp_mint_address"`
	PoolCoinTokenAccount   string        `json:"pool_coin_token_account"`
	PoolPcTokenAccount     string        `json:"pool_pc_token_account"`
	PoolWithdrawQueue      string        `json:"pool_withdraw_queue"`
	PoolTempLpTokenAccount string        `json:"pool_temp_lp_token_account"`
	SerumProgram           string        `json:"serum_program"`
	SerumMarket            string        `json:"serum_market"`
	SerumCoinVaultAccount  string        `json:"serum_coin_vault_account"`
	SerumPcVaultAccount    string        `json:"serum_pc_vault_account"`
	SerumVaultSigner       string        `json:"serum_vault_signer"`
	UserLpTokenAccount     string        `json:"user_lp_token_account"`
	UserCoinTokenAccount   string        `json:"user_coin_token_account"`
	UserPcTokenAccount     string        `json:"user_pc_token_account"`
	SerumEventQueue        string        `json:"serum_event_queue"`
	SerumBids              string        `json:"serum_bids"`
	SerumAsks              string        `json:"serum_asks"`
}

func (e *RaydiumAmmV4WithdrawEvent) EventType() EventType       { return EventTypeRaydiumAmmV4Withdraw }
func (e *RaydiumAmmV4WithdrawEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Raydium CPMM 事件结构体
// ============================================================

// RaydiumCpmmSwapEvent Raydium CPMM 交换事件
type RaydiumCpmmSwapEvent struct {
	Metadata          EventMetadata `json:"metadata"`
	PoolID            string        `json:"pool_id"`
	InputAmount       uint64        `json:"input_amount"`
	OutputAmount      uint64        `json:"output_amount"`
	InputVaultBefore  uint64        `json:"input_vault_before"`
	OutputVaultBefore uint64        `json:"output_vault_before"`
	InputTransferFee  uint64        `json:"input_transfer_fee"`
	OutputTransferFee uint64        `json:"output_transfer_fee"`
	BaseInput         bool          `json:"base_input"`
}

func (e *RaydiumCpmmSwapEvent) EventType() EventType       { return EventTypeRaydiumCpmmSwap }
func (e *RaydiumCpmmSwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumCpmmDepositEvent Raydium CPMM 存款事件
type RaydiumCpmmDepositEvent struct {
	Metadata      EventMetadata `json:"metadata"`
	Pool          string        `json:"pool"`
	User          string        `json:"user"`
	LpTokenAmount uint64        `json:"lp_token_amount"`
	Token0Amount  uint64        `json:"token0_amount"`
	Token1Amount  uint64        `json:"token1_amount"`
}

func (e *RaydiumCpmmDepositEvent) EventType() EventType       { return EventTypeRaydiumCpmmDeposit }
func (e *RaydiumCpmmDepositEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumCpmmWithdrawEvent Raydium CPMM 取款事件
type RaydiumCpmmWithdrawEvent struct {
	Metadata      EventMetadata `json:"metadata"`
	Pool          string        `json:"pool"`
	User          string        `json:"user"`
	LpTokenAmount uint64        `json:"lp_token_amount"`
	Token0Amount  uint64        `json:"token0_amount"`
	Token1Amount  uint64        `json:"token1_amount"`
}

func (e *RaydiumCpmmWithdrawEvent) EventType() EventType       { return EventTypeRaydiumCpmmWithdraw }
func (e *RaydiumCpmmWithdrawEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumCpmmInitializeEvent Raydium CPMM 初始化事件
type RaydiumCpmmInitializeEvent struct {
	Metadata    EventMetadata `json:"metadata"`
	Pool        string        `json:"pool"`
	Creator     string        `json:"creator"`
	InitAmount0 uint64        `json:"init_amount0"`
	InitAmount1 uint64        `json:"init_amount1"`
}

func (e *RaydiumCpmmInitializeEvent) EventType() EventType       { return EventTypeRaydiumCpmmInitialize }
func (e *RaydiumCpmmInitializeEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Orca Whirlpool 事件结构体
// ============================================================

// OrcaWhirlpoolSwapEvent Orca Whirlpool 交换事件
type OrcaWhirlpoolSwapEvent struct {
	Metadata          EventMetadata `json:"metadata"`
	Whirlpool         string        `json:"whirlpool"`
	AToB              bool          `json:"a_to_b"`
	PreSqrtPrice      string        `json:"pre_sqrt_price"`
	PostSqrtPrice     string        `json:"post_sqrt_price"`
	InputAmount       uint64        `json:"input_amount"`
	OutputAmount      uint64        `json:"output_amount"`
	InputTransferFee  uint64        `json:"input_transfer_fee"`
	OutputTransferFee uint64        `json:"output_transfer_fee"`
	LpFee             uint64        `json:"lp_fee"`
	ProtocolFee       uint64        `json:"protocol_fee"`
}

func (e *OrcaWhirlpoolSwapEvent) EventType() EventType       { return EventTypeOrcaWhirlpoolSwap }
func (e *OrcaWhirlpoolSwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// OrcaWhirlpoolLiquidityIncreasedEvent Orca Whirlpool 增加流动性事件
type OrcaWhirlpoolLiquidityIncreasedEvent struct {
	Metadata          EventMetadata `json:"metadata"`
	Whirlpool         string        `json:"whirlpool"`
	Position          string        `json:"position"`
	TickLowerIndex    int32         `json:"tick_lower_index"`
	TickUpperIndex    int32         `json:"tick_upper_index"`
	Liquidity         string        `json:"liquidity"`
	TokenAAmount      uint64        `json:"token_a_amount"`
	TokenBAmount      uint64        `json:"token_b_amount"`
	TokenATransferFee uint64        `json:"token_a_transfer_fee"`
	TokenBTransferFee uint64        `json:"token_b_transfer_fee"`
}

func (e *OrcaWhirlpoolLiquidityIncreasedEvent) EventType() EventType {
	return EventTypeOrcaWhirlpoolLiquidityIncreased
}
func (e *OrcaWhirlpoolLiquidityIncreasedEvent) GetMetadata() EventMetadata { return e.Metadata }

// OrcaWhirlpoolLiquidityDecreasedEvent Orca Whirlpool 减少流动性事件
type OrcaWhirlpoolLiquidityDecreasedEvent struct {
	Metadata          EventMetadata `json:"metadata"`
	Whirlpool         string        `json:"whirlpool"`
	Position          string        `json:"position"`
	TickLowerIndex    int32         `json:"tick_lower_index"`
	TickUpperIndex    int32         `json:"tick_upper_index"`
	Liquidity         string        `json:"liquidity"`
	TokenAAmount      uint64        `json:"token_a_amount"`
	TokenBAmount      uint64        `json:"token_b_amount"`
	TokenATransferFee uint64        `json:"token_a_transfer_fee"`
	TokenBTransferFee uint64        `json:"token_b_transfer_fee"`
}

func (e *OrcaWhirlpoolLiquidityDecreasedEvent) EventType() EventType {
	return EventTypeOrcaWhirlpoolLiquidityDecreased
}
func (e *OrcaWhirlpoolLiquidityDecreasedEvent) GetMetadata() EventMetadata { return e.Metadata }

// OrcaWhirlpoolPoolInitializedEvent Orca Whirlpool 池子初始化事件
type OrcaWhirlpoolPoolInitializedEvent struct {
	Metadata         EventMetadata `json:"metadata"`
	Whirlpool        string        `json:"whirlpool"`
	WhirlpoolsConfig string        `json:"whirlpools_config"`
	TokenMintA       string        `json:"token_mint_a"`
	TokenMintB       string        `json:"token_mint_b"`
	TickSpacing      uint16        `json:"tick_spacing"`
	TokenProgramA    string        `json:"token_program_a"`
	TokenProgramB    string        `json:"token_program_b"`
	DecimalsA        uint8         `json:"decimals_a"`
	DecimalsB        uint8         `json:"decimals_b"`
	InitialSqrtPrice string        `json:"initial_sqrt_price"`
}

func (e *OrcaWhirlpoolPoolInitializedEvent) EventType() EventType {
	return EventTypeOrcaWhirlpoolPoolInitialized
}
func (e *OrcaWhirlpoolPoolInitializedEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Meteora DLMM 事件结构体
// ============================================================

// MeteoraDlmmSwapEvent Meteora DLMM 交换事件
type MeteoraDlmmSwapEvent struct {
	Metadata    EventMetadata `json:"metadata"`
	Pool        string        `json:"pool"`
	From        string        `json:"from"`
	StartBinID  int32         `json:"start_bin_id"`
	EndBinID    int32         `json:"end_bin_id"`
	AmountIn    uint64        `json:"amount_in"`
	AmountOut   uint64        `json:"amount_out"`
	SwapForY    bool          `json:"swap_for_y"`
	Fee         uint64        `json:"fee"`
	ProtocolFee uint64        `json:"protocol_fee"`
	FeeBps      string        `json:"fee_bps"`
	HostFee     uint64        `json:"host_fee"`
}

func (e *MeteoraDlmmSwapEvent) EventType() EventType       { return EventTypeMeteoraDlmmSwap }
func (e *MeteoraDlmmSwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmAddLiquidityEvent Meteora DLMM 添加流动性事件
type MeteoraDlmmAddLiquidityEvent struct {
	Metadata    EventMetadata `json:"metadata"`
	Pool        string        `json:"pool"`
	From        string        `json:"from"`
	Position    string        `json:"position"`
	Amounts     []uint64      `json:"amounts"`
	ActiveBinID int32         `json:"active_bin_id"`
}

func (e *MeteoraDlmmAddLiquidityEvent) EventType() EventType       { return EventTypeMeteoraDlmmAddLiquidity }
func (e *MeteoraDlmmAddLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmRemoveLiquidityEvent Meteora DLMM 移除流动性事件
type MeteoraDlmmRemoveLiquidityEvent struct {
	Metadata    EventMetadata `json:"metadata"`
	Pool        string        `json:"pool"`
	From        string        `json:"from"`
	Position    string        `json:"position"`
	Amounts     []uint64      `json:"amounts"`
	ActiveBinID int32         `json:"active_bin_id"`
}

func (e *MeteoraDlmmRemoveLiquidityEvent) EventType() EventType {
	return EventTypeMeteoraDlmmRemoveLiquidity
}
func (e *MeteoraDlmmRemoveLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmInitializePoolEvent Meteora DLMM 初始化池子事件
type MeteoraDlmmInitializePoolEvent struct {
	Metadata    EventMetadata `json:"metadata"`
	Pool        string        `json:"pool"`
	Creator     string        `json:"creator"`
	ActiveBinID int32         `json:"active_bin_id"`
	BinStep     uint16        `json:"bin_step"`
}

func (e *MeteoraDlmmInitializePoolEvent) EventType() EventType {
	return EventTypeMeteoraDlmmInitializePool
}
func (e *MeteoraDlmmInitializePoolEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmInitializeBinArrayEvent Meteora DLMM 初始化 Bin Array 事件
type MeteoraDlmmInitializeBinArrayEvent struct {
	Metadata EventMetadata `json:"metadata"`
	Pool     string        `json:"pool"`
	BinArray string        `json:"bin_array"`
	Index    uint64        `json:"index"`
}

func (e *MeteoraDlmmInitializeBinArrayEvent) EventType() EventType {
	return EventTypeMeteoraDlmmInitializeBinArray
}
func (e *MeteoraDlmmInitializeBinArrayEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmCreatePositionEvent Meteora DLMM 创建仓位事件
type MeteoraDlmmCreatePositionEvent struct {
	Metadata   EventMetadata `json:"metadata"`
	Pool       string        `json:"pool"`
	Position   string        `json:"position"`
	Owner      string        `json:"owner"`
	LowerBinID int32         `json:"lower_bin_id"`
	Width      uint32        `json:"width"`
}

func (e *MeteoraDlmmCreatePositionEvent) EventType() EventType {
	return EventTypeMeteoraDlmmCreatePosition
}
func (e *MeteoraDlmmCreatePositionEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmClosePositionEvent Meteora DLMM 关闭仓位事件
type MeteoraDlmmClosePositionEvent struct {
	Metadata EventMetadata `json:"metadata"`
	Pool     string        `json:"pool"`
	Position string        `json:"position"`
	Owner    string        `json:"owner"`
}

func (e *MeteoraDlmmClosePositionEvent) EventType() EventType {
	return EventTypeMeteoraDlmmClosePosition
}
func (e *MeteoraDlmmClosePositionEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDlmmClaimFeeEvent Meteora DLMM 收取费用事件
type MeteoraDlmmClaimFeeEvent struct {
	Metadata EventMetadata `json:"metadata"`
	Pool     string        `json:"pool"`
	Position string        `json:"position"`
	Owner    string        `json:"owner"`
	FeeX     uint64        `json:"fee_x"`
	FeeY     uint64        `json:"fee_y"`
}

func (e *MeteoraDlmmClaimFeeEvent) EventType() EventType       { return EventTypeMeteoraDlmmClaimFee }
func (e *MeteoraDlmmClaimFeeEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Meteora Pools 事件结构体
// ============================================================

// MeteoraPoolsSwapEvent Meteora Pools 交换事件
type MeteoraPoolsSwapEvent struct {
	Metadata  EventMetadata `json:"metadata"`
	InAmount  uint64        `json:"in_amount"`
	OutAmount uint64        `json:"out_amount"`
	TradeFee  uint64        `json:"trade_fee"`
	AdminFee  uint64        `json:"admin_fee"`
	HostFee   uint64        `json:"host_fee"`
}

func (e *MeteoraPoolsSwapEvent) EventType() EventType       { return EventTypeMeteoraPoolsSwap }
func (e *MeteoraPoolsSwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraPoolsAddLiquidityEvent Meteora Pools 添加流动性事件
type MeteoraPoolsAddLiquidityEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	LpMintAmount uint64        `json:"lp_mint_amount"`
	TokenAAmount uint64        `json:"token_a_amount"`
	TokenBAmount uint64        `json:"token_b_amount"`
}

func (e *MeteoraPoolsAddLiquidityEvent) EventType() EventType {
	return EventTypeMeteoraPoolsAddLiquidity
}
func (e *MeteoraPoolsAddLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraPoolsRemoveLiquidityEvent Meteora Pools 移除流动性事件
type MeteoraPoolsRemoveLiquidityEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	LpUnmintAmount  uint64        `json:"lp_unmint_amount"`
	TokenAOutAmount uint64        `json:"token_a_out_amount"`
	TokenBOutAmount uint64        `json:"token_b_out_amount"`
}

func (e *MeteoraPoolsRemoveLiquidityEvent) EventType() EventType {
	return EventTypeMeteoraPoolsRemoveLiquidity
}
func (e *MeteoraPoolsRemoveLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraPoolsBootstrapLiquidityEvent Meteora Pools 引导流动性事件
type MeteoraPoolsBootstrapLiquidityEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	LpMintAmount uint64        `json:"lp_mint_amount"`
	TokenAAmount uint64        `json:"token_a_amount"`
	TokenBAmount uint64        `json:"token_b_amount"`
	Pool         string        `json:"pool"`
}

func (e *MeteoraPoolsBootstrapLiquidityEvent) EventType() EventType {
	return EventTypeMeteoraPoolsBootstrapLiquidity
}
func (e *MeteoraPoolsBootstrapLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraPoolsPoolCreatedEvent Meteora Pools 创建池子事件
type MeteoraPoolsPoolCreatedEvent struct {
	Metadata   EventMetadata `json:"metadata"`
	LpMint     string        `json:"lp_mint"`
	TokenAMint string        `json:"token_a_mint"`
	TokenBMint string        `json:"token_b_mint"`
	PoolType   uint8         `json:"pool_type"`
	Pool       string        `json:"pool"`
}

func (e *MeteoraPoolsPoolCreatedEvent) EventType() EventType       { return EventTypeMeteoraPoolsPoolCreated }
func (e *MeteoraPoolsPoolCreatedEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraPoolsSetPoolFeesEvent Meteora Pools 设置池子费用事件
type MeteoraPoolsSetPoolFeesEvent struct {
	Metadata                 EventMetadata `json:"metadata"`
	TradeFeeNumerator        uint64        `json:"trade_fee_numerator"`
	TradeFeeDenominator      uint64        `json:"trade_fee_denominator"`
	OwnerTradeFeeNumerator   uint64        `json:"owner_trade_fee_numerator"`
	OwnerTradeFeeDenominator uint64        `json:"owner_trade_fee_denominator"`
	Pool                     string        `json:"pool"`
}

func (e *MeteoraPoolsSetPoolFeesEvent) EventType() EventType       { return EventTypeMeteoraPoolsSetPoolFees }
func (e *MeteoraPoolsSetPoolFeesEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Meteora DAMM v2 事件结构体
// ============================================================

// MeteoraDammV2SwapEvent Meteora DAMM v2 交换事件
type MeteoraDammV2SwapEvent struct {
	Metadata         EventMetadata `json:"metadata"`
	Pool             string        `json:"pool"`
	TradeDirection   uint8         `json:"trade_direction"`
	HasReferral      bool          `json:"has_referral"`
	AmountIn         uint64        `json:"amount_in"`
	MinimumAmountOut uint64        `json:"minimum_amount_out"`
	OutputAmount     uint64        `json:"output_amount"`
	NextSqrtPrice    string        `json:"next_sqrt_price"`
	LpFee            uint64        `json:"lp_fee"`
	ProtocolFee      uint64        `json:"protocol_fee"`
	PartnerFee       uint64        `json:"partner_fee"`
	ReferralFee      uint64        `json:"referral_fee"`
	ActualAmountIn   uint64        `json:"actual_amount_in"`
	CurrentTimestamp uint64        `json:"current_timestamp"`
	TokenAVault      string        `json:"token_a_vault"`
	TokenBVault      string        `json:"token_b_vault"`
	TokenAMint       string        `json:"token_a_mint"`
	TokenBMint       string        `json:"token_b_mint"`
	TokenAProgram    string        `json:"token_a_program"`
	TokenBProgram    string        `json:"token_b_program"`
}

func (e *MeteoraDammV2SwapEvent) EventType() EventType       { return EventTypeMeteoraDammV2Swap }
func (e *MeteoraDammV2SwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDammV2CreatePositionEvent Meteora DAMM v2 创建仓位事件
type MeteoraDammV2CreatePositionEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	Owner           string        `json:"owner"`
	Position        string        `json:"position"`
	PositionNftMint string        `json:"position_nft_mint"`
}

func (e *MeteoraDammV2CreatePositionEvent) EventType() EventType {
	return EventTypeMeteoraDammV2CreatePosition
}
func (e *MeteoraDammV2CreatePositionEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDammV2ClosePositionEvent Meteora DAMM v2 关闭仓位事件
type MeteoraDammV2ClosePositionEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	Owner           string        `json:"owner"`
	Position        string        `json:"position"`
	PositionNftMint string        `json:"position_nft_mint"`
}

func (e *MeteoraDammV2ClosePositionEvent) EventType() EventType {
	return EventTypeMeteoraDammV2ClosePosition
}
func (e *MeteoraDammV2ClosePositionEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDammV2AddLiquidityEvent Meteora DAMM v2 添加流动性事件
type MeteoraDammV2AddLiquidityEvent struct {
	Metadata              EventMetadata `json:"metadata"`
	Pool                  string        `json:"pool"`
	Position              string        `json:"position"`
	Owner                 string        `json:"owner"`
	LiquidityDelta        string        `json:"liquidity_delta"`
	TokenAAmountThreshold uint64        `json:"token_a_amount_threshold"`
	TokenBAmountThreshold uint64        `json:"token_b_amount_threshold"`
	TokenAAmount          uint64        `json:"token_a_amount"`
	TokenBAmount          uint64        `json:"token_b_amount"`
	TotalAmountA          uint64        `json:"total_amount_a"`
	TotalAmountB          uint64        `json:"total_amount_b"`
}

func (e *MeteoraDammV2AddLiquidityEvent) EventType() EventType {
	return EventTypeMeteoraDammV2AddLiquidity
}
func (e *MeteoraDammV2AddLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDammV2RemoveLiquidityEvent Meteora DAMM v2 移除流动性事件
type MeteoraDammV2RemoveLiquidityEvent struct {
	Metadata              EventMetadata `json:"metadata"`
	Pool                  string        `json:"pool"`
	Position              string        `json:"position"`
	Owner                 string        `json:"owner"`
	LiquidityDelta        string        `json:"liquidity_delta"`
	TokenAAmountThreshold uint64        `json:"token_a_amount_threshold"`
	TokenBAmountThreshold uint64        `json:"token_b_amount_threshold"`
	TokenAAmount          uint64        `json:"token_a_amount"`
	TokenBAmount          uint64        `json:"token_b_amount"`
}

func (e *MeteoraDammV2RemoveLiquidityEvent) EventType() EventType {
	return EventTypeMeteoraDammV2RemoveLiquidity
}
func (e *MeteoraDammV2RemoveLiquidityEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDammV2InitializePoolEvent Meteora DAMM v2 初始化池子事件
type MeteoraDammV2InitializePoolEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	Position        string        `json:"position"`
	PositionNftMint string        `json:"position_nft_mint"`
	TokenAMint      string        `json:"token_a_mint"`
	TokenBMint      string        `json:"token_b_mint"`
	Creator         string        `json:"creator"`
	Payer           string        `json:"payer"`
	AlphaVault      string        `json:"alpha_vault"`
	PoolFees        any           `json:"pool_fees"`
	SqrtMinPrice    string        `json:"sqrt_min_price"`
	SqrtMaxPrice    string        `json:"sqrt_max_price"`
	ActivationType  uint8         `json:"activation_type"`
	CollectFeeMode  uint8         `json:"collect_fee_mode"`
	Liquidity       string        `json:"liquidity"`
	SqrtPrice       string        `json:"sqrt_price"`
	ActivationPoint uint64        `json:"activation_point"`
	TokenAFlag      uint8         `json:"token_a_flag"`
	TokenBFlag      uint8         `json:"token_b_flag"`
	TokenAAmount    uint64        `json:"token_a_amount"`
	TokenBAmount    uint64        `json:"token_b_amount"`
	TotalAmountA    uint64        `json:"total_amount_a"`
	TotalAmountB    uint64        `json:"total_amount_b"`
	PoolType        uint8         `json:"pool_type"`
}

func (e *MeteoraDammV2InitializePoolEvent) EventType() EventType {
	return EventTypeMeteoraDammV2InitializePool
}
func (e *MeteoraDammV2InitializePoolEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDbcSwapEvent Meteora DBC 交易事件
type MeteoraDbcSwapEvent struct {
	Metadata          EventMetadata `json:"metadata"`
	Pool              string        `json:"pool"`
	Config            string        `json:"config"`
	TradeDirection    uint8         `json:"trade_direction"`
	HasReferral       bool          `json:"has_referral"`
	AmountIn          uint64        `json:"amount_in"`
	MinimumAmountOut  uint64        `json:"minimum_amount_out"`
	ActualInputAmount uint64        `json:"actual_input_amount"`
	OutputAmount      uint64        `json:"output_amount"`
	NextSqrtPrice     string        `json:"next_sqrt_price"`
	TradingFee        uint64        `json:"trading_fee"`
	ProtocolFee       uint64        `json:"protocol_fee"`
	ReferralFee       uint64        `json:"referral_fee"`
	CurrentTimestamp  uint64        `json:"current_timestamp"`
}

func (e *MeteoraDbcSwapEvent) EventType() EventType       { return EventTypeMeteoraDbcSwap }
func (e *MeteoraDbcSwapEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDbcInitializePoolEvent Meteora DBC 初始化池子事件
type MeteoraDbcInitializePoolEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	Pool            string        `json:"pool"`
	Config          string        `json:"config"`
	Creator         string        `json:"creator"`
	BaseMint        string        `json:"base_mint"`
	PoolType        uint8         `json:"pool_type"`
	ActivationPoint uint64        `json:"activation_point"`
}

func (e *MeteoraDbcInitializePoolEvent) EventType() EventType {
	return EventTypeMeteoraDbcInitializePool
}
func (e *MeteoraDbcInitializePoolEvent) GetMetadata() EventMetadata { return e.Metadata }

// MeteoraDbcCurveCompleteEvent Meteora DBC 曲线完成事件
type MeteoraDbcCurveCompleteEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	Pool         string        `json:"pool"`
	Config       string        `json:"config"`
	BaseReserve  uint64        `json:"base_reserve"`
	QuoteReserve uint64        `json:"quote_reserve"`
}

func (e *MeteoraDbcCurveCompleteEvent) EventType() EventType {
	return EventTypeMeteoraDbcCurveComplete
}
func (e *MeteoraDbcCurveCompleteEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// RaydiumLaunchlab 事件结构体
// ============================================================

// RaydiumLaunchlabTradeEvent RaydiumLaunchlab 交易事件
type RaydiumLaunchlabTradeEvent struct {
	Metadata       EventMetadata `json:"metadata"`
	PoolState      string        `json:"pool_state"`
	User           string        `json:"user"`
	AmountIn       uint64        `json:"amount_in"`
	AmountOut      uint64        `json:"amount_out"`
	IsBuy          bool          `json:"is_buy"`
	TradeDirection string        `json:"trade_direction"`
	ExactIn        bool          `json:"exact_in"`
}

func (e *RaydiumLaunchlabTradeEvent) EventType() EventType       { return EventTypeRaydiumLaunchlabTrade }
func (e *RaydiumLaunchlabTradeEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumLaunchlabPoolCreateEvent RaydiumLaunchlab 创建池子事件
type RaydiumLaunchlabPoolCreateEvent struct {
	Metadata      EventMetadata             `json:"metadata"`
	BaseMintParam RaydiumLaunchlabMintParam `json:"base_mint_param"`
	PoolState     string                    `json:"pool_state"`
	Creator       string                    `json:"creator"`
}

// RaydiumLaunchlabMintParam RaydiumLaunchlab mint 参数
type RaydiumLaunchlabMintParam struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Uri      string `json:"uri"`
	Decimals uint8  `json:"decimals"`
}

func (e *RaydiumLaunchlabPoolCreateEvent) EventType() EventType {
	return EventTypeRaydiumLaunchlabPoolCreate
}
func (e *RaydiumLaunchlabPoolCreateEvent) GetMetadata() EventMetadata { return e.Metadata }

// RaydiumLaunchlabMigrateAmmEvent RaydiumLaunchlab 迁移 AMM 事件
type RaydiumLaunchlabMigrateAmmEvent struct {
	Metadata        EventMetadata `json:"metadata"`
	OldPool         string        `json:"old_pool"`
	NewPool         string        `json:"new_pool"`
	User            string        `json:"user"`
	LiquidityAmount uint64        `json:"liquidity_amount"`
}

func (e *RaydiumLaunchlabMigrateAmmEvent) EventType() EventType {
	return EventTypeRaydiumLaunchlabMigrateAmm
}
func (e *RaydiumLaunchlabMigrateAmmEvent) GetMetadata() EventMetadata { return e.Metadata }

// ============================================================
// Account 事件结构体
// ============================================================

type RaydiumClmmAmmConfigAccountEvent struct {
	Metadata  EventMetadata        `json:"metadata"`
	Pubkey    string               `json:"pubkey"`
	AmmConfig RaydiumClmmAmmConfig `json:"amm_config"`
}

func (e *RaydiumClmmAmmConfigAccountEvent) EventType() EventType {
	return EventTypeAccountRaydiumClmmAmmConfig
}
func (e *RaydiumClmmAmmConfigAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmAmmConfig struct {
	Bump            uint8    `json:"bump"`
	Index           uint16   `json:"index"`
	Owner           string   `json:"owner"`
	ProtocolFeeRate uint32   `json:"protocol_fee_rate"`
	TradeFeeRate    uint32   `json:"trade_fee_rate"`
	TickSpacing     uint16   `json:"tick_spacing"`
	FundFeeRate     uint32   `json:"fund_fee_rate"`
	PaddingU32      uint32   `json:"padding_u32"`
	FundOwner       string   `json:"fund_owner"`
	Padding         []uint64 `json:"padding"`
}

type RaydiumClmmPoolStateAccountEvent struct {
	Metadata  EventMetadata        `json:"metadata"`
	Pubkey    string               `json:"pubkey"`
	PoolState RaydiumClmmPoolState `json:"pool_state"`
}

func (e *RaydiumClmmPoolStateAccountEvent) EventType() EventType {
	return EventTypeAccountRaydiumClmmPoolState
}
func (e *RaydiumClmmPoolStateAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmPoolState struct {
	Bump                []uint8                   `json:"bump"`
	AmmConfig           string                    `json:"amm_config"`
	Owner               string                    `json:"owner"`
	TokenMint0          string                    `json:"token_mint_0"`
	TokenMint1          string                    `json:"token_mint_1"`
	TokenVault0         string                    `json:"token_vault_0"`
	TokenVault1         string                    `json:"token_vault_1"`
	ObservationKey      string                    `json:"observation_key"`
	MintDecimals0       uint8                     `json:"mint_decimals_0"`
	MintDecimals1       uint8                     `json:"mint_decimals_1"`
	TickSpacing         uint16                    `json:"tick_spacing"`
	Liquidity           string                    `json:"liquidity"`
	SqrtPriceX64        string                    `json:"sqrt_price_x64"`
	TickCurrent         int32                     `json:"tick_current"`
	Padding3            uint16                    `json:"padding3"`
	Padding4            uint16                    `json:"padding4"`
	FeeGrowthGlobal0X64 string                    `json:"fee_growth_global_0_x64"`
	FeeGrowthGlobal1X64 string                    `json:"fee_growth_global_1_x64"`
	ProtocolFeesToken0  uint64                    `json:"protocol_fees_token_0"`
	ProtocolFeesToken1  uint64                    `json:"protocol_fees_token_1"`
	Padding5            []string                  `json:"padding5"`
	Status              uint8                     `json:"status"`
	FeeOn               uint8                     `json:"fee_on"`
	Padding             []uint8                   `json:"padding"`
	RewardInfos         []RaydiumClmmRewardInfo   `json:"reward_infos"`
	TickArrayBitmap     []uint64                  `json:"tick_array_bitmap"`
	Padding6            []uint64                  `json:"padding6"`
	FundFeesToken0      uint64                    `json:"fund_fees_token_0"`
	FundFeesToken1      uint64                    `json:"fund_fees_token_1"`
	OpenTime            uint64                    `json:"open_time"`
	RecentEpoch         uint64                    `json:"recent_epoch"`
	DynamicFeeInfo      RaydiumClmmDynamicFeeInfo `json:"dynamic_fee_info"`
	Padding1            []uint64                  `json:"padding1"`
	Padding2            []uint64                  `json:"padding2"`
}

type RaydiumClmmRewardInfo struct {
	RewardState           uint8  `json:"reward_state"`
	OpenTime              uint64 `json:"open_time"`
	EndTime               uint64 `json:"end_time"`
	LastUpdateTime        uint64 `json:"last_update_time"`
	EmissionsPerSecondX64 string `json:"emissions_per_second_x64"`
	RewardTotalEmitted    uint64 `json:"reward_total_emitted"`
	RewardClaimed         uint64 `json:"reward_claimed"`
	TokenMint             string `json:"token_mint"`
	TokenVault            string `json:"token_vault"`
	Authority             string `json:"authority"`
	RewardGrowthGlobalX64 string `json:"reward_growth_global_x64"`
}

type RaydiumClmmDynamicFeeInfo struct {
	FilterPeriod              uint16  `json:"filter_period"`
	DecayPeriod               uint16  `json:"decay_period"`
	ReductionFactor           uint16  `json:"reduction_factor"`
	DynamicFeeControl         uint32  `json:"dynamic_fee_control"`
	MaxVolatilityAccumulator  uint32  `json:"max_volatility_accumulator"`
	TickSpacingIndexReference int32   `json:"tick_spacing_index_reference"`
	VolatilityReference       uint32  `json:"volatility_reference"`
	VolatilityAccumulator     uint32  `json:"volatility_accumulator"`
	LastUpdateTimestamp       uint64  `json:"last_update_timestamp"`
	Padding                   []uint8 `json:"padding"`
}

type RaydiumClmmTickArrayStateAccountEvent struct {
	Metadata       EventMetadata             `json:"metadata"`
	Pubkey         string                    `json:"pubkey"`
	TickArrayState RaydiumClmmTickArrayState `json:"tick_array_state"`
}

func (e *RaydiumClmmTickArrayStateAccountEvent) EventType() EventType {
	return EventTypeAccountRaydiumClmmTickArrayState
}
func (e *RaydiumClmmTickArrayStateAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumClmmTickArrayState struct {
	PoolID               string            `json:"pool_id"`
	StartTickIndex       int32             `json:"start_tick_index"`
	Ticks                []RaydiumClmmTick `json:"ticks"`
	InitializedTickCount uint8             `json:"initialized_tick_count"`
	RecentEpoch          uint64            `json:"recent_epoch"`
	Padding              []uint8           `json:"padding"`
}

type RaydiumClmmTick struct {
	Tick                      int32    `json:"tick"`
	LiquidityNet              string   `json:"liquidity_net"`
	LiquidityGross            string   `json:"liquidity_gross"`
	FeeGrowthOutside0X64      string   `json:"fee_growth_outside_0_x64"`
	FeeGrowthOutside1X64      string   `json:"fee_growth_outside_1_x64"`
	RewardGrowthsOutsideX64   []string `json:"reward_growths_outside_x64"`
	OrderPhase                uint64   `json:"order_phase"`
	OrdersAmount              uint64   `json:"orders_amount"`
	PartFilledOrdersRemaining uint64   `json:"part_filled_orders_remaining"`
	UnfilledRatioX64          string   `json:"unfilled_ratio_x64"`
	Padding                   []uint32 `json:"padding"`
}

type RaydiumCpmmAmmConfigAccountEvent struct {
	Metadata  EventMetadata        `json:"metadata"`
	Pubkey    string               `json:"pubkey"`
	AmmConfig RaydiumCpmmAmmConfig `json:"amm_config"`
}

func (e *RaydiumCpmmAmmConfigAccountEvent) EventType() EventType {
	return EventTypeAccountRaydiumCpmmAmmConfig
}
func (e *RaydiumCpmmAmmConfigAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumCpmmAmmConfig struct {
	Bump              uint8    `json:"bump"`
	DisableCreatePool bool     `json:"disable_create_pool"`
	Index             uint16   `json:"index"`
	TradeFeeRate      uint64   `json:"trade_fee_rate"`
	ProtocolFeeRate   uint64   `json:"protocol_fee_rate"`
	FundFeeRate       uint64   `json:"fund_fee_rate"`
	CreatePoolFee     uint64   `json:"create_pool_fee"`
	ProtocolOwner     string   `json:"protocol_owner"`
	FundOwner         string   `json:"fund_owner"`
	CreatorFeeRate    uint64   `json:"creator_fee_rate"`
	Padding           []uint64 `json:"padding"`
}

type RaydiumCpmmPoolStateAccountEvent struct {
	Metadata  EventMetadata        `json:"metadata"`
	Pubkey    string               `json:"pubkey"`
	PoolState RaydiumCpmmPoolState `json:"pool_state"`
}

func (e *RaydiumCpmmPoolStateAccountEvent) EventType() EventType {
	return EventTypeAccountRaydiumCpmmPoolState
}
func (e *RaydiumCpmmPoolStateAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type RaydiumCpmmPoolState struct {
	AmmConfig          string   `json:"amm_config"`
	PoolCreator        string   `json:"pool_creator"`
	Token0Vault        string   `json:"token_0_vault"`
	Token1Vault        string   `json:"token_1_vault"`
	LpMint             string   `json:"lp_mint"`
	Token0Mint         string   `json:"token_0_mint"`
	Token1Mint         string   `json:"token_1_mint"`
	Token0Program      string   `json:"token_0_program"`
	Token1Program      string   `json:"token_1_program"`
	ObservationKey     string   `json:"observation_key"`
	AuthBump           uint8    `json:"auth_bump"`
	Status             uint8    `json:"status"`
	LpMintDecimals     uint8    `json:"lp_mint_decimals"`
	Mint0Decimals      uint8    `json:"mint_0_decimals"`
	Mint1Decimals      uint8    `json:"mint_1_decimals"`
	LpSupply           uint64   `json:"lp_supply"`
	ProtocolFeesToken0 uint64   `json:"protocol_fees_token_0"`
	ProtocolFeesToken1 uint64   `json:"protocol_fees_token_1"`
	FundFeesToken0     uint64   `json:"fund_fees_token_0"`
	FundFeesToken1     uint64   `json:"fund_fees_token_1"`
	OpenTime           uint64   `json:"open_time"`
	RecentEpoch        uint64   `json:"recent_epoch"`
	CreatorFeeOn       uint8    `json:"creator_fee_on"`
	EnableCreatorFee   bool     `json:"enable_creator_fee"`
	Padding1           []uint8  `json:"padding1"`
	CreatorFeesToken0  uint64   `json:"creator_fees_token_0"`
	CreatorFeesToken1  uint64   `json:"creator_fees_token_1"`
	Padding            []uint64 `json:"padding"`
}

type OrcaWhirlpoolAccountEvent struct {
	Metadata  EventMetadata        `json:"metadata"`
	Pubkey    string               `json:"pubkey"`
	Whirlpool OrcaWhirlpoolAccount `json:"whirlpool"`
}

func (e *OrcaWhirlpoolAccountEvent) EventType() EventType       { return EventTypeAccountOrcaWhirlpool }
func (e *OrcaWhirlpoolAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type OrcaWhirlpoolAccount struct {
	WhirlpoolsConfig           string                    `json:"whirlpools_config"`
	WhirlpoolBump              uint8                     `json:"whirlpool_bump"`
	TickSpacing                uint16                    `json:"tick_spacing"`
	TickSpacingSeed            []uint8                   `json:"tick_spacing_seed"`
	FeeRate                    uint16                    `json:"fee_rate"`
	ProtocolFeeRate            uint16                    `json:"protocol_fee_rate"`
	Liquidity                  string                    `json:"liquidity"`
	SqrtPrice                  string                    `json:"sqrt_price"`
	TickCurrentIndex           int32                     `json:"tick_current_index"`
	ProtocolFeeOwedA           uint64                    `json:"protocol_fee_owed_a"`
	ProtocolFeeOwedB           uint64                    `json:"protocol_fee_owed_b"`
	TokenMintA                 string                    `json:"token_mint_a"`
	TokenVaultA                string                    `json:"token_vault_a"`
	FeeGrowthGlobalA           string                    `json:"fee_growth_global_a"`
	TokenMintB                 string                    `json:"token_mint_b"`
	TokenVaultB                string                    `json:"token_vault_b"`
	FeeGrowthGlobalB           string                    `json:"fee_growth_global_b"`
	RewardLastUpdatedTimestamp uint64                    `json:"reward_last_updated_timestamp"`
	RewardInfos                []OrcaWhirlpoolRewardInfo `json:"reward_infos"`
}

type OrcaWhirlpoolRewardInfo struct {
	Mint                  string `json:"mint"`
	Vault                 string `json:"vault"`
	Authority             string `json:"authority"`
	EmissionsPerSecondX64 string `json:"emissions_per_second_x64"`
	GrowthGlobalX64       string `json:"growth_global_x64"`
}

type OrcaPositionAccountEvent struct {
	Metadata EventMetadata       `json:"metadata"`
	Pubkey   string              `json:"pubkey"`
	Position OrcaPositionAccount `json:"position"`
}

func (e *OrcaPositionAccountEvent) EventType() EventType       { return EventTypeAccountOrcaPosition }
func (e *OrcaPositionAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type OrcaPositionAccount struct {
	Whirlpool            string                   `json:"whirlpool"`
	PositionMint         string                   `json:"position_mint"`
	Liquidity            string                   `json:"liquidity"`
	TickLowerIndex       int32                    `json:"tick_lower_index"`
	TickUpperIndex       int32                    `json:"tick_upper_index"`
	FeeGrowthCheckpointA string                   `json:"fee_growth_checkpoint_a"`
	FeeOwedA             uint64                   `json:"fee_owed_a"`
	FeeGrowthCheckpointB string                   `json:"fee_growth_checkpoint_b"`
	FeeOwedB             uint64                   `json:"fee_owed_b"`
	RewardInfos          []OrcaPositionRewardInfo `json:"reward_infos"`
}

type OrcaPositionRewardInfo struct {
	GrowthInsideCheckpoint string `json:"growth_inside_checkpoint"`
	AmountOwed             uint64 `json:"amount_owed"`
}

type OrcaTickArrayAccountEvent struct {
	Metadata  EventMetadata        `json:"metadata"`
	Pubkey    string               `json:"pubkey"`
	TickArray OrcaTickArrayAccount `json:"tick_array"`
}

func (e *OrcaTickArrayAccountEvent) EventType() EventType       { return EventTypeAccountOrcaTickArray }
func (e *OrcaTickArrayAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type OrcaTickArrayAccount struct {
	StartTickIndex int32      `json:"start_tick_index"`
	Ticks          []OrcaTick `json:"ticks"`
	Whirlpool      string     `json:"whirlpool"`
}

type OrcaTick struct {
	Initialized          bool     `json:"initialized"`
	LiquidityNet         string   `json:"liquidity_net"`
	LiquidityGross       string   `json:"liquidity_gross"`
	FeeGrowthOutsideA    string   `json:"fee_growth_outside_a"`
	FeeGrowthOutsideB    string   `json:"fee_growth_outside_b"`
	RewardGrowthsOutside []string `json:"reward_growths_outside"`
}

type OrcaFeeTierAccountEvent struct {
	Metadata EventMetadata      `json:"metadata"`
	Pubkey   string             `json:"pubkey"`
	FeeTier  OrcaFeeTierAccount `json:"fee_tier"`
}

func (e *OrcaFeeTierAccountEvent) EventType() EventType       { return EventTypeAccountOrcaFeeTier }
func (e *OrcaFeeTierAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type OrcaFeeTierAccount struct {
	WhirlpoolsConfig string `json:"whirlpools_config"`
	TickSpacing      uint16 `json:"tick_spacing"`
	DefaultFeeRate   uint16 `json:"default_fee_rate"`
}

type OrcaWhirlpoolsConfigAccountEvent struct {
	Metadata EventMetadata               `json:"metadata"`
	Pubkey   string                      `json:"pubkey"`
	Config   OrcaWhirlpoolsConfigAccount `json:"config"`
}

func (e *OrcaWhirlpoolsConfigAccountEvent) EventType() EventType {
	return EventTypeAccountOrcaWhirlpoolsConfig
}
func (e *OrcaWhirlpoolsConfigAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

type OrcaWhirlpoolsConfigAccount struct {
	FeeAuthority                  string `json:"fee_authority"`
	CollectProtocolFeesAuthority  string `json:"collect_protocol_fees_authority"`
	RewardEmissionsSuperAuthority string `json:"reward_emissions_super_authority"`
	DefaultProtocolFeeRate        uint16 `json:"default_protocol_fee_rate"`
}

// TokenInfoEvent Token Mint 信息事件
type TokenInfoEvent struct {
	Metadata   EventMetadata `json:"metadata"`
	Pubkey     string        `json:"pubkey"`
	Executable bool          `json:"executable"`
	Lamports   uint64        `json:"lamports"`
	Owner      string        `json:"owner"`
	RentEpoch  uint64        `json:"rent_epoch"`
	Supply     uint64        `json:"supply"`
	Decimals   uint8         `json:"decimals"`
}

func (e *TokenInfoEvent) EventType() EventType       { return EventTypeTokenInfo }
func (e *TokenInfoEvent) GetMetadata() EventMetadata { return e.Metadata }

// TokenAccountEvent Token 账户事件
type TokenAccountEvent struct {
	Metadata   EventMetadata `json:"metadata"`
	Pubkey     string        `json:"pubkey"`
	Executable bool          `json:"executable"`
	Lamports   uint64        `json:"lamports"`
	Owner      string        `json:"owner"`
	RentEpoch  uint64        `json:"rent_epoch"`
	Amount     uint64        `json:"amount"`
}

func (e *TokenAccountEvent) EventType() EventType       { return EventTypeTokenAccount }
func (e *TokenAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

// NonceAccountEvent Nonce 账户事件
type NonceAccountEvent struct {
	Metadata   EventMetadata `json:"metadata"`
	Pubkey     string        `json:"pubkey"`
	Executable bool          `json:"executable"`
	Lamports   uint64        `json:"lamports"`
	Owner      string        `json:"owner"`
	RentEpoch  uint64        `json:"rent_epoch"`
	Nonce      string        `json:"nonce"`
	Authority  string        `json:"authority"`
}

func (e *NonceAccountEvent) EventType() EventType       { return EventTypeNonceAccount }
func (e *NonceAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpSwapGlobalConfigAccountEvent PumpSwap 全局配置账户事件
type PumpSwapGlobalConfigAccountEvent struct {
	Metadata EventMetadata                   `json:"metadata"`
	Pubkey   string                          `json:"pubkey"`
	Config   PumpSwapGlobalConfigAccountData `json:"config"`
}

// PumpSwapGlobalConfigAccountData PumpSwap 全局配置数据
type PumpSwapGlobalConfigAccountData struct {
	Admin                        string   `json:"admin"`
	LpFeeBasisPoints             uint64   `json:"lp_fee_basis_points"`
	ProtocolFeeBasisPoints       uint64   `json:"protocol_fee_basis_points"`
	DisableFlags                 uint8    `json:"disable_flags"`
	ProtocolFeeRecipients        []string `json:"protocol_fee_recipients"`
	CoinCreatorFeeBasisPoints    uint64   `json:"coin_creator_fee_basis_points"`
	AdminSetCoinCreatorAuthority string   `json:"admin_set_coin_creator_authority"`
	WhitelistPda                 string   `json:"whitelist_pda"`
	ReservedFeeRecipient         string   `json:"reserved_fee_recipient"`
	MayhemModeEnabled            bool     `json:"mayhem_mode_enabled"`
	ReservedFeeRecipients        []string `json:"reserved_fee_recipients"`
}

func (e *PumpSwapGlobalConfigAccountEvent) EventType() EventType {
	return EventTypeAccountPumpSwapGlobalConfig
}
func (e *PumpSwapGlobalConfigAccountEvent) GetMetadata() EventMetadata { return e.Metadata }

// PumpSwapPoolAccountEvent PumpSwap 池子账户事件
type PumpSwapPoolAccountEvent struct {
	Metadata EventMetadata           `json:"metadata"`
	Pubkey   string                  `json:"pubkey"`
	Pool     PumpSwapPoolAccountData `json:"pool"`
}

// PumpSwapPoolAccountData PumpSwap 池子数据
type PumpSwapPoolAccountData struct {
	PoolBump              uint8  `json:"pool_bump"`
	Index                 uint16 `json:"index"`
	Creator               string `json:"creator"`
	BaseMint              string `json:"base_mint"`
	QuoteMint             string `json:"quote_mint"`
	LpMint                string `json:"lp_mint"`
	PoolBaseTokenAccount  string `json:"pool_base_token_account"`
	PoolQuoteTokenAccount string `json:"pool_quote_token_account"`
	LpSupply              uint64 `json:"lp_supply"`
	CoinCreator           string `json:"coin_creator"`
	IsMayhemMode          bool   `json:"is_mayhem_mode"`
	IsCashbackCoin        bool   `json:"is_cashback_coin"`
	VirtualQuoteReserves  string `json:"virtual_quote_reserves"`
}

func (e *PumpSwapPoolAccountEvent) EventType() EventType       { return EventTypeAccountPumpSwapPool }
func (e *PumpSwapPoolAccountEvent) GetMetadata() EventMetadata { return e.Metadata }
