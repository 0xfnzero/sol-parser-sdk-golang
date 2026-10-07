package solparser

// RPC 路径下对齐 Rust `account_dispatcher::fill_accounts_with_owned_keys` 与 `common_filler::fill_data`。

func buildRpcProgramInvokes(msg *RpcMessage, meta *RpcTransactionMeta) map[string][][2]int32 {
	m := make(map[string][][2]int32)
	if msg == nil {
		return m
	}
	fullKeys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
	for i, ix := range msg.Instructions {
		if int(ix.ProgramIDIndex) < len(fullKeys) {
			pid := fullKeys[ix.ProgramIDIndex]
			m[pid] = append(m[pid], [2]int32{int32(i), -1})
		}
	}
	if meta != nil {
		for _, g := range meta.InnerInstructions {
			for j, ix := range g.Instructions {
				if int(ix.ProgramIDIndex) < len(fullKeys) {
					pid := fullKeys[ix.ProgramIDIndex]
					m[pid] = append(m[pid], [2]int32{int32(g.Index), int32(j)})
				}
			}
		}
	}
	return m
}

func getRpcInstructionData(msg *RpcMessage, meta *RpcTransactionMeta, inv [2]int32) []byte {
	if msg == nil {
		return nil
	}
	if inv[1] >= 0 {
		if meta == nil {
			return nil
		}
		for _, g := range meta.InnerInstructions {
			if g.Index == uint32(inv[0]) && int(inv[1]) < len(g.Instructions) {
				return g.Instructions[inv[1]].Data
			}
		}
		return nil
	}
	if int(inv[0]) >= 0 && int(inv[0]) < len(msg.Instructions) {
		return msg.Instructions[inv[0]].Data
	}
	return nil
}

func fillRpcDexEvents(events []DexEvent, msg *RpcMessage, meta *RpcTransactionMeta) {
	if len(events) == 0 || msg == nil {
		return
	}
	invokes := buildRpcProgramInvokes(msg, meta)
	feesInv := invokes[GrpcPumpSwapFeesProgramID]

	for i := range events {
		fillRpcOneEvent(&events[i], msg, meta, invokes, feesInv)
	}
}

// fillRpcDexEventsPump is kept for compatibility with existing internal tests and callers.
func fillRpcDexEventsPump(events []DexEvent, msg *RpcMessage, meta *RpcTransactionMeta) {
	fillRpcDexEvents(events, msg, meta)
}

func fillRpcOneEvent(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32, feesInv [][2]int32) {
	switch ev.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn:
		pumpInv := invokes[PUMPFUN_PROGRAM_ID]
		get := rpcAccountGetter(msg, meta, pumpInv)
		if get == nil {
			return
		}
		tr, ok := ev.Data.(*PumpFunTradeEvent)
		if !ok {
			return
		}
		fillRpcPumpFunTrade(tr, get)
		fillPumpFunBalances(tr, msg, meta)
	case EventTypePumpFunCreate:
		pumpInv := invokes[PUMPFUN_PROGRAM_ID]
		get, isCreateV2 := rpcPumpFunCreateAccountGetter(msg, meta, pumpInv, ev)
		if get == nil {
			return
		}
		if c, ok := ev.Data.(*PumpFunCreateEvent); ok {
			fillRpcPumpFunCreate(c, get, isCreateV2)
		}
	case EventTypePumpFunCreateV2:
		pumpInv := invokes[PUMPFUN_PROGRAM_ID]
		get, _ := rpcPumpFunCreateAccountGetter(msg, meta, pumpInv, ev)
		if get == nil {
			return
		}
		if c, ok := ev.Data.(*PumpFunCreateV2TokenEvent); ok {
			fillRpcPumpFunCreateV2(c, get)
		}
	case EventTypePumpSwapBuy:
		swapInv := invokes[PUMPSWAP_PROGRAM_ID]
		get := rpcAccountGetter(msg, meta, swapInv)
		if get != nil {
			if b, ok := ev.Data.(*PumpSwapBuyEvent); ok {
				fillRpcPumpSwapBuy(b, get)
			}
		}
		fillPumpSwapIsPumpPool(ev, msg, meta, feesInv)
	case EventTypePumpSwapSell:
		swapInv := invokes[PUMPSWAP_PROGRAM_ID]
		get := rpcAccountGetter(msg, meta, swapInv)
		if get != nil {
			if s, ok := ev.Data.(*PumpSwapSellEvent); ok {
				fillRpcPumpSwapSell(s, get)
			}
		}
		fillPumpSwapIsPumpPool(ev, msg, meta, feesInv)
	case EventTypePumpSwapCreatePool:
		swapInv := invokes[PUMPSWAP_PROGRAM_ID]
		get := rpcAccountGetter(msg, meta, swapInv)
		if get != nil {
			if c, ok := ev.Data.(*PumpSwapCreatePoolEvent); ok {
				fillRpcPumpSwapCreatePool(c, get)
			}
		}
	case EventTypeRaydiumClmmSwap:
		fillRpcRaydiumClmmSwap(ev, msg, meta, invokes)
	case EventTypeRaydiumClmmCreatePool:
		fillRpcRaydiumClmmCreatePool(ev, msg, meta, invokes)
	case EventTypeRaydiumClmmOpenPosition:
		fillRpcRaydiumClmmOpenPosition(ev, msg, meta, invokes)
	case EventTypeRaydiumClmmClosePosition:
		fillRpcRaydiumClmmClosePosition(ev, msg, meta, invokes)
	case EventTypeRaydiumClmmIncreaseLiquidity:
		fillRpcRaydiumClmmIncreaseLiquidity(ev, msg, meta, invokes)
	case EventTypeRaydiumClmmDecreaseLiquidity:
		fillRpcRaydiumClmmDecreaseLiquidity(ev, msg, meta, invokes)
	case EventTypeRaydiumCpmmSwap:
		fillRpcRaydiumCpmmSwap(ev, msg, meta, invokes)
	case EventTypeRaydiumCpmmDeposit:
		fillRpcRaydiumCpmmDeposit(ev, msg, meta, invokes)
	case EventTypeRaydiumCpmmWithdraw:
		fillRpcRaydiumCpmmWithdraw(ev, msg, meta, invokes)
	case EventTypeRaydiumCpmmInitialize:
		fillRpcRaydiumCpmmInitialize(ev, msg, meta, invokes)
	case EventTypeRaydiumAmmV4Swap:
		fillRpcRaydiumAmmV4Swap(ev, msg, meta, invokes)
	case EventTypeRaydiumAmmV4Deposit:
		fillRpcRaydiumAmmV4Deposit(ev, msg, meta, invokes)
	case EventTypeRaydiumAmmV4Withdraw:
		fillRpcRaydiumAmmV4Withdraw(ev, msg, meta, invokes)
	case EventTypeOrcaWhirlpoolLiquidityIncreased:
		fillRpcOrcaWhirlpoolLiquidityIncreased(ev, msg, meta, invokes)
	case EventTypeOrcaWhirlpoolLiquidityDecreased:
		fillRpcOrcaWhirlpoolLiquidityDecreased(ev, msg, meta, invokes)
	case EventTypeMeteoraDammV2InitializePool:
		fillRpcMeteoraDammV2InitializePool(ev, msg, meta, invokes)
	case EventTypeMeteoraDlmmSwap:
		fillRpcMeteoraDlmmSwap(ev, msg, meta, invokes)
	case EventTypeRaydiumLaunchlabTrade:
		fillRpcRaydiumLaunchlabTrade(ev, msg, meta, invokes)
	case EventTypeRaydiumLaunchlabPoolCreate:
		fillRpcRaydiumLaunchlabPoolCreate(ev, msg, meta, invokes)
	}
}

func isDefaultPubkeyString(s string) bool {
	return s == "" || s == zeroPubkey
}

func fillStringFromAccount(dst *string, get func(int) string, idx int) {
	if dst == nil || !isDefaultPubkeyString(*dst) {
		return
	}
	if v := get(idx); !isDefaultPubkeyString(v) {
		*dst = v
	}
}

func fillRpcPumpFunTrade(tr *PumpFunTradeEvent, get func(int) string) {
	if tr == nil {
		return
	}
	isV2 := tr.IxName == "buy_v2" || tr.IxName == "sell_v2" || tr.IxName == "buy_exact_quote_in_v2" ||
		(!isDefaultPubkeyString(tr.Mint) && get(1) == tr.Mint)
	isSell := tr.IxName == "sell" || tr.IxName == "sell_v2" || !tr.IsBuy

	if isV2 {
		fillStringFromAccount(&tr.Global, get, 0)
		fillStringFromAccount(&tr.QuoteMint, get, 2)
		fillStringFromAccount(&tr.FeeRecipient, get, 6)
		fillStringFromAccount(&tr.BondingCurve, get, 10)
		fillStringFromAccount(&tr.AssociatedBondingCurve, get, 11)
		fillStringFromAccount(&tr.AssociatedQuoteBondingCurve, get, 12)
		fillStringFromAccount(&tr.User, get, 13)
		fillStringFromAccount(&tr.AssociatedUser, get, 14)
		fillStringFromAccount(&tr.AssociatedQuoteUser, get, 15)
		fillStringFromAccount(&tr.TokenProgram, get, 3)
		fillStringFromAccount(&tr.QuoteTokenProgram, get, 4)
		fillStringFromAccount(&tr.AssociatedTokenProgram, get, 5)
		fillStringFromAccount(&tr.CreatorVault, get, 16)
		fillStringFromAccount(&tr.AssociatedQuoteFeeRecipient, get, 7)
		fillStringFromAccount(&tr.BuybackFeeRecipient, get, 8)
		fillStringFromAccount(&tr.AssociatedQuoteBuybackFeeRecipient, get, 9)
		fillStringFromAccount(&tr.AssociatedCreatorVault, get, 17)
		fillStringFromAccount(&tr.SharingConfig, get, 18)
		if isSell {
			fillStringFromAccount(&tr.SystemProgram, get, 23)
			fillStringFromAccount(&tr.EventAuthority, get, 24)
			fillStringFromAccount(&tr.Program, get, 25)
			fillStringFromAccount(&tr.UserVolumeAccumulator, get, 19)
			fillStringFromAccount(&tr.AssociatedUserVolumeAccumulator, get, 20)
			fillStringFromAccount(&tr.FeeConfig, get, 21)
			fillStringFromAccount(&tr.FeeProgram, get, 22)
		} else {
			fillStringFromAccount(&tr.SystemProgram, get, 24)
			fillStringFromAccount(&tr.EventAuthority, get, 25)
			fillStringFromAccount(&tr.Program, get, 26)
			fillStringFromAccount(&tr.GlobalVolumeAccumulator, get, 19)
			fillStringFromAccount(&tr.UserVolumeAccumulator, get, 20)
			fillStringFromAccount(&tr.AssociatedUserVolumeAccumulator, get, 21)
			fillStringFromAccount(&tr.FeeConfig, get, 22)
			fillStringFromAccount(&tr.FeeProgram, get, 23)
		}
		return
	}

	fillStringFromAccount(&tr.Global, get, 0)
	fillStringFromAccount(&tr.FeeRecipient, get, 1)
	fillStringFromAccount(&tr.BondingCurve, get, 3)
	fillStringFromAccount(&tr.AssociatedBondingCurve, get, 4)
	fillStringFromAccount(&tr.AssociatedUser, get, 5)
	fillStringFromAccount(&tr.User, get, 6)
	fillStringFromAccount(&tr.SystemProgram, get, 7)
	if tr.IsBuy {
		fillStringFromAccount(&tr.TokenProgram, get, 8)
		fillStringFromAccount(&tr.CreatorVault, get, 9)
		fillStringFromAccount(&tr.EventAuthority, get, 10)
		fillStringFromAccount(&tr.Program, get, 11)
		fillStringFromAccount(&tr.GlobalVolumeAccumulator, get, 12)
		fillStringFromAccount(&tr.UserVolumeAccumulator, get, 13)
		fillStringFromAccount(&tr.FeeConfig, get, 14)
		fillStringFromAccount(&tr.FeeProgram, get, 15)
		fillStringFromAccount(&tr.BondingCurveV2, get, 16)
		fillStringFromAccount(&tr.BuybackFeeRecipient, get, 17)
		if isDefaultPubkeyString(tr.Account) {
			fillStringFromAccount(&tr.Account, get, 17)
		}
	} else {
		fillStringFromAccount(&tr.CreatorVault, get, 8)
		fillStringFromAccount(&tr.TokenProgram, get, 9)
		fillStringFromAccount(&tr.EventAuthority, get, 10)
		fillStringFromAccount(&tr.Program, get, 11)
		fillStringFromAccount(&tr.FeeConfig, get, 12)
		fillStringFromAccount(&tr.FeeProgram, get, 13)
		a16 := get(16)
		if !isDefaultPubkeyString(a16) {
			fillStringFromAccount(&tr.UserVolumeAccumulator, get, 14)
			fillStringFromAccount(&tr.BondingCurveV2, get, 15)
			fillStringFromAccount(&tr.BuybackFeeRecipient, get, 16)
			fillStringFromAccount(&tr.Account, get, 16)
		} else if tr.IsCashbackCoin {
			fillStringFromAccount(&tr.UserVolumeAccumulator, get, 14)
			fillStringFromAccount(&tr.BondingCurveV2, get, 15)
		} else {
			fillStringFromAccount(&tr.BondingCurveV2, get, 14)
			fillStringFromAccount(&tr.BuybackFeeRecipient, get, 15)
			fillStringFromAccount(&tr.Account, get, 15)
		}
	}
}

// The all-zero public key is a real account for the System Program, not an
// unknown value. Preserve it when filling a verified create instruction.
func fillPumpFunCreateSystemProgram(dst *string, get func(int) string, index int) {
	if isDefaultPubkeyString(*dst) && get(index) == zeroPubkey {
		*dst = zeroPubkey
		return
	}
	fillStringFromAccount(dst, get, index)
}

func fillRpcPumpFunCreate(c *PumpFunCreateEvent, get func(int) string, isCreateV2 bool) {
	if c == nil {
		return
	}
	fillStringFromAccount(&c.Mint, get, 0)
	fillStringFromAccount(&c.BondingCurve, get, 2)
	if isCreateV2 {
		fillStringFromAccount(&c.User, get, 5)
		fillStringFromAccount(&c.MintAuthority, get, 1)
		fillStringFromAccount(&c.AssociatedBondingCurve, get, 3)
		fillStringFromAccount(&c.Global, get, 4)
		fillPumpFunCreateSystemProgram(&c.SystemProgram, get, 6)
		fillStringFromAccount(&c.TokenProgram, get, 7)
		fillStringFromAccount(&c.AssociatedTokenProgram, get, 8)
		fillStringFromAccount(&c.MayhemProgramID, get, 9)
		fillStringFromAccount(&c.GlobalParams, get, 10)
		fillStringFromAccount(&c.SolVault, get, 11)
		fillStringFromAccount(&c.MayhemState, get, 12)
		fillStringFromAccount(&c.MayhemTokenVault, get, 13)
		fillStringFromAccount(&c.EventAuthority, get, 14)
		fillStringFromAccount(&c.Program, get, 15)
		if c.IxName == "" {
			c.IxName = "create_v2"
		}
	} else {
		fillStringFromAccount(&c.User, get, 7)
		fillStringFromAccount(&c.MintAuthority, get, 1)
		fillStringFromAccount(&c.AssociatedBondingCurve, get, 3)
		fillStringFromAccount(&c.Global, get, 4)
		fillPumpFunCreateSystemProgram(&c.SystemProgram, get, 8)
		fillStringFromAccount(&c.TokenProgram, get, 9)
		fillStringFromAccount(&c.AssociatedTokenProgram, get, 10)
		fillStringFromAccount(&c.EventAuthority, get, 12)
		fillStringFromAccount(&c.Program, get, 13)
		if c.IxName == "" {
			c.IxName = "create"
		}
	}
}

func fillRpcPumpFunCreateV2(c *PumpFunCreateV2TokenEvent, get func(int) string) {
	if c == nil {
		return
	}
	fillStringFromAccount(&c.Mint, get, 0)
	fillStringFromAccount(&c.BondingCurve, get, 2)
	fillStringFromAccount(&c.User, get, 5)
	fillStringFromAccount(&c.MintAuthority, get, 1)
	fillStringFromAccount(&c.AssociatedBondingCurve, get, 3)
	fillStringFromAccount(&c.Global, get, 4)
	fillPumpFunCreateSystemProgram(&c.SystemProgram, get, 6)
	fillStringFromAccount(&c.TokenProgram, get, 7)
	fillStringFromAccount(&c.AssociatedTokenProgram, get, 8)
	fillStringFromAccount(&c.MayhemProgramID, get, 9)
	fillStringFromAccount(&c.GlobalParams, get, 10)
	fillStringFromAccount(&c.SolVault, get, 11)
	fillStringFromAccount(&c.MayhemState, get, 12)
	fillStringFromAccount(&c.MayhemTokenVault, get, 13)
	fillStringFromAccount(&c.EventAuthority, get, 14)
	fillStringFromAccount(&c.Program, get, 15)
}

func fillRpcPumpSwapBuy(b *PumpSwapBuyEvent, get func(int) string) {
	if b == nil {
		return
	}
	fillRpcPumpSwapTradeCommon(&b.Pool, &b.User, &b.BaseMint, &b.QuoteMint, &b.UserBaseTokenAccount, &b.UserQuoteTokenAccount, &b.PoolBaseTokenAccount, &b.PoolQuoteTokenAccount, &b.ProtocolFeeRecipient, &b.ProtocolFeeRecipientTokenAccount, &b.BaseTokenProgram, &b.QuoteTokenProgram, &b.CoinCreatorVaultAta, &b.CoinCreatorVaultAuthority, get)
	if !isDefaultPubkeyString(get(26)) {
		fillStringFromAccount(&b.PoolV2, get, 24)
		fillStringFromAccount(&b.FeeRecipient, get, 25)
		fillStringFromAccount(&b.FeeRecipientQuoteTokenAccount, get, 26)
		return
	}
	if !isDefaultPubkeyString(get(25)) {
		fillStringFromAccount(&b.PoolV2, get, 23)
		fillStringFromAccount(&b.FeeRecipient, get, 24)
		fillStringFromAccount(&b.FeeRecipientQuoteTokenAccount, get, 25)
		return
	}
	fillStringFromAccount(&b.PoolV2, get, 23)
}

func fillRpcPumpSwapSell(s *PumpSwapSellEvent, get func(int) string) {
	if s == nil {
		return
	}
	fillRpcPumpSwapTradeCommon(&s.Pool, &s.User, &s.BaseMint, &s.QuoteMint, &s.UserBaseTokenAccount, &s.UserQuoteTokenAccount, &s.PoolBaseTokenAccount, &s.PoolQuoteTokenAccount, &s.ProtocolFeeRecipient, &s.ProtocolFeeRecipientTokenAccount, &s.BaseTokenProgram, &s.QuoteTokenProgram, &s.CoinCreatorVaultAta, &s.CoinCreatorVaultAuthority, get)
	if !isDefaultPubkeyString(get(25)) {
		fillStringFromAccount(&s.PoolV2, get, 23)
		fillStringFromAccount(&s.FeeRecipient, get, 24)
		fillStringFromAccount(&s.FeeRecipientQuoteTokenAccount, get, 25)
		return
	}
	if !isDefaultPubkeyString(get(23)) {
		fillStringFromAccount(&s.PoolV2, get, 21)
		fillStringFromAccount(&s.FeeRecipient, get, 22)
		fillStringFromAccount(&s.FeeRecipientQuoteTokenAccount, get, 23)
		return
	}
	fillStringFromAccount(&s.PoolV2, get, 21)
}

func fillRpcPumpSwapTradeCommon(
	pool, user, baseMint, quoteMint, userBase, userQuote, poolBase, poolQuote, protocolFee, protocolFeeToken, baseProgram, quoteProgram, creatorVault, creatorVaultAuthority *string,
	get func(int) string,
) {
	fillStringFromAccount(pool, get, 0)
	fillStringFromAccount(user, get, 1)
	fillStringFromAccount(baseMint, get, 3)
	fillStringFromAccount(quoteMint, get, 4)
	fillStringFromAccount(userBase, get, 5)
	fillStringFromAccount(userQuote, get, 6)
	fillStringFromAccount(poolBase, get, 7)
	fillStringFromAccount(poolQuote, get, 8)
	fillStringFromAccount(protocolFee, get, 9)
	fillStringFromAccount(protocolFeeToken, get, 10)
	fillStringFromAccount(baseProgram, get, 11)
	fillStringFromAccount(quoteProgram, get, 12)
	fillStringFromAccount(creatorVault, get, 17)
	fillStringFromAccount(creatorVaultAuthority, get, 18)
}

func fillRpcPumpSwapCreatePool(c *PumpSwapCreatePoolEvent, get func(int) string) {
	if c == nil {
		return
	}
	fillStringFromAccount(&c.Pool, get, 0)
	fillStringFromAccount(&c.Creator, get, 2)
	fillStringFromAccount(&c.BaseMint, get, 3)
	fillStringFromAccount(&c.QuoteMint, get, 4)
	fillStringFromAccount(&c.LpMint, get, 5)
	fillStringFromAccount(&c.UserBaseTokenAccount, get, 6)
	fillStringFromAccount(&c.UserQuoteTokenAccount, get, 7)
}

func rpcGetterForProgram(msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32, programID string) func(int) string {
	return rpcAccountGetter(msg, meta, invokes[programID])
}

// Select the creating invocation by discriminator and mint, never by account count.
// A transaction can contain create_v2 followed by a buy with a larger account list.
func rpcPumpFunCreateAccountGetter(msg *RpcMessage, meta *RpcTransactionMeta, list [][2]int32, ev *DexEvent) (func(int) string, bool) {
	if msg == nil || ev == nil {
		return nil, false
	}
	var mint string
	v2Only := ev.Type == EventTypePumpFunCreateV2
	switch c := ev.Data.(type) {
	case *PumpFunCreateEvent:
		if c == nil {
			return nil, false
		}
		mint = c.Mint
	case *PumpFunCreateV2TokenEvent:
		if c == nil {
			return nil, false
		}
		mint = c.Mint
	default:
		return nil, false
	}
	var selected func(int) string
	var selectedV2 bool
	for _, inv := range list {
		data := getRpcInstructionData(msg, meta, inv)
		if len(data) < 8 {
			continue
		}
		disc := disc8FromBytes(data[:8])
		v2 := disc == instrPumpOuterCreateV2
		if !v2 && (v2Only || disc != instrPumpOuterCreate) {
			continue
		}
		minimum := 14
		if v2 {
			minimum = 16
		}
		if rpcInvokeAccountLen(msg, meta, inv) < minimum {
			continue
		}
		get := rpcAccountGetter(msg, meta, [][2]int32{inv})
		if get == nil || (!isDefaultPubkeyString(mint) && get(0) != mint) {
			continue
		}
		// An event without a usable mint cannot safely choose between multiple creates.
		if selected != nil {
			return nil, false
		}
		selected, selectedV2 = get, v2
	}
	return selected, selectedV2
}

func disc8FromBytes(b []byte) uint64 {
	if len(b) < 8 {
		return 0
	}
	return uint64(b[0]) |
		uint64(b[1])<<8 |
		uint64(b[2])<<16 |
		uint64(b[3])<<24 |
		uint64(b[4])<<32 |
		uint64(b[5])<<40 |
		uint64(b[6])<<48 |
		uint64(b[7])<<56
}

func fillRpcRaydiumClmmSwap(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumClmmSwapEvent); ok {
		fillStringFromAccount(&e.PoolState, get, 2)
		fillStringFromAccount(&e.Sender, get, 0)
		fillStringFromAccount(&e.TokenAccount0, get, 3)
		fillStringFromAccount(&e.TokenAccount1, get, 4)
	}
}

func fillRpcRaydiumClmmCreatePool(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumClmmCreatePoolEvent); ok {
		fillStringFromAccount(&e.Creator, get, 0)
		fillStringFromAccount(&e.Pool, get, 2)
		fillStringFromAccount(&e.Token0Mint, get, 3)
		fillStringFromAccount(&e.Token1Mint, get, 4)
		fillStringFromAccount(&e.TokenVault0, get, 5)
		fillStringFromAccount(&e.TokenVault1, get, 6)
	}
}

func fillRpcRaydiumClmmOpenPosition(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumClmmOpenPositionEvent); ok {
		fillStringFromAccount(&e.User, get, 1)
		fillStringFromAccount(&e.PositionNftMint, get, 2)
		fillStringFromAccount(&e.Pool, get, 5)
	}
}

func fillRpcRaydiumClmmClosePosition(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumClmmClosePositionEvent); ok {
		fillStringFromAccount(&e.User, get, 0)
		fillStringFromAccount(&e.PositionNftMint, get, 1)
	}
}

func fillRpcRaydiumClmmIncreaseLiquidity(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumClmmIncreaseLiquidityEvent); ok {
		fillStringFromAccount(&e.User, get, 0)
		fillStringFromAccount(&e.PositionNftMint, get, 1)
		fillStringFromAccount(&e.Pool, get, 2)
	}
}

func fillRpcRaydiumClmmDecreaseLiquidity(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumClmmDecreaseLiquidityEvent); ok {
		fillStringFromAccount(&e.User, get, 0)
		fillStringFromAccount(&e.PositionNftMint, get, 1)
		fillStringFromAccount(&e.Pool, get, 3)
	}
}

func fillRpcRaydiumCpmmDeposit(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CPMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumCpmmDepositEvent); ok {
		fillStringFromAccount(&e.User, get, 0)
	}
}

func fillRpcRaydiumCpmmSwap(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CPMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumCpmmSwapEvent); ok {
		fillStringFromAccount(&e.PoolID, get, 3)
	}
}

func fillRpcRaydiumCpmmWithdraw(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CPMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumCpmmWithdrawEvent); ok {
		fillStringFromAccount(&e.User, get, 0)
	}
}

func fillRpcRaydiumCpmmInitialize(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_CPMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumCpmmInitializeEvent); ok {
		fillStringFromAccount(&e.Creator, get, 0)
		fillStringFromAccount(&e.Pool, get, 3)
	}
}

func fillRpcRaydiumAmmV4Swap(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	e, ok := ev.Data.(*RaydiumAmmV4SwapEvent)
	if !ok {
		return
	}
	var get func(int) string
	for _, inv := range invokes[RAYDIUM_AMM_V4_PROGRAM_ID] {
		candidate := rpcAccountGetter(msg, meta, [][2]int32{inv})
		if candidate == nil || (!isDefaultPubkeyString(e.Amm) && candidate(1) != e.Amm) {
			continue
		}
		if get != nil {
			for i := 0; i < 18; i++ {
				if get(i) != candidate(i) {
					return
				}
			}
		}
		get = candidate
	}
	if get == nil {
		return
	}
	fillStringFromAccount(&e.Amm, get, 1)
	modern := isDefaultPubkeyString(get(8)) && !isDefaultPubkeyString(get(7))
	if modern {
		fillStringFromAccount(&e.TokenProgram, get, 0)
		fillStringFromAccount(&e.AmmAuthority, get, 2)
		fillStringFromAccount(&e.PoolCoinTokenAccount, get, 3)
		fillStringFromAccount(&e.PoolPcTokenAccount, get, 4)
		fillStringFromAccount(&e.UserSourceTokenAccount, get, 5)
		fillStringFromAccount(&e.UserDestinationTokenAccount, get, 6)
		fillStringFromAccount(&e.UserSourceOwner, get, 7)
	} else if isDefaultPubkeyString(get(17)) {
		fillStringFromAccount(&e.UserSourceOwner, get, 16)
	} else {
		fillStringFromAccount(&e.UserSourceOwner, get, 17)
	}
}

func fillRpcRaydiumAmmV4Deposit(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_AMM_V4_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumAmmV4DepositEvent); ok {
		fillStringFromAccount(&e.TokenProgram, get, 0)
		fillStringFromAccount(&e.AmmAuthority, get, 2)
	}
}

func fillRpcRaydiumAmmV4Withdraw(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, RAYDIUM_AMM_V4_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumAmmV4WithdrawEvent); ok {
		fillStringFromAccount(&e.TokenProgram, get, 0)
		fillStringFromAccount(&e.AmmAuthority, get, 2)
		fillStringFromAccount(&e.AmmOpenOrders, get, 3)
	}
}

func fillRpcOrcaWhirlpoolLiquidityIncreased(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, ORCA_WHIRLPOOL_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*OrcaWhirlpoolLiquidityIncreasedEvent); ok {
		fillStringFromAccount(&e.Position, get, 3)
	}
}

func fillRpcOrcaWhirlpoolLiquidityDecreased(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, ORCA_WHIRLPOOL_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*OrcaWhirlpoolLiquidityDecreasedEvent); ok {
		fillStringFromAccount(&e.Position, get, 3)
	}
}

func fillRpcMeteoraDammV2InitializePool(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, METEORA_DAMM_V2_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*MeteoraDammV2InitializePoolEvent); ok {
		fillStringFromAccount(&e.Creator, get, 0)
		fillStringFromAccount(&e.PositionNftMint, get, 1)
		fillStringFromAccount(&e.Pool, get, 6)
		fillStringFromAccount(&e.Position, get, 7)
		fillStringFromAccount(&e.TokenAMint, get, 8)
		fillStringFromAccount(&e.TokenBMint, get, 9)
	}
}

func fillRpcMeteoraDlmmSwap(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	get := rpcGetterForProgram(msg, meta, invokes, METEORA_DLMM_PROGRAM_ID)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*MeteoraDlmmSwapEvent); ok {
		fillStringFromAccount(&e.UserTokenIn, get, 4)
		fillStringFromAccount(&e.UserTokenOut, get, 5)
		fillStringFromAccount(&e.TokenXMint, get, 6)
		fillStringFromAccount(&e.TokenYMint, get, 7)
	}
}

func fillPumpFunBalances(tr *PumpFunTradeEvent, msg *RpcMessage, meta *RpcTransactionMeta) {
	if tr == nil || msg == nil || meta == nil {
		return
	}
	fullKeys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
	if idx := indexOfAccountKey(fullKeys, tr.User); idx >= 0 {
		if idx < len(meta.PreBalances) {
			v := meta.PreBalances[idx]
			tr.PreSolBalance = &v
		}
		if idx < len(meta.PostBalances) {
			v := meta.PostBalances[idx]
			tr.PostSolBalance = &v
		}
	}
	tokenIdx := indexOfAccountKey(fullKeys, tr.AssociatedUser)
	if tokenIdx < 0 || isDefaultPubkeyString(tr.AssociatedUser) {
		return
	}
	pre, preOK := findTokenBalanceAmount(meta.PreTokenBalances, uint32(tokenIdx))
	post, postOK := findTokenBalanceAmount(meta.PostTokenBalances, uint32(tokenIdx))
	if !preOK && !postOK {
		return
	}
	if !preOK {
		pre = 0
	}
	if !postOK {
		post = 0
	}
	tr.PreTokenBalance = &pre
	tr.PostTokenBalance = &post
}

func indexOfAccountKey(keys []string, key string) int {
	if isDefaultPubkeyString(key) {
		return -1
	}
	for i, k := range keys {
		if k == key {
			return i
		}
	}
	return -1
}

func findTokenBalanceAmount(balances []RpcTokenBalance, accountIndex uint32) (uint64, bool) {
	for _, b := range balances {
		if b.AccountIndex == accountIndex {
			var amt uint64
			for _, c := range b.UiTokenAmount.Amount {
				if c < '0' || c > '9' {
					return 0, false
				}
				amt = amt*10 + uint64(c-'0')
			}
			return amt, true
		}
	}
	return 0, false
}

// A shared program can appear for several pools/users in one transaction.
// Only fill unambiguous account context anchored to the actual event's pool.
func rpcLaunchlabPoolGetter(msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32, pool string, index int) func(int) string {
	if isDefaultPubkeyString(pool) {
		return nil
	}
	var selected func(int) string
	for _, inv := range invokes[RAYDIUM_LAUNCHLAB_PROGRAM_ID] {
		candidate := rpcAccountGetter(msg, meta, [][2]int32{inv})
		if candidate == nil || candidate(index) != pool {
			continue
		}
		if selected != nil {
			for i := 0; i < 18; i++ {
				if selected(i) != candidate(i) {
					return nil
				}
			}
		}
		selected = candidate
	}
	return selected
}

func fillRpcRaydiumLaunchlabTrade(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	trade, ok := ev.Data.(*RaydiumLaunchlabTradeEvent)
	if !ok {
		return
	}
	get := rpcLaunchlabPoolGetter(msg, meta, invokes, trade.PoolState, 4)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumLaunchlabTradeEvent); ok {
		fillStringFromAccount(&e.User, get, 0)
		fillStringFromAccount(&e.GlobalConfig, get, 2)
		fillStringFromAccount(&e.PlatformConfig, get, 3)
		fillStringFromAccount(&e.PoolState, get, 4)
		fillStringFromAccount(&e.UserBaseToken, get, 5)
		fillStringFromAccount(&e.UserQuoteToken, get, 6)
		fillStringFromAccount(&e.BaseVault, get, 7)
		fillStringFromAccount(&e.QuoteVault, get, 8)
		fillStringFromAccount(&e.BaseMint, get, 9)
		fillStringFromAccount(&e.QuoteMint, get, 10)
		fillStringFromAccount(&e.BaseTokenProgram, get, 11)
		fillStringFromAccount(&e.QuoteTokenProgram, get, 12)
		fillStringFromAccount(&e.SystemProgram, get, 15)
		fillStringFromAccount(&e.PlatformAssociatedAccount, get, 16)
		fillStringFromAccount(&e.CreatorAssociatedAccount, get, 17)
	}
}

func fillRpcRaydiumLaunchlabPoolCreate(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32) {
	created, ok := ev.Data.(*RaydiumLaunchlabPoolCreateEvent)
	if !ok {
		return
	}
	get := rpcLaunchlabPoolGetter(msg, meta, invokes, created.PoolState, 5)
	if get == nil {
		return
	}
	if e, ok := ev.Data.(*RaydiumLaunchlabPoolCreateEvent); ok {
		fillStringFromAccount(&e.Creator, get, 1)
		fillStringFromAccount(&e.PoolState, get, 5)
		fillStringFromAccount(&e.Payer, get, 0)
		fillStringFromAccount(&e.GlobalConfig, get, 2)
		fillStringFromAccount(&e.PlatformConfig, get, 3)
		fillStringFromAccount(&e.BaseMint, get, 6)
		fillStringFromAccount(&e.QuoteMint, get, 7)
		fillStringFromAccount(&e.BaseVault, get, 8)
		fillStringFromAccount(&e.QuoteVault, get, 9)
		if get(17) == RAYDIUM_LAUNCHLAB_PROGRAM_ID {
			fillStringFromAccount(&e.BaseTokenProgram, get, 11)
			fillStringFromAccount(&e.QuoteTokenProgram, get, 12)
		} else if get(14) == RAYDIUM_LAUNCHLAB_PROGRAM_ID {
			fillStringFromAccount(&e.BaseTokenProgram, get, 10)
			fillStringFromAccount(&e.QuoteTokenProgram, get, 11)
		}
	}
}

func fillPumpSwapIsPumpPool(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, feesInv [][2]int32) {
	if len(feesInv) == 0 {
		return
	}
	last := feesInv[len(feesInv)-1]
	data := getRpcInstructionData(msg, meta, last)
	if len(data) < 10 {
		return
	}
	isPump, ok := readBool(data, 9)
	if !ok {
		return
	}
	switch ev.Type {
	case EventTypePumpSwapBuy:
		if b, ok := ev.Data.(*PumpSwapBuyEvent); ok {
			b.IsPumpPool = isPump
		}
	case EventTypePumpSwapSell:
		if s, ok := ev.Data.(*PumpSwapSellEvent); ok {
			s.IsPumpPool = isPump
		}
	}
}

// rpcAccountGetter 对齐 Rust `find_instruction_invoke` + 账户索引解析：选取账户数最多的一次 invoke。
func rpcAccountGetter(msg *RpcMessage, meta *RpcTransactionMeta, list [][2]int32) func(int) string {
	if len(list) == 0 || msg == nil {
		return nil
	}
	fullKeys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
	best := list[0]
	bestN := -1
	for _, inv := range list {
		n := rpcInvokeAccountLen(msg, meta, inv)
		if n > bestN {
			bestN = n
			best = inv
		}
	}
	if best[1] >= 0 && meta == nil {
		return nil
	}
	return func(i int) string {
		var accounts []byte
		if best[1] >= 0 {
			for _, g := range meta.InnerInstructions {
				if g.Index == uint32(best[0]) && int(best[1]) < len(g.Instructions) {
					accounts = g.Instructions[best[1]].Accounts
					break
				}
			}
		} else if int(best[0]) < len(msg.Instructions) {
			accounts = msg.Instructions[best[0]].Accounts
		}
		if i < 0 || i >= len(accounts) {
			return ""
		}
		idx := int(accounts[i])
		if idx < len(fullKeys) {
			return fullKeys[idx]
		}
		return ""
	}
}

func rpcInvokeAccountLen(msg *RpcMessage, meta *RpcTransactionMeta, inv [2]int32) int {
	if meta == nil {
		if int(inv[0]) < len(msg.Instructions) {
			return len(msg.Instructions[inv[0]].Accounts)
		}
		return 0
	}
	if inv[1] >= 0 {
		for _, g := range meta.InnerInstructions {
			if g.Index == uint32(inv[0]) && int(inv[1]) < len(g.Instructions) {
				return len(g.Instructions[inv[1]].Accounts)
			}
		}
		return 0
	}
	if int(inv[0]) < len(msg.Instructions) {
		return len(msg.Instructions[inv[0]].Accounts)
	}
	return 0
}
