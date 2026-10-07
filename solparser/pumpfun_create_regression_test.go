package solparser

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mr-tron/base58"
)

func TestPumpFunCreateMainnetAccountAndLogRegression(t *testing.T) {
	for _, name := range []string{"create_v2_with_dev_buy", "legacy_create_2024"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/pumpfun_create_mainnet/" + name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			type instruction struct {
				ProgramIDIndex uint32
				Accounts       []int
				Data           string
				StackHeight    *uint32
			}
			var raw struct {
				Slot        uint64
				BlockTime   *int64
				Transaction struct {
					Signatures []string
					Message    struct {
						AccountKeys  []string
						Instructions []instruction
					}
				}
				Meta struct {
					LogMessages       []string
					InnerInstructions []struct {
						Index        uint32
						Instructions []instruction
					}
					LoadedAddresses RpcLoadedAddresses
				}
			}
			if err = json.Unmarshal(data, &raw); err != nil {
				t.Fatal(err)
			}
			decode := func(s string) []byte {
				if s == "" {
					return nil
				}
				b, e := base58.Decode(s)
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			convert := func(i instruction) RpcCompiledInstruction {
				a := make([]byte, len(i.Accounts))
				for j, v := range i.Accounts {
					a[j] = byte(v)
				}
				return RpcCompiledInstruction{ProgramIDIndex: i.ProgramIDIndex, Accounts: a, Data: decode(i.Data), StackHeight: i.StackHeight}
			}
			msg := &RpcMessage{AccountKeys: raw.Transaction.Message.AccountKeys}
			for _, i := range raw.Transaction.Message.Instructions {
				msg.Instructions = append(msg.Instructions, convert(i))
			}
			meta := &RpcTransactionMeta{LogMessages: raw.Meta.LogMessages, LoadedAddresses: &raw.Meta.LoadedAddresses}
			for _, g := range raw.Meta.InnerInstructions {
				out := RpcInnerInstructionGroup{Index: g.Index}
				for _, i := range g.Instructions {
					out.Instructions = append(out.Instructions, convert(i))
				}
				meta.InnerInstructions = append(meta.InnerInstructions, out)
			}
			sig := raw.Transaction.Signatures[0]
			tx := &RpcTransactionResponse{Slot: raw.Slot, BlockTime: raw.BlockTime, Meta: meta, Transaction: &RpcTransaction{Signatures: raw.Transaction.Signatures, Message: msg}}
			events, pe := ParseRpcTransaction(tx, sig, nil, 0)
			if pe != nil {
				t.Fatal(pe)
			}
			keys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
			var creating *RpcCompiledInstruction
			for i := range msg.Instructions {
				ix := &msg.Instructions[i]
				if keys[ix.ProgramIDIndex] != PUMPFUN_PROGRAM_ID {
					continue
				}
				disc := disc8FromBytes(ix.Data)
				if disc == instrPumpOuterCreate || disc == instrPumpOuterCreateV2 {
					creating = ix
					break
				}
			}
			if creating == nil {
				t.Fatal("fixture has no creating instruction")
			}
			get := func(i int) string { return keys[creating.Accounts[i]] }
			v2 := disc8FromBytes(creating.Data) == instrPumpOuterCreateV2
			found := 0
			for _, ev := range events {
				// Log and instruction creates are deduplicated into one public event.
				c, ok := ev.Data.(*PumpFunCreateEvent)
				if !ok {
					continue
				}
				found++
				user, system, token, ata := 7, 8, 9, 10
				if v2 {
					user, system, token, ata = 5, 6, 7, 8
				}
				if c.Mint != get(0) || c.BondingCurve != get(2) || c.User != get(user) || c.SystemProgram != get(system) || c.TokenProgram != get(token) || c.AssociatedTokenProgram != get(ata) {
					t.Fatalf("create fields were filled from a different instruction: %+v", c)
				}
				if !isDefaultPubkeyString(c.QuoteMint) || !isDefaultPubkeyString(c.QuoteVault) || !isDefaultPubkeyString(c.QuoteTokenProgram) {
					t.Fatalf("unrelated accounts leaked into quote fields: %+v", c)
				}
			}
			if found != 1 {
				t.Fatalf("create count=%d, want 1", found)
			}
			logs := ParseLogsOnly(raw.Meta.LogMessages, sig, raw.Slot, nil)
			count := 0
			for _, ev := range logs {
				if c, ok := ev.Data.(*PumpFunCreateEvent); ok {
					count++
					if c.Mint != get(0) || c.BondingCurve != get(2) || c.User != get(mapBoolIndex(v2, 5, 7)) {
						t.Fatalf("wrong create log: %+v", c)
					}
				}
			}
			if count != 1 {
				t.Fatalf("log create count=%d, want 1", count)
			}
		})
	}
}

func TestPumpFunCreateDoesNotInferQuoteFromRemainingAccounts(t *testing.T) {
	ev := ParsePumpfunInstruction(pumpfunCreateV2Instruction(false, false, 0, false), pumpfunV2TestAccounts(19), "sig", 1, 0, nil, 0)
	c := ev.Data.(*PumpFunCreateV2TokenEvent)
	if !isDefaultPubkeyString(c.QuoteMint) || !isDefaultPubkeyString(c.QuoteVault) || !isDefaultPubkeyString(c.QuoteTokenProgram) {
		t.Fatalf("remaining accounts interpreted as quote fields: %+v", c)
	}
}

func TestPumpFunCreateAccountGetterSelectsMintAndDiscriminator(t *testing.T) {
	for _, inner := range []bool{false, true} {
		msg := &RpcMessage{AccountKeys: []string{PUMPFUN_PROGRAM_ID}}
		var invokes [][2]int32
		meta := &RpcTransactionMeta{}
		for j, mint := range []string{"firstMint", "secondMint"} {
			a := pumpfunV2TestAccounts(16)
			a[0] = mint
			indices := make([]byte, len(a))
			for i, key := range a {
				indices[i] = byte(len(msg.AccountKeys))
				msg.AccountKeys = append(msg.AccountKeys, key)
			}
			ix := RpcCompiledInstruction{ProgramIDIndex: 0, Accounts: indices, Data: discBytes(instrPumpOuterCreateV2)}
			if inner {
				if len(meta.InnerInstructions) == 0 {
					meta.InnerInstructions = append(meta.InnerInstructions, RpcInnerInstructionGroup{Index: 0})
				}
				meta.InnerInstructions[0].Instructions = append(meta.InnerInstructions[0].Instructions, ix)
				invokes = append(invokes, [2]int32{0, int32(j)})
			} else {
				msg.Instructions = append(msg.Instructions, ix)
				invokes = append(invokes, [2]int32{int32(j), -1})
			}
		}
		ev := DexEvent{Type: EventTypePumpFunCreateV2, Data: &PumpFunCreateV2TokenEvent{Mint: "secondMint"}}
		get, v2 := rpcPumpFunCreateAccountGetter(msg, meta, invokes, &ev)
		if get == nil || !v2 || get(0) != "secondMint" {
			t.Fatal("did not select the matching mint's create_v2")
		}
		ev.Data = &PumpFunCreateV2TokenEvent{}
		if get, _ = rpcPumpFunCreateAccountGetter(msg, meta, invokes, &ev); get != nil {
			t.Fatal("ambiguous creates without mint must not be guessed")
		}
		ev.Data = &PumpFunCreateV2TokenEvent{Mint: "unrelatedMint"}
		if get, _ = rpcPumpFunCreateAccountGetter(msg, meta, invokes, &ev); get != nil {
			t.Fatal("unrelated mint matched")
		}
	}
	// A buy cannot fill a create, even if it is the only PumpFun invocation.
	msg := rpcMessageWithProgramInvoke(append(pumpfunV2TestAccounts(19), PUMPFUN_PROGRAM_ID), PUMPFUN_PROGRAM_ID, sequentialAccounts(19))
	msg.Instructions[0].Data = discBytes(instrPumpOuterBuy)
	ev := DexEvent{Type: EventTypePumpFunCreate, Data: &PumpFunCreateEvent{}}
	if get, _ := rpcPumpFunCreateAccountGetter(msg, nil, [][2]int32{{0, -1}}, &ev); get != nil {
		t.Fatal("buy selected as a create")
	}
}

func TestPumpFunHistoricalCreateLogBounds(t *testing.T) {
	prefix := appendPumpfunString(nil, "Historical")
	prefix = appendPumpfunString(prefix, "OLD")
	prefix = appendPumpfunString(prefix, "https://example.invalid/old")
	body := append([]byte{}, prefix...)
	for _, v := range []byte{1, 2, 3} {
		body = append(body, pumpfunTestPubkey(v)...)
	}
	ev := parseCreateFromData(body, EventMetadata{})
	c, ok := ev.Data.(*PumpFunCreateEvent)
	if !ok || c.Mint != ReadPubkey(pumpfunTestPubkey(1), 0) || c.BondingCurve != ReadPubkey(pumpfunTestPubkey(2), 0) || c.User != ReadPubkey(pumpfunTestPubkey(3), 0) {
		t.Fatalf("historical log not decoded: %+v", ev)
	}
	if c.Timestamp != 0 || c.VirtualTokenReserves != 0 || !isDefaultPubkeyString(c.Creator) || !isDefaultPubkeyString(c.TokenProgram) || !isDefaultPubkeyString(c.QuoteMint) {
		t.Fatal("historical event fabricated unavailable fields")
	}
	for n := 0; n < 201; n++ {
		if n == 96 {
			continue
		}
		data := append(append([]byte{}, prefix...), make([]byte, n)...)
		if ev := parseCreateFromData(data, EventMetadata{}); ev.Type != "" {
			t.Fatalf("accepted incomplete/unsupported create layout of %d bytes", n)
		}
	}
}

func TestPumpFunCreateAccountFillPreservesDecodedQuote(t *testing.T) {
	keys := append(pumpfunV2TestAccounts(19), PUMPFUN_PROGRAM_ID)
	msg := rpcMessageWithProgramInvoke(keys, PUMPFUN_PROGRAM_ID, sequentialAccounts(19))
	msg.Instructions[0].Data = discBytes(instrPumpOuterCreateV2)
	c := &PumpFunCreateV2TokenEvent{Mint: keys[0], QuoteMint: "decodedUSDC", QuoteVault: "decodedVault", QuoteTokenProgram: "decodedTokenProgram", VirtualQuoteReserves: 123}
	events := []DexEvent{{Type: EventTypePumpFunCreateV2, Data: c}}
	fillRpcDexEvents(events, msg, nil)
	if c.QuoteMint != "decodedUSDC" || c.QuoteVault != "decodedVault" || c.QuoteTokenProgram != "decodedTokenProgram" || c.VirtualQuoteReserves != 123 {
		t.Fatal("authoritative quote fields were overwritten")
	}
	if c.TokenProgram != keys[7] {
		t.Fatal("verified create_v2 accounts were not filled")
	}
}
