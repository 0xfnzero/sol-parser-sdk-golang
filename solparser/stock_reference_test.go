package solparser

import (
	"bytes"
	"github.com/mr-tron/base58"
	"testing"
)

func TestStockReferenceOrderAndReset(t *testing.T) {
	mint, pool := base58.Encode(bytes.Repeat([]byte{21}, 32)), base58.Encode(bytes.Repeat([]byte{22}, 32))
	config := DefaultStockPriceConfig()
	config.SolUsdcDex = StockRaydiumCpmm
	config.Catalog = StockTokenCatalog{Tokens: []StockToken{{Mint: mint, Symbol: "TEST", Decimals: 6}}, Pools: []StockPool{{Pool: pool, Mint: mint, QuoteMint: StockUSDC, DexType: StockRaydiumCpmm}}}
	tracker, err := NewStockPriceTracker(config)
	if err != nil {
		t.Fatal(err)
	}
	swap := func(reference bool, index uint64) DexEvent {
		e := &RaydiumCpmmSwapEvent{Metadata: EventMetadata{Slot: 100, TxIndex: index, GrpcRecvUs: 1000}, ParseStatus: TradeParseExecutionConfirmed, PoolID: pool, InputMint: mint, OutputMint: StockUSDC, InputAmount: 1_000_000, OutputAmount: 10_000_000, InputVaultBefore: 1, OutputVaultBefore: 1}
		if reference {
			e.PoolID = DefaultStockSolUsdcPool
			e.InputMint = StockWSOL
			e.InputAmount = 1_000_000_000
			e.OutputAmount = 100_000_000
		}
		return DexEvent{Type: EventTypeRaydiumCpmmSwap, Data: e}
	}
	tracker.Process(swap(true, 1), 1000)
	if p := tracker.Process(swap(false, 1), 1000); p == nil || p.PriceSOL != nil {
		t.Fatal("same transaction reference")
	}
	value := float64(500)
	if err = tracker.UpdateLiquidity(StockPoolLiquidity{Pool: pool, Slot: 100, ObservedUs: 1000, Source: "coherent-test", TvlUSDC: &value}); err != nil {
		t.Fatal(err)
	}
	p := tracker.Process(swap(false, 2), 1000)
	for _, impact := range []uint16{10000, 65535} {
		if tracker.UpdateLiquidity(StockPoolLiquidity{Pool: pool, Slot: 100, ObservedUs: 1000, Source: "test", DepthMaxPriceImpactBps: &impact}) == nil {
			t.Fatal("accepted invalid impact bound")
		}
	}
	if tracker.UpdateLiquidity(StockPoolLiquidity{Pool: pool, Slot: 100, ObservedUs: 1000, Source: " \t\n"}) == nil {
		t.Fatal("accepted blank source")
	}
	if p == nil || p.PriceSOL == nil || *p.PriceSOL != 0.1 || p.Liquidity == nil || *p.Liquidity.TvlUSDC != 500 {
		t.Fatal(p)
	}
	*p.Liquidity.TvlUSDC = 1
	value = 2
	if *tracker.Process(swap(false, 3), 1000).Liquidity.TvlUSDC != 500 {
		t.Fatal("aliased liquidity")
	}
	tracker.Reset()
	p = tracker.Process(swap(false, 4), 1000)
	if p == nil || p.PriceSOL != nil || p.Liquidity != nil {
		t.Fatal("stale reference")
	}
}
