// Native Go signed-wire throughput/tail samples. No RPC or submission.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/0xfnzero/sol-parser-sdk-golang/examples/internal/latency"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	path := "examples/hotpath_latency/config.json"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	config, e := latency.Load(path)
	if e != nil {
		return e
	}
	raw, e := os.ReadFile("examples/fixtures/pump_shred_routes.json")
	if e != nil {
		return e
	}
	var fixtures []struct {
		Wire     string
		Slot     uint64
		Writable []string `json:"loaded_writable"`
		Readonly []string `json:"loaded_readonly"`
		Expected []json.RawMessage
	}
	if e = json.Unmarshal(raw, &fixtures); e != nil {
		return e
	}
	wires := make([][]byte, len(fixtures))
	for i, f := range fixtures {
		wires[i], e = base64.StdEncoding.DecodeString(f.Wire)
		if e != nil {
			return e
		}
	}
	result, e := latency.Measure("six-signed-wire-route-batch", config, func() error {
		for i, f := range fixtures {
			events, e := sdk.DexEventsFromShredTransactionWireWithLoadedAddresses(wires[i], f.Writable, f.Readonly, "", f.Slot, 888, nil, 999, nil)
			if e != nil {
				return e
			}
			if len(events) != len(f.Expected) {
				return fmt.Errorf("fixture %d events %d expected%d", i, len(events), len(f.Expected))
			}
		}
		return nil
	})
	if e != nil {
		return e
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		Config       latency.Config `json:"config"`
		Transactions int            `json:"transactions_per_op"`
		Result       latency.Result `json:"result"`
		RPCCalls     int            `json:"rpc_calls"`
	}{config, len(fixtures), result, 0})
}
