package solparser

import (
	"sort"
)

// rpcIndexedEvent 与 Rust `parse_instructions_enhanced` 中 (outer_idx, inner_idx, DexEvent) 对应。
type rpcIndexedEvent struct {
	OuterIdx       int
	InnerIdx       *int // nil identifies an outer instruction.
	StackHeight    *uint32
	IsDlmmEventCPI bool
	Event          DexEvent
}

func uint32Ptr(value uint32) *uint32 { return &value }

func isDlmmEvent(event DexEvent) bool {
	switch event.Type {
	case EventTypeMeteoraDlmmSwap,
		EventTypeMeteoraDlmmAddLiquidity,
		EventTypeMeteoraDlmmRemoveLiquidity,
		EventTypeMeteoraDlmmInitializePool,
		EventTypeMeteoraDlmmInitializeBinArray,
		EventTypeMeteoraDlmmCreatePosition,
		EventTypeMeteoraDlmmClosePosition,
		EventTypeMeteoraDlmmClaimFee:
		return true
	default:
		return false
	}
}

// mergeRpcInstructionEvents 合并同一 outer_idx 下的外层与内层事件（对齐 Rust `merge_instruction_events`）。
// 排序键：(outerIdx, 0 为外层，1+innerJ 为内层) 以保证 **外层先于同槽内层**（修正仅按 unwrap_or(MAX) 时内层会排在前面问题）。
func mergeRpcInstructionEvents(events []rpcIndexedEvent) []DexEvent {
	if len(events) == 0 {
		return nil
	}
	sort.Slice(events, func(i, j int) bool {
		ai, aj := events[i].OuterIdx, events[j].OuterIdx
		if ai != aj {
			return ai < aj
		}
		si := secondaryMergeKey(events[i].InnerIdx)
		sj := secondaryMergeKey(events[j].InnerIdx)
		return si < sj
	})

	out := make([]DexEvent, 0, len(events))
	outerTargetIdx := -1
	outerTargetOuter := -1
	type dlmmTarget struct {
		outerIdx    int
		stackHeight *uint32
		resultIdx   int
	}
	var dlmmTargets [8]dlmmTarget
	dlmmTargetsLen := 0

	for _, e := range events {
		if e.InnerIdx == nil {
			out = append(out, e.Event)
			outerTargetIdx = len(out) - 1
			outerTargetOuter = e.OuterIdx
			dlmmTargetsLen = 0
			if isDlmmEvent(e.Event) {
				dlmmTargets[0] = dlmmTarget{e.OuterIdx, e.StackHeight, outerTargetIdx}
				dlmmTargetsLen = 1
			}
			continue
		}

		if e.IsDlmmEventCPI {
			merged := false
			for i := dlmmTargetsLen - 1; i >= 0; i-- {
				target := dlmmTargets[i]
				directChild := target.stackHeight == nil || e.StackHeight == nil ||
					*e.StackHeight == *target.stackHeight+1
				if target.outerIdx == e.OuterIdx && directChild {
					dlmmTargetsLen = i + 1
					merged = tryMergeDexEvents(&out[target.resultIdx], e.Event)
					break
				}
			}
			if !merged {
				out = append(out, e.Event)
			}
			continue
		}

		targetIdx := -1
		if outerTargetOuter == e.OuterIdx && outerTargetIdx >= 0 &&
			tryMergeDexEvents(&out[outerTargetIdx], e.Event) {
			targetIdx = outerTargetIdx
		} else {
			out = append(out, e.Event)
			targetIdx = len(out) - 1
		}

		if isDlmmEvent(e.Event) {
			if e.StackHeight == nil {
				dlmmTargetsLen = 0
			} else {
				for dlmmTargetsLen > 0 {
					last := dlmmTargets[dlmmTargetsLen-1]
					if last.outerIdx != e.OuterIdx ||
						(last.stackHeight != nil && *last.stackHeight >= *e.StackHeight) {
						dlmmTargetsLen--
					} else {
						break
					}
				}
			}
			if dlmmTargetsLen < len(dlmmTargets) {
				dlmmTargets[dlmmTargetsLen] = dlmmTarget{e.OuterIdx, e.StackHeight, targetIdx}
				dlmmTargetsLen++
			}
		}
	}
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
	tryMergeDexEvents(base, inner)
}

func tryMergeDexEvents(base *DexEvent, inner DexEvent) bool {
	if base == nil || base.Type == "" || inner.Type == "" {
		return false
	}
	switch base.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn:
		if inner.Type != EventTypePumpFunTrade && inner.Type != EventTypePumpFunBuy &&
			inner.Type != EventTypePumpFunSell && inner.Type != EventTypePumpFunBuyExactSolIn {
			return false
		}
		b, ok1 := base.Data.(*PumpFunTradeEvent)
		i, ok2 := inner.Data.(*PumpFunTradeEvent)
		if !ok1 || !ok2 {
			return false
		}
		mergePumpfunTrade(b, i)
		return true
	case EventTypePumpFunCreate:
		if inner.Type != EventTypePumpFunCreate {
			return false
		}
		b, ok1 := base.Data.(*PumpFunCreateEvent)
		i, ok2 := inner.Data.(*PumpFunCreateEvent)
		if !ok1 || !ok2 {
			return false
		}
		mergePumpfunCreate(b, i)
		return true
	case EventTypePumpFunCreateV2:
		if inner.Type != EventTypePumpFunCreateV2 {
			return false
		}
		b, ok1 := base.Data.(*PumpFunCreateV2TokenEvent)
		i, ok2 := inner.Data.(*PumpFunCreateV2TokenEvent)
		if !ok1 || !ok2 {
			return false
		}
		mergePumpfunCreateV2(b, i)
		return true
	case EventTypePumpFunMigrate:
		if inner.Type != EventTypePumpFunMigrate {
			return false
		}
		b, ok1 := base.Data.(*PumpFunMigrateEvent)
		i, ok2 := inner.Data.(*PumpFunMigrateEvent)
		if !ok1 || !ok2 {
			return false
		}
		mergePumpfunMigrate(b, i)
		return true
	case EventTypePumpSwapBuy:
		if inner.Type != EventTypePumpSwapBuy {
			return false
		}
		b, ok1 := base.Data.(*PumpSwapBuyEvent)
		i, ok2 := inner.Data.(*PumpSwapBuyEvent)
		if !ok1 || !ok2 {
			return false
		}
		supplementPumpSwapBuy(b, i)
		return true
	case EventTypePumpSwapSell:
		if inner.Type != EventTypePumpSwapSell {
			return false
		}
		b, ok1 := base.Data.(*PumpSwapSellEvent)
		i, ok2 := inner.Data.(*PumpSwapSellEvent)
		if !ok1 || !ok2 {
			return false
		}
		supplementPumpSwapSell(b, i)
		return true
	case EventTypePumpSwapCreatePool:
		if inner.Type != EventTypePumpSwapCreatePool {
			return false
		}
		b, ok1 := base.Data.(*PumpSwapCreatePoolEvent)
		i, ok2 := inner.Data.(*PumpSwapCreatePoolEvent)
		if !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypePumpSwapLiquidityAdded:
		if inner.Type != EventTypePumpSwapLiquidityAdded {
			return false
		}
		b, ok1 := base.Data.(*PumpSwapLiquidityAddedEvent)
		i, ok2 := inner.Data.(*PumpSwapLiquidityAddedEvent)
		if !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypePumpSwapLiquidityRemoved:
		if inner.Type != EventTypePumpSwapLiquidityRemoved {
			return false
		}
		b, ok1 := base.Data.(*PumpSwapLiquidityRemovedEvent)
		i, ok2 := inner.Data.(*PumpSwapLiquidityRemovedEvent)
		if !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypeMeteoraDlmmSwap:
		b, ok1 := base.Data.(*MeteoraDlmmSwapEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmSwapEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypeMeteoraDlmmAddLiquidity:
		b, ok1 := base.Data.(*MeteoraDlmmAddLiquidityEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmAddLiquidityEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypeMeteoraDlmmRemoveLiquidity:
		b, ok1 := base.Data.(*MeteoraDlmmRemoveLiquidityEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmRemoveLiquidityEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypeMeteoraDlmmInitializePool:
		b, ok1 := base.Data.(*MeteoraDlmmInitializePoolEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmInitializePoolEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		creator, activeBinID := b.Creator, b.ActiveBinID
		*b = *i
		b.Creator, b.ActiveBinID = creator, activeBinID
		return true
	case EventTypeMeteoraDlmmInitializeBinArray:
		b, ok1 := base.Data.(*MeteoraDlmmInitializeBinArrayEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmInitializeBinArrayEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	case EventTypeMeteoraDlmmCreatePosition:
		b, ok1 := base.Data.(*MeteoraDlmmCreatePositionEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmCreatePositionEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		lowerBinID, width := b.LowerBinID, b.Width
		*b = *i
		b.LowerBinID, b.Width = lowerBinID, width
		return true
	case EventTypeMeteoraDlmmClosePosition:
		b, ok1 := base.Data.(*MeteoraDlmmClosePositionEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmClosePositionEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		pool := b.Pool
		*b = *i
		b.Pool = pool
		return true
	case EventTypeMeteoraDlmmClaimFee:
		b, ok1 := base.Data.(*MeteoraDlmmClaimFeeEvent)
		i, ok2 := inner.Data.(*MeteoraDlmmClaimFeeEvent)
		if inner.Type != base.Type || !ok1 || !ok2 {
			return false
		}
		*b = *i
		return true
	}
	return false
}

func mergePumpfunTrade(base, inner *PumpFunTradeEvent) {
	leg := inner.SolAmount != 0 || inner.TokenAmount != 0

	putStringIfSet(&base.Mint, inner.Mint)
	putStringIfSet(&base.User, inner.User)
	putStringIfSet(&base.FeeRecipient, inner.FeeRecipient)
	putStringIfSet(&base.Creator, inner.Creator)

	if leg {
		base.SolAmount = inner.SolAmount
		base.TokenAmount = inner.TokenAmount
		base.IsBuy = inner.IsBuy
		base.Timestamp = inner.Timestamp
		base.VirtualSolReserves = inner.VirtualSolReserves
		base.VirtualTokenReserves = inner.VirtualTokenReserves
		base.RealSolReserves = inner.RealSolReserves
		base.RealTokenReserves = inner.RealTokenReserves
		base.FeeBasisPoints = inner.FeeBasisPoints
		base.Fee = inner.Fee
		base.CreatorFeeBasisPoints = inner.CreatorFeeBasisPoints
		base.CreatorFee = inner.CreatorFee
		base.TrackVolume = inner.TrackVolume
		base.TotalUnclaimedTokens = inner.TotalUnclaimedTokens
		base.TotalClaimedTokens = inner.TotalClaimedTokens
		base.CurrentSolVolume = inner.CurrentSolVolume
		base.LastUpdateTimestamp = inner.LastUpdateTimestamp
		putStringIfSet(&base.IxName, inner.IxName)
		base.MayhemMode = inner.MayhemMode
		putUint64IfNonzero(&base.CashbackFeeBasisPoints, inner.CashbackFeeBasisPoints)
		putUint64IfNonzero(&base.Cashback, inner.Cashback)
		putUint64IfNonzero(&base.BuybackFeeBasisPoints, inner.BuybackFeeBasisPoints)
		putUint64IfNonzero(&base.BuybackFee, inner.BuybackFee)
		if len(base.Shareholders) == 0 && len(inner.Shareholders) != 0 {
			base.Shareholders = inner.Shareholders
		}
		putStringIfSet(&base.QuoteMint, inner.QuoteMint)
		putUint64IfNonzero(&base.QuoteAmount, inner.QuoteAmount)
		putUint64IfNonzero(&base.VirtualQuoteReserves, inner.VirtualQuoteReserves)
		putUint64IfNonzero(&base.RealQuoteReserves, inner.RealQuoteReserves)
		base.IsCashbackCoin = inner.IsCashbackCoin
	} else {
		putUint64IfNonzero(&base.Fee, inner.Fee)
		putUint64IfNonzero(&base.CreatorFee, inner.CreatorFee)
		putUint64IfNonzero(&base.FeeBasisPoints, inner.FeeBasisPoints)
		putUint64IfNonzero(&base.CreatorFeeBasisPoints, inner.CreatorFeeBasisPoints)
		putUint64IfNonzero(&base.VirtualSolReserves, inner.VirtualSolReserves)
		putUint64IfNonzero(&base.VirtualTokenReserves, inner.VirtualTokenReserves)
		putUint64IfNonzero(&base.RealSolReserves, inner.RealSolReserves)
		putUint64IfNonzero(&base.RealTokenReserves, inner.RealTokenReserves)
		putUint64IfNonzero(&base.TotalUnclaimedTokens, inner.TotalUnclaimedTokens)
		putUint64IfNonzero(&base.TotalClaimedTokens, inner.TotalClaimedTokens)
		putUint64IfNonzero(&base.CurrentSolVolume, inner.CurrentSolVolume)
		putUint64IfNonzero(&base.CashbackFeeBasisPoints, inner.CashbackFeeBasisPoints)
		putUint64IfNonzero(&base.Cashback, inner.Cashback)
		putUint64IfNonzero(&base.BuybackFeeBasisPoints, inner.BuybackFeeBasisPoints)
		putUint64IfNonzero(&base.BuybackFee, inner.BuybackFee)
		if len(base.Shareholders) == 0 && len(inner.Shareholders) != 0 {
			base.Shareholders = inner.Shareholders
		}
		putStringIfSet(&base.QuoteMint, inner.QuoteMint)
		putUint64IfNonzero(&base.QuoteAmount, inner.QuoteAmount)
		putUint64IfNonzero(&base.VirtualQuoteReserves, inner.VirtualQuoteReserves)
		putUint64IfNonzero(&base.RealQuoteReserves, inner.RealQuoteReserves)
		putInt64IfNonzero(&base.Timestamp, inner.Timestamp)
		putInt64IfNonzero(&base.LastUpdateTimestamp, inner.LastUpdateTimestamp)
		putStringIfSet(&base.IxName, inner.IxName)
		if inner.TrackVolume {
			base.TrackVolume = true
		}
		if inner.MayhemMode {
			base.MayhemMode = true
		}
		if inner.IsCashbackCoin {
			base.IsCashbackCoin = true
		}
	}

	putUint64IfNonzero(&base.Amount, inner.Amount)
	putUint64IfNonzero(&base.MaxSolCost, inner.MaxSolCost)
	putUint64IfNonzero(&base.MinSolOutput, inner.MinSolOutput)
	putUint64IfNonzero(&base.SpendableSolIn, inner.SpendableSolIn)
	putUint64IfNonzero(&base.SpendableQuoteIn, inner.SpendableQuoteIn)
	putUint64IfNonzero(&base.MinTokensOut, inner.MinTokensOut)
	if inner.IsCreatedBuy {
		base.IsCreatedBuy = true
	}
	supplementPumpfunTradeAccountFields(base, inner)
}

func putStringIfSet(dst *string, src string) {
	if src != "" && src != zeroPubkey {
		*dst = src
	}
}

func putUint64IfNonzero(dst *uint64, src uint64) {
	if src != 0 {
		*dst = src
	}
}

func putInt64IfNonzero(dst *int64, src int64) {
	if src != 0 {
		*dst = src
	}
}

// supplementPumpfunTradeAccountFields 合并外层/内层时补全 bonding_curve 等（内层常带完整指令账户）。
func supplementPumpfunTradeAccountFields(base, inner *PumpFunTradeEvent) {
	pick := func(dst *string, src string) {
		if (*dst == "" || *dst == zeroPubkey) && src != "" && src != zeroPubkey {
			*dst = src
		}
	}
	pick(&base.BondingCurve, inner.BondingCurve)
	pick(&base.BondingCurveV2, inner.BondingCurveV2)
	pick(&base.AssociatedBondingCurve, inner.AssociatedBondingCurve)
	pick(&base.AssociatedUser, inner.AssociatedUser)
	pick(&base.SystemProgram, inner.SystemProgram)
	pick(&base.TokenProgram, inner.TokenProgram)
	pick(&base.QuoteTokenProgram, inner.QuoteTokenProgram)
	pick(&base.AssociatedTokenProgram, inner.AssociatedTokenProgram)
	pick(&base.CreatorVault, inner.CreatorVault)
	pick(&base.Global, inner.Global)
	pick(&base.QuoteMint, inner.QuoteMint)
	pick(&base.AssociatedQuoteFeeRecipient, inner.AssociatedQuoteFeeRecipient)
	pick(&base.BuybackFeeRecipient, inner.BuybackFeeRecipient)
	pick(&base.AssociatedQuoteBuybackFeeRecipient, inner.AssociatedQuoteBuybackFeeRecipient)
	pick(&base.AssociatedQuoteBondingCurve, inner.AssociatedQuoteBondingCurve)
	pick(&base.AssociatedQuoteUser, inner.AssociatedQuoteUser)
	pick(&base.AssociatedCreatorVault, inner.AssociatedCreatorVault)
	pick(&base.SharingConfig, inner.SharingConfig)
	pick(&base.EventAuthority, inner.EventAuthority)
	pick(&base.Program, inner.Program)
	pick(&base.GlobalVolumeAccumulator, inner.GlobalVolumeAccumulator)
	pick(&base.UserVolumeAccumulator, inner.UserVolumeAccumulator)
	pick(&base.AssociatedUserVolumeAccumulator, inner.AssociatedUserVolumeAccumulator)
	pick(&base.FeeConfig, inner.FeeConfig)
	pick(&base.FeeProgram, inner.FeeProgram)
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
	base.IsCashbackEnabled = inner.IsCashbackEnabled
	base.QuoteMint = inner.QuoteMint
	base.QuoteVault = inner.QuoteVault
	base.QuoteTokenProgram = inner.QuoteTokenProgram
	base.VirtualQuoteReserves = inner.VirtualQuoteReserves
}

func mergePumpfunCreateV2(base, inner *PumpFunCreateV2TokenEvent) {
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
	base.IsCashbackEnabled = inner.IsCashbackEnabled
	base.QuoteMint = inner.QuoteMint
	base.QuoteVault = inner.QuoteVault
	base.QuoteTokenProgram = inner.QuoteTokenProgram
	base.VirtualQuoteReserves = inner.VirtualQuoteReserves
	base.MintAuthority = inner.MintAuthority
	base.AssociatedBondingCurve = inner.AssociatedBondingCurve
	base.Global = inner.Global
	base.SystemProgram = inner.SystemProgram
	base.AssociatedTokenProgram = inner.AssociatedTokenProgram
	base.MayhemProgramID = inner.MayhemProgramID
	base.GlobalParams = inner.GlobalParams
	base.SolVault = inner.SolVault
	base.MayhemState = inner.MayhemState
	base.MayhemTokenVault = inner.MayhemTokenVault
	base.EventAuthority = inner.EventAuthority
	base.Program = inner.Program
	base.ObservedFeeRecipient = inner.ObservedFeeRecipient
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
