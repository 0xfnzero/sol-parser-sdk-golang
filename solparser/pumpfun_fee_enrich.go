package solparser

func pumpfunTradeEvent(ev DexEvent) *PumpFunTradeEvent {
	if t, ok := ev.Data.(*PumpFunTradeEvent); ok {
		return t
	}
	return nil
}

func enrichCreateV2ObservedFeeRecipient(events []DexEvent) {
	mintToFee := make(map[string]string)
	for _, ev := range events {
		t := pumpfunTradeEvent(ev)
		if t == nil || t.Mint == "" || t.Mint == zeroPubkey || t.FeeRecipient == "" || t.FeeRecipient == zeroPubkey {
			continue
		}
		buyLike := (ev.Type == EventTypePumpFunTrade && t.IsBuy) ||
			ev.Type == EventTypePumpFunBuy ||
			ev.Type == EventTypePumpFunBuyExactSolIn
		if buyLike {
			if _, ok := mintToFee[t.Mint]; !ok {
				mintToFee[t.Mint] = t.FeeRecipient
			}
		}
	}
	if len(mintToFee) == 0 {
		return
	}
	for _, ev := range events {
		if ev.Type != EventTypePumpFunCreateV2 {
			continue
		}
		c, ok := ev.Data.(*PumpFunCreateV2TokenEvent)
		if !ok || c.ObservedFeeRecipient != "" {
			continue
		}
		if fee, ok := mintToFee[c.Mint]; ok {
			c.ObservedFeeRecipient = fee
		}
	}
}

func enrichPumpfunTradesFromCreateInstructions(events []DexEvent) {
	type flags struct {
		cashback bool
		mayhem   bool
	}
	mintFlags := make(map[string]flags)
	for _, ev := range events {
		switch c := ev.Data.(type) {
		case *PumpFunCreateEvent:
			if c.Mint != "" && c.Mint != zeroPubkey {
				if _, ok := mintFlags[c.Mint]; !ok {
					mintFlags[c.Mint] = flags{cashback: c.IsCashbackEnabled, mayhem: c.IsMayhemMode}
				}
			}
		case *PumpFunCreateV2TokenEvent:
			if c.Mint != "" && c.Mint != zeroPubkey {
				if _, ok := mintFlags[c.Mint]; !ok {
					mintFlags[c.Mint] = flags{cashback: c.IsCashbackEnabled, mayhem: c.IsMayhemMode}
				}
			}
		}
	}
	if len(mintFlags) == 0 {
		return
	}
	for _, ev := range events {
		t := pumpfunTradeEvent(ev)
		if t == nil || t.Mint == "" || t.Mint == zeroPubkey {
			continue
		}
		f, ok := mintFlags[t.Mint]
		if !ok {
			continue
		}
		t.IsCashbackCoin = t.IsCashbackCoin || f.cashback
		t.MayhemMode = t.MayhemMode || f.mayhem
		if f.cashback {
			t.TrackVolume = true
		}
	}
}

func enrichPumpfunSameTxPostMerge(events []DexEvent) {
	enrichCreateV2ObservedFeeRecipient(events)
	enrichPumpfunTradesFromCreateInstructions(events)
}
