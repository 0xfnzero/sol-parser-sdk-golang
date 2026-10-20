// Run from the repository root: go run ./examples/wrapped_pumpfun_replay
// Offline input replay; no RPC. Rules are explicitly loaded at startup.
package main

import (
	"encoding/json"
	"fmt"
	parser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
)

func main() {
	data, err := os.ReadFile("solparser/testdata/wrapped_pumpfun_inputs.json")
	if err != nil {
		panic(err)
	}
	var f struct {
		Rules  []parser.WrappedPumpfunRule
		Inputs []struct{ Raw parser.RawInstructionEvent }
	}
	if err = json.Unmarshal(data, &f); err != nil {
		panic(err)
	}
	for _, input := range f.Inputs {
		event := parser.DexEvent{Type: parser.EventTypeRawInstruction, Data: &input.Raw}
		for _, rule := range f.Rules {
			if parsed := parser.ParseWrappedPumpfunInstruction(&input.Raw, &rule, nil); parsed.Type != "" {
				event = parsed
				break
			}
		}
		out, err := json.Marshal(event)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}
