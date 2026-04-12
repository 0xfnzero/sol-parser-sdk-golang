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
	case *PumpFunCreateV2TokenEvent:
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

// SetRecentBlockhash 写入事件的 RecentBlockhash（与 Rust `parse_instructions_enhanced` 在 merge 后填充一致）。
func (e *DexEvent) SetRecentBlockhash(h string) {
	if e == nil || h == "" {
		return
	}
	switch d := e.Data.(type) {
	case *PumpFunTradeEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpFunCreateEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpFunCreateV2TokenEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpFunMigrateEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapBuyEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapSellEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapCreatePoolEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapLiquidityAddedEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapLiquidityRemovedEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumClmmSwapEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumClmmIncreaseLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumClmmDecreaseLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumClmmCreatePoolEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumClmmCollectFeeEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumCpmmSwapEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumCpmmDepositEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumCpmmWithdrawEvent:
		d.Metadata.RecentBlockhash = h
	case *RaydiumCpmmInitializeEvent:
		d.Metadata.RecentBlockhash = h
	case *OrcaWhirlpoolSwapEvent:
		d.Metadata.RecentBlockhash = h
	case *OrcaWhirlpoolLiquidityIncreasedEvent:
		d.Metadata.RecentBlockhash = h
	case *OrcaWhirlpoolLiquidityDecreasedEvent:
		d.Metadata.RecentBlockhash = h
	case *OrcaWhirlpoolPoolInitializedEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmSwapEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmAddLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmRemoveLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmInitializePoolEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmInitializeBinArrayEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmCreatePositionEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmClosePositionEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDlmmClaimFeeEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraPoolsSwapEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraPoolsAddLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraPoolsRemoveLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraPoolsBootstrapLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraPoolsPoolCreatedEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraPoolsSetPoolFeesEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDammV2SwapEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDammV2CreatePositionEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDammV2ClosePositionEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDammV2AddLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDammV2RemoveLiquidityEvent:
		d.Metadata.RecentBlockhash = h
	case *MeteoraDammV2InitializePoolEvent:
		d.Metadata.RecentBlockhash = h
	case *BonkTradeEvent:
		d.Metadata.RecentBlockhash = h
	case *BonkPoolCreateEvent:
		d.Metadata.RecentBlockhash = h
	case *BonkMigrateAmmEvent:
		d.Metadata.RecentBlockhash = h
	case *TokenInfoEvent:
		d.Metadata.RecentBlockhash = h
	case *TokenAccountEvent:
		d.Metadata.RecentBlockhash = h
	case *NonceAccountEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapGlobalConfigAccountEvent:
		d.Metadata.RecentBlockhash = h
	case *PumpSwapPoolAccountEvent:
		d.Metadata.RecentBlockhash = h
	default:
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

// AsPumpFunCreateV2 返回 PumpFunCreateV2TokenEvent
func (e DexEvent) AsPumpFunCreateV2() *PumpFunCreateV2TokenEvent {
	if p, ok := e.Data.(*PumpFunCreateV2TokenEvent); ok {
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
