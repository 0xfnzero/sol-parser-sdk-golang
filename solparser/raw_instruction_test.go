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
		if f.RawInstructionCount != nil {
			parsed, err := DexEventsFromShredTransactionWireWithLoadedAddresses(wire, f.LoadedWritable, f.LoadedReadonly, "", 1, 2, nil, 3, nil)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, e := range parsed {
				if e.Type == EventTypeRawInstruction {
					count++
				}
				if _, ok := e.Data.(*PumpFunTradeEvent); ok {
					t.Fatal("wrapper inferred as trade")
				}
			}
			if count != *f.RawInstructionCount {
				t.Fatal("raw count", count)
			}
		}
		if f.Signature[:4] == "3nMX" || f.Signature[:4] == "2r4u" {
			events = DexEventsFromShredTransactionWire(wire, "", 1, 2, nil, 3, nil)
			want := 3
			if f.Signature[:4] == "3nMX" {
				want = 2
			}
			if len(events) != want {
				t.Fatal("default wrapper count")
			}
			for _, event := range events {
				if event.Type != EventTypeRawInstruction {
					t.Fatal("inferred wrapper")
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
	events := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, nil, false, nil)
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
		{EventTypePumpFunTrade, []int{0, 3}}, {EventTypeRaydiumCpmmSwap, []int{1, 3}}, {EventTypePumpFunBuy, []int{}}, {EventTypeRawInstruction, []int{0, 1, 2, 3, 4}},
	} {
		events := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, EventTypeFilterIncludeOnly([]EventType{test.event}), false, nil)
		indices := []int{}
		for _, event := range events {
			indices = append(indices, event.Data.(*RawInstructionEvent).InstructionIndex)
		}
		if !reflect.DeepEqual(indices, test.indices) {
			t.Fatalf("%s: got %v want %v", test.event, indices, test.indices)
		}
	}
	events := parseShredKeys(tx, tx.Message.AccountKeys, "", 1, 2, nil, 3, nil, false, nil)
	raw := events[3].Data.(*RawInstructionEvent)
	if len(raw.Dexes) != 2 || raw.Dexes[0].DexType != "pumpfun" || raw.Dexes[1].DexType != "raydium_cpmm" || raw.Dexes[0].Pool != nil || raw.Accounts[2] != nil {
		t.Fatal(raw)
	}
	if rawInstructionMatchesFilter(raw.Dexes, EventTypeFilterExclude([]EventType{EventTypeRawInstruction})) {
		t.Fatal("raw exclusion ignored")
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
