package solparser

type logInstrDedupKey struct {
	kind       uint8
	a          string
	b          string
	c          string
	flag       bool
	lane       uint8
	occurrence uint16
}

const (
	dedupPumpFunTrade uint8 = iota + 1
	dedupPumpFunCreate
	dedupPumpFunCreateV2
	dedupPumpFunMigrate
	dedupRaydiumLaunchlabTrade
	dedupRaydiumLaunchlabPoolCreate
	dedupRaydiumLaunchlabMigrateAmm
	dedupPumpSwapBuy
	dedupPumpSwapSell
	dedupPumpSwapCreatePool
	dedupPumpSwapLiquidityAdded
	dedupPumpSwapLiquidityRemoved
	dedupRaydiumClmmSwap
	dedupRaydiumAmmV4Swap
	dedupMeteoraDlmmSwap
)

type pumpfunLaneBase struct {
	mint  string
	user  string
	isBuy bool
	lane  uint8
}

func pumpfunIxLane(ixName string) uint8 {
	switch ixName {
	case "sell", "sell_v2":
		return 1
	case "buy_exact_sol_in", "buy_exact_quote_in", "buy_exact_quote_in_v2":
		return 2
	default:
		return 0
	}
}

func nextPumpfunOccurrence(base pumpfunLaneBase, counts map[pumpfunLaneBase]uint16) uint16 {
	occ := counts[base]
	counts[base] = occ + 1
	return occ
}

func dedupeKey(ev DexEvent, pumpfunLaneCounts map[pumpfunLaneBase]uint16) (logInstrDedupKey, bool) {
	switch ev.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn:
		t, ok := ev.Data.(*PumpFunTradeEvent)
		if !ok || t == nil {
			return logInstrDedupKey{}, false
		}
		base := pumpfunLaneBase{
			mint:  t.Mint,
			user:  t.User,
			isBuy: t.IsBuy,
			lane:  pumpfunIxLane(t.IxName),
		}
		occ := nextPumpfunOccurrence(base, pumpfunLaneCounts)
		return logInstrDedupKey{
			kind:       dedupPumpFunTrade,
			a:          base.mint,
			b:          base.user,
			flag:       base.isBuy,
			lane:       base.lane,
			occurrence: occ,
		}, true
	case EventTypePumpFunCreate:
		if c, ok := ev.Data.(*PumpFunCreateEvent); ok && c != nil {
			return logInstrDedupKey{kind: dedupPumpFunCreate, a: c.Mint}, true
		}
	case EventTypePumpFunCreateV2:
		if c, ok := ev.Data.(*PumpFunCreateV2TokenEvent); ok && c != nil {
			return logInstrDedupKey{kind: dedupPumpFunCreateV2, a: c.Mint}, true
		}
	case EventTypePumpFunMigrate:
		if m, ok := ev.Data.(*PumpFunMigrateEvent); ok && m != nil {
			return logInstrDedupKey{kind: dedupPumpFunMigrate, a: m.Mint, b: m.Pool, c: m.User}, true
		}
	case EventTypeRaydiumLaunchlabTrade:
		if t, ok := ev.Data.(*RaydiumLaunchlabTradeEvent); ok && t != nil {
			return logInstrDedupKey{kind: dedupRaydiumLaunchlabTrade, a: t.PoolState, b: t.User, flag: t.IsBuy}, true
		}
	case EventTypeRaydiumLaunchlabPoolCreate:
		if p, ok := ev.Data.(*RaydiumLaunchlabPoolCreateEvent); ok && p != nil {
			return logInstrDedupKey{kind: dedupRaydiumLaunchlabPoolCreate, a: p.PoolState}, true
		}
	case EventTypeRaydiumLaunchlabMigrateAmm:
		if m, ok := ev.Data.(*RaydiumLaunchlabMigrateAmmEvent); ok && m != nil {
			return logInstrDedupKey{kind: dedupRaydiumLaunchlabMigrateAmm, a: m.OldPool, b: m.NewPool, c: m.User}, true
		}
	case EventTypePumpSwapBuy:
		if b, ok := ev.Data.(*PumpSwapBuyEvent); ok && b != nil {
			return logInstrDedupKey{kind: dedupPumpSwapBuy, a: b.Pool, b: b.User}, true
		}
	case EventTypePumpSwapSell:
		if s, ok := ev.Data.(*PumpSwapSellEvent); ok && s != nil {
			return logInstrDedupKey{kind: dedupPumpSwapSell, a: s.Pool, b: s.User}, true
		}
	case EventTypePumpSwapCreatePool:
		if c, ok := ev.Data.(*PumpSwapCreatePoolEvent); ok && c != nil {
			return logInstrDedupKey{kind: dedupPumpSwapCreatePool, a: c.Pool, b: c.BaseMint, c: c.QuoteMint}, true
		}
	case EventTypePumpSwapLiquidityAdded:
		if a, ok := ev.Data.(*PumpSwapLiquidityAddedEvent); ok && a != nil {
			return logInstrDedupKey{kind: dedupPumpSwapLiquidityAdded, a: a.Pool, b: a.User}, true
		}
	case EventTypePumpSwapLiquidityRemoved:
		if r, ok := ev.Data.(*PumpSwapLiquidityRemovedEvent); ok && r != nil {
			return logInstrDedupKey{kind: dedupPumpSwapLiquidityRemoved, a: r.Pool, b: r.User}, true
		}
	case EventTypeRaydiumClmmSwap:
		if s, ok := ev.Data.(*RaydiumClmmSwapEvent); ok && s != nil {
			return logInstrDedupKey{kind: dedupRaydiumClmmSwap, a: s.PoolState, flag: s.ZeroForOne}, true
		}
	case EventTypeRaydiumAmmV4Swap:
		if s, ok := ev.Data.(*RaydiumAmmV4SwapEvent); ok && s != nil {
			return logInstrDedupKey{kind: dedupRaydiumAmmV4Swap, a: s.Amm}, true
		}
	case EventTypeMeteoraDlmmSwap:
		if s, ok := ev.Data.(*MeteoraDlmmSwapEvent); ok && s != nil {
			return logInstrDedupKey{kind: dedupMeteoraDlmmSwap, a: s.Pool, b: s.From, flag: s.SwapForY}, true
		}
	}
	return logInstrDedupKey{}, false
}

func fillStringIfDefault(dst *string, src string) {
	if (*dst == "" || *dst == zeroPubkey) && src != "" && src != zeroPubkey {
		*dst = src
	}
}

func fillUint64IfDefault(dst *uint64, src uint64) {
	if *dst == 0 && src != 0 {
		*dst = src
	}
}

func mergeGrpcInstructionIntoLog(log *DexEvent, ix DexEvent) {
	if log == nil {
		return
	}
	switch log.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn:
		l, ok1 := log.Data.(*PumpFunTradeEvent)
		i, ok2 := ix.Data.(*PumpFunTradeEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.FeeRecipient, i.FeeRecipient)
			fillStringIfDefault(&l.Creator, i.Creator)
			fillStringIfDefault(&l.Account, i.Account)
			supplementPumpfunTradeAccountFields(l, i)
			fillUint64IfDefault(&l.Amount, i.Amount)
			fillUint64IfDefault(&l.MaxSolCost, i.MaxSolCost)
			fillUint64IfDefault(&l.MinSolOutput, i.MinSolOutput)
			fillUint64IfDefault(&l.SpendableSolIn, i.SpendableSolIn)
			fillUint64IfDefault(&l.SpendableQuoteIn, i.SpendableQuoteIn)
			fillUint64IfDefault(&l.MinTokensOut, i.MinTokensOut)
			fillUint64IfDefault(&l.QuoteAmount, i.QuoteAmount)
			fillUint64IfDefault(&l.VirtualQuoteReserves, i.VirtualQuoteReserves)
			fillUint64IfDefault(&l.RealQuoteReserves, i.RealQuoteReserves)
			if l.IxName == "" && i.IxName != "" {
				l.IxName = i.IxName
			}
			l.IsCreatedBuy = l.IsCreatedBuy || i.IsCreatedBuy
		}
	case EventTypePumpFunCreate:
		l, ok1 := log.Data.(*PumpFunCreateEvent)
		i, ok2 := ix.Data.(*PumpFunCreateEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.Name, i.Name)
			fillStringIfDefault(&l.Symbol, i.Symbol)
			fillStringIfDefault(&l.Uri, i.Uri)
			fillStringIfDefault(&l.BondingCurve, i.BondingCurve)
			fillStringIfDefault(&l.User, i.User)
			fillStringIfDefault(&l.Creator, i.Creator)
			fillStringIfDefault(&l.TokenProgram, i.TokenProgram)
			fillStringIfDefault(&l.QuoteMint, i.QuoteMint)
			fillUint64IfDefault(&l.VirtualQuoteReserves, i.VirtualQuoteReserves)
		}
	case EventTypePumpFunCreateV2:
		l, ok1 := log.Data.(*PumpFunCreateV2TokenEvent)
		i, ok2 := ix.Data.(*PumpFunCreateV2TokenEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.Name, i.Name)
			fillStringIfDefault(&l.Symbol, i.Symbol)
			fillStringIfDefault(&l.Uri, i.Uri)
			fillStringIfDefault(&l.BondingCurve, i.BondingCurve)
			fillStringIfDefault(&l.User, i.User)
			fillStringIfDefault(&l.Creator, i.Creator)
			fillStringIfDefault(&l.TokenProgram, i.TokenProgram)
			fillStringIfDefault(&l.QuoteMint, i.QuoteMint)
			fillUint64IfDefault(&l.VirtualQuoteReserves, i.VirtualQuoteReserves)
			fillStringIfDefault(&l.MintAuthority, i.MintAuthority)
			fillStringIfDefault(&l.AssociatedBondingCurve, i.AssociatedBondingCurve)
			fillStringIfDefault(&l.Global, i.Global)
			fillStringIfDefault(&l.SystemProgram, i.SystemProgram)
			fillStringIfDefault(&l.AssociatedTokenProgram, i.AssociatedTokenProgram)
			fillStringIfDefault(&l.MayhemProgramID, i.MayhemProgramID)
			fillStringIfDefault(&l.GlobalParams, i.GlobalParams)
			fillStringIfDefault(&l.SolVault, i.SolVault)
			fillStringIfDefault(&l.MayhemState, i.MayhemState)
			fillStringIfDefault(&l.MayhemTokenVault, i.MayhemTokenVault)
			fillStringIfDefault(&l.EventAuthority, i.EventAuthority)
			fillStringIfDefault(&l.Program, i.Program)
			fillStringIfDefault(&l.ObservedFeeRecipient, i.ObservedFeeRecipient)
		}
	case EventTypePumpFunMigrate:
		l, ok1 := log.Data.(*PumpFunMigrateEvent)
		i, ok2 := ix.Data.(*PumpFunMigrateEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.BondingCurve, i.BondingCurve)
			fillStringIfDefault(&l.Pool, i.Pool)
			fillStringIfDefault(&l.User, i.User)
		}
	case EventTypePumpSwapBuy:
		l, ok1 := log.Data.(*PumpSwapBuyEvent)
		i, ok2 := ix.Data.(*PumpSwapBuyEvent)
		if ok1 && ok2 && l != nil && i != nil {
			mergePumpSwapBuyLogPreferred(l, i)
		}
	case EventTypePumpSwapSell:
		l, ok1 := log.Data.(*PumpSwapSellEvent)
		i, ok2 := ix.Data.(*PumpSwapSellEvent)
		if ok1 && ok2 && l != nil && i != nil {
			mergePumpSwapSellLogPreferred(l, i)
		}
	case EventTypePumpSwapCreatePool:
		l, ok1 := log.Data.(*PumpSwapCreatePoolEvent)
		i, ok2 := ix.Data.(*PumpSwapCreatePoolEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.Creator, i.Creator)
			fillStringIfDefault(&l.Pool, i.Pool)
			fillStringIfDefault(&l.LpMint, i.LpMint)
			fillStringIfDefault(&l.UserBaseTokenAccount, i.UserBaseTokenAccount)
			fillStringIfDefault(&l.UserQuoteTokenAccount, i.UserQuoteTokenAccount)
			fillStringIfDefault(&l.CoinCreator, i.CoinCreator)
		}
	case EventTypePumpSwapLiquidityAdded:
		l, ok1 := log.Data.(*PumpSwapLiquidityAddedEvent)
		i, ok2 := ix.Data.(*PumpSwapLiquidityAddedEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.UserBaseTokenAccount, i.UserBaseTokenAccount)
			fillStringIfDefault(&l.UserQuoteTokenAccount, i.UserQuoteTokenAccount)
			fillStringIfDefault(&l.UserPoolTokenAccount, i.UserPoolTokenAccount)
		}
	case EventTypePumpSwapLiquidityRemoved:
		l, ok1 := log.Data.(*PumpSwapLiquidityRemovedEvent)
		i, ok2 := ix.Data.(*PumpSwapLiquidityRemovedEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.UserBaseTokenAccount, i.UserBaseTokenAccount)
			fillStringIfDefault(&l.UserQuoteTokenAccount, i.UserQuoteTokenAccount)
			fillStringIfDefault(&l.UserPoolTokenAccount, i.UserPoolTokenAccount)
		}
	case EventTypeRaydiumClmmSwap:
		l, ok1 := log.Data.(*RaydiumClmmSwapEvent)
		i, ok2 := ix.Data.(*RaydiumClmmSwapEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.TokenAccount0, i.TokenAccount0)
			fillStringIfDefault(&l.TokenAccount1, i.TokenAccount1)
			fillStringIfDefault(&l.Sender, i.Sender)
		}
	case EventTypeRaydiumAmmV4Swap:
		l, ok1 := log.Data.(*RaydiumAmmV4SwapEvent)
		i, ok2 := ix.Data.(*RaydiumAmmV4SwapEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.TokenProgram, i.TokenProgram)
			fillStringIfDefault(&l.AmmAuthority, i.AmmAuthority)
			fillStringIfDefault(&l.AmmOpenOrders, i.AmmOpenOrders)
			fillStringIfDefault(&l.PoolCoinTokenAccount, i.PoolCoinTokenAccount)
			fillStringIfDefault(&l.PoolPcTokenAccount, i.PoolPcTokenAccount)
			fillStringIfDefault(&l.SerumProgram, i.SerumProgram)
			fillStringIfDefault(&l.SerumMarket, i.SerumMarket)
			fillStringIfDefault(&l.SerumBids, i.SerumBids)
			fillStringIfDefault(&l.SerumAsks, i.SerumAsks)
			fillStringIfDefault(&l.SerumEventQueue, i.SerumEventQueue)
			fillStringIfDefault(&l.SerumCoinVaultAccount, i.SerumCoinVaultAccount)
			fillStringIfDefault(&l.SerumPcVaultAccount, i.SerumPcVaultAccount)
			fillStringIfDefault(&l.SerumVaultSigner, i.SerumVaultSigner)
			fillStringIfDefault(&l.UserSourceTokenAccount, i.UserSourceTokenAccount)
			fillStringIfDefault(&l.UserDestinationTokenAccount, i.UserDestinationTokenAccount)
		}
	case EventTypeRaydiumLaunchlabPoolCreate:
		l, ok1 := log.Data.(*RaydiumLaunchlabPoolCreateEvent)
		i, ok2 := ix.Data.(*RaydiumLaunchlabPoolCreateEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.Creator, i.Creator)
			fillStringIfDefault(&l.BaseMintParam.Name, i.BaseMintParam.Name)
			fillStringIfDefault(&l.BaseMintParam.Symbol, i.BaseMintParam.Symbol)
			fillStringIfDefault(&l.BaseMintParam.Uri, i.BaseMintParam.Uri)
		}
	case EventTypeRaydiumLaunchlabMigrateAmm:
		l, ok1 := log.Data.(*RaydiumLaunchlabMigrateAmmEvent)
		i, ok2 := ix.Data.(*RaydiumLaunchlabMigrateAmmEvent)
		if ok1 && ok2 && l != nil && i != nil {
			fillStringIfDefault(&l.OldPool, i.OldPool)
			fillStringIfDefault(&l.NewPool, i.NewPool)
			fillStringIfDefault(&l.User, i.User)
		}
	}
}

func mergePumpSwapBuyLogPreferred(log, ix *PumpSwapBuyEvent) {
	fillStringIfDefault(&log.UserBaseTokenAccount, ix.UserBaseTokenAccount)
	fillStringIfDefault(&log.UserQuoteTokenAccount, ix.UserQuoteTokenAccount)
	fillStringIfDefault(&log.ProtocolFeeRecipient, ix.ProtocolFeeRecipient)
	fillStringIfDefault(&log.ProtocolFeeRecipientTokenAccount, ix.ProtocolFeeRecipientTokenAccount)
	fillStringIfDefault(&log.CoinCreator, ix.CoinCreator)
	fillStringIfDefault(&log.BaseMint, ix.BaseMint)
	fillStringIfDefault(&log.QuoteMint, ix.QuoteMint)
	fillStringIfDefault(&log.PoolBaseTokenAccount, ix.PoolBaseTokenAccount)
	fillStringIfDefault(&log.PoolQuoteTokenAccount, ix.PoolQuoteTokenAccount)
	fillStringIfDefault(&log.CoinCreatorVaultAta, ix.CoinCreatorVaultAta)
	fillStringIfDefault(&log.CoinCreatorVaultAuthority, ix.CoinCreatorVaultAuthority)
	fillStringIfDefault(&log.BaseTokenProgram, ix.BaseTokenProgram)
	fillStringIfDefault(&log.QuoteTokenProgram, ix.QuoteTokenProgram)
	fillStringIfDefault(&log.PoolV2, ix.PoolV2)
	fillStringIfDefault(&log.FeeRecipient, ix.FeeRecipient)
	fillStringIfDefault(&log.FeeRecipientQuoteTokenAccount, ix.FeeRecipientQuoteTokenAccount)
	if log.IxName == "" && ix.IxName != "" {
		log.IxName = ix.IxName
	}
}

func mergePumpSwapSellLogPreferred(log, ix *PumpSwapSellEvent) {
	fillStringIfDefault(&log.UserBaseTokenAccount, ix.UserBaseTokenAccount)
	fillStringIfDefault(&log.UserQuoteTokenAccount, ix.UserQuoteTokenAccount)
	fillStringIfDefault(&log.ProtocolFeeRecipient, ix.ProtocolFeeRecipient)
	fillStringIfDefault(&log.ProtocolFeeRecipientTokenAccount, ix.ProtocolFeeRecipientTokenAccount)
	fillStringIfDefault(&log.CoinCreator, ix.CoinCreator)
	fillStringIfDefault(&log.BaseMint, ix.BaseMint)
	fillStringIfDefault(&log.QuoteMint, ix.QuoteMint)
	fillStringIfDefault(&log.PoolBaseTokenAccount, ix.PoolBaseTokenAccount)
	fillStringIfDefault(&log.PoolQuoteTokenAccount, ix.PoolQuoteTokenAccount)
	fillStringIfDefault(&log.CoinCreatorVaultAta, ix.CoinCreatorVaultAta)
	fillStringIfDefault(&log.CoinCreatorVaultAuthority, ix.CoinCreatorVaultAuthority)
	fillStringIfDefault(&log.BaseTokenProgram, ix.BaseTokenProgram)
	fillStringIfDefault(&log.QuoteTokenProgram, ix.QuoteTokenProgram)
	fillStringIfDefault(&log.PoolV2, ix.PoolV2)
	fillStringIfDefault(&log.FeeRecipient, ix.FeeRecipient)
	fillStringIfDefault(&log.FeeRecipientQuoteTokenAccount, ix.FeeRecipientQuoteTokenAccount)
}

// DedupeLogInstructionEvents folds Yellowstone/RPC log and instruction parse results with
// the same business keys as the Rust SDK. Log data remains authoritative; instruction data
// only fills missing account-like fields.
func DedupeLogInstructionEvents(logEvents []DexEvent, instrEvents []DexEvent) []DexEvent {
	out := make([]DexEvent, 0, len(logEvents)+len(instrEvents))
	idxByKey := make(map[logInstrDedupKey]int, len(logEvents))
	logPumpfunCounts := make(map[pumpfunLaneBase]uint16)
	ixPumpfunCounts := make(map[pumpfunLaneBase]uint16)

	for _, ev := range logEvents {
		if k, ok := dedupeKey(ev, logPumpfunCounts); ok {
			idxByKey[k] = len(out)
		}
		out = append(out, ev)
	}

	for _, ev := range instrEvents {
		k, ok := dedupeKey(ev, ixPumpfunCounts)
		if !ok {
			out = append(out, ev)
			continue
		}
		if idx, exists := idxByKey[k]; exists {
			mergeGrpcInstructionIntoLog(&out[idx], ev)
			continue
		}
		idxByKey[k] = len(out)
		out = append(out, ev)
	}

	return out
}
