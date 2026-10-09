package solparser

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/mr-tron/base58"
	"net/http"
	"os"
	"strconv"
	"testing"
)

func TestSignedTransactionReplay(t *testing.T) {
	data, err := os.ReadFile("testdata/signed_replay_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name, Wire string
			RPC        struct{ Meta map[string]any }
			Expected   struct {
				Signature        string
				SignatureCount   int `json:"signature_count"`
				InstructionCount int `json:"instruction_count"`
				Version          any
			}
		}
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, err := base64.StdEncoding.DecodeString(c.Wire)
			if err != nil {
				t.Fatal(err)
			}
			tx, n, err := DecodeWireTransaction(wire, 0, true)
			if err != nil {
				t.Fatal(err)
			}
			count := c.Expected.SignatureCount
			if int(wire[0]) != count || count >= 128 || n != len(wire) || len(tx.Signatures) != count || tx.Signatures[0] != c.Expected.Signature || len(tx.Message.Instructions) != c.Expected.InstructionCount {
				t.Fatal("signed fixture structure mismatch")
			}
			message := wire[1+64*count:]
			for i := 0; i < count; i++ {
				key, err := base58.Decode(tx.Message.AccountKeys[i])
				if err != nil {
					t.Fatal(err)
				}
				sig := wire[1+64*i : 1+64*(i+1)]
				if base58.Encode(sig) != tx.Signatures[i] || !ed25519.Verify(key, message, sig) {
					t.Fatal("invalid original signature")
				}
				badSig := append([]byte(nil), sig...)
				badSig[0] ^= 1
				badMsg := append([]byte(nil), message...)
				badMsg[len(badMsg)-1] ^= 1
				if ed25519.Verify(key, message, badSig) || ed25519.Verify(key, badMsg, sig) {
					t.Fatal("tamper accepted")
				}
			}
			for _, bad := range [][]byte{wire[:len(wire)-1], append(append([]byte(nil), wire...), 0)} {
				if _, _, err := DecodeWireTransaction(bad, 0, true); err == nil {
					t.Fatal("malformed wire accepted")
				}
			}
			replay := func(meta map[string]any) *TransactionRoute {
				response, err := json.Marshal(map[string]any{"result": map[string]any{"value": meta}})
				if err != nil {
					t.Fatal(err)
				}
				route, err := AnalyzeSimulationRoutes(wire, response, nil)
				if err != nil {
					t.Fatal(err)
				}
				return route
			}
			// Cached historical metadata is replayed locally; generated legacy metadata is synthetic.
			route := replay(c.RPC.Meta)
			if route.Signature != c.Expected.Signature || !route.Succeeded {
				t.Fatal("replay signature or success mismatch")
			}
			expectedLegs := 1
			if c.Name == "generated_legacy_two_signers" {
				expectedLegs = 0
			}
			if len(route.Legs) != expectedLegs {
				t.Fatal("DEX legs mismatch")
			}
			if expectedLegs > 0 {
				leg := route.Legs[0]
				protocol, amount := "PumpSwap", uint64(10000000)
				if c.Name == "pumpfun" {
					protocol, amount = "PumpFun", 30765521374696
				}
				if leg.Protocol != protocol || leg.SpecifiedAmount != amount || leg.ActualOutputAmount != nil {
					t.Fatal("historical intent/evidence mismatch")
				}
				if c.Name == "pumpfun" {
					if leg.ActualInputAmount != nil {
						t.Fatal("ambiguous input attributed")
					}
				} else if leg.ActualInputAmount == nil || *leg.ActualInputAmount != 10000000 {
					t.Fatal("input evidence missing")
				}
				c.RPC.Meta["err"] = map[string]any{"InstructionError": []any{0, "Custom"}}
				failed := replay(c.RPC.Meta)
				if failed.Succeeded {
					t.Fatal("failed metadata reported successful")
				}
				for _, l := range failed.Legs {
					if l.ActualInputAmount != nil || l.ActualOutputAmount != nil {
						t.Fatal("failed metadata reports actual settlement")
					}
				}
			}
		})
	}
}

func TestSignedMultiALTCPIAndRealFailedReplay(t *testing.T) {
	data, err := os.ReadFile("testdata/signed_dex_replay_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name, Wire string
			RPC        map[string]any
			Expected   struct {
				LookupCount int `json:"lookup_count"`
				Succeeded   bool
				Legs        []struct {
					Protocol, Pool     string
					InputMint          string  `json:"input_mint"`
					OutputMint         string  `json:"output_mint"`
					SpecifiedAmount    string  `json:"specified_amount"`
					ActualOutputAmount *string `json:"actual_output_amount"`
				}
			}
		}
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, err := base64.StdEncoding.DecodeString(c.Wire)
			if err != nil {
				t.Fatal(err)
			}
			tx, consumed, err := DecodeWireTransaction(wire, 0, true)
			if err != nil {
				t.Fatal(err)
			}
			if consumed != len(wire) || len(tx.Message.AddressTableLookups) != c.Expected.LookupCount {
				t.Fatal("multi ALT structure mismatch")
			}
			count := len(tx.Signatures)
			if int(wire[0]) != count || count >= 128 {
				t.Fatal("fixture signature count")
			}
			for i, signature := range tx.Signatures {
				key, err := base58.Decode(tx.Message.AccountKeys[i])
				if err != nil {
					t.Fatal(err)
				}
				sig, err := base58.Decode(signature)
				if err != nil {
					t.Fatal(err)
				}
				if !ed25519.Verify(key, wire[1+64*count:], sig) {
					t.Fatal("invalid original signature")
				}
			}
			response, _ := json.Marshal(map[string]any{"result": map[string]any{"value": c.RPC["meta"]}})
			route, err := AnalyzeSimulationRoutes(wire, response, nil)
			if err != nil {
				t.Fatal(err)
			}
			if route.Signature != tx.Signatures[0] || route.Succeeded != c.Expected.Succeeded || len(route.Legs) != len(c.Expected.Legs) {
				t.Fatal("route status/signature/leg count mismatch")
			}
			for i, expected := range c.Expected.Legs {
				leg := route.Legs[i]
				input, err := strconv.ParseUint(expected.SpecifiedAmount, 10, 64)
				if err != nil {
					t.Fatal(err)
				}
				if leg.Protocol != expected.Protocol || leg.Pool != expected.Pool || leg.InputMint == nil || *leg.InputMint != expected.InputMint || leg.OutputMint == nil || *leg.OutputMint != expected.OutputMint || leg.SpecifiedAmount != input {
					t.Fatal("protocol/pool/mints/intent mismatch")
				}
				if route.Succeeded {
					output, err := strconv.ParseUint(*expected.ActualOutputAmount, 10, 64)
					if err != nil {
						t.Fatal(err)
					}
					if leg.ActualInputAmount == nil || *leg.ActualInputAmount != input || leg.ActualOutputAmount == nil || *leg.ActualOutputAmount != output || leg.Position.StackHeight == nil || *leg.Position.StackHeight != 2 {
						t.Fatal("CPI settlement mismatch")
					}
				} else if leg.ActualInputAmount != nil || leg.ActualOutputAmount != nil {
					t.Fatal("failed transaction attributed settlement")
				}
			}
			if !route.Succeeded && (len(route.Transfers) != 0 || len(route.NativeTokenActions) != 0) {
				t.Fatal("failed transfers exposed")
			}
			// The RPC test adapter is fed the exact decoded, independently verified message.
			c.RPC["transaction"] = map[string]any{"signatures": tx.Signatures, "message": tx.Message}
			raw, err := json.Marshal(c.RPC)
			if err != nil {
				t.Fatal(err)
			}
			events, parseErr := ParseRpcTransaction(pumpUpgradeRPC(t, raw), tx.Signatures[0], nil, 0)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			expectedEvents := 3
			if !route.Succeeded {
				expectedEvents = 0
			}
			if len(events) != expectedEvents {
				t.Fatalf("events=%d want=%d", len(events), expectedEvents)
			}
			for i, event := range events {
				expected := c.Expected.Legs[i]
				if event.GetMetadata().Signature != route.Signature || string(event.Type) != expected.Protocol+"Swap" {
					t.Fatal("event signature/protocol mismatch")
				}
				var pool string
				var input, output uint64
				switch body := event.Data.(type) {
				case *RaydiumCpmmSwapEvent:
					pool, input, output = body.PoolID, body.InputAmount, body.OutputAmount
				case *OrcaWhirlpoolSwapEvent:
					pool, input, output = body.Whirlpool, body.InputAmount, body.OutputAmount
				case *MeteoraDlmmSwapEvent:
					pool, input, output = body.Pool, body.AmountIn, body.AmountOut
				case *RaydiumClmmSwapEvent:
					pool = body.PoolState
					if body.ZeroForOne {
						input, output = body.Amount0, body.Amount1
					} else {
						input, output = body.Amount1, body.Amount0
					}
				default:
					t.Fatalf("unexpected event type %T", body)
				}
				if pool != expected.Pool || strconv.FormatUint(input, 10) != expected.SpecifiedAmount || strconv.FormatUint(output, 10) != *expected.ActualOutputAmount {
					t.Fatal("event pool/settlement mismatch")
				}
			}
		})
	}
}

func TestOfficialSanitizerSignedWireBoundaryContract(t *testing.T) {
	for _, filename := range []string{"signed_wire_boundaries_20261009.json", "signed_alt_load_rejections_20261009.json"} {
		data, err := os.ReadFile("testdata/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		var corpus struct {
			Cases []struct {
				Name, Wire     string
				ValidStructure bool `json:"valid_structure"`
				ValidSignature bool `json:"valid_signature"`
			}
		}
		if err := json.Unmarshal(data, &corpus); err != nil {
			t.Fatal(err)
		}
		for _, c := range corpus.Cases {
			t.Run(c.Name, func(t *testing.T) {
				wire, err := base64.StdEncoding.DecodeString(c.Wire)
				if err != nil {
					t.Fatal(err)
				}
				tx, consumed, err := DecodeWireTransaction(wire, 0, true)
				if !c.ValidStructure {
					if err == nil {
						t.Fatal("official sanitizer rejects accepted wire")
					}
					return
				}
				if err != nil || consumed != len(wire) {
					t.Fatalf("valid structural decode: %v", err)
				}
				valid := true
				for i, signature := range tx.Signatures {
					key, err := base58.Decode(tx.Message.AccountKeys[i])
					if err != nil {
						t.Fatal(err)
					}
					sig, err := base58.Decode(signature)
					if err != nil {
						t.Fatal(err)
					}
					valid = valid && ed25519.Verify(key, wire[1+64*len(tx.Signatures):], sig)
				}
				if valid != c.ValidSignature {
					t.Fatal("independent signature validity mismatch")
				}
				framed := append(append([]byte{9}, wire...), 10)
				if _, n, err := DecodeWireTransaction(framed, 1, false); err != nil || n != len(wire) {
					t.Fatalf("Entry slice contract: n=%d err=%v", n, err)
				}
			})
		}
	}
}

func TestSignedCustomizablePoolPublicParseBoundaries(t *testing.T) {
	data, err := os.ReadFile("testdata/signed_clmm_boundaries_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name, Wire string
			Valid      bool `json:"valid_instruction"`
			RPC        map[string]any
		}
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		wire, err := base64.StdEncoding.DecodeString(c.Wire)
		if err != nil {
			t.Fatal(err)
		}
		tx, n, err := DecodeWireTransaction(wire, 0, true)
		if err != nil || n != len(wire) {
			t.Fatalf("decode %s: %v", c.Name, err)
		}
		key, err := base58.Decode(tx.Message.AccountKeys[0])
		if err != nil || !ed25519.Verify(key, wire[65:], wire[1:65]) {
			t.Fatal("signature")
		}
		c.RPC["transaction"] = map[string]any{"signatures": tx.Signatures, "message": tx.Message}
		raw, err := json.Marshal(c.RPC)
		if err != nil {
			t.Fatal(err)
		}
		events, parseErr := ParseRpcTransaction(pumpUpgradeRPC(t, raw), tx.Signatures[0], nil, 0)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		want := 0
		if c.Valid {
			want = 1
		}
		if len(events) != want {
			t.Fatalf("%s events=%d want=%d", c.Name, len(events), want)
		}
	}
}

// Deny the HTTP transport boundary while replaying signed bank transactions.
type parserDenyNetwork struct{ t *testing.T }

func (g parserDenyNetwork) RoundTrip(*http.Request) (*http.Response, error) {
	g.t.Error("Parser hot path attempted HTTP/RPC")
	return nil, errors.New("parser network denied")
}

func TestCapturedCPMMSignedBankFeeAndFailureReplay(t *testing.T) {
	data, err := os.ReadFile("testdata/signed_cpmm_bank_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name, Wire  string
			RPC         map[string]any
			Succeeded   bool
			GrossInput  uint64 `json:"gross_input"`
			NetOutput   uint64 `json:"net_output"`
			VaultCredit uint64 `json:"vault_credit"`
			VaultDebit  uint64 `json:"vault_debit"`
		}
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = parserDenyNetwork{t}
	defer func() { http.DefaultTransport = original }()
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, err := base64.StdEncoding.DecodeString(c.Wire)
			if err != nil {
				t.Fatal(err)
			}
			tx, n, err := DecodeWireTransaction(wire, 0, true)
			if err != nil || n != len(wire) {
				t.Fatalf("wire: %v", err)
			}
			key, err := base58.Decode(tx.Message.AccountKeys[0])
			if err != nil || !ed25519.Verify(key, wire[65:], wire[1:65]) {
				t.Fatal("signature")
			}
			c.RPC["transaction"] = map[string]any{"signatures": tx.Signatures, "message": tx.Message}
			raw, err := json.Marshal(c.RPC)
			if err != nil {
				t.Fatal(err)
			}
			events, parseErr := ParseRpcTransaction(pumpUpgradeRPC(t, raw), tx.Signatures[0], nil, 0)
			want := 0
			if c.Succeeded {
				want = 1
			}
			if parseErr != nil || len(events) != want {
				t.Fatalf("events=%d err=%v", len(events), parseErr)
			}
			response, err := json.Marshal(map[string]any{"result": map[string]any{"value": c.RPC["meta"]}})
			if err != nil {
				t.Fatal(err)
			}
			route, err := AnalyzeSimulationRoutes(wire, response, nil)
			if err != nil || route.Succeeded != c.Succeeded || len(route.Legs) != 1 {
				t.Fatalf("route: %v", err)
			}
			leg := route.Legs[0]
			if leg.Protocol != "RaydiumCpmm" {
				t.Fatal("wrong protocol")
			}
			if c.Succeeded {
				event, ok := events[0].Data.(*RaydiumCpmmSwapEvent)
				if !ok || event.InputAmount != c.VaultCredit || event.OutputAmount != c.VaultDebit || event.Metadata.Signature != tx.Signatures[0] {
					t.Fatal("vault amounts/signature")
				}
				if leg.ActualInputAmount == nil || *leg.ActualInputAmount != c.GrossInput {
					t.Fatal("trader gross input")
				}
				if c.Name == "fee-output" {
					if leg.ActualOutputAmount != nil {
						t.Fatal("unknown Token-2022 fee must not be guessed")
					}
				} else if leg.ActualOutputAmount == nil || *leg.ActualOutputAmount != c.NetOutput {
					t.Fatal("trader net output")
				}
			} else if leg.ActualInputAmount != nil || leg.ActualOutputAmount != nil || len(route.Transfers) != 0 || len(route.NativeTokenActions) != 0 {
				t.Fatal("failed settlement attributed")
			}
		})
	}
}

func TestSamePoolMultilegBankInvocationSettlement(t *testing.T) {
	original := http.DefaultTransport
	http.DefaultTransport = parserDenyNetwork{t}
	defer func() { http.DefaultTransport = original }()
	for _, filename := range []string{"signed_multileg_bank_20261009.json", "signed_identical_alt_bank_20261009.json"} {
		data, err := os.ReadFile("testdata/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		var corpus struct {
			Cases []struct {
				Name, Wire string
				RPC        map[string]any
				Tables     []struct {
					Key       string
					Addresses []string
				} `json:"lookup_tables"`
				Identity  map[string]string `json:"identity"`
				Succeeded bool
				Inputs    []uint64 `json:"requested_gross_inputs"`
				FeeOutput []bool   `json:"fee_output_legs"`
				Legs      []struct {
					GrossInput  uint64 `json:"gross_input"`
					NetOutput   uint64 `json:"net_output"`
					VaultCredit uint64 `json:"vault_credit"`
					VaultDebit  uint64 `json:"vault_debit"`
				}
			}
		}
		if err := json.Unmarshal(data, &corpus); err != nil {
			t.Fatal(err)
		}
		for _, c := range corpus.Cases {
			t.Run(c.Name, func(t *testing.T) {
				wire, err := base64.StdEncoding.DecodeString(c.Wire)
				if err != nil {
					t.Fatal(err)
				}
				tx, n, err := DecodeWireTransaction(wire, 0, true)
				if err != nil || n != len(wire) {
					t.Fatalf("wire %v", err)
				}
				for i, sig := range tx.Signatures {
					key, _ := base58.Decode(tx.Message.AccountKeys[i])
					signature, _ := base58.Decode(sig)
					if !ed25519.Verify(key, wire[1+64*len(tx.Signatures):], signature) {
						t.Fatal("signature")
					}
				}

				if c.Tables != nil {
					tables := map[string][]string{}
					for _, table := range c.Tables {
						tables[table.Key] = table.Addresses
					}
					for _, side := range []string{"writable", "readonly"} {
						expected := []any{}
						for _, lookup := range tx.Message.AddressTableLookups {
							indexes := lookup.WritableIndexes
							if side == "readonly" {
								indexes = lookup.ReadonlyIndexes
							}
							for _, index := range indexes {
								expected = append(expected, tables[lookup.AccountKey][index])
							}
						}
						want, _ := json.Marshal(expected)
						got, _ := json.Marshal(c.RPC["meta"].(map[string]any)["loadedAddresses"].(map[string]any)[side])
						if string(got) != string(want) {
							t.Fatal("runtime ALT load order")
						}
					}
				}

				// Fault-injected metadata; the original wire/signatures are unchanged.
				if len(c.Tables) > 0 {
					for _, side := range []string{"writable", "readonly"} {
						for _, action := range []string{"short", "extra"} {
							rawMeta, _ := json.Marshal(c.RPC["meta"])
							bad := map[string]any{}
							json.Unmarshal(rawMeta, &bad)
							loaded := bad["loadedAddresses"].(map[string]any)
							addresses := loaded[side].([]any)
							if action == "short" {
								addresses = addresses[:len(addresses)-1]
							} else {
								addresses = append(addresses, addresses[0])
							}
							loaded[side] = addresses
							response, _ := json.Marshal(map[string]any{"result": map[string]any{"value": bad}})
							if _, err := AnalyzeSimulationRoutes(wire, response, nil); err == nil {
								t.Fatal("malformed ALT count accepted")
							}
						}
					}
				}
				c.RPC["transaction"] = map[string]any{"signatures": tx.Signatures, "message": tx.Message}
				raw, _ := json.Marshal(c.RPC)
				events, parseErr := ParseRpcTransaction(pumpUpgradeRPC(t, raw), tx.Signatures[0], nil, 0)
				want := 0
				if c.Succeeded {
					want = 2
				}
				if parseErr != nil || len(events) != want {
					t.Fatalf("events %d err %v", len(events), parseErr)
				}
				response, _ := json.Marshal(map[string]any{"result": map[string]any{"value": c.RPC["meta"]}})
				route, err := AnalyzeSimulationRoutes(wire, response, nil)
				if err != nil || len(route.Legs) != 2 || route.Succeeded != c.Succeeded {
					t.Fatalf("route %v", err)
				}
				if route.Legs[0].Pool != route.Legs[1].Pool || route.Legs[0].Trader != route.Legs[1].Trader {
					t.Fatal("pool/trader")
				}
				for i, leg := range route.Legs {
					if c.Identity != nil {
						got := map[string]string{"pool": leg.Pool, "trader": leg.Trader, "input_account": leg.InputAccount, "output_account": leg.OutputAccount}
						if leg.InputMint != nil {
							got["input_mint"] = *leg.InputMint
						}
						if leg.OutputMint != nil {
							got["output_mint"] = *leg.OutputMint
						}
						for key, value := range c.Identity {
							if got[key] != value {
								t.Fatalf("bank identity %s", key)
							}
						}
					}
					if leg.Position.OuterIndex != i+1 || leg.SpecifiedAmount != c.Inputs[i] {
						t.Fatal("invocation attribution")
					}
					if c.Succeeded {
						e := c.Legs[i]
						if leg.ActualInputAmount == nil || *leg.ActualInputAmount != e.GrossInput {
							t.Fatal("input")
						}
						if c.FeeOutput[i] {
							if leg.ActualOutputAmount != nil {
								t.Fatal("unknown fee guessed")
							}
						} else if leg.ActualOutputAmount == nil || *leg.ActualOutputAmount != e.NetOutput {
							t.Fatal("output")
						}
						event, ok := events[i].Data.(*RaydiumCpmmSwapEvent)
						if !ok || event.InputAmount != e.VaultCredit || event.OutputAmount != e.VaultDebit {
							t.Fatal("event vault amounts")
						}
					} else if leg.ActualInputAmount != nil || leg.ActualOutputAmount != nil {
						t.Fatal("rollback settlement")
					}
				}
			})
		}
	}
}

func TestActualSignedNestedCPIBankCannotInventSettlement(t *testing.T) {
	data, err := os.ReadFile("testdata/signed_nested_cpi_bank_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name, Wire string
			RPC        map[string]any
			Succeeded  bool
			Deltas     map[string]int64 `json:"token_deltas"`
		}
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	original := http.DefaultTransport
	http.DefaultTransport = parserDenyNetwork{t}
	defer func() { http.DefaultTransport = original }()
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, err := base64.StdEncoding.DecodeString(c.Wire)
			if err != nil {
				t.Fatal(err)
			}
			tx, n, err := DecodeWireTransaction(wire, 0, true)
			if err != nil || n != len(wire) {
				t.Fatalf("wire %v", err)
			}
			for i, sig := range tx.Signatures {
				key, _ := base58.Decode(tx.Message.AccountKeys[i])
				signature, _ := base58.Decode(sig)
				if !ed25519.Verify(key, wire[1+64*len(tx.Signatures):], signature) {
					t.Fatal("signature")
				}
			}
			c.RPC["transaction"] = map[string]any{"signatures": tx.Signatures, "message": tx.Message}
			raw, _ := json.Marshal(c.RPC)
			events, parseErr := ParseRpcTransaction(pumpUpgradeRPC(t, raw), tx.Signatures[0], nil, 0)
			want := 0
			if c.Succeeded {
				want = 1
			}
			if parseErr != nil || len(events) != want {
				t.Fatalf("events %d error %v", len(events), parseErr)
			}
			response, _ := json.Marshal(map[string]any{"result": map[string]any{"value": c.RPC["meta"]}})
			route, err := AnalyzeSimulationRoutes(wire, response, nil)
			if err != nil || route.Succeeded != c.Succeeded || len(route.Legs) != 1 {
				t.Fatalf("route %v", err)
			}
			leg := route.Legs[0]
			wantHeight := 3
			if c.Succeeded {
				wantHeight = 2
			}
			if leg.Position.StackHeight == nil || *leg.Position.StackHeight != wantHeight || leg.SpecifiedAmount != 10001 {
				t.Fatal("nested invocation")
			}
			if c.Succeeded {
				if leg.ActualInputAmount == nil || *leg.ActualInputAmount != 10001 || leg.ActualOutputAmount == nil || *leg.ActualOutputAmount != 468207 {
					t.Fatal("actual settlement")
				}
				if c.Deltas[leg.InputAccount] != -10001 || c.Deltas[leg.OutputAccount] != 468207 {
					t.Fatal("bank balance deltas")
				}
				event, ok := events[0].Data.(*RaydiumCpmmSwapEvent)
				if !ok || event.InputAmount != 9800 || event.OutputAmount != 468207 {
					t.Fatal("vault event amounts")
				}
			} else {
				if leg.ActualInputAmount != nil || leg.ActualOutputAmount != nil {
					t.Fatal("rolled-back settlement")
				}
				for _, delta := range c.Deltas {
					if delta != 0 {
						t.Fatal("bank did not roll back")
					}
				}
			}
		})
	}
}

func TestMixedSignedLegacyV0LegacyFramedStreamIsolation(t *testing.T) {
	type replayCase struct {
		Name, Wire string
		RPC        map[string]any
	}
	load := func(file, name string) replayCase {
		data, err := os.ReadFile("testdata/" + file)
		if err != nil {
			t.Fatal(err)
		}
		var corpus struct{ Cases []replayCase }
		if err := json.Unmarshal(data, &corpus); err != nil {
			t.Fatal(err)
		}
		for _, c := range corpus.Cases {
			if c.Name == name {
				return c
			}
		}
		t.Fatal("case missing")
		return replayCase{}
	}
	legacy := load("signed_replay_20261009.json", "generated_legacy_two_signers")
	v0 := load("signed_identical_alt_bank_20261009.json", "identical-intents-two-alt")
	cases := []replayCase{legacy, v0, legacy}
	var stream []byte
	var wires [][]byte
	for _, c := range cases {
		wire, err := base64.StdEncoding.DecodeString(c.Wire)
		if err != nil {
			t.Fatal(err)
		}
		wires = append(wires, wire)
		stream = append(stream, wire...)
	}
	if _, _, err := DecodeWireTransaction(stream, 0, true); err == nil {
		t.Fatal("complete decoder accepted multiple frames")
	}
	offset := 0
	for i, c := range cases {
		tx, n, err := DecodeWireTransaction(stream, offset, false)
		if err != nil || n != len(wires[i]) {
			t.Fatalf("frame %d: %v n=%d", i, err, n)
		}
		single, _, err := DecodeWireTransaction(wires[i], 0, true)
		if err != nil {
			t.Fatal(err)
		}
		a, _ := json.Marshal(tx)
		b, _ := json.Marshal(single)
		if string(a) != string(b) || len(tx.Signatures) != 2 {
			t.Fatal("frame contamination")
		}
		for signer, sig := range tx.Signatures {
			key, _ := base58.Decode(tx.Message.AccountKeys[signer])
			signature, _ := base58.Decode(sig)
			if !ed25519.Verify(key, wires[i][129:], signature) {
				t.Fatal("signature")
			}
		}
		c.RPC["transaction"] = map[string]any{"signatures": tx.Signatures, "message": tx.Message}
		raw, _ := json.Marshal(c.RPC)
		events, parseErr := ParseRpcTransaction(pumpUpgradeRPC(t, raw), tx.Signatures[0], nil, 0)
		want := 0
		if i == 1 {
			want = 2
		}
		if parseErr != nil || len(events) != want {
			t.Fatalf("frame%d events=%d err=%v", i, len(events), parseErr)
		}
		offset += n
	}
	if offset != len(stream) {
		t.Fatal("unconsumed stream")
	}
	damaged := stream[:len(stream)-1]
	if _, n, err := DecodeWireTransaction(damaged, 0, false); err != nil || n != len(wires[0]) {
		t.Fatal("first damaged")
	}
	if _, n, err := DecodeWireTransaction(damaged, len(wires[0]), false); err != nil || n != len(wires[1]) {
		t.Fatal("second damaged")
	}
	if _, _, err := DecodeWireTransaction(damaged, len(wires[0])+len(wires[1]), false); err == nil {
		t.Fatal("truncated last accepted")
	}
}
