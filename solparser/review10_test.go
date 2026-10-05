package solparser

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func TestReviewClosedLiquidity(t *testing.T) {
	data := make([]byte, 148)
	copy(data, []byte{17, 216, 246, 142, 225, 199, 218, 56})
	a := &AccountData{Owner: ORCA_WHIRLPOOL_PROGRAM_ID, Lamports: 1, Data: data}
	if ParseLiquidityAccount(a, EventMetadata{}).Type == "" {
		t.Fatal("missing live candidate")
	}
	a.Lamports = 0
	if ParseLiquidityAccount(a, EventMetadata{}).Type != "" {
		t.Fatal("closed candidate")
	}
	a.Lamports = 1
	a.Executable = true
	if ParseLiquidityAccount(a, EventMetadata{}).Type != "" {
		t.Fatal("executable candidate")
	}
}

func TestReviewPrefundedATA(t *testing.T) {
	raw, err := os.ReadFile("testdata/review10_prefunded_ata.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Wire     string
		Response json.RawMessage
	}
	json.Unmarshal(raw, &c)
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	r, err := AnalyzeSimulationRoutes(wire, c.Response, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Succeeded || len(r.Legs) != 0 || len(r.Transfers) != 0 {
		t.Fatal("fictitious fills")
	}
	if _, _, err := simulationAccountSetup(zeroPubkey, "allocate", map[string]any{"space": "-1"}); err == nil {
		t.Fatal("invalid allocation accepted")
	}
}

func TestReviewCompiledBounds(t *testing.T) {
	raw, err := os.ReadFile("testdata/review10_compiled_bounds.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name        string
		Transaction json.RawMessage
		Valid       bool
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := AnalyzeRPCTransactionRoutes(c.Transaction, nil)
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v err=%v", c.Valid, err)
			}
		})
	}
}

func TestReviewWireBounds(t *testing.T) {
	raw, err := os.ReadFile("testdata/review10_wire_bounds.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name  string
		Wire  string
		Valid bool
	}
	if err = json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, _ := base64.StdEncoding.DecodeString(c.Wire)
			_, _, err := DecodeWireTransaction(wire, 0, true)
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v err=%v", c.Valid, err)
			}
		})
	}
}
