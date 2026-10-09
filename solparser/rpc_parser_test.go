package solparser

import (
	"encoding/base64"
	"encoding/binary"
	"testing"
)

func TestParseRpcTransactionUsesBlockTransactionIndexForLogEvents(t *testing.T) {
	tx := &RpcTransactionResponse{
		Slot:             9,
		TransactionIndex: 42,
		Meta: &RpcTransactionMeta{
			LogMessages: []string{pumpFeesUpdateAdminLogForTest()},
		},
		Transaction: &RpcTransaction{
			Message: &RpcMessage{},
		},
	}

	events, parseErr := parseRpcTransactionImpl(tx, "sig", nil, 100)
	if parseErr != nil {
		t.Fatalf("unexpected parse error: %v", parseErr)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if got := events[0].GetMetadata().TxIndex; got != 42 {
		t.Fatalf("tx_index: got %d want 42", got)
	}
}

func rpcPumpfunTradePayload(ixName string) []byte {
	data := []byte{}
	data = append(data, pumpfunTestPubkey(70)...)
	data = appendPumpfunU64(data, 10)
	data = appendPumpfunU64(data, 20)
	data = append(data, 1)
	data = append(data, pumpfunTestPubkey(71)...)
	data = appendPumpfunI64(data, 30)
	for _, v := range []uint64{40, 50, 60, 70} {
		data = appendPumpfunU64(data, v)
	}
	data = append(data, pumpfunTestPubkey(72)...)
	data = appendPumpfunU64(data, 80)
	data = appendPumpfunU64(data, 90)
	data = append(data, pumpfunTestPubkey(73)...)
	data = appendPumpfunU64(data, 100)
	data = appendPumpfunU64(data, 110)
	data = append(data, 0)
	for _, v := range []uint64{120, 130, 140} {
		data = appendPumpfunU64(data, v)
	}
	data = appendPumpfunI64(data, 150)
	data = appendPumpfunString(data, ixName)
	return data
}

func rpcPumpfunInnerTrade(ixName string) []byte {
	data := make([]byte, 16, 16+256)
	binary.LittleEndian.PutUint64(data[:8], discPumpTrade)
	copy(data[8:16], eventCPISuffix[:])
	return append(data, rpcPumpfunTradePayload(ixName)...)
}

func rpcPumpfunCreateLogForDetection(name string) string {
	data := []byte{27, 114, 169, 77, 222, 235, 99, 118}
	for _, value := range []string{name, "SDK", "https://example.invalid"} {
		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(value)))
		data = append(data, length...)
		data = append(data, []byte(value)...)
	}
	for _, seed := range []byte{70, 80, 90} {
		for index := 0; index < 32; index++ {
			data = append(data, seed)
		}
	}
	return "Program data: " + base64.StdEncoding.EncodeToString(data)
}

func rpcPumpfunTradeLog(ixName string) string {
	data := make([]byte, 8, 8+256)
	binary.LittleEndian.PutUint64(data[:8], discPumpTrade)
	data = append(data, rpcPumpfunTradePayload(ixName)...)
	return "Program data: " + base64.StdEncoding.EncodeToString(data)
}

func TestParseRpcTransactionMergesOuterAndInnerPumpfunInstructions(t *testing.T) {
	outerData := make([]byte, 8)
	binary.LittleEndian.PutUint64(outerData[:8], instrPumpOuterBuy)
	outerData = appendPumpfunU64(outerData, 123)
	outerData = appendPumpfunU64(outerData, 456)
	outerData = append(outerData, 0)

	keys := []string{PUMPFUN_PROGRAM_ID}
	for i := 1; i <= 18; i++ {
		keys = append(keys, ReadPubkey(pumpfunTestPubkey(byte(i)), 0))
	}

	keys[2] = ReadPubkey(pumpfunTestPubkey(70), 0)
	keys[6] = ReadPubkey(pumpfunTestPubkey(71), 0)

	tx := &RpcTransactionResponse{
		Slot:             7,
		TransactionIndex: 42,
		Meta: &RpcTransactionMeta{
			InnerInstructions: []RpcInnerInstructionGroup{
				{
					Index: 0,
					Instructions: []RpcCompiledInstruction{
						{ProgramIDIndex: 0, Accounts: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Data: rpcPumpfunInnerTrade("buy")},
					},
				},
			},
		},
		Transaction: &RpcTransaction{
			Message: &RpcMessage{
				AccountKeys:     keys,
				RecentBlockhash: "11111111111111111111111111111111",
				Instructions: []RpcCompiledInstruction{
					{ProgramIDIndex: 0, Accounts: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Data: outerData},
				},
			},
		},
	}

	events, parseErr := parseRpcTransactionImpl(tx, "sig", nil, 99)
	if parseErr != nil {
		t.Fatalf("unexpected parse error: %v", parseErr)
	}
	if len(events) != 1 || events[0].Type != EventTypePumpFunBuy {
		t.Fatalf("expected one PumpFunBuy event, got %+v", events)
	}
	trade, ok := events[0].Data.(*PumpFunTradeEvent)
	if !ok {
		t.Fatalf("expected PumpFunTradeEvent, got %T", events[0].Data)
	}
	if trade.SolAmount != 10 || trade.TokenAmount != 20 || trade.Amount != 123 ||
		trade.MaxSolCost != 456 || trade.BondingCurve != keys[3] {
		t.Fatalf("unexpected merged trade: %+v", trade)
	}
	if trade.Metadata.TxIndex != 42 || trade.Metadata.RecentBlockhash != "11111111111111111111111111111111" {
		t.Fatalf("metadata not preserved: %+v", trade.Metadata)
	}
}

func TestParseRpcTransactionMarksPumpfunLogTradeCreatedBuyFromWholeTransaction(t *testing.T) {
	tx := &RpcTransactionResponse{
		Slot:             7,
		TransactionIndex: 7,
		Meta: &RpcTransactionMeta{
			LogMessages: []string{
				"Program " + PUMPFUN_PROGRAM_ID + " invoke [1]",
				rpcPumpfunTradeLog("buy"),
				rpcPumpfunCreateLogForDetection("x"),
				"Program " + PUMPFUN_PROGRAM_ID + " success",
			},
		},
		Transaction: &RpcTransaction{Message: &RpcMessage{}},
	}

	events, parseErr := parseRpcTransactionImpl(tx, "sig", &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypePumpFunBuy}}, 99)
	if parseErr != nil {
		t.Fatalf("unexpected parse error: %v", parseErr)
	}
	if len(events) != 1 || events[0].Type != EventTypePumpFunBuy {
		t.Fatalf("expected one PumpFunBuy event, got %+v", events)
	}
	trade := events[0].Data.(*PumpFunTradeEvent)
	if !trade.IsCreatedBuy {
		t.Fatalf("expected is_created_buy from whole transaction create detection: %+v", trade)
	}
}

func TestPumpCreateDetectionScopeAndTradeOnlyOutput(t *testing.T) {
	const foreign = "MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr"
	create := rpcPumpfunCreateLogForDetection("x")
	scope := func(program, line string) []string {
		return []string{"Program " + program + " invoke [1]", line, "Program " + program + " success"}
	}
	cases := []struct {
		name     string
		prefix   []string
		expected bool
	}{
		{"Memo text", scope(foreign, `Program log: Memo: "Program data: G3KpTd7rY3Y"`), false},
		{"foreign data", scope(foreign, create), false},
		{"unscoped data", []string{create}, false},
		{"truncated", scope(PUMPFUN_PROGRAM_ID, "Program data: G3KpTd7rY3Y"), false},
		{"malformed base64", scope(PUMPFUN_PROGRAM_ID, create+"!"), false},
		{"nested Pump", []string{"Program " + foreign + " invoke [1]", "Program " + PUMPFUN_PROGRAM_ID + " invoke [2]", create, "Program " + PUMPFUN_PROGRAM_ID + " success", "Program " + foreign + " success"}, true},
		{"Pump event CPI", []string{"Program " + PUMPFUN_PROGRAM_ID + " invoke [1]", "Program " + PUMPFUN_PROGRAM_ID + " invoke [2]", create, "Program " + PUMPFUN_PROGRAM_ID + " success", "Program " + PUMPFUN_PROGRAM_ID + " success"}, true},
		{"foreign child", []string{"Program " + PUMPFUN_PROGRAM_ID + " invoke [1]", "Program " + foreign + " invoke [2]", create, "Program " + foreign + " success", "Program " + PUMPFUN_PROGRAM_ID + " success"}, false},
		{"scope after failed child", []string{"Program " + PUMPFUN_PROGRAM_ID + " invoke [1]", "Program " + foreign + " invoke [2]", create, "Program " + foreign + " failed: custom program error: 1", create, "Program " + PUMPFUN_PROGRAM_ID + " success"}, true},
		{"failed scope removed", []string{"Program " + PUMPFUN_PROGRAM_ID + " invoke [1]", "Program " + PUMPFUN_PROGRAM_ID + " failed: custom program error: 1", create}, false},
		{"quoted invoke", []string{"Program log: Program " + PUMPFUN_PROGRAM_ID + " invoke [1]", create}, false},
	}
	for _, name := range []string{"", "x", "xx", "xxx"} {
		cases = append(cases, struct {
			name     string
			prefix   []string
			expected bool
		}{"valid name:" + name, scope(PUMPFUN_PROGRAM_ID, rpcPumpfunCreateLogForDetection(name)), true})
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := DetectPumpfunCreateFromLogs(test.prefix); got != test.expected {
				t.Fatalf("detect=%v want=%v", got, test.expected)
			}
			logs := append(append([]string{}, test.prefix...), scope(PUMPFUN_PROGRAM_ID, rpcPumpfunTradeLog("buy"))...)
			tx := &RpcTransactionResponse{Slot: 7, Meta: &RpcTransactionMeta{LogMessages: logs}, Transaction: &RpcTransaction{Message: &RpcMessage{}}}
			events, err := ParseRpcTransaction(tx, "sig", &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypePumpFunBuy}}, 0)
			if err != nil || len(events) != 1 {
				t.Fatalf("events=%+v err=%v", events, err)
			}
			if got := events[0].Data.(*PumpFunTradeEvent).IsCreatedBuy; got != test.expected {
				t.Fatalf("buy flag=%v want=%v", got, test.expected)
			}
		})
	}
}
