package solparser

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func TestLaunchlabCurrentLogFieldsAndInvalidTail(t *testing.T) {
	d := make([]byte, 139)
	for i := 0; i < 13; i++ {
		binary.LittleEndian.PutUint64(d[32+i*8:], uint64(100+i))
	}
	d[137] = 2
	d[138] = 1
	ev := parseRaydiumLaunchlabTradeFromData(d, EventMetadata{})
	e, ok := ev.Data.(*RaydiumLaunchlabTradeEvent)
	if !ok {
		t.Fatal("missing trade")
	}
	if e.VirtualBase != 101 || e.RealQuoteAfter != 106 || e.ProtocolFee != 109 || e.PlatformFee != 110 || e.CreatorFee != 111 || e.ShareFee != 112 || e.PoolStatus != "Trade" {
		t.Fatal(e)
	}
	for _, o := range []int{136, 137, 138} {
		bad := append([]byte{}, d...)
		bad[o] = 3
		if parseRaydiumLaunchlabTradeFromData(bad, EventMetadata{}).Type != "" {
			t.Fatal("accepted invalid enum")
		}
	}
	if parseRaydiumLaunchlabTradeFromData(append(d, 0), EventMetadata{}).Type != "" {
		t.Fatal("accepted appended bytes")
	}
}
func TestLaunchlabMigrationIdentities(t *testing.T) {
	accounts := make([]string, 32)
	for i := range accounts {
		accounts[i] = fmt.Sprintf("account_%d", i)
	}
	cp := []byte{136, 92, 200, 103, 28, 218, 144, 140}
	ev := ParseRaydiumLaunchlabInstruction(cp, accounts, "sig", 1, 0, nil, 0)
	e, ok := ev.Data.(*RaydiumLaunchlabMigrateAmmEvent)
	if !ok {
		t.Fatal("missing migration")
	}
	if e.OldPool != accounts[17] || e.NewPool != accounts[5] || e.PlatformConfig != accounts[3] || e.QuoteMint != accounts[2] || e.DestinationProgram != accounts[4] || e.LiquidityAmountKnown || e.LiquidityAmount != 0 {
		t.Fatal(e)
	}
	if ParseRaydiumLaunchlabInstruction(cp, accounts[:27], "sig", 1, 0, nil, 0).Type != "" {
		t.Fatal("accepted truncated accounts")
	}
	amm := make([]byte, 17)
	copy(amm, []byte{207, 82, 192, 145, 254, 207, 145, 223})
	ev = ParseRaydiumLaunchlabInstruction(amm, accounts, "sig", 1, 0, nil, 0)
	e, ok = ev.Data.(*RaydiumLaunchlabMigrateAmmEvent)
	if !ok || e.OldPool != accounts[23] || e.NewPool != accounts[13] || e.DestinationProgram != accounts[12] || e.PlatformConfig != zeroPubkey || e.LiquidityAmountKnown {
		t.Fatal(e)
	}
}
func TestLaunchlabPoolAnchoredAccountFill(t *testing.T) {
	keys := make([]string, 36)
	for i := range keys {
		keys[i] = fmt.Sprintf("key_%d", i)
	}
	first, second := make([]byte, 18), make([]byte, 18)
	for i := 0; i < 18; i++ {
		first[i] = byte(i)
		second[i] = byte(i + 18)
	}
	msg := &RpcMessage{AccountKeys: keys, Instructions: []RpcCompiledInstruction{{Accounts: first}, {Accounts: second}}}
	invokes := map[string][][2]int32{RAYDIUM_LAUNCHLAB_PROGRAM_ID: {{0, -1}, {1, -1}}}
	event := DexEvent{Type: EventTypeRaydiumLaunchlabTrade, Data: &RaydiumLaunchlabTradeEvent{PoolState: keys[22], User: zeroPubkey}}
	fillRpcRaydiumLaunchlabTrade(&event, msg, nil, invokes)
	e := event.Data.(*RaydiumLaunchlabTradeEvent)
	if e.BaseMint != keys[27] || e.User != keys[18] {
		t.Fatal(e)
	}
	second[4] = first[4]
	msg.Instructions[1].Accounts = second
	event.Data = &RaydiumLaunchlabTradeEvent{PoolState: keys[4], User: zeroPubkey}
	fillRpcRaydiumLaunchlabTrade(&event, msg, nil, invokes)
	e = event.Data.(*RaydiumLaunchlabTradeEvent)
	if e.User != zeroPubkey || e.PlatformConfig != "" {
		t.Fatal("guessed ambiguous user", e)
	}
}
