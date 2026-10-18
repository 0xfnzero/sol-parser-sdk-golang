// go run ./examples/raw_instruction_filter — offline signed transaction replay.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	_, source, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(source), "../../solparser/testdata/raw_instructions.json"))
	if err != nil {
		panic(err)
	}
	var cases []struct {
		Signature string
		Wire      string
		Slot      uint64
		Writable  []string `json:"loaded_writable"`
		Readonly  []string `json:"loaded_readonly"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		panic(err)
	}
	for _, c := range cases {
		wire, err := base64.StdEncoding.DecodeString(c.Wire)
		if err != nil {
			panic(err)
		}
		for _, kind := range []sdk.EventType{sdk.EventTypePumpFunTrade, sdk.EventTypeRaydiumCpmmSwap, sdk.EventTypeRawInstruction} {
			events, err := sdk.DexEventsFromShredTransactionWireWithLoadedAddresses(wire, c.Writable, c.Readonly, "", c.Slot, 0, nil, 0, sdk.EventTypeFilterIncludeOnly([]sdk.EventType{kind}))
			if err != nil {
				panic(err)
			}
			raw := []*sdk.RawInstructionEvent{}
			for _, event := range events {
				if e, ok := event.Data.(*sdk.RawInstructionEvent); ok {
					raw = append(raw, e)
				}
			}
			data, err := json.Marshal(map[string]any{"signature": c.Signature, "filter": kind, "raw_instructions": raw})
			if err != nil {
				panic(err)
			}
			fmt.Println(string(data))
		}
	}
}
