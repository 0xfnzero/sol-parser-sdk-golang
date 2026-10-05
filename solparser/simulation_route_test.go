package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mr-tron/base58"
)

type simulationCase struct {
	Name     string
	Wire     string
	Response json.RawMessage
	Expected json.RawMessage
}

func simulationCases(t *testing.T) []simulationCase {
	return simulationCasesFile(t, "cached_tip_routes_20261002.json")
}

func simulationCasesFile(t *testing.T, file string) []simulationCase {
	t.Helper()
	data, err := os.ReadFile("testdata/" + file)
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct{ Cases []simulationCase }
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	return corpus.Cases
}

func TestSimulationMainnetEvidence(t *testing.T) {
	cases := append(simulationCases(t), simulationCasesFile(t, "simulation_routes_live_20261002.json")...)
	cases = append(cases, simulationCasesFile(t, "simulation_ata_20261002.json")...)
	cases = append(cases, simulationCasesFile(t, "pumpswap_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "damm_v2_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "cached_damm_v2_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "pumpfun_current_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "pumpfun_settlement_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "pumpfun_funded_sell_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "pumpfun_usdc_multihop_mainnet_simulations_20261004.json")...)
	cases = append(cases, simulationCasesFile(t, "pumpfun_concentrated_multihop_mainnet_simulations_20261005.json")...)
	cases = append(cases, simulationCasesFile(t, "cpmm_lp_simulations_20261006.json")...)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			wire, err := base64.StdEncoding.DecodeString(c.Wire)
			if err != nil {
				t.Fatal(err)
			}
			route, err := AnalyzeSimulationRoutes(wire, c.Response, nil)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := json.Marshal(route)
			if err != nil {
				t.Fatal(err)
			}
			var got, want any
			json.Unmarshal(actual, &got)
			json.Unmarshal(c.Expected, &want)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("evidence mismatch\nwant %s\ngot %s", c.Expected, actual)
			}
		})
	}
}

func TestFailedPumpFunV2Intent(t *testing.T) {
	c := simulationCasesFile(t, "pumpfun_current_mainnet_simulations_20261004.json")[0]
	original := []byte{194, 171, 28, 70, 104, 77, 91, 47}
	for i, disc := range [][]byte{original, {184, 23, 238, 97, 103, 197, 211, 61}, {93, 246, 130, 60, 231, 233, 64, 178}} {
		wire, e := base64.StdEncoding.DecodeString(c.Wire)
		if e != nil {
			t.Fatal(e)
		}
		if bytes.Count(wire, original) != 1 {
			t.Fatal("missing unique discriminator")
		}
		wire = bytes.Replace(wire, original, disc, 1)
		var response map[string]any
		if e = json.Unmarshal(c.Response, &response); e != nil {
			t.Fatal(e)
		}
		response["result"].(map[string]any)["value"].(map[string]any)["err"] = map[string]any{"InstructionError": []any{0, "Custom"}}
		raw, _ := json.Marshal(response)
		r, e := AnalyzeSimulationRoutes(wire, raw, nil)
		if e != nil {
			t.Fatal(e)
		}
		if r.Succeeded || len(r.Legs) != 1 {
			t.Fatal("failed intent lost")
		}
		leg := r.Legs[0]
		if leg.AmountSpecifiedIsInput != (i != 1) || leg.ActualInputAmount != nil || leg.ActualOutputAmount != nil {
			t.Fatal("incorrect failed amounts")
		}
	}
}

func TestSimulationATAInitializationMint(t *testing.T) {
	c := simulationCasesFile(t, "simulation_ata_20261002.json")[0]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	var root map[string]any
	json.Unmarshal(c.Response, &root)
	value := root["result"].(map[string]any)["value"].(map[string]any)
	group := value["innerInstructions"].([]any)[0].(map[string]any)
	ixs := group["instructions"].([]any)
	ix := ixs[3].(map[string]any)
	info := ix["parsed"].(map[string]any)["info"].(map[string]any)
	value["preTokenBalances"], value["postTokenBalances"] = []any{}, []any{}
	group["instructions"] = append(ixs, map[string]any{"programId": ix["programId"], "stackHeight": 2, "parsed": map[string]any{"type": "transfer", "info": map[string]any{"source": info["account"], "destination": info["owner"], "authority": info["owner"], "amount": "1"}}})
	raw, _ := json.Marshal(root)
	route, err := AnalyzeSimulationRoutes(wire, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(route.Transfers) != 1 || route.Transfers[0].Mint == nil || *route.Transfers[0].Mint != info["mint"] || *route.Transfers[0].Position.InnerIndex != 4 {
		t.Fatal("lost initialization mint or CPI position")
	}
}

func TestSimulationRejectsInvalidATASetup(t *testing.T) {
	c := simulationCasesFile(t, "simulation_ata_20261002.json")[0]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	for _, kind := range []string{"extension", "owner", "lamports"} {
		t.Run(kind, func(t *testing.T) {
			var root map[string]any
			json.Unmarshal(c.Response, &root)
			ixs := root["result"].(map[string]any)["value"].(map[string]any)["innerInstructions"].([]any)[0].(map[string]any)["instructions"].([]any)
			info := func(i int) map[string]any {
				return ixs[i].(map[string]any)["parsed"].(map[string]any)["info"].(map[string]any)
			}
			switch kind {
			case "extension":
				info(0)["extensionTypes"] = []any{"unknown"}
			case "owner":
				info(3)["owner"] = "1"
			case "lamports":
				info(1)["lamports"] = -1
			}
			raw, _ := json.Marshal(root)
			if _, err := AnalyzeSimulationRoutes(wire, raw, nil); err == nil {
				t.Fatal("invalid setup accepted")
			}
		})
	}
}

func TestCompiledRouteRequiresExecutionStatus(t *testing.T) {
	c := simulationCases(t)[0]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	tx, _, _ := DecodeWireTransaction(wire, 0, true)
	raw, _ := json.Marshal(map[string]any{"transaction": tx, "meta": map[string]any{"innerInstructions": []any{}}})
	if _, err := AnalyzeRPCTransactionRoutes(raw, nil); err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatal("absent execution status accepted")
	}
}

func TestSimulationRejectsInvalidMetadata(t *testing.T) {
	c := simulationCases(t)[0]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	cases := map[string]func(map[string]any){
		"missing err":        func(v map[string]any) { delete(v, "err") },
		"missing inner":      func(v map[string]any) { delete(v, "innerInstructions") },
		"invalid groups":     func(v map[string]any) { v["innerInstructions"] = "bad" },
		"duplicate group":    func(v map[string]any) { a := v["innerInstructions"].([]any); v["innerInstructions"] = append(a, a[0]) },
		"string index":       func(v map[string]any) { v["innerInstructions"].([]any)[0].(map[string]any)["index"] = "5" },
		"group outside wire": func(v map[string]any) { v["innerInstructions"].([]any)[0].(map[string]any)["index"] = 9999 },
		"bad stack":          func(v map[string]any) { simulationFirstIx(v)["stackHeight"] = 1 },
		"unknown parsed":     func(v map[string]any) { simulationFirstIx(v)["parsed"].(map[string]any)["type"] = "unsupported" },
		"unknown account": func(v map[string]any) {
			simulationFirstIx(v)["parsed"].(map[string]any)["info"].(map[string]any)["source"] = "unknown"
		},
		"u64 overflow": func(v map[string]any) {
			simulationFirstIx(v)["parsed"].(map[string]any)["info"].(map[string]any)["amount"] = "18446744073709551616"
		},
		"negative amount": func(v map[string]any) {
			simulationFirstIx(v)["parsed"].(map[string]any)["info"].(map[string]any)["amount"] = "-1"
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			var root map[string]any
			json.Unmarshal(c.Response, &root)
			mutate(root["result"].(map[string]any)["value"].(map[string]any))
			raw, _ := json.Marshal(root)
			if _, err := AnalyzeSimulationRoutes(wire, raw, nil); err == nil {
				t.Fatal("accepted invalid metadata")
			}
		})
	}
	if _, err := AnalyzeSimulationRoutes(wire, append(c.Response, []byte(" {}")...), nil); err == nil {
		t.Fatal("accepted trailing response")
	}
}

func simulationFirstIx(v map[string]any) map[string]any {
	return v["innerInstructions"].([]any)[0].(map[string]any)["instructions"].([]any)[0].(map[string]any)
}

func TestSimulationFailedIntents(t *testing.T) {
	c := simulationCases(t)[0]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	var root map[string]any
	json.Unmarshal(c.Response, &root)
	root["result"].(map[string]any)["value"].(map[string]any)["err"] = map[string]any{"InstructionError": []any{5, "Custom"}}
	raw, _ := json.Marshal(root)
	route, err := AnalyzeSimulationRoutes(wire, raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if route.Succeeded || len(route.Legs) == 0 {
		t.Fatal("failed intent lost")
	}
	for _, leg := range route.Legs {
		if leg.ActualInputAmount != nil || leg.ActualOutputAmount != nil {
			t.Fatal("failed fill reported")
		}
	}
}

func TestSimulationRawAndCompiledCPI(t *testing.T) {
	c := simulationCases(t)[0]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	tx, _, _ := DecodeWireTransaction(wire, 0, true)
	for _, compiled := range []bool{false, true} {
		var root map[string]any
		json.Unmarshal(c.Response, &root)
		ix := simulationFirstIx(root["result"].(map[string]any)["value"].(map[string]any))
		info := ix["parsed"].(map[string]any)["info"].(map[string]any)
		amount, err := simulationUnsigned(info["amount"], 64)
		if err != nil {
			t.Fatal(err)
		}
		data := make([]byte, 9)
		data[0] = 3
		binary.LittleEndian.PutUint64(data[1:], amount)
		ix["data"] = base58.Encode(data)
		accounts := []any{info["source"], info["destination"], info["authority"]}
		delete(ix, "parsed")
		if compiled {
			for i, key := range tx.Message.AccountKeys {
				if key == ix["programId"] {
					ix["programIdIndex"] = i
				}
			}
			delete(ix, "programId")
			for j, a := range accounts {
				for i, key := range tx.Message.AccountKeys {
					if key == a {
						accounts[j] = i
						break
					}
				}
			}
		}
		ix["accounts"] = accounts
		raw, _ := json.Marshal(root)
		route, err := AnalyzeSimulationRoutes(wire, raw, nil)
		if err != nil {
			t.Fatal(err)
		}
		actual, _ := json.Marshal(route)
		var got, want any
		json.Unmarshal(actual, &got)
		json.Unmarshal(c.Expected, &want)
		if !reflect.DeepEqual(got, want) {
			t.Fatal("CPI changed route")
		}
	}
}

func TestSimulationExplicitFee(t *testing.T) {
	c := simulationCases(t)[2]
	wire, _ := base64.StdEncoding.DecodeString(c.Wire)
	original, err := AnalyzeSimulationRoutes(wire, c.Response, nil)
	if err != nil {
		t.Fatal(err)
	}
	if original.Legs[1].ActualOutputAmount != nil {
		t.Fatal("invented unknown fee")
	}
	var root map[string]any
	json.Unmarshal(c.Response, &root)
	groups := root["result"].(map[string]any)["value"].(map[string]any)["innerInstructions"].([]any)
	for _, g := range groups {
		for _, raw := range g.(map[string]any)["instructions"].([]any) {
			ix := raw.(map[string]any)
			p, _ := ix["parsed"].(map[string]any)
			info, _ := p["info"].(map[string]any)
			if ix["programId"] != routeToken2022 || info["destination"] != original.Legs[1].OutputAccount {
				continue
			}
			amount, _ := simulationUnsigned(info["tokenAmount"].(map[string]any)["amount"], 64)
			p["type"] = "transferCheckedWithFee"
			info["feeAmount"] = "7"
			encoded, _ := json.Marshal(root)
			route, err := AnalyzeSimulationRoutes(wire, encoded, nil)
			if err != nil {
				t.Fatal(err)
			}
			if route.Legs[1].ActualOutputAmount == nil || *route.Legs[1].ActualOutputAmount != amount-7 {
				t.Fatal("incorrect explicit fee")
			}
			delete(info, "feeAmount")
			encoded, _ = json.Marshal(root)
			if _, err := AnalyzeSimulationRoutes(wire, encoded, nil); err == nil {
				t.Fatal("accepted missing fee")
			}
			return
		}
	}
	t.Fatal("fee evidence not found")
}

func TestSimulationALTRequiresLoadedAddresses(t *testing.T) {
	raw, err := os.ReadFile("testdata/stonkfun_routes_0_7_7.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct{ Transaction json.RawMessage }
	}
	json.Unmarshal(raw, &corpus)
	for _, c := range corpus.Cases {
		var root map[string]any
		json.Unmarshal(c.Transaction, &root)
		if result, ok := root["result"].(map[string]any); ok {
			root = result
		}
		encoded := root["transaction"].([]any)[0].(string)
		wire, _ := base64.StdEncoding.DecodeString(encoded)
		tx, _, err := DecodeWireTransaction(wire, 0, true)
		if err != nil {
			t.Fatal(err)
		}
		if len(tx.Message.AddressTableLookups) == 0 {
			continue
		}
		if _, err := AnalyzeSimulationRoutes(wire, simulationCases(t)[0].Response, nil); err == nil || !strings.Contains(err.Error(), "ALT") {
			t.Fatal("ALT accepted without addresses")
		}
		return
	}
	t.Fatal("ALT evidence missing")
}

func TestDammSwap2ModeIntent(t *testing.T) {
	c := simulationCasesFile(t, "cached_damm_v2_mainnet_simulations_20261004.json")[0]
	original := make([]byte, 25)
	copy(original, []byte{65, 75, 63, 76, 235, 91, 91, 136})
	binary.LittleEndian.PutUint64(original[8:16], 10000)
	binary.LittleEndian.PutUint64(original[16:24], 1)
	for mode := byte(0); mode <= 3; mode++ {
		wire, e := base64.StdEncoding.DecodeString(c.Wire)
		if e != nil {
			t.Fatal(e)
		}
		at := bytes.Index(wire, original)
		if at < 0 {
			t.Fatal("swap2 not found")
		}
		wire[at+24] = mode
		r, e := AnalyzeSimulationRoutes(wire, c.Response, nil)
		if e != nil {
			t.Fatal(e)
		}
		if mode == 3 {
			if len(r.Legs) != 0 {
				t.Fatal("unknown mode accepted")
			}
		} else {
			if len(r.Legs) != 1 || r.Legs[0].AmountSpecifiedIsInput != (mode != 2) || r.Legs[0].SpecifiedAmount != 10000 || r.Legs[0].OtherAmountThreshold != 1 {
				t.Fatal("swap2 intent mismatch")
			}
		}
	}
}

func TestSupplySimulationCPIEncoding(t *testing.T) {
	for _, program := range []string{routeToken, routeToken2022} {
		for _, row := range []struct {
			kind             string
			opcode           byte
			minting, checked bool
		}{
			{"mintTo", 7, true, false}, {"burn", 8, false, false}, {"mintToChecked", 14, true, true}, {"burnChecked", 15, false, true},
		} {
			t.Run(program+row.kind, func(t *testing.T) {
				info := map[string]any{"mint": "mint", "account": "account", "signers": []any{"signer"}}
				authority := "multisigAuthority"
				if row.minting {
					authority = "multisigMintAuthority"
				}
				info[authority] = "authority"
				amount := map[string]any{"amount": "18446744073709551615", "decimals": json.Number("255")}
				if row.checked {
					info["tokenAmount"] = amount
				} else {
					info["amount"] = amount["amount"]
				}
				accounts, data, err := simulationAccountSetup(program, row.kind, info)
				if err != nil {
					t.Fatal(err)
				}
				wantAccounts := []any{"account", "mint", "authority", "signer"}
				if row.minting {
					wantAccounts[0], wantAccounts[1] = wantAccounts[1], wantAccounts[0]
				}
				want := []byte{row.opcode, 255, 255, 255, 255, 255, 255, 255, 255}
				if row.checked {
					want = append(want, 255)
				}
				if !reflect.DeepEqual(accounts, wantAccounts) || !bytes.Equal(data, want) {
					t.Fatalf("wrong supply CPI: %v %x", accounts, data)
				}
				for _, invalid := range []any{"18446744073709551616", "-1", true, json.Number("1.5")} {
					if row.checked {
						amount["amount"] = invalid
					} else {
						info["amount"] = invalid
					}
					if _, _, err := simulationAccountSetup(program, row.kind, info); err == nil {
						t.Fatal("accepted invalid supply amount")
					}
				}
			})
		}
	}
}
