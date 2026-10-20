package main

import (
	"encoding/json"
	"github.com/0xfnzero/sol-parser-sdk-golang/examples/internal/replay"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"math"
	"os"
	"testing"
)

func TestCapturedStockPrices(t *testing.T) {
	raw, err := os.ReadFile("../../solparser/testdata/stock_price_replay_20261020.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct{ Cases []replay.Case }
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	expected := []float64{343.6402644024645, 175.28120314435913, 904.3107197730765}
	if len(doc.Cases) != len(expected) {
		t.Fatal("capture count")
	}
	for i, c := range doc.Cases {
		data := c.RPC
		if data == nil {
			data = c.WireRPC
		}
		if data == nil {
			data = c.Raw
		}
		tx, err := replay.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		events, parseErr := sdk.ParseRpcTransaction(tx, c.Signature, nil, 1_000_000)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		config := sdk.DefaultStockPriceConfig()
		config.Catalog = sdk.ScreenedStockCatalog()
		tracker, err := sdk.NewStockPriceTracker(config)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, e := range events {
			if p := tracker.Process(e, 1_000_000); p != nil {
				count++
				if p.PriceUSDC == nil || math.Abs(*p.PriceUSDC-expected[i]) > 1e-8 || p.Liquidity != nil {
					t.Fatal(p)
				}
			}
		}
		if count != 1 {
			t.Fatal(c.Name, count)
		}
		tracker.Reset()
		for _, e := range events {
			if tracker.Process(e, 100_000_000) != nil {
				t.Fatal("stale price")
			}
		}
	}
}
