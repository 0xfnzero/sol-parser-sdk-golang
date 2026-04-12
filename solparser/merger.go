package solparser

import (
	"sort"
)

// rpcIndexedEvent 与 Rust `parse_instructions_enhanced` 中 (outer_idx, inner_idx, DexEvent) 对应。
type rpcIndexedEvent struct {
	OuterIdx int
	InnerIdx *int // nil 表示外层指令
	Event    DexEvent
}

// mergeRpcInstructionEvents 合并同一 outer_idx 下的外层与内层事件（对齐 Rust `merge_instruction_events`）。
// 排序键：(outerIdx, 0 为外层，1+innerJ 为内层) 以保证 **外层先于同槽内层**（修正仅按 unwrap_or(MAX) 时内层会排在前面问题）。
func mergeRpcInstructionEvents(events []rpcIndexedEvent) []DexEvent {
	if len(events) == 0 {
		return nil
	}
	sort.SliceStable(events, func(i, j int) bool {
		ai, aj := events[i].OuterIdx, events[j].OuterIdx
		if ai != aj {
			return ai < aj
		}
		si := secondaryMergeKey(events[i].InnerIdx)
		sj := secondaryMergeKey(events[j].InnerIdx)
		return si < sj
	})

	out := make([]DexEvent, 0, len(events))
	var pendingOuter *DexEvent
	var pendingOuterIdx int

	flushPending := func() {
		if pendingOuter != nil {
			out = append(out, *pendingOuter)
			pendingOuter = nil
		}
	}

	for _, e := range events {
		if e.InnerIdx == nil {
			flushPending()
			pendingOuterIdx = e.OuterIdx
			ev := e.Event
			pendingOuter = &ev
			continue
		}
		if pendingOuter != nil && pendingOuterIdx == e.OuterIdx {
			mergeDexEvents(pendingOuter, e.Event)
			out = append(out, *pendingOuter)
			pendingOuter = nil
		} else {
			flushPending()
			out = append(out, e.Event)
		}
	}
	flushPending()
	return out
}

func secondaryMergeKey(inner *int) int {
	if inner == nil {
		return 0
	}
	return 1 + *inner
}

// mergeDexEvents 对齐 Rust `core::merger::merge_events`（Pump 系子集）。
func mergeDexEvents(base *DexEvent, inner DexEvent) {
	if base == nil || base.Type == "" || inner.Type == "" {
		return
	}
	switch base.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn:
		if inner.Type != EventTypePumpFunTrade && inner.Type != EventTypePumpFunBuy &&
			inner.Type != EventTypePumpFunSell && inner.Type != EventTypePumpFunBuyExactSolIn {
			return
		}
		b, ok1 := base.Data.(*PumpFunTradeEvent)
		i, ok2 := inner.Data.(*PumpFunTradeEvent)
		if !ok1 || !ok2 {
			return
		}
		mergePumpfunTrade(b, i)
	case EventTypePumpFunCreate:
		if inner.Type != EventTypePumpFunCreate {
			return
		}
		b, ok1 := base.Data.(*PumpFunCreateEvent)
		i, ok2 := inner.Data.(*PumpFunCreateEvent)
		if !ok1 || !ok2 {
			return
		}
		mergePumpfunCreate(b, i)
	case EventTypePumpFunMigrate:
		if inner.Type != EventTypePumpFunMigrate {
			return
		}
		b, ok1 := base.Data.(*PumpFunMigrateEvent)
		i, ok2 := inner.Data.(*PumpFunMigrateEvent)
		if !ok1 || !ok2 {
			return
		}
		mergePumpfunMigrate(b, i)
	case EventTypePumpSwapBuy:
		if inner.Type != EventTypePumpSwapBuy {
			return
		}
		b, ok1 := base.Data.(*PumpSwapBuyEvent)
		i, ok2 := inner.Data.(*PumpSwapBuyEvent)
		if !ok1 || !ok2 {
			return
		}
		supplementPumpSwapBuy(b, i)
	case EventTypePumpSwapSell:
		if inner.Type != EventTypePumpSwapSell {
			return
		}
		b, ok1 := base.Data.(*PumpSwapSellEvent)
		i, ok2 := inner.Data.(*PumpSwapSellEvent)
		if !ok1 || !ok2 {
			return
		}
		supplementPumpSwapSell(b, i)
	case EventTypePumpSwapCreatePool:
		if inner.Type != EventTypePumpSwapCreatePool {
			return
		}
		b, ok1 := base.Data.(*PumpSwapCreatePoolEvent)
		i, ok2 := inner.Data.(*PumpSwapCreatePoolEvent)
		if !ok1 || !ok2 {
			return
		}
		*b = *i
	case EventTypePumpSwapLiquidityAdded:
		if inner.Type != EventTypePumpSwapLiquidityAdded {
			return
		}
		b, ok1 := base.Data.(*PumpSwapLiquidityAddedEvent)
		i, ok2 := inner.Data.(*PumpSwapLiquidityAddedEvent)
		if !ok1 || !ok2 {
			return
		}
		*b = *i
	case EventTypePumpSwapLiquidityRemoved:
		if inner.Type != EventTypePumpSwapLiquidityRemoved {
			return
		}
		b, ok1 := base.Data.(*PumpSwapLiquidityRemovedEvent)
		i, ok2 := inner.Data.(*PumpSwapLiquidityRemovedEvent)
		if !ok1 || !ok2 {
			return
		}
		*b = *i
	}
}

func mergePumpfunTrade(base, inner *PumpFunTradeEvent) {
	base.Mint = inner.Mint
	base.SolAmount = inner.SolAmount
	base.TokenAmount = inner.TokenAmount
	base.IsBuy = inner.IsBuy
	base.User = inner.User
	base.Timestamp = inner.Timestamp
	base.VirtualSolReserves = inner.VirtualSolReserves
	base.VirtualTokenReserves = inner.VirtualTokenReserves
	base.RealSolReserves = inner.RealSolReserves
	base.RealTokenReserves = inner.RealTokenReserves
	base.FeeRecipient = inner.FeeRecipient
	base.FeeBasisPoints = inner.FeeBasisPoints
	base.Fee = inner.Fee
	base.Creator = inner.Creator
	base.CreatorFeeBasisPoints = inner.CreatorFeeBasisPoints
	base.CreatorFee = inner.CreatorFee
	base.TrackVolume = inner.TrackVolume
	base.TotalUnclaimedTokens = inner.TotalUnclaimedTokens
	base.TotalClaimedTokens = inner.TotalClaimedTokens
	base.CurrentSolVolume = inner.CurrentSolVolume
	base.LastUpdateTimestamp = inner.LastUpdateTimestamp
	base.IxName = inner.IxName
	base.IsCreatedBuy = inner.IsCreatedBuy
	base.MayhemMode = inner.MayhemMode
	base.CashbackFeeBasisPoints = inner.CashbackFeeBasisPoints
	base.Cashback = inner.Cashback
	base.IsCashbackCoin = inner.IsCashbackCoin
	supplementPumpfunTradeAccountFields(base, inner)
}

// supplementPumpfunTradeAccountFields 合并外层/内层时补全 bonding_curve 等（内层常带完整指令账户）。
func supplementPumpfunTradeAccountFields(base, inner *PumpFunTradeEvent) {
	pick := func(dst *string, src string) {
		if (*dst == "" || *dst == zeroPubkey) && src != "" && src != zeroPubkey {
			*dst = src
		}
	}
	pick(&base.BondingCurve, inner.BondingCurve)
	pick(&base.AssociatedBondingCurve, inner.AssociatedBondingCurve)
	pick(&base.TokenProgram, inner.TokenProgram)
	pick(&base.CreatorVault, inner.CreatorVault)
}

func mergePumpfunCreate(base, inner *PumpFunCreateEvent) {
	base.Name = inner.Name
	base.Symbol = inner.Symbol
	base.Uri = inner.Uri
	base.Mint = inner.Mint
	base.BondingCurve = inner.BondingCurve
	base.User = inner.User
	base.Creator = inner.Creator
	base.Timestamp = inner.Timestamp
	base.VirtualTokenReserves = inner.VirtualTokenReserves
	base.VirtualSolReserves = inner.VirtualSolReserves
	base.RealTokenReserves = inner.RealTokenReserves
	base.TokenTotalSupply = inner.TokenTotalSupply
	base.TokenProgram = inner.TokenProgram
	base.IsMayhemMode = inner.IsMayhemMode
}

func mergePumpfunMigrate(base, inner *PumpFunMigrateEvent) {
	base.User = inner.User
	base.Mint = inner.Mint
	base.MintAmount = inner.MintAmount
	base.SolAmount = inner.SolAmount
	base.PoolMigrationFee = inner.PoolMigrationFee
	base.BondingCurve = inner.BondingCurve
	base.Timestamp = inner.Timestamp
	base.Pool = inner.Pool
}
