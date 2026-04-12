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
			Metadata:    EventMetadata{},
			IsBuy:       true,
			IxName:      "buy_exact_sol_in",
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
