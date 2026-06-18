package solparser

import "testing"

// 回归：Program data 日志解析出的 PumpFun Trade（无指令账户）须能通过 fillRpcDexEventsPump
// 用交易 message 中任意一次 PumpFun invoke 的账户表补全 bonding_curve 等字段。
func TestFillRpcDexEventsPump_LogDerivedBuyExactSolIn(t *testing.T) {
	// 与 IDL buy 一致：0 global … 3 bonding_curve, 4 associated_bonding_curve, 8 token_program, 9 creator_vault
	keys := []string{
		"g0", "g1", "mintMINT", "bcBONDING", "abcASSOC", "auser", "userUSER",
		"SysProg1111111111111111111111111111111",
		"TokenKEK111111111111111111111111111111",
		"CreatorVaultAddrVVVVVVVVVVVVVVVVVVVVVVVV",
		"e10", "e11", "e12", "e13", "e14", "e15",
		PUMPFUN_PROGRAM_ID,
	}
	progIx := uint32(len(keys) - 1)
	accs := make([]byte, 16)
	for i := range accs {
		accs[i] = byte(i)
	}
	msg := &RpcMessage{
		AccountKeys: keys,
		Instructions: []RpcCompiledInstruction{
			{
				ProgramIDIndex: progIx,
				Accounts:       accs,
				Data:           []byte{1, 2, 3},
			},
		},
	}
	ev := DexEvent{
		Type: EventTypePumpFunBuyExactSolIn,
		Data: &PumpFunTradeEvent{
			Metadata:     EventMetadata{},
			IsBuy:        true,
			IxName:       "buy_exact_sol_in",
			BondingCurve: "",
		},
	}
	events := []DexEvent{ev}
	fillRpcDexEventsPump(events, msg, nil)

	tr := events[0].Data.(*PumpFunTradeEvent)
	if tr.BondingCurve != "bcBONDING" {
		t.Fatalf("bonding_curve: got %q want bcBONDING", tr.BondingCurve)
	}
	if tr.AssociatedBondingCurve != "abcASSOC" {
		t.Fatalf("associated_bonding_curve: got %q", tr.AssociatedBondingCurve)
	}
	if tr.TokenProgram != "TokenKEK111111111111111111111111111111" {
		t.Fatalf("token_program: got %q", tr.TokenProgram)
	}
	if tr.CreatorVault != "CreatorVaultAddrVVVVVVVVVVVVVVVVVVVVVVVV" {
		t.Fatalf("creator_vault: got %q", tr.CreatorVault)
	}
	if tr.User != "userUSER" {
		t.Fatalf("user: got %q", tr.User)
	}
}

func TestFillRpcDexEventsPumpFunCreateAccounts(t *testing.T) {
	keys := []string{
		"mint", "mintAuthority", "bondingCurve", "associatedBondingCurve", "global",
		"mplTokenMetadata", "metadata", "user", "systemProgram", "tokenProgram",
		"associatedTokenProgram", "rent", "eventAuthority", "program", PUMPFUN_PROGRAM_ID,
	}
	msg := rpcMessageWithProgramInvoke(keys, PUMPFUN_PROGRAM_ID, sequentialAccounts(14))
	msg.Instructions[0].Data = discBytes(instrPumpOuterCreate)
	events := []DexEvent{{
		Type: EventTypePumpFunCreate,
		Data: &PumpFunCreateEvent{
			Mint:                   zeroPubkey,
			BondingCurve:           zeroPubkey,
			User:                   zeroPubkey,
			TokenProgram:           zeroPubkey,
			MintAuthority:          zeroPubkey,
			AssociatedBondingCurve: zeroPubkey,
			Global:                 zeroPubkey,
			SystemProgram:          zeroPubkey,
			AssociatedTokenProgram: zeroPubkey,
			EventAuthority:         zeroPubkey,
			Program:                zeroPubkey,
		},
	}}

	fillRpcDexEvents(events, msg, nil)

	create := events[0].Data.(*PumpFunCreateEvent)
	if create.IxName != "create" || create.Mint != "mint" || create.MintAuthority != "mintAuthority" ||
		create.BondingCurve != "bondingCurve" || create.AssociatedBondingCurve != "associatedBondingCurve" ||
		create.Global != "global" || create.User != "user" || create.SystemProgram != "systemProgram" ||
		create.TokenProgram != "tokenProgram" || create.AssociatedTokenProgram != "associatedTokenProgram" ||
		create.EventAuthority != "eventAuthority" || create.Program != "program" {
		t.Fatalf("unexpected PumpFun Create account fill: %+v", create)
	}
}

func TestFillRpcDexEventsRaydiumClmmLogDerivedSwap(t *testing.T) {
	keys := []string{
		"userPAYER", "ammConfig", "poolSTATE", "tokenAccount0", "tokenAccount1",
		"inputVault", "outputVault", "observation", RAYDIUM_CLMM_PROGRAM_ID,
	}
	msg := rpcMessageWithProgramInvoke(keys, RAYDIUM_CLMM_PROGRAM_ID, []byte{0, 1, 2, 3, 4, 5, 6, 7})
	events := []DexEvent{{
		Type: EventTypeRaydiumClmmSwap,
		Data: &RaydiumClmmSwapEvent{
			PoolState:     zeroPubkey,
			Sender:        zeroPubkey,
			TokenAccount0: zeroPubkey,
			TokenAccount1: zeroPubkey,
		},
	}}

	fillRpcDexEvents(events, msg, nil)

	swap := events[0].Data.(*RaydiumClmmSwapEvent)
	if swap.PoolState != "poolSTATE" || swap.Sender != "userPAYER" ||
		swap.TokenAccount0 != "tokenAccount0" || swap.TokenAccount1 != "tokenAccount1" {
		t.Fatalf("unexpected CLMM account fill: %+v", swap)
	}
}

func TestFillRpcDexEventsPumpSwapLogDerivedBuy(t *testing.T) {
	keys := []string{
		"pool", "user", "authority", "baseMint", "quoteMint", "userBase",
		"userQuote", "poolBase", "poolQuote", "protocolFee", "protocolFeeToken",
		"baseTokenProgram", "quoteTokenProgram", "sys", "ata", "eventAuthority",
		"program", "coinCreatorVaultAta", "coinCreatorVaultAuthority", "extra19",
		"extra20", "extra21", "extra22", "poolV2", "feeRecipient",
		"feeRecipientQuoteTokenAccount", PUMPSWAP_PROGRAM_ID,
	}
	msg := rpcMessageWithProgramInvoke(keys, PUMPSWAP_PROGRAM_ID, sequentialAccounts(26))
	events := []DexEvent{{
		Type: EventTypePumpSwapBuy,
		Data: &PumpSwapBuyEvent{
			Pool:                          zeroPubkey,
			User:                          zeroPubkey,
			BaseMint:                      zeroPubkey,
			QuoteMint:                     zeroPubkey,
			FeeRecipientQuoteTokenAccount: zeroPubkey,
		},
	}}

	fillRpcDexEvents(events, msg, nil)

	buy := events[0].Data.(*PumpSwapBuyEvent)
	if buy.Pool != "pool" || buy.User != "user" ||
		buy.BaseMint != "baseMint" || buy.QuoteMint != "quoteMint" ||
		buy.PoolV2 != "poolV2" || buy.FeeRecipient != "feeRecipient" ||
		buy.FeeRecipientQuoteTokenAccount != "feeRecipientQuoteTokenAccount" {
		t.Fatalf("unexpected PumpSwap account fill: %+v", buy)
	}
}

func TestFillRpcDexEventsMeteoraDammV2InitializePoolAccounts(t *testing.T) {
	keys := []string{
		"creator", "positionNftMint", "a2", "a3", "a4", "a5", "pool",
		"position", "tokenAMint", "tokenBMint", METEORA_DAMM_V2_PROGRAM_ID,
	}
	msg := rpcMessageWithProgramInvoke(keys, METEORA_DAMM_V2_PROGRAM_ID, sequentialAccounts(10))
	events := []DexEvent{{
		Type: EventTypeMeteoraDammV2InitializePool,
		Data: &MeteoraDammV2InitializePoolEvent{
			Creator:         zeroPubkey,
			PositionNftMint: zeroPubkey,
			Pool:            zeroPubkey,
			Position:        zeroPubkey,
			TokenAMint:      zeroPubkey,
			TokenBMint:      zeroPubkey,
		},
	}}

	fillRpcDexEvents(events, msg, nil)

	init := events[0].Data.(*MeteoraDammV2InitializePoolEvent)
	if init.Creator != "creator" || init.PositionNftMint != "positionNftMint" ||
		init.Pool != "pool" || init.Position != "position" ||
		init.TokenAMint != "tokenAMint" || init.TokenBMint != "tokenBMint" {
		t.Fatalf("unexpected DAMM InitializePool account fill: %+v", init)
	}
}

func rpcMessageWithProgramInvoke(keys []string, programID string, accounts []byte) *RpcMessage {
	progIdx := -1
	for i, key := range keys {
		if key == programID {
			progIdx = i
			break
		}
	}
	if progIdx < 0 {
		panic("program id missing from keys")
	}
	return &RpcMessage{
		AccountKeys: keys,
		Instructions: []RpcCompiledInstruction{
			{
				ProgramIDIndex: uint32(progIdx),
				Accounts:       accounts,
				Data:           []byte{1, 2, 3},
			},
		},
	}
}

func sequentialAccounts(n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = byte(i)
	}
	return out
}

func discBytes(d uint64) []byte {
	return []byte{
		byte(d),
		byte(d >> 8),
		byte(d >> 16),
		byte(d >> 24),
		byte(d >> 32),
		byte(d >> 40),
		byte(d >> 48),
		byte(d >> 56),
	}
}
