package solparser

func pumpfunTradeEvent(ev DexEvent) *PumpFunTradeEvent {
	if t, ok := ev.Data.(*PumpFunTradeEvent); ok {
		return t
	}
	return nil
}

func emptyPubkeyOrString(v string) bool {
	return v == "" || v == zeroPubkey
}

func fillStringIfEmpty(dst *string, src string) {
	if emptyPubkeyOrString(*dst) && !emptyPubkeyOrString(src) {
		*dst = src
	}
}

func fillUint64IfZero(dst *uint64, src uint64) {
	if *dst == 0 && src != 0 {
		*dst = src
	}
}

func fillInt64IfZero(dst *int64, src int64) {
	if *dst == 0 && src != 0 {
		*dst = src
	}
}

func enrichCreateV2FromCreateEvents(events []DexEvent) {
	creates := make(map[string]*PumpFunCreateEvent)
	for _, ev := range events {
		if ev.Type != EventTypePumpFunCreate {
			continue
		}
		c, ok := ev.Data.(*PumpFunCreateEvent)
		if !ok || c == nil || emptyPubkeyOrString(c.Mint) {
			continue
		}
		if _, exists := creates[c.Mint]; !exists {
			creates[c.Mint] = c
		}
	}
	if len(creates) == 0 {
		return
	}

	for _, ev := range events {
		if ev.Type != EventTypePumpFunCreateV2 {
			continue
		}
		c2, ok := ev.Data.(*PumpFunCreateV2TokenEvent)
		if !ok || c2 == nil {
			continue
		}
		c, exists := creates[c2.Mint]
		if !exists {
			continue
		}

		fillStringIfEmpty(&c2.Name, c.Name)
		fillStringIfEmpty(&c2.Symbol, c.Symbol)
		fillStringIfEmpty(&c2.Uri, c.Uri)
		fillStringIfEmpty(&c2.BondingCurve, c.BondingCurve)
		fillStringIfEmpty(&c2.User, c.User)
		fillStringIfEmpty(&c2.Creator, c.Creator)
		fillStringIfEmpty(&c2.TokenProgram, c.TokenProgram)
		fillStringIfEmpty(&c2.QuoteMint, c.QuoteMint)
		fillInt64IfZero(&c2.Timestamp, c.Timestamp)
		fillUint64IfZero(&c2.VirtualTokenReserves, c.VirtualTokenReserves)
		fillUint64IfZero(&c2.VirtualSolReserves, c.VirtualSolReserves)
		fillUint64IfZero(&c2.RealTokenReserves, c.RealTokenReserves)
		fillUint64IfZero(&c2.TokenTotalSupply, c.TokenTotalSupply)
		fillUint64IfZero(&c2.VirtualQuoteReserves, c.VirtualQuoteReserves)
		c2.IsCashbackEnabled = c2.IsCashbackEnabled || c.IsCashbackEnabled
		c2.IsMayhemMode = c2.IsMayhemMode || c.IsMayhemMode
	}
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
		if !ok || !emptyPubkeyOrString(c.ObservedFeeRecipient) {
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
	enrichCreateV2FromCreateEvents(events)
	enrichCreateV2ObservedFeeRecipient(events)
	enrichPumpfunTradesFromCreateInstructions(events)
}
