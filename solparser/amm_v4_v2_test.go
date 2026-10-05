package solparser

import (
	"encoding/binary"
	"testing"
)

func TestAmmV4V2EightAccountEvent(t *testing.T) {
	for _, tag := range []byte{16, 17} {
		keys := []string{"token", "pool", "authority", "coin-vault", "pc-vault", "source", "destination", "owner"}
		d := make([]byte, 17)
		d[0] = tag
		binary.LittleEndian.PutUint64(d[1:], 1001)
		binary.LittleEndian.PutUint64(d[9:], 500)
		e := ParseRaydiumAmmV4Instruction(d, keys, "signature", 123, 4, nil, 890)
		p, ok := e.Data.(*RaydiumAmmV4SwapEvent)
		if !ok || p.Amm != keys[1] || p.PoolCoinTokenAccount != keys[3] || p.PoolPcTokenAccount != keys[4] || p.UserSourceTokenAccount != keys[5] || p.UserDestinationTokenAccount != keys[6] || p.UserSourceOwner != keys[7] || p.AmmOpenOrders != "11111111111111111111111111111111" || p.SerumProgram != "11111111111111111111111111111111" {
			t.Fatal(e)
		}
		if tag == 16 && (p.AmountIn != 1001 || p.MinimumAmountOut != 500 || p.MaxAmountIn != 0 || p.AmountOut != 0) {
			t.Fatal(p)
		}
		if tag == 17 && (p.MaxAmountIn != 1001 || p.AmountOut != 500 || p.AmountIn != 0 || p.MinimumAmountOut != 0) {
			t.Fatal(p)
		}
		if ParseRaydiumAmmV4Instruction(d, keys[:7], "s", 1, 0, nil, 0).Data != nil || ParseRaydiumAmmV4Instruction(d[:16], keys, "s", 1, 0, nil, 0).Data != nil {
			t.Fatal("truncated V2 accepted")
		}
	}
}
func TestAmmV4LogContextV2Unambiguous(t *testing.T) {
	for _, mode := range []string{"single", "anchored", "ambiguous"} {
		keys := []string{"t1", "p1", "a1", "c1", "q1", "s1", "d1", "u1", "t2", "p2", "a2", "c2", "q2", "s2", "d2", "u2"}
		msg := &RpcMessage{AccountKeys: keys, Instructions: []RpcCompiledInstruction{{Accounts: []byte{0, 1, 2, 3, 4, 5, 6, 7}}, {Accounts: []byte{8, 9, 10, 11, 12, 13, 14, 15}}}}
		list := [][2]int32{{0, -1}}
		if mode != "single" {
			list = append(list, [2]int32{1, -1})
		}
		p := &RaydiumAmmV4SwapEvent{}
		if mode == "anchored" {
			p.Amm = "p2"
		}
		e := DexEvent{Type: EventTypeRaydiumAmmV4Swap, Data: p}
		fillRpcRaydiumAmmV4Swap(&e, msg, nil, map[string][][2]int32{RAYDIUM_AMM_V4_PROGRAM_ID: list})
		if mode == "ambiguous" {
			if p.Amm != "" || p.UserSourceOwner != "" {
				t.Fatal(p)
			}
		} else {
			offset := 0
			if mode == "anchored" {
				offset = 8
			}
			if p.UserSourceOwner != keys[offset+7] || p.PoolCoinTokenAccount != keys[offset+3] {
				t.Fatal(p)
			}
		}
	}
}
