package solparser

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestPumpTradeLogAccountContext(t *testing.T) {
	raw, err := os.ReadFile("../tests/fixtures/pump_upgrade/pump_trade_account_context.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Program string
		Cases   []struct {
			Name         string
			Keys         []string
			Mint, User   string
			Buy          bool
			Expected     map[string]string
			Instructions []struct {
				Accounts []int
				Data     string
			}
		}
	}
	if err = json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			var instructions []RpcCompiledInstruction
			for _, ix := range c.Instructions {
				data, err := hex.DecodeString(ix.Data)
				if err != nil {
					t.Fatal(err)
				}
				accounts := make([]byte, len(ix.Accounts))
				for j, k := range ix.Accounts {
					accounts[j] = byte(k)
				}
				var pid uint32
				for j, k := range c.Keys {
					if k == f.Program {
						pid = uint32(j)
					}
				}
				instructions = append(instructions, RpcCompiledInstruction{ProgramIDIndex: pid, Accounts: accounts, Data: data})
			}
			for _, inner := range []bool{false, true} {
				msg, meta := &RpcMessage{AccountKeys: c.Keys}, &RpcTransactionMeta{}
				if inner {
					meta.InnerInstructions = []RpcInnerInstructionGroup{{Index: 0, Instructions: instructions}}
				} else {
					msg.Instructions = instructions
				}
				body := &PumpFunTradeEvent{Mint: c.Mint, User: c.User, IsBuy: c.Buy}
				fillRpcDexEvents([]DexEvent{{Type: EventTypePumpFunTrade, Data: body}}, msg, meta)
				data, _ := json.Marshal(body)
				var got map[string]any
				json.Unmarshal(data, &got)
				for field, want := range c.Expected {
					if got[field] != want {
						t.Fatalf("inner=%v %s: got %v want %s", inner, field, got[field], want)
					}
				}
			}
			for _, mode := range []string{"duplicate", "wrong_mint", "wrong_user", "wrong_direction", "truncated", "foreign_instruction"} {
				ix := instructions[0]
				ix.Accounts = append([]byte(nil), ix.Accounts...)
				if mode == "truncated" {
					ix.Accounts = ix.Accounts[:len(ix.Accounts)-1]
				}
				if mode == "foreign_instruction" {
					ix.Data = make([]byte, 24)
				}
				msg := &RpcMessage{AccountKeys: c.Keys, Instructions: []RpcCompiledInstruction{ix}}
				if mode == "duplicate" {
					msg.Instructions = append(msg.Instructions, ix)
				}
				body := &PumpFunTradeEvent{Mint: c.Mint, User: c.User, IsBuy: c.Buy}
				if mode == "wrong_mint" {
					body.Mint = c.Keys[len(c.Keys)-2]
				}
				if mode == "wrong_user" {
					body.User = c.Keys[len(c.Keys)-2]
				}
				if mode == "wrong_direction" {
					body.IsBuy = !body.IsBuy
				}
				fillRpcDexEvents([]DexEvent{{Type: EventTypePumpFunTrade, Data: body}}, msg, &RpcTransactionMeta{})
				if !isDefaultPubkeyString(body.AssociatedUser) {
					t.Fatal("unsafe context", mode, body.AssociatedUser)
				}
			}
		})
	}
}
