package solparser

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestPumpFunNativeDebitEvidence(t *testing.T) {
	depth, childDepth := 1, 2
	ix := routeIx{Position: InstructionPosition{StackHeight: &depth}, Accounts: make([]string, 17)}
	ix.Accounts[2] = routeWSOL
	ix.Accounts[6] = "fee"
	ix.Accounts[8] = "buyback"
	ix.Accounts[10] = "pool"
	ix.Accounts[16] = "creator"
	leg := &RouteSwapLeg{Protocol: "PumpFun", Trader: "wallet", InputAccount: "wallet", Pool: "pool"}
	payment := func(dest string, n uint64) routeIx {
		data := make([]byte, 12)
		binary.LittleEndian.PutUint32(data, 2)
		binary.LittleEndian.PutUint64(data[4:], n)
		return routeIx{Position: InstructionPosition{StackHeight: &childDepth}, Program: zeroPubkey, Accounts: []string{"wallet", dest}, Data: data}
	}
	children := []routeIx{payment("pool", 9876), payment("fee", 47), payment("buyback", 47), payment("creator", 30)}
	amount := pumpfunNativeDebit(ix, children, leg)
	if amount == nil || *amount != 10000 {
		t.Fatal("incorrect protocol debit")
	}
	for _, bad := range [][]routeIx{{payment("fee", 1)}, {payment("pool", 1), payment("unknown", 1)}, {payment("pool", math.MaxUint64), payment("fee", 1)}} {
		if pumpfunNativeDebit(ix, bad, leg) != nil {
			t.Fatal("ambiguous or overflowing evidence accepted")
		}
	}
	children[0].Position.StackHeight = nil
	if pumpfunNativeDebit(ix, children, leg) != nil {
		t.Fatal("missing pool depth accepted")
	}
	nested := 3
	children[0].Position.StackHeight = &nested
	if pumpfunNativeDebit(ix, children, leg) != nil {
		t.Fatal("nested pool payment accepted")
	}
}
