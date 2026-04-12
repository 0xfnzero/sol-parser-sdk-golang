package solparser

import "reflect"

// ApplyMetadataTxIndex 将事件中嵌套 Metadata.TxIndex 设为 slot 内交易序号（如 Geyser SubscribeUpdateTransactionInfo.index）。
func ApplyMetadataTxIndex(events []DexEvent, txIndex uint64) {
	for i := range events {
		setEventMetadataTxIndex(&events[i], txIndex)
	}
}

func setEventMetadataTxIndex(ev *DexEvent, txIndex uint64) {
	d := ev.Data
	if d == nil {
		return
	}
	v := reflect.ValueOf(d)
	if v.Kind() != reflect.Ptr {
		return
	}
	v = v.Elem()
	md := v.FieldByName("Metadata")
	if !md.IsValid() || md.Kind() != reflect.Struct {
		return
	}
	txF := md.FieldByName("TxIndex")
	if txF.IsValid() && txF.CanSet() && txF.Kind() == reflect.Uint64 {
		txF.SetUint(txIndex)
	}
}

// CoalescePumpSwapBuySellBySignature 合并同一笔交易中、同一 signature 下重复的 PumpSwap Buy/Sell：
// 指令路径带账户（mint、池子 ATA），日志路径带链上 Program data 数值；合并后字段齐全。
func CoalescePumpSwapBuySellBySignature(events []DexEvent) []DexEvent {
	mergedBuy := map[string]*PumpSwapBuyEvent{}
	mergedSell := map[string]*PumpSwapSellEvent{}

	for _, ev := range events {
		switch ev.Type {
		case EventTypePumpSwapBuy:
			p := ev.Data.(*PumpSwapBuyEvent)
			sig := p.Metadata.Signature
			if cur, ok := mergedBuy[sig]; ok {
				supplementPumpSwapBuy(cur, p)
			} else {
				c := *p
				mergedBuy[sig] = &c
			}
		case EventTypePumpSwapSell:
			p := ev.Data.(*PumpSwapSellEvent)
			sig := p.Metadata.Signature
			if cur, ok := mergedSell[sig]; ok {
				supplementPumpSwapSell(cur, p)
			} else {
				c := *p
				mergedSell[sig] = &c
			}
		}
	}

	out := make([]DexEvent, 0, len(events))
	seenBuy := map[string]bool{}
	seenSell := map[string]bool{}
	for _, ev := range events {
		switch ev.Type {
		case EventTypePumpSwapBuy:
			p := ev.Data.(*PumpSwapBuyEvent)
			sig := p.Metadata.Signature
			if seenBuy[sig] {
				continue
			}
			seenBuy[sig] = true
			out = append(out, DexEvent{Type: EventTypePumpSwapBuy, Data: mergedBuy[sig]})
		case EventTypePumpSwapSell:
			p := ev.Data.(*PumpSwapSellEvent)
			sig := p.Metadata.Signature
			if seenSell[sig] {
				continue
			}
			seenSell[sig] = true
			out = append(out, DexEvent{Type: EventTypePumpSwapSell, Data: mergedSell[sig]})
		default:
			out = append(out, ev)
		}
	}
	return out
}

func supplementString(dst *string, src string) {
	if *dst == "" && src != "" {
		*dst = src
	}
}

func supplementU64(dst *uint64, src uint64) {
	if src != 0 && *dst == 0 {
		*dst = src
	}
}

func supplementI64(dst *int64, src int64) {
	if src != 0 && *dst == 0 {
		*dst = src
	}
}

func supplementPumpSwapBuy(dst, src *PumpSwapBuyEvent) {
	supplementString(&dst.BaseMint, src.BaseMint)
	supplementString(&dst.QuoteMint, src.QuoteMint)
	supplementString(&dst.PoolBaseTokenAccount, src.PoolBaseTokenAccount)
	supplementString(&dst.PoolQuoteTokenAccount, src.PoolQuoteTokenAccount)
	supplementString(&dst.CoinCreatorVaultAta, src.CoinCreatorVaultAta)
	supplementString(&dst.CoinCreatorVaultAuthority, src.CoinCreatorVaultAuthority)
	supplementString(&dst.BaseTokenProgram, src.BaseTokenProgram)
	supplementString(&dst.QuoteTokenProgram, src.QuoteTokenProgram)

	supplementI64(&dst.Timestamp, src.Timestamp)
	supplementU64(&dst.BaseAmountOut, src.BaseAmountOut)
	supplementU64(&dst.MaxQuoteAmountIn, src.MaxQuoteAmountIn)
	supplementU64(&dst.UserBaseTokenReserves, src.UserBaseTokenReserves)
	supplementU64(&dst.UserQuoteTokenReserves, src.UserQuoteTokenReserves)
	supplementU64(&dst.PoolBaseTokenReserves, src.PoolBaseTokenReserves)
	supplementU64(&dst.PoolQuoteTokenReserves, src.PoolQuoteTokenReserves)
	supplementU64(&dst.QuoteAmountIn, src.QuoteAmountIn)
	supplementU64(&dst.LpFeeBasisPoints, src.LpFeeBasisPoints)
	supplementU64(&dst.LpFee, src.LpFee)
	supplementU64(&dst.ProtocolFeeBasisPoints, src.ProtocolFeeBasisPoints)
	supplementU64(&dst.ProtocolFee, src.ProtocolFee)
	supplementU64(&dst.QuoteAmountInWithLpFee, src.QuoteAmountInWithLpFee)
	supplementU64(&dst.UserQuoteAmountIn, src.UserQuoteAmountIn)
	supplementU64(&dst.CoinCreatorFeeBasisPoints, src.CoinCreatorFeeBasisPoints)
	supplementU64(&dst.CoinCreatorFee, src.CoinCreatorFee)
	supplementU64(&dst.TotalUnclaimedTokens, src.TotalUnclaimedTokens)
	supplementU64(&dst.TotalClaimedTokens, src.TotalClaimedTokens)
	supplementU64(&dst.CurrentSolVolume, src.CurrentSolVolume)
	supplementI64(&dst.LastUpdateTimestamp, src.LastUpdateTimestamp)
	supplementU64(&dst.MinBaseAmountOut, src.MinBaseAmountOut)
	supplementU64(&dst.CashbackFeeBasisPoints, src.CashbackFeeBasisPoints)
	supplementU64(&dst.Cashback, src.Cashback)

	if dst.IxName == "" {
		dst.IxName = src.IxName
	}
	dst.MayhemMode = dst.MayhemMode || src.MayhemMode
	dst.TrackVolume = dst.TrackVolume || src.TrackVolume
	dst.IsCashbackCoin = dst.IsCashbackCoin || src.IsCashbackCoin
	dst.IsPumpPool = dst.IsPumpPool || src.IsPumpPool

	supplementString(&dst.Pool, src.Pool)
	supplementString(&dst.User, src.User)
	supplementString(&dst.UserBaseTokenAccount, src.UserBaseTokenAccount)
	supplementString(&dst.UserQuoteTokenAccount, src.UserQuoteTokenAccount)
	supplementString(&dst.ProtocolFeeRecipient, src.ProtocolFeeRecipient)
	supplementString(&dst.ProtocolFeeRecipientTokenAccount, src.ProtocolFeeRecipientTokenAccount)
	supplementString(&dst.CoinCreator, src.CoinCreator)
}

func supplementPumpSwapSell(dst, src *PumpSwapSellEvent) {
	supplementString(&dst.BaseMint, src.BaseMint)
	supplementString(&dst.QuoteMint, src.QuoteMint)
	supplementString(&dst.PoolBaseTokenAccount, src.PoolBaseTokenAccount)
	supplementString(&dst.PoolQuoteTokenAccount, src.PoolQuoteTokenAccount)
	supplementString(&dst.CoinCreatorVaultAta, src.CoinCreatorVaultAta)
	supplementString(&dst.CoinCreatorVaultAuthority, src.CoinCreatorVaultAuthority)
	supplementString(&dst.BaseTokenProgram, src.BaseTokenProgram)
	supplementString(&dst.QuoteTokenProgram, src.QuoteTokenProgram)

	supplementI64(&dst.Timestamp, src.Timestamp)
	supplementU64(&dst.BaseAmountIn, src.BaseAmountIn)
	supplementU64(&dst.MinQuoteAmountOut, src.MinQuoteAmountOut)
	supplementU64(&dst.UserBaseTokenReserves, src.UserBaseTokenReserves)
	supplementU64(&dst.UserQuoteTokenReserves, src.UserQuoteTokenReserves)
	supplementU64(&dst.PoolBaseTokenReserves, src.PoolBaseTokenReserves)
	supplementU64(&dst.PoolQuoteTokenReserves, src.PoolQuoteTokenReserves)
	supplementU64(&dst.QuoteAmountOut, src.QuoteAmountOut)
	supplementU64(&dst.LpFeeBasisPoints, src.LpFeeBasisPoints)
	supplementU64(&dst.LpFee, src.LpFee)
	supplementU64(&dst.ProtocolFeeBasisPoints, src.ProtocolFeeBasisPoints)
	supplementU64(&dst.ProtocolFee, src.ProtocolFee)
	supplementU64(&dst.QuoteAmountOutWithoutLpFee, src.QuoteAmountOutWithoutLpFee)
	supplementU64(&dst.UserQuoteAmountOut, src.UserQuoteAmountOut)
	supplementU64(&dst.CoinCreatorFeeBasisPoints, src.CoinCreatorFeeBasisPoints)
	supplementU64(&dst.CoinCreatorFee, src.CoinCreatorFee)
	supplementU64(&dst.CashbackFeeBasisPoints, src.CashbackFeeBasisPoints)
	supplementU64(&dst.Cashback, src.Cashback)

	supplementString(&dst.Pool, src.Pool)
	supplementString(&dst.User, src.User)
	supplementString(&dst.UserBaseTokenAccount, src.UserBaseTokenAccount)
	supplementString(&dst.UserQuoteTokenAccount, src.UserQuoteTokenAccount)
	supplementString(&dst.ProtocolFeeRecipient, src.ProtocolFeeRecipient)
	supplementString(&dst.ProtocolFeeRecipientTokenAccount, src.ProtocolFeeRecipientTokenAccount)
	supplementString(&dst.CoinCreator, src.CoinCreator)
}
