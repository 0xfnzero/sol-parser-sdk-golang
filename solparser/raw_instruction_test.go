package solparser

import (
	"encoding/base64"
	"encoding/json"
	"github.com/mr-tron/base58"
	"os"
	"reflect"
	"testing"
)

func TestRawSignedInstructions(t *testing.T) {
	data, err := os.ReadFile("testdata/raw_instructions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Wire                string
		Signature           string
		RawInstructionCount *int     `json:"raw_instruction_count"`
		LoadedWritable      []string `json:"loaded_writable"`
		LoadedReadonly      []string `json:"loaded_readonly"`
	}
	if err = json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, f := range fixtures {
		wire, _ := base64.StdEncoding.DecodeString(f.Wire)
		tx, _, err := DecodeWireTransaction(wire, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		keys := append(append(append([]string{}, tx.Message.AccountKeys...), f.LoadedWritable...), f.LoadedReadonly...)
		events, err := DexEventsFromShredTransactionWireWithLoadedAddresses(wire, f.LoadedWritable, f.LoadedReadonly, "", 1, 2, nil, 3, EventTypeFilterIncludeOnly([]EventType{EventTypeRawInstruction}))
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != len(tx.Message.Instructions) {
			t.Fatal("not all outer instructions returned")
		}
		for i, event := range events {
			e, ok := event.Data.(*RawInstructionEvent)
			if !ok {
				t.Fatal("not raw")
			}
			if e.Metadata.TxIndex != 2 || e.Metadata.GrpcRecvUs != 3 {
				t.Fatal("raw metadata was reset")
			}
			ix := tx.Message.Instructions[i]
			if e.ProgramID != keys[ix.ProgramIDIndex] || !reflect.DeepEqual(e.AccountIndices, ix.Accounts) {
				t.Fatal("instruction identity lost")
			}
			for j, k := range ix.Accounts {
				if e.Accounts[j] == nil || *e.Accounts[j] != keys[k] {
					t.Fatal("account order changed")
				}
			}
			data, _ := base58.Decode(ix.Data)
			for j, b := range data {
				if e.InstructionData[j] != int(b) {
					t.Fatal("data changed")
				}
			}
			encoded, err := json.Marshal(e)
			if err != nil {
				t.Fatal(err)
			}
			var restored RawInstructionEvent
			if err = json.Unmarshal(encoded, &restored); err != nil || !reflect.DeepEqual(e, &restored) {
				t.Fatal("JSON roundtrip")
			}
		}
		defaults, err := DexEventsFromShredTransactionWireWithLoadedAddresses(wire, f.LoadedWritable, f.LoadedReadonly, "", 1, 2, nil, 3, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, event := range defaults {
			if event.Type == EventTypeRawInstruction {
				t.Fatal("raw requires explicit subscription")
			}
			if e, ok := event.Data.(*PumpFunTradeEvent); ok && e.ThirdProgram != nil {
				if e.ParseStatus != TradeParseUnresolved || e.SolAmount != 0 {
					t.Fatal("invented execution")
				}
				raw := e.ThirdProgram
				data, _ := base58.Decode(tx.Message.Instructions[raw.InstructionIndex].Data)
				for j, b := range data {
					if raw.InstructionData[j] != int(b) {
						t.Fatal("changed payload")
					}
				}
			}
		}
	}
}
func TestRawUnknownChangedInstruction(t *testing.T) {
	data, _ := os.ReadFile("testdata/raw_instructions.json")
	var f []struct{ Wire string }
	if err := json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	wire, _ := base64.StdEncoding.DecodeString(f[0].Wire)
	tx, _, err := DecodeWireTransaction(wire, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	ix := tx.Message.Instructions[1]
	ix.ProgramIDIndex = 0
	ix.Accounts = []int{ix.Accounts[2], ix.Accounts[1], ix.Accounts[2], 250}
	ix.Data = base58.Encode([]byte{255, 0, 1})
	tx.Message.Instructions = tx.Message.Instructions[:1]
	tx.Message.Instructions[0] = ix
	events := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, EventTypeFilterIncludeOnly([]EventType{EventTypeRawInstruction}), false, nil)
	if len(events) != 1 {
		t.Fatal("unknown instruction dropped")
	}
	e := events[0].Data.(*RawInstructionEvent)
	if e.Accounts[3] != nil || *e.Accounts[0] != *e.Accounts[2] || !reflect.DeepEqual(e.InstructionData, []int{255, 0, 1}) {
		t.Fatal("raw data changed")
	}
}

func TestRawDexSubscriptionScope(t *testing.T) {
	data, _ := os.ReadFile("testdata/raw_instructions.json")
	var fixtures []struct{ Wire string }
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	wire, _ := base64.StdEncoding.DecodeString(fixtures[0].Wire)
	tx, _, err := DecodeWireTransaction(wire, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	tx.Message.AccountKeys = []string{tx.Message.AccountKeys[0], PUMPFUN_PROGRAM_ID, RAYDIUM_CPMM_PROGRAM_ID}
	template := tx.Message.Instructions[0]
	tx.Message.Instructions = nil
	for _, indices := range [][]int{{1, 1}, {2}, {}, {1, 2, 250}, {250}} {
		ix := template
		ix.ProgramIDIndex = 0
		ix.Accounts = indices
		ix.Data = base58.Encode([]byte{255})
		tx.Message.Instructions = append(tx.Message.Instructions, ix)
	}
	for _, test := range []struct {
		event   EventType
		indices []int
	}{
		{EventTypePumpFunTrade, []int{0, 3}}, {EventTypeRaydiumCpmmSwap, []int{}}, {EventTypePumpFunBuy, []int{}}, {EventTypeRawInstruction, []int{0, 1, 2, 3, 4}},
	} {
		events := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, EventTypeFilterIncludeOnly([]EventType{test.event}), false, nil)
		indices := []int{}
		for _, event := range events {
			switch e := event.Data.(type) {
			case *RawInstructionEvent:
				indices = append(indices, e.InstructionIndex)
			case *PumpFunTradeEvent:
				indices = append(indices, e.ThirdProgram.InstructionIndex)
			default:
				t.Fatal("unexpected event")
			}
		}
		if !reflect.DeepEqual(indices, test.indices) {
			t.Fatalf("%s: got %v want %v", test.event, indices, test.indices)
		}
	}
	events := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, nil, false, nil)
	if len(events) != 2 {
		t.Fatal("wrong candidate count")
	}
	raw := events[1].Data.(*PumpFunTradeEvent).ThirdProgram
	if raw.Accounts[2] != nil {
		t.Fatal("missing ALT must remain null")
	}
	excluded := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, EventTypeFilterExclude([]EventType{EventTypeRawInstruction}), false, nil)
	if len(excluded) != 2 {
		t.Fatal("raw exclusion must not suppress Pump callback")
	}
}

func TestRawDexEnumJSON(t *testing.T) {
	for _, dex := range rawDexPrograms {
		payload, err := json.Marshal(dex.name)
		if err != nil {
			t.Fatal(err)
		}
		var restored RawInstructionDexType
		if err := json.Unmarshal(payload, &restored); err != nil {
			t.Fatal(err)
		}
		if restored != dex.name {
			t.Fatal("enum changed on JSON roundtrip")
		}
		var value string
		if err := json.Unmarshal(payload, &value); err != nil || value != string(dex.name) {
			t.Fatal("JSON format changed")
		}
	}
	for _, payload := range []string{`"typo"`, `null`, `123`} {
		var kind RawInstructionDexType
		if json.Unmarshal([]byte(payload), &kind) == nil {
			t.Fatalf("accepted invalid enum: %s", payload)
		}
	}
}
