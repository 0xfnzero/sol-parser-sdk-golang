package solparser

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestPumpSwapMainnetLogAccountContext(t *testing.T) {
	raw, err := os.ReadFile("../tests/fixtures/pump_upgrade/account_context.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Program string
		Cases   []struct {
			Signature    string
			Keys         []string
			Instructions []struct {
				Accounts []int
				Data     string
			}
			Events []struct {
				Buy        bool
				Pool, User string
				Expected   map[string]string
			}
		}
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Cases {
		for _, inner := range []bool{false, true} {
			var programIndex uint32
			for i, k := range c.Keys {
				if k == f.Program {
					programIndex = uint32(i)
				}
			}
			var instructions []RpcCompiledInstruction
			for _, i := range c.Instructions {
				data, e := hex.DecodeString(i.Data)
				if e != nil {
					t.Fatal(e)
				}
				accounts := make([]byte, len(i.Accounts))
				for j, k := range i.Accounts {
					accounts[j] = byte(k)
				}
				instructions = append(instructions, RpcCompiledInstruction{ProgramIDIndex: programIndex, Accounts: accounts, Data: data})
			}
			msg := &RpcMessage{AccountKeys: c.Keys}
			meta := &RpcTransactionMeta{}
			if inner {
				meta.InnerInstructions = []RpcInnerInstructionGroup{{Index: 0, Instructions: instructions}}
			} else {
				msg.Instructions = instructions
			}
			for _, expected := range c.Events {
				var event DexEvent
				if expected.Buy {
					event = DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{Pool: expected.Pool, User: expected.User}}
				} else {
					event = DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: expected.Pool, User: expected.User}}
				}
				fillRpcDexEvents([]DexEvent{event}, msg, meta)
				data, e := json.Marshal(event.Data)
				if e != nil {
					t.Fatal(e)
				}
				var body map[string]any
				if e = json.Unmarshal(data, &body); e != nil {
					t.Fatal(e)
				}
				for field, key := range expected.Expected {
					if body[field] != key {
						t.Fatalf("%s inner=%v %s: got %v want %s", c.Signature, inner, field, body[field], key)
					}
				}
			}
		}
	}
}

func TestPumpSwapTradeContextRejectsAmbiguityAndMismatches(t *testing.T) {
	keys := make([]string, 17)
	for i := range keys {
		keys[i] = fmt.Sprint("key", i)
	}
	accounts := make([]byte, 17)
	for i := range accounts {
		accounts[i] = byte(i)
	}
	for _, mode := range []string{"duplicate", "wrong_user", "wrong_direction", "truncated", "foreign_instruction"} {
		ix := RpcCompiledInstruction{Accounts: append([]byte(nil), accounts...), Data: []byte{184, 23, 238, 97, 103, 197, 211, 61}}
		if mode == "truncated" {
			ix.Accounts = ix.Accounts[:16]
		}
		if mode == "foreign_instruction" {
			ix.Data = make([]byte, 24)
		}
		msg := &RpcMessage{AccountKeys: keys, Instructions: []RpcCompiledInstruction{ix}}
		invokes := [][2]int32{{0, -1}}
		if mode == "duplicate" {
			msg.Instructions = append(msg.Instructions, ix)
			invokes = append(invokes, [2]int32{1, -1})
		}
		user := keys[1]
		if mode == "wrong_user" {
			user = keys[2]
		}
		if get := rpcPumpSwapTradeGetter(msg, nil, invokes, keys[0], user, mode != "wrong_direction"); get != nil {
			t.Fatal("unsafe context accepted", mode)
		}
	}
}
