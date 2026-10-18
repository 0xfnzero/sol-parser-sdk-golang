// Native Go driver. The Go VM runs separately to isolate its Yellowstone protobuf registry.
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type execution struct {
	Name     string
	Alt      bool
	Reject   bool
	Wire     string
	Logs     []string
	Writable []string
	Readonly []string
	Input    string
	Output   string
	Amount   uint64
	Minimum  uint64
}

func check(ok bool, s string) {
	if !ok {
		panic(s)
	}
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func runBank(configPath string) int {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Dir(file)
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = filepath.Join(root, "engine")
	if configPath != "" {
		absolute, e := filepath.Abs(configPath)
		must(e)
		cmd.Args = append(cmd.Args, absolute)
	}
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	must(err)
	var cases []execution
	must(json.Unmarshal(output, &cases))
	check(len(cases) == 20, "matrix size")
	data, err := os.ReadFile(filepath.Join(root, "../offline_bank/expected.json"))
	must(err)
	var oracle map[string][]map[string]string
	must(json.Unmarshal(data, &oracle))
	for _, c := range cases {
		raw, e := base64.StdEncoding.DecodeString(c.Wire)
		must(e)
		decoded, _, e := sdk.DecodeWireTransaction(raw, 0, true)
		must(e)
		events, e := sdk.DexEventsFromShredTransactionWireWithLoadedAddresses(raw, c.Writable, c.Readonly, "", 0, 0, nil, 0, nil)
		must(e)
		check(len(events) == 1, "route count")
		intent, ok := events[0].Data.(*sdk.PumpRouteIntentEvent)
		check(ok && len(intent.Hops) == 2 && intent.InstructionIndex == 1, "route intent")
		check(intent.InputAccount == c.Input && intent.OutputAccount == c.Output, "input/output")
		check(*intent.AmountIn == c.Amount && *intent.MinAmountOut == c.Minimum, "intent amount")
		count := 1
		if c.Alt {
			count = 0
		}
		check(len(sdk.DexEventsFromShredTransactionWire(raw, "", 0, 0, nil, 0, nil)) == count, "strict ALT")

		// Use local execution status to ensure failed inner logs never become settled fills.
		rpc := &sdk.RpcTransactionResponse{Transaction: &sdk.RpcTransaction{Signatures: decoded.Signatures, Message: &sdk.RpcMessage{AccountKeys: decoded.Message.AccountKeys, RecentBlockhash: decoded.Message.RecentBlockhash}}, Meta: &sdk.RpcTransactionMeta{LogMessages: c.Logs, LoadedAddresses: &sdk.RpcLoadedAddresses{Writable: c.Writable, Readonly: c.Readonly}}}
		if c.Reject {
			rpc.Meta.Err = map[string]any{"InstructionError": []any{1, map[string]int{"Custom": 6004}}}
		}
		settled, parseErr := sdk.ParseRpcTransaction(rpc, decoded.Signatures[0], nil, 0)
		check(parseErr == nil, "status-aware parse")
		countFills := 0
		for _, event := range settled {
			if event.Type == sdk.EventTypePumpFunTrade || event.Type == sdk.EventTypePumpSwapBuy || event.Type == sdk.EventTypePumpSwapSell {
				countFills++
			}
		}
		wantFills := 2
		if c.Reject {
			wantFills = 0
		}
		check(countFills == wantFills, "failed parent became settled fills")
		var fills []sdk.DexEvent
		for _, line := range c.Logs {
			ev := sdk.ParseLog(line, decoded.Signatures[0], 0, 0, nil, 0, false, "")
			if ev.Type == sdk.EventTypePumpFunTrade || ev.Type == sdk.EventTypePumpSwapBuy || ev.Type == sdk.EventTypePumpSwapSell {
				fills = append(fills, ev)
			}
		}
		check(len(fills) == len(oracle[c.Name]), "fill count")
		for i, event := range fills {
			payload, e := json.Marshal(event.Data)
			must(e)
			var body map[string]any
			decoder := json.NewDecoder(bytes.NewReader(payload))
			decoder.UseNumber()
			must(decoder.Decode(&body))
			for field, want := range oracle[c.Name][i] {
				if field == "type" {
					check(string(event.Type) == want, "fill type")
					continue
				}
				if field == "user" {
					want = intent.User
				}
				check(fmt.Sprint(body[field]) == want, fmt.Sprintf("%s %s got %v want %s", c.Name, field, body[field], want))
			}
		}
	}
	fmt.Println("20 native Go bank scenarios executed twice; SDK route and fill parsing passed, zero RPC.")
	return len(cases)
}
func main() {
	config := ""
	if len(os.Args) > 1 {
		config = os.Args[1]
	}
	runBank(config)
}
