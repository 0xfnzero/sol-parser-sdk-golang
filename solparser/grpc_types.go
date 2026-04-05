package solparser

import pb "sol-parser-sdk-golang/proto"

// OrderMode gRPC 订阅顺序模式
type OrderMode string

const (
	OrderModeUnordered        OrderMode = "Unordered"
	OrderModeOrdered          OrderMode = "Ordered"
	OrderModeStreamingOrdered OrderMode = "StreamingOrdered"
	OrderModeMicroBatch       OrderMode = "MicroBatch"
)

// CommitmentLevel Solana 确认级别
type CommitmentLevel int32

const (
	CommitmentLevelProcessed CommitmentLevel = 0
	CommitmentLevelConfirmed CommitmentLevel = 1
	CommitmentLevelFinalized CommitmentLevel = 2
)

// ClientConfig gRPC 客户端配置
type ClientConfig struct {
	EnableMetrics           bool
	ConnectionTimeoutMs     int
	RequestTimeoutMs        int
	EnableTLS               bool
	MaxRetries              int
	RetryDelayMs            int
	MaxConcurrentStreams    int
	KeepAliveIntervalMs     int
	KeepAliveTimeoutMs      int
	BufferSize              int
	OrderMode               OrderMode
	OrderTimeoutMs          int
	MicroBatchUs            int
}

// DefaultClientConfig 返回默认客户端配置
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		EnableMetrics:           false,
		ConnectionTimeoutMs:     8000,
		RequestTimeoutMs:        15000,
		EnableTLS:               true,
		MaxRetries:              3,
		RetryDelayMs:            1000,
		MaxConcurrentStreams:    100,
		KeepAliveIntervalMs:     30000,
		KeepAliveTimeoutMs:      5000,
		BufferSize:              8192,
		OrderMode:               OrderModeUnordered,
		OrderTimeoutMs:          100,
		MicroBatchUs:            100,
	}
}

// TransactionFilter 交易过滤器
type TransactionFilter struct {
	AccountInclude []string
	AccountExclude []string
	AccountRequired []string
	Vote           *bool
	Failed         *bool
	Signature      string
}

// NewTransactionFilter 创建新的交易过滤器
func NewTransactionFilter() TransactionFilter {
	return TransactionFilter{
		AccountInclude:  []string{},
		AccountExclude:  []string{},
		AccountRequired: []string{},
	}
}

// EventType 事件类型
type EventType string

const (
	EventTypeBlockMeta                   EventType = "BlockMeta"
	EventTypeBonkTrade                   EventType = "BonkTrade"
	EventTypeBonkPoolCreate              EventType = "BonkPoolCreate"
	EventTypeBonkMigrateAmm              EventType = "BonkMigrateAmm"
	EventTypePumpFunTrade                EventType = "PumpFunTrade"
	EventTypePumpFunBuy                  EventType = "PumpFunBuy"
	EventTypePumpFunSell                 EventType = "PumpFunSell"
	EventTypePumpFunBuyExactSolIn        EventType = "PumpFunBuyExactSolIn"
	EventTypePumpFunCreate               EventType = "PumpFunCreate"
	EventTypePumpFunCreateV2             EventType = "PumpFunCreateV2"
	EventTypePumpFunComplete             EventType = "PumpFunComplete"
	EventTypePumpFunMigrate              EventType = "PumpFunMigrate"
	EventTypePumpSwapBuy                 EventType = "PumpSwapBuy"
	EventTypePumpSwapSell                EventType = "PumpSwapSell"
	EventTypePumpSwapCreatePool          EventType = "PumpSwapCreatePool"
	EventTypePumpSwapLiquidityAdded      EventType = "PumpSwapLiquidityAdded"
	EventTypePumpSwapLiquidityRemoved    EventType = "PumpSwapLiquidityRemoved"
	EventTypeMeteoraDammV2Swap           EventType = "MeteoraDammV2Swap"
	EventTypeMeteoraDammV2AddLiquidity   EventType = "MeteoraDammV2AddLiquidity"
	EventTypeMeteoraDammV2RemoveLiquidity EventType = "MeteoraDammV2RemoveLiquidity"
	EventTypeMeteoraDammV2CreatePosition EventType = "MeteoraDammV2CreatePosition"
	EventTypeMeteoraDammV2ClosePosition  EventType = "MeteoraDammV2ClosePosition"
	EventTypeMeteoraDammV2InitializePool EventType = "MeteoraDammV2InitializePool"
	EventTypeTokenAccount                EventType = "TokenAccount"
	EventTypeTokenInfo                   EventType = "TokenInfo"
	EventTypeNonceAccount                EventType = "NonceAccount"
	EventTypeAccountPumpSwapGlobalConfig EventType = "AccountPumpSwapGlobalConfig"
	EventTypeAccountPumpSwapPool         EventType = "AccountPumpSwapPool"

	// Raydium CLMM
	EventTypeRaydiumClmmSwap                        EventType = "RaydiumClmmSwap"
	EventTypeRaydiumClmmIncreaseLiquidity           EventType = "RaydiumClmmIncreaseLiquidity"
	EventTypeRaydiumClmmDecreaseLiquidity           EventType = "RaydiumClmmDecreaseLiquidity"
	EventTypeRaydiumClmmCreatePool                  EventType = "RaydiumClmmCreatePool"
	EventTypeRaydiumClmmOpenPosition                EventType = "RaydiumClmmOpenPosition"
	EventTypeRaydiumClmmOpenPositionWithTokenExtNft EventType = "RaydiumClmmOpenPositionWithTokenExtNft"
	EventTypeRaydiumClmmClosePosition               EventType = "RaydiumClmmClosePosition"
	EventTypeRaydiumClmmCollectFee                  EventType = "RaydiumClmmCollectFee"

	// Raydium CPMM
	EventTypeRaydiumCpmmSwap       EventType = "RaydiumCpmmSwap"
	EventTypeRaydiumCpmmDeposit    EventType = "RaydiumCpmmDeposit"
	EventTypeRaydiumCpmmWithdraw   EventType = "RaydiumCpmmWithdraw"
	EventTypeRaydiumCpmmInitialize EventType = "RaydiumCpmmInitialize"

	// Raydium AMM V4
	EventTypeRaydiumAmmV4Swap        EventType = "RaydiumAmmV4Swap"
	EventTypeRaydiumAmmV4Deposit     EventType = "RaydiumAmmV4Deposit"
	EventTypeRaydiumAmmV4Withdraw    EventType = "RaydiumAmmV4Withdraw"
	EventTypeRaydiumAmmV4WithdrawPnl EventType = "RaydiumAmmV4WithdrawPnl"
	EventTypeRaydiumAmmV4Initialize2 EventType = "RaydiumAmmV4Initialize2"

	// Orca Whirlpool
	EventTypeOrcaWhirlpoolSwap               EventType = "OrcaWhirlpoolSwap"
	EventTypeOrcaWhirlpoolLiquidityIncreased EventType = "OrcaWhirlpoolLiquidityIncreased"
	EventTypeOrcaWhirlpoolLiquidityDecreased EventType = "OrcaWhirlpoolLiquidityDecreased"
	EventTypeOrcaWhirlpoolPoolInitialized    EventType = "OrcaWhirlpoolPoolInitialized"

	// Meteora Pools (AMM)
	EventTypeMeteoraPoolsSwap               EventType = "MeteoraPoolsSwap"
	EventTypeMeteoraPoolsAddLiquidity       EventType = "MeteoraPoolsAddLiquidity"
	EventTypeMeteoraPoolsRemoveLiquidity    EventType = "MeteoraPoolsRemoveLiquidity"
	EventTypeMeteoraPoolsBootstrapLiquidity EventType = "MeteoraPoolsBootstrapLiquidity"
	EventTypeMeteoraPoolsPoolCreated        EventType = "MeteoraPoolsPoolCreated"
	EventTypeMeteoraPoolsSetPoolFees        EventType = "MeteoraPoolsSetPoolFees"

	// Meteora DLMM
	EventTypeMeteoraDlmmSwap               EventType = "MeteoraDlmmSwap"
	EventTypeMeteoraDlmmAddLiquidity       EventType = "MeteoraDlmmAddLiquidity"
	EventTypeMeteoraDlmmRemoveLiquidity    EventType = "MeteoraDlmmRemoveLiquidity"
	EventTypeMeteoraDlmmInitializePool     EventType = "MeteoraDlmmInitializePool"
	EventTypeMeteoraDlmmInitializeBinArray EventType = "MeteoraDlmmInitializeBinArray"
	EventTypeMeteoraDlmmCreatePosition     EventType = "MeteoraDlmmCreatePosition"
	EventTypeMeteoraDlmmClosePosition      EventType = "MeteoraDlmmClosePosition"
	EventTypeMeteoraDlmmClaimFee           EventType = "MeteoraDlmmClaimFee"

	// PumpSwap Trade alias
	EventTypePumpSwapTrade EventType = "PumpSwapTrade"
)

// EventTypeFilter 事件类型过滤器接口
type EventTypeFilter interface {
	ShouldInclude(eventType EventType) bool
}

// IncludeOnlyFilter 仅包含指定类型的事件过滤器
type IncludeOnlyFilter struct {
	IncludeOnly []EventType
}

// ShouldInclude 判断是否包含指定事件类型
func (f *IncludeOnlyFilter) ShouldInclude(eventType EventType) bool {
	// 空列表表示包含所有类型
	if len(f.IncludeOnly) == 0 {
		return true
	}
	for _, t := range f.IncludeOnly {
		if t == eventType {
			return true
		}
		// PumpFunTrade 包含 PumpFunBuy, PumpFunSell, PumpFunBuyExactSolIn
		if eventType == EventTypePumpFunTrade {
			if t == EventTypePumpFunBuy || t == EventTypePumpFunSell || t == EventTypePumpFunBuyExactSolIn {
				return true
			}
		}
	}
	return false
}

// ExcludeFilter 排除指定类型的事件过滤器
type ExcludeFilter struct {
	ExcludeTypes []EventType
}

// ShouldInclude 判断是否包含指定事件类型
func (f *ExcludeFilter) ShouldInclude(eventType EventType) bool {
	for _, t := range f.ExcludeTypes {
		if t == eventType {
			return false
		}
	}
	return true
}

// EventTypeFilterIncludeOnly 创建仅包含指定类型的事件过滤器
func EventTypeFilterIncludeOnly(types []EventType) EventTypeFilter {
	return &IncludeOnlyFilter{IncludeOnly: types}
}

// EventTypeFilterExclude 创建排除指定类型的事件过滤器
func EventTypeFilterExclude(types []EventType) EventTypeFilter {
	return &ExcludeFilter{ExcludeTypes: types}
}

// EventTypeFilterIncludesPumpfun 判断过滤器是否包含 PumpFun 相关类型
func EventTypeFilterIncludesPumpfun(filter EventTypeFilter) bool {
	pumpfunTypes := []EventType{
		EventTypePumpFunTrade,
		EventTypePumpFunBuy,
		EventTypePumpFunSell,
		EventTypePumpFunBuyExactSolIn,
		EventTypePumpFunCreate,
		EventTypePumpFunCreateV2,
		EventTypePumpFunComplete,
		EventTypePumpFunMigrate,
	}
	for _, t := range pumpfunTypes {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterIncludesPumpswap 判断过滤器是否包含 PumpSwap 相关类型
func EventTypeFilterIncludesPumpswap(filter EventTypeFilter) bool {
	pumpswapTypes := []EventType{
		EventTypePumpSwapBuy,
		EventTypePumpSwapSell,
		EventTypePumpSwapCreatePool,
		EventTypePumpSwapLiquidityAdded,
		EventTypePumpSwapLiquidityRemoved,
	}
	for _, t := range pumpswapTypes {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterIncludesMeteoraDammV2 判断过滤器是否包含 Meteora DAMM V2 相关类型
func EventTypeFilterIncludesMeteoraDammV2(filter EventTypeFilter) bool {
	meteoraTypes := []EventType{
		EventTypeMeteoraDammV2Swap,
		EventTypeMeteoraDammV2AddLiquidity,
		EventTypeMeteoraDammV2CreatePosition,
		EventTypeMeteoraDammV2ClosePosition,
		EventTypeMeteoraDammV2InitializePool,
		EventTypeMeteoraDammV2RemoveLiquidity,
	}
	for _, t := range meteoraTypes {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterAllowsInstructionParsing 判断过滤器是否允许指令解析
func EventTypeFilterAllowsInstructionParsing(includeOnly []EventType) bool {
	ixTypes := []EventType{
		EventTypePumpFunMigrate,
		EventTypeMeteoraDammV2Swap,
		EventTypeMeteoraDammV2AddLiquidity,
		EventTypeMeteoraDammV2CreatePosition,
		EventTypeMeteoraDammV2ClosePosition,
		EventTypeMeteoraDammV2InitializePool,
		EventTypeMeteoraDammV2RemoveLiquidity,
	}
	for _, t := range ixTypes {
		for _, include := range includeOnly {
			if t == include {
				return true
			}
		}
	}
	return false
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	Accounts         map[string]*SubscribeRequestFilterAccounts
	Slots            map[string]*SubscribeRequestFilterSlots
	Transactions     map[string]*SubscribeRequestFilterTransactions
	TransactionsStatus map[string]*SubscribeRequestFilterTransactions
	Blocks           map[string]*SubscribeRequestFilterBlocks
	BlocksMeta       map[string]*SubscribeRequestFilterBlocksMeta
	Entry            map[string]*SubscribeRequestFilterEntry
	Commitment       *CommitmentLevel
	AccountsDataSlice []*SubscribeRequestAccountsDataSlice
	Ping             *SubscribeRequestPing
	FromSlot         *uint64
}

// SubscribeRequestFilterAccounts 账户过滤器
type SubscribeRequestFilterAccounts struct {
	Account            []string
	Owner              []string
	Filters            []*SubscribeRequestFilterAccountsFilter
	NonemptyTxnSignature *bool
}

// SubscribeRequestFilterAccountsFilter 账户过滤条件
type SubscribeRequestFilterAccountsFilter struct {
	Memcmp         *SubscribeRequestFilterAccountsFilterMemcmp
	Datasize       *uint64
	TokenAccountState *bool
	Lamports       *SubscribeRequestFilterAccountsFilterLamports
}

// SubscribeRequestFilterAccountsFilterMemcmp Memcmp 过滤器
type SubscribeRequestFilterAccountsFilterMemcmp struct {
	Offset uint64
	Bytes  []byte
	Base58 string
	Base64 string
}

// SubscribeRequestFilterAccountsFilterLamports Lamports 过滤器
type SubscribeRequestFilterAccountsFilterLamports struct {
	Eq *uint64
	Ne *uint64
	Lt *uint64
	Gt *uint64
}

// SubscribeRequestFilterSlots Slot 过滤器
type SubscribeRequestFilterSlots struct {
	FilterByCommitment *bool
	InterslotUpdates   *bool
}

// SubscribeRequestFilterTransactions 交易过滤器（proto 定义）
type SubscribeRequestFilterTransactions struct {
	Vote            *bool
	Failed          *bool
	Signature       string
	AccountInclude  []string
	AccountExclude  []string
	AccountRequired []string
}

// SubscribeRequestFilterBlocks 区块过滤器
type SubscribeRequestFilterBlocks struct {
	AccountInclude      []string
	IncludeTransactions *bool
	IncludeAccounts     *bool
	IncludeEntries      *bool
}

// SubscribeRequestFilterBlocksMeta 区块元数据过滤器
type SubscribeRequestFilterBlocksMeta struct{}

// SubscribeRequestFilterEntry Entry 过滤器
type SubscribeRequestFilterEntry struct{}

// SubscribeRequestAccountsDataSlice 账户数据切片
type SubscribeRequestAccountsDataSlice struct {
	Offset uint64
	Length uint64
}

// SubscribeRequestPing Ping 请求
type SubscribeRequestPing struct {
	ID int32
}

// SubscribeUpdate 订阅更新
type SubscribeUpdate struct {
	Filters           []string
	Account           *SubscribeUpdateAccount
	Slot              *SubscribeUpdateSlot
	Transaction       *SubscribeUpdateTransaction
	TransactionStatus *SubscribeUpdateTransactionStatus
	Block             *SubscribeUpdateBlock
	Ping              *SubscribeUpdatePing
	Pong              *SubscribeUpdatePong
	BlockMeta         *SubscribeUpdateBlockMeta
	Entry             *SubscribeUpdateEntry
	CreatedAt         *int64 // Unix timestamp in microseconds
}

// SubscribeUpdateAccount 账户更新
type SubscribeUpdateAccount struct {
	Account   *SubscribeUpdateAccountInfo
	Slot      uint64
	IsStartup bool
}

// SubscribeUpdateAccountInfo 账户信息
type SubscribeUpdateAccountInfo struct {
	Pubkey        []byte
	Lamports      uint64
	Owner         []byte
	Executable    bool
	RentEpoch     uint64
	Data          []byte
	WriteVersion  uint64
	TxnSignature  []byte
}

// SubscribeUpdateSlot Slot 更新
type SubscribeUpdateSlot struct {
	Slot      uint64
	Parent    *uint64
	Status    SlotStatus
	DeadError *string
}

// SlotStatus Slot 状态
type SlotStatus int32

const (
	SlotStatusProcessed      SlotStatus = 0
	SlotStatusConfirmed      SlotStatus = 1
	SlotStatusFinalized      SlotStatus = 2
	SlotStatusFirstShredReceived SlotStatus = 3
	SlotStatusCompleted      SlotStatus = 4
	SlotStatusCreatedBank    SlotStatus = 5
	SlotStatusDead           SlotStatus = 6
)

// SubscribeUpdateTransaction 交易更新
type SubscribeUpdateTransaction struct {
	Transaction *SubscribeUpdateTransactionInfo
	Slot        uint64
}

// SubscribeUpdateTransactionInfo 交易信息
type SubscribeUpdateTransactionInfo struct {
	Signature      []byte
	IsVote         bool
	Transaction    *pb.Transaction
	Meta           *pb.TransactionStatusMeta
	Index          uint64
}

// SubscribeUpdateTransactionStatus 交易状态更新
type SubscribeUpdateTransactionStatus struct {
	Slot      uint64
	Signature []byte
	IsVote    bool
	Index     uint64
	Err       []byte
}

// SubscribeUpdateBlock 区块更新
type SubscribeUpdateBlock struct {
	Slot                   uint64
	Blockhash              string
	ParentSlot             uint64
	ParentBlockhash        string
	ExecutedTransactionCount uint64
	Transactions           []*SubscribeUpdateTransactionInfo
}

// SubscribeUpdatePing Ping 更新
type SubscribeUpdatePing struct{}

// SubscribeUpdatePong Pong 更新
type SubscribeUpdatePong struct {
	ID int32
}

// SubscribeUpdateBlockMeta 区块元数据更新
type SubscribeUpdateBlockMeta struct {
	Slot                   uint64
	Blockhash              string
	ParentSlot             uint64
	ParentBlockhash        string
	ExecutedTransactionCount uint64
}

// SubscribeUpdateEntry Entry 更新
type SubscribeUpdateEntry struct {
	Slot                   uint64
	Index                  uint64
	NumHashes              uint64
	Hash                   []byte
	ExecutedTransactionCount uint64
	StartingTransactionIndex uint64
}

// GetLatestBlockhashRequest 获取最新区块哈希请求
type GetLatestBlockhashRequest struct {
	Commitment *CommitmentLevel
}

// GetLatestBlockhashResponse 获取最新区块哈希响应
type GetLatestBlockhashResponse struct {
	Slot               uint64
	Blockhash          string
	LastValidBlockHeight uint64
}

// GetBlockHeightRequest 获取区块高度请求
type GetBlockHeightRequest struct {
	Commitment *CommitmentLevel
}

// GetBlockHeightResponse 获取区块高度响应
type GetBlockHeightResponse struct {
	BlockHeight uint64
}

// GetSlotRequest 获取 Slot 请求
type GetSlotRequest struct {
	Commitment *CommitmentLevel
}

// GetSlotResponse 获取 Slot 响应
type GetSlotResponse struct {
	Slot uint64
}

// GetVersionRequest 获取版本请求
type GetVersionRequest struct{}

// GetVersionResponse 获取版本响应
type GetVersionResponse struct {
	Version string
}

// IsBlockhashValidRequest 验证区块哈希请求
type IsBlockhashValidRequest struct {
	Blockhash  string
	Commitment *CommitmentLevel
}

// IsBlockhashValidResponse 验证区块哈希响应
type IsBlockhashValidResponse struct {
	Slot  uint64
	Valid bool
}

// PingRequest Ping 请求
type PingRequest struct {
	Count int32
}

// PongResponse Pong 响应
type PongResponse struct {
	Count int32
}

// SubscribeReplayInfoRequest 订阅重放信息请求
type SubscribeReplayInfoRequest struct{}

// SubscribeReplayInfoResponse 订阅重放信息响应
type SubscribeReplayInfoResponse struct {
	FirstAvailable *uint64
}

// EventTypeFilterIncludesOrcaWhirlpool 判断过滤器是否包含 Orca Whirlpool 相关类型
func EventTypeFilterIncludesOrcaWhirlpool(filter EventTypeFilter) bool {
	for _, t := range []EventType{
		EventTypeOrcaWhirlpoolSwap, EventTypeOrcaWhirlpoolLiquidityIncreased,
		EventTypeOrcaWhirlpoolLiquidityDecreased, EventTypeOrcaWhirlpoolPoolInitialized,
	} {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterIncludesBonk 判断过滤器是否包含 Bonk 相关类型
func EventTypeFilterIncludesBonk(filter EventTypeFilter) bool {
	for _, t := range []EventType{
		EventTypeBonkTrade, EventTypeBonkPoolCreate, EventTypeBonkMigrateAmm,
	} {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterIncludesRaydiumClmm 判断过滤器是否包含 Raydium CLMM 相关类型
func EventTypeFilterIncludesRaydiumClmm(filter EventTypeFilter) bool {
	for _, t := range []EventType{
		EventTypeRaydiumClmmSwap, EventTypeRaydiumClmmIncreaseLiquidity,
		EventTypeRaydiumClmmDecreaseLiquidity, EventTypeRaydiumClmmCreatePool,
		EventTypeRaydiumClmmOpenPosition, EventTypeRaydiumClmmOpenPositionWithTokenExtNft,
		EventTypeRaydiumClmmClosePosition, EventTypeRaydiumClmmCollectFee,
	} {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterIncludesRaydiumCpmm 判断过滤器是否包含 Raydium CPMM 相关类型
func EventTypeFilterIncludesRaydiumCpmm(filter EventTypeFilter) bool {
	for _, t := range []EventType{
		EventTypeRaydiumCpmmSwap, EventTypeRaydiumCpmmDeposit,
		EventTypeRaydiumCpmmWithdraw, EventTypeRaydiumCpmmInitialize,
	} {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}

// EventTypeFilterIncludesRaydiumAmmV4 判断过滤器是否包含 Raydium AMM V4 相关类型
func EventTypeFilterIncludesRaydiumAmmV4(filter EventTypeFilter) bool {
	for _, t := range []EventType{
		EventTypeRaydiumAmmV4Swap, EventTypeRaydiumAmmV4Deposit,
		EventTypeRaydiumAmmV4Withdraw, EventTypeRaydiumAmmV4WithdrawPnl,
		EventTypeRaydiumAmmV4Initialize2,
	} {
		if filter.ShouldInclude(t) {
			return true
		}
	}
	return false
}
