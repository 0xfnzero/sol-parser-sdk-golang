package solparser

import (
	"reflect"
	"testing"
)

func TestTruncatedSourceMapping(t *testing.T) {
	msg := &RpcMessage{Instructions: []RpcCompiledInstruction{{ProgramIDIndex: 0}, {ProgramIDIndex: 0}}}
	if got := logOuterIndices([]string{"Program router invoke [1]", "Program data: x"}, msg, []string{"router"}); !reflect.DeepEqual(got, []int{-1, -1}) {
		t.Fatal(got)
	}
}
func TestExecutionSourcePayload(t *testing.T) {
	msg := &RpcMessage{Instructions: []RpcCompiledInstruction{{ProgramIDIndex: 0, Accounts: []byte{0, 0, 250}, Data: []byte{1, 2, 3}}}}
	trade := &PumpFunTradeEvent{}
	markTradeExecutionSource(DexEvent{Type: EventTypePumpFunTrade, Data: trade}, 0, msg, []string{"router"}, &RpcTransactionMeta{})
	raw := trade.ThirdProgram
	if trade.ParseStatus != TradeParseExecutionConfirmed || raw == nil || raw.InstructionIndex != 0 || raw.Accounts[2] != nil || *raw.Accounts[0] != *raw.Accounts[1] || !reflect.DeepEqual(raw.InstructionData, []int{1, 2, 3}) {
		t.Fatal(trade)
	}
}
func TestOtherDexOccurrenceMismatch(t *testing.T) {
	fixtures := []DexEvent{
		{Type: EventTypeRaydiumClmmSwap, Data: &RaydiumClmmSwapEvent{PoolState: "pool"}},
		{Type: EventTypeRaydiumCpmmSwap, Data: &RaydiumCpmmSwapEvent{PoolID: "pool"}},
		{Type: EventTypeOrcaWhirlpoolSwap, Data: &OrcaWhirlpoolSwapEvent{Whirlpool: "pool"}},
		{Type: EventTypeMeteoraDlmmSwap, Data: &MeteoraDlmmSwapEvent{Pool: "pool"}},
		{Type: EventTypeRaydiumLaunchlabTrade, Data: &RaydiumLaunchlabTradeEvent{PoolState: "pool"}},
	}
	for _, e := range fixtures {
		for _, reversed := range []bool{false, true} {
			logs, instructions := []DexEvent{e}, []DexEvent{e, e}
			if reversed {
				logs, instructions = instructions, logs
			}
			if got := DedupeLogInstructionEvents(logs, instructions); len(got) != 3 {
				t.Fatal(e.Type, len(got))
			}
		}
	}
}
