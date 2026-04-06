package solparser

import "encoding/json"

// DexEvent 是强类型的 DEX 事件容器
type DexEvent struct {
	Type EventType
	Data any // 具体的事件类型，如 *PumpFunTradeEvent
}

// GetMetadata 返回事件元数据
func (e DexEvent) GetMetadata() EventMetadata {
	switch d := e.Data.(type) {
	case *PumpFunTradeEvent:
		return d.Metadata
	case *PumpFunCreateEvent:
		return d.Metadata
	case *PumpFunMigrateEvent:
		return d.Metadata
	case *PumpSwapBuyEvent:
		return d.Metadata
	case *PumpSwapSellEvent:
		return d.Metadata
	case *PumpSwapCreatePoolEvent:
		return d.Metadata
	case *PumpSwapLiquidityAddedEvent:
		return d.Metadata
	case *PumpSwapLiquidityRemovedEvent:
		return d.Metadata
	case *RaydiumClmmSwapEvent:
		return d.Metadata
	case *RaydiumClmmIncreaseLiquidityEvent:
		return d.Metadata
	case *RaydiumClmmDecreaseLiquidityEvent:
		return d.Metadata
	case *RaydiumClmmCreatePoolEvent:
		return d.Metadata
	case *RaydiumClmmCollectFeeEvent:
		return d.Metadata
	case *RaydiumCpmmSwapEvent:
		return d.Metadata
	case *RaydiumCpmmDepositEvent:
		return d.Metadata
	case *RaydiumCpmmWithdrawEvent:
		return d.Metadata
	case *RaydiumCpmmInitializeEvent:
		return d.Metadata
	case *OrcaWhirlpoolSwapEvent:
		return d.Metadata
	case *OrcaWhirlpoolLiquidityIncreasedEvent:
		return d.Metadata
	case *OrcaWhirlpoolLiquidityDecreasedEvent:
		return d.Metadata
	case *OrcaWhirlpoolPoolInitializedEvent:
		return d.Metadata
	case *MeteoraDlmmSwapEvent:
		return d.Metadata
	case *MeteoraDlmmAddLiquidityEvent:
		return d.Metadata
	case *MeteoraDlmmRemoveLiquidityEvent:
		return d.Metadata
	case *MeteoraDlmmInitializePoolEvent:
		return d.Metadata
	case *MeteoraDlmmInitializeBinArrayEvent:
		return d.Metadata
	case *MeteoraDlmmCreatePositionEvent:
		return d.Metadata
	case *MeteoraDlmmClosePositionEvent:
		return d.Metadata
	case *MeteoraDlmmClaimFeeEvent:
		return d.Metadata
	case *MeteoraPoolsSwapEvent:
		return d.Metadata
	case *MeteoraPoolsAddLiquidityEvent:
		return d.Metadata
	case *MeteoraPoolsRemoveLiquidityEvent:
		return d.Metadata
	case *MeteoraPoolsBootstrapLiquidityEvent:
		return d.Metadata
	case *MeteoraPoolsPoolCreatedEvent:
		return d.Metadata
	case *MeteoraPoolsSetPoolFeesEvent:
		return d.Metadata
	case *MeteoraDammV2SwapEvent:
		return d.Metadata
	case *MeteoraDammV2CreatePositionEvent:
		return d.Metadata
	case *MeteoraDammV2ClosePositionEvent:
		return d.Metadata
	case *MeteoraDammV2AddLiquidityEvent:
		return d.Metadata
	case *MeteoraDammV2RemoveLiquidityEvent:
		return d.Metadata
	case *MeteoraDammV2InitializePoolEvent:
		return d.Metadata
	case *BonkTradeEvent:
		return d.Metadata
	case *BonkPoolCreateEvent:
		return d.Metadata
	case *BonkMigrateAmmEvent:
		return d.Metadata
	case *TokenInfoEvent:
		return d.Metadata
	case *TokenAccountEvent:
		return d.Metadata
	case *NonceAccountEvent:
		return d.Metadata
	case *PumpSwapGlobalConfigAccountEvent:
		return d.Metadata
	case *PumpSwapPoolAccountEvent:
		return d.Metadata
	default:
		return EventMetadata{}
	}
}

// MarshalJSON 实现 JSON 序列化
func (e DexEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		string(e.Type): e.Data,
	})
}

// ============================================================
// 类型断言便捷方法
// ============================================================

// AsPumpFunTrade 返回 PumpFunTradeEvent，类型不匹配返回 nil
func (e DexEvent) AsPumpFunTrade() *PumpFunTradeEvent {
	if p, ok := e.Data.(*PumpFunTradeEvent); ok {
		return p
	}
	return nil
}

// AsPumpFunCreate 返回 PumpFunCreateEvent
func (e DexEvent) AsPumpFunCreate() *PumpFunCreateEvent {
	if p, ok := e.Data.(*PumpFunCreateEvent); ok {
		return p
	}
	return nil
}

// AsPumpFunMigrate 返回 PumpFunMigrateEvent
func (e DexEvent) AsPumpFunMigrate() *PumpFunMigrateEvent {
	if p, ok := e.Data.(*PumpFunMigrateEvent); ok {
		return p
	}
	return nil
}

// AsPumpSwapBuy 返回 PumpSwapBuyEvent
func (e DexEvent) AsPumpSwapBuy() *PumpSwapBuyEvent {
	if p, ok := e.Data.(*PumpSwapBuyEvent); ok {
		return p
	}
	return nil
}

// AsPumpSwapSell 返回 PumpSwapSellEvent
func (e DexEvent) AsPumpSwapSell() *PumpSwapSellEvent {
	if p, ok := e.Data.(*PumpSwapSellEvent); ok {
		return p
	}
	return nil
}

// AsPumpSwapCreatePool 返回 PumpSwapCreatePoolEvent
func (e DexEvent) AsPumpSwapCreatePool() *PumpSwapCreatePoolEvent {
	if p, ok := e.Data.(*PumpSwapCreatePoolEvent); ok {
		return p
	}
	return nil
}

// AsRaydiumClmmSwap 返回 RaydiumClmmSwapEvent
func (e DexEvent) AsRaydiumClmmSwap() *RaydiumClmmSwapEvent {
	if p, ok := e.Data.(*RaydiumClmmSwapEvent); ok {
		return p
	}
	return nil
}

// AsRaydiumCpmmSwap 返回 RaydiumCpmmSwapEvent
func (e DexEvent) AsRaydiumCpmmSwap() *RaydiumCpmmSwapEvent {
	if p, ok := e.Data.(*RaydiumCpmmSwapEvent); ok {
		return p
	}
	return nil
}

// AsOrcaWhirlpoolSwap 返回 OrcaWhirlpoolSwapEvent
func (e DexEvent) AsOrcaWhirlpoolSwap() *OrcaWhirlpoolSwapEvent {
	if p, ok := e.Data.(*OrcaWhirlpoolSwapEvent); ok {
		return p
	}
	return nil
}

// AsMeteoraDlmmSwap 返回 MeteoraDlmmSwapEvent
func (e DexEvent) AsMeteoraDlmmSwap() *MeteoraDlmmSwapEvent {
	if p, ok := e.Data.(*MeteoraDlmmSwapEvent); ok {
		return p
	}
	return nil
}

// AsMeteoraDammV2Swap 返回 MeteoraDammV2SwapEvent
func (e DexEvent) AsMeteoraDammV2Swap() *MeteoraDammV2SwapEvent {
	if p, ok := e.Data.(*MeteoraDammV2SwapEvent); ok {
		return p
	}
	return nil
}

// AsBonkTrade 返回 BonkTradeEvent
func (e DexEvent) AsBonkTrade() *BonkTradeEvent {
	if p, ok := e.Data.(*BonkTradeEvent); ok {
		return p
	}
	return nil
}

// IsPumpFun 判断是否为 PumpFun 相关事件
func (e DexEvent) IsPumpFun() bool {
	switch e.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell,
		EventTypePumpFunBuyExactSolIn, EventTypePumpFunCreate, EventTypePumpFunMigrate:
		return true
	default:
		return false
	}
}

// IsPumpSwap 判断是否为 PumpSwap 相关事件
func (e DexEvent) IsPumpSwap() bool {
	switch e.Type {
	case EventTypePumpSwapBuy, EventTypePumpSwapSell, EventTypePumpSwapCreatePool,
		EventTypePumpSwapLiquidityAdded, EventTypePumpSwapLiquidityRemoved:
		return true
	default:
		return false
	}
}

// IsTrade 判断是否为交易事件
func (e DexEvent) IsTrade() bool {
	switch e.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn,
		EventTypePumpSwapBuy, EventTypePumpSwapSell,
		EventTypeRaydiumClmmSwap, EventTypeRaydiumCpmmSwap,
		EventTypeOrcaWhirlpoolSwap,
		EventTypeMeteoraDlmmSwap, EventTypeMeteoraPoolsSwap, EventTypeMeteoraDammV2Swap,
		EventTypeBonkTrade:
		return true
	default:
		return false
	}
}
