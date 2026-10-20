package solparser

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"testing"
)

func TestWrappedPumpfunRealInputs(t *testing.T) {
	data, err := os.ReadFile("testdata/wrapped_pumpfun_inputs.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Rules  []WrappedPumpfunRule
		Inputs []struct {
			Raw    RawInstructionEvent
			Actual []struct {
				Data     []byte
				Accounts []string
			} `json:"actual_cpis"`
		}
	}
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, input := range f.Inputs {
		for _, rule := range f.Rules {
			event := ParseWrappedPumpfunInstruction(&input.Raw, &rule, nil)
			if event.Type == "" {
				continue
			}
			if len(input.Actual) != 1 {
				t.Fatal("ambiguous CPI")
			}
			actual := input.Actual[0]
			trade := event.Data.(*PumpFunTradeEvent)
			if trade.SpendableQuoteIn != binary.LittleEndian.Uint64(actual.Data[8:16]) || trade.MinTokensOut != binary.LittleEndian.Uint64(actual.Data[16:24]) || trade.Mint != actual.Accounts[1] || trade.User != actual.Accounts[13] {
				t.Fatal("CPI mismatch")
			}
			if trade.ThirdProgram == nil || trade.ThirdProgram.ProgramID != input.Raw.ProgramID || trade.SolAmount != 0 || trade.QuoteAmount != 0 || trade.TokenAmount != 0 {
				t.Fatal("wrong provenance or fabricated execution")
			}
			input.Raw.Accounts[0] = nil
			if ParseWrappedPumpfunInstruction(&input.Raw, &rule, nil).Type != "" {
				t.Fatal("missing account accepted")
			}
			count++
			break
		}
	}
	if count != 5 {
		t.Fatalf("restored %d, expected 5", count)
	}
}
