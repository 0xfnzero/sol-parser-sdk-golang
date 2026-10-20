package main

import (
	"encoding/json"
	"github.com/0xfnzero/sol-parser-sdk-golang/examples/internal/replay"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
)

func main() {
	cases, err := replay.Load("solparser/testdata/cross_dex_stock_routes_20261020.json")
	if err != nil {
		panic(err)
	}
	catalog := sdk.ScreenedStockCatalog()
	for _, c := range cases {
		route, err := sdk.AnalyzeRPCTransactionRoutes(c.RPC, nil)
		if err != nil {
			panic(err)
		}
		// Example boundary conversion only; trading hot paths use typed identities directly.
		wire, err := json.Marshal(route)
		if err != nil {
			panic(err)
		}
		var identity struct {
			Succeeded bool                `json:"succeeded"`
			Legs      []sdk.StockRouteLeg `json:"legs"`
		}
		if err = json.Unmarshal(wire, &identity); err != nil {
			panic(err)
		}
		legs := []sdk.RouteSwapLeg{}
		for _, i := range catalog.ConversionLegIndices(identity.Succeeded, identity.Legs) {
			legs = append(legs, route.Legs[i])
		}
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"name": c.Name, "signature": c.Signature, "replay": true, "route": route, "conversion_legs": legs}); err != nil {
			panic(err)
		}
	}
}
