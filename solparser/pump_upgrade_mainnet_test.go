package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mr-tron/base58"
	"os"
	"testing"
)

type pumpUpgradeRawIx struct {
	ProgramIDIndex uint32
	Accounts       []int
	Data           string
	StackHeight    *uint32
}

func pumpUpgradeRPC(t *testing.T, b json.RawMessage) *RpcTransactionResponse {
	t.Helper()
	var raw struct {
		Slot        uint64
		BlockTime   *int64
		Transaction struct {
			Signatures []string
			Message    struct {
				AccountKeys     []string
				RecentBlockhash string
				Instructions    []pumpUpgradeRawIx
			}
		}
		Meta struct {
			Err               any
			LogMessages       []string
			LoadedAddresses   RpcLoadedAddresses
			InnerInstructions []struct {
				Index        uint32
				Instructions []pumpUpgradeRawIx
			}
		}
	}
	if e := json.Unmarshal(b, &raw); e != nil {
		t.Fatal(e)
	}
	convert := func(i pumpUpgradeRawIx) RpcCompiledInstruction {
		var data []byte
		var e error
		if i.Data != "" {
			data, e = base58.Decode(i.Data)
		}
		if e != nil {
			t.Fatal(e)
		}
		indices := make([]byte, len(i.Accounts))
		for j, v := range i.Accounts {
			if v < 0 || v > 255 {
				t.Fatal("invalid index")
			}
			indices[j] = byte(v)
		}
		return RpcCompiledInstruction{ProgramIDIndex: i.ProgramIDIndex, Accounts: indices, Data: data, StackHeight: i.StackHeight}
	}
	msg := &RpcMessage{AccountKeys: raw.Transaction.Message.AccountKeys, RecentBlockhash: raw.Transaction.Message.RecentBlockhash}
	for _, i := range raw.Transaction.Message.Instructions {
		msg.Instructions = append(msg.Instructions, convert(i))
	}
	meta := &RpcTransactionMeta{Err: raw.Meta.Err, LogMessages: raw.Meta.LogMessages, LoadedAddresses: &raw.Meta.LoadedAddresses}
	for _, g := range raw.Meta.InnerInstructions {
		out := RpcInnerInstructionGroup{Index: g.Index}
		for _, i := range g.Instructions {
			out.Instructions = append(out.Instructions, convert(i))
		}
		meta.InnerInstructions = append(meta.InnerInstructions, out)
	}
	return &RpcTransactionResponse{Slot: raw.Slot, BlockTime: raw.BlockTime, Transaction: &RpcTransaction{Signatures: raw.Transaction.Signatures, Message: msg}, Meta: meta}
}

func TestCapturedPumpUpgradeMainnet(t *testing.T) {
	b, e := os.ReadFile("../tests/fixtures/pump_upgrade/mainnet.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Cases []struct {
			Signature string
			Raw       json.RawMessage
			Expected  []struct {
				Name string
				Data map[string]any
			}
		}
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	extra, e := os.ReadFile("../tests/fixtures/pump_upgrade/simulation_events.json")
	if e != nil {
		t.Fatal(e)
	}
	g := f
	g.Cases = nil
	if e = json.Unmarshal(extra, &g); e != nil {
		t.Fatal(e)
	}
	f.Cases = append(f.Cases, g.Cases...)
	for _, c := range f.Cases {
		t.Run(c.Signature[:12], func(t *testing.T) {
			tx := pumpUpgradeRPC(t, c.Raw)
			events, pe := ParseRpcTransaction(tx, c.Signature, nil, 0)
			if pe != nil {
				t.Fatal(pe)
			}
			if tx.Meta.Err != nil {
				if len(events) != 0 {
					t.Fatal("failed transaction emitted events")
				}
				return
			}
			bodies := []map[string]any{}
			for _, ev := range events {
				b, e := json.Marshal(ev.Data)
				if e != nil {
					t.Fatal(e)
				}
				var d map[string]any
				dec := json.NewDecoder(bytes.NewReader(b))
				dec.UseNumber()
				if e = dec.Decode(&d); e != nil {
					t.Fatal(e)
				}
				bodies = append(bodies, d)
			}
			for _, expected := range c.Expected {
				identity := "pool"
				if expected.Name == "TradeEvent" {
					identity = "mint"
				}
				matches := []map[string]any{}
				for _, d := range bodies {
					if d[identity] != expected.Data[identity] || d["user"] != expected.Data["user"] || (expected.Data["is_buy"] != nil && d["is_buy"] != expected.Data["is_buy"]) {
						continue
					}
					if expected.Name == "BuyEvent" && d["base_amount_out"] == nil {
						continue
					}
					if expected.Name == "SellEvent" && d["base_amount_in"] == nil {
						continue
					}
					matches = append(matches, d)
				}
				if len(matches) != 1 {
					t.Fatalf("expected one %s, got %d", expected.Name, len(matches))
				}
				for k, v := range expected.Data {
					if k == "shareholders" {
						continue
					}
					got := matches[0][k]
					if k == "quote_mint" && v == "11111111111111111111111111111111" && (got == v || got == "So11111111111111111111111111111111111111111" || got == "So11111111111111111111111111111111111111112") {
						continue
					}
					if fmt.Sprint(got) != fmt.Sprint(v) {
						t.Errorf("%s got %v want %v", k, got, v)
					}
				}
			}
		})
	}
}
