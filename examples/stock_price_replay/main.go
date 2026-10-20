package main

import (
	"encoding/json"
	"github.com/0xfnzero/sol-parser-sdk-golang/examples/internal/replay"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"os"
)

func main() {
	cases, err := replay.Load("solparser/testdata/stock_price_replay_20261020.json")
	if err != nil {
		panic(err)
	}
	config := sdk.DefaultStockPriceConfig()
	config.Catalog = sdk.ScreenedStockCatalog()
	tracker, err := sdk.NewStockPriceTracker(config)
	if err != nil {
		panic(err)
	}
	for _, c := range cases {
		tracker.Reset()
		tx, err := replay.Decode(c.RPC)
		if err != nil {
			panic(err)
		}
		events, parseErr := sdk.ParseRpcTransaction(tx, c.Signature, nil, 1_000_000)
		if parseErr != nil {
			panic(parseErr)
		}
		prices := []*sdk.StockTokenPriceEvent{}
		for _, event := range events {
			if price := tracker.Process(event, 1_000_000); price != nil {
				prices = append(prices, price)
			}
		}
		if err = json.NewEncoder(os.Stdout).Encode(map[string]any{"name": c.Name, "signature": c.Signature, "replay": true, "prices": prices}); err != nil {
			panic(err)
		}
	}
}
