package solparser

import (
	"errors"
	"github.com/mr-tron/base58"
	"math"
	"math/big"
	"strings"
)

const StockWSOL = "So11111111111111111111111111111111111111112"
const StockUSDC = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
const DefaultStockSolUsdcPool = "Czfq3xZZDmsdGdUyrNLtRhGc47cXcZtLG4crryfu44zE"

type StockPriceConfig struct {
	Catalog     StockTokenCatalog
	SolUsdcPool string
	SolUsdcDex  StockDexType
	MaxAgeUs    int64
	MaxSlotLag  uint64
}

func DefaultStockPriceConfig() StockPriceConfig {
	return StockPriceConfig{SolUsdcPool: DefaultStockSolUsdcPool, SolUsdcDex: StockOrcaWhirlpool, MaxAgeUs: 30_000_000, MaxSlotLag: 75}
}

type StockPoolLiquidity struct {
	Pool                   string   `json:"pool"`
	Slot                   uint64   `json:"slot"`
	ObservedUs             int64    `json:"observed_us"`
	Source                 string   `json:"source"`
	TvlUSDC                *float64 `json:"tvl_usdc"`
	BuyDepthUSDC           *float64 `json:"buy_depth_usdc"`
	SellDepthUSDC          *float64 `json:"sell_depth_usdc"`
	DepthMaxPriceImpactBps *uint16  `json:"depth_max_price_impact_bps"`
}
type SolUsdcReference struct {
	Pool       string        `json:"pool"`
	UsdcPerSol float64       `json:"usdc_per_sol"`
	Metadata   EventMetadata `json:"metadata"`
}
type StockTokenPriceEvent struct {
	Metadata         EventMetadata       `json:"metadata"`
	Mint             string              `json:"mint"`
	Symbol           string              `json:"symbol"`
	Pool             string              `json:"pool"`
	DexType          StockDexType        `json:"dex_type"`
	QuoteMint        string              `json:"quote_mint"`
	PriceInQuote     float64             `json:"price_in_quote"`
	PriceSOL         *float64            `json:"price_sol"`
	PriceUSDC        *float64            `json:"price_usdc"`
	StockAmountRaw   uint64              `json:"stock_amount_raw"`
	QuoteAmountRaw   uint64              `json:"quote_amount_raw"`
	Basis            string              `json:"basis"`
	SolUsdcReference *SolUsdcReference   `json:"sol_usdc_reference"`
	Liquidity        *StockPoolLiquidity `json:"liquidity"`
}
type stockPosition struct{ slot, index uint64 }

func (a stockPosition) before(b stockPosition) bool {
	return a.slot < b.slot || (a.slot == b.slot && a.index < b.index)
}

type trackedStockPool struct {
	pool  StockPool
	scale float64
}

// StockPriceTracker is single-consumer, network-free state. Reset after stream gaps/fork changes.
type StockPriceTracker struct {
	config    StockPriceConfig
	tokens    map[string]StockToken
	pools     map[string]trackedStockPool
	positions map[string]stockPosition
	liquidity map[string]StockPoolLiquidity
	reference *SolUsdcReference
}

func validStockKey(key string) bool {
	raw, err := base58.Decode(key)
	if err != nil || len(raw) != 32 {
		return false
	}
	for _, b := range raw {
		if b != 0 {
			return true
		}
	}
	return false
}
func NewStockPriceTracker(config StockPriceConfig) (*StockPriceTracker, error) {
	if !validStockKey(config.SolUsdcPool) || stockDexPrograms[config.SolUsdcDex] == "" || config.MaxAgeUs <= 0 {
		return nil, errors.New("invalid stock reference or freshness")
	}
	t := &StockPriceTracker{config: config, tokens: map[string]StockToken{}, pools: map[string]trackedStockPool{}}
	t.config.Catalog = StockTokenCatalog{}
	for _, token := range config.Catalog.Tokens {
		_, duplicate := t.tokens[token.Mint]
		if !validStockKey(token.Mint) || token.Mint == StockWSOL || token.Mint == StockUSDC || token.Decimals > 18 || duplicate {
			return nil, errors.New("invalid or duplicate stock token")
		}
		t.tokens[token.Mint] = token
	}
	for _, pool := range config.Catalog.Pools {
		token, ok := t.tokens[pool.Mint]
		_, duplicate := t.pools[pool.Pool]
		if !ok || !validStockKey(pool.Pool) || pool.Pool == config.SolUsdcPool || duplicate || stockDexPrograms[pool.DexType] == "" || (pool.QuoteMint != StockWSOL && pool.QuoteMint != StockUSDC) {
			return nil, errors.New("invalid or duplicate stock pool")
		}
		decimals := 6
		if pool.QuoteMint == StockWSOL {
			decimals = 9
		}
		// Only identity fields are needed in event processing; no fee JSON is copied per trade.
		pool.CatalogFeeSnapshot = nil
		t.pools[pool.Pool] = trackedStockPool{pool, math.Pow10(int(token.Decimals) - decimals)}
	}
	t.Reset()
	return t, nil
}
func (t *StockPriceTracker) Reset() {
	t.reference = nil
	t.positions = map[string]stockPosition{}
	t.liquidity = map[string]StockPoolLiquidity{}
}
func cloneLiquidity(s StockPoolLiquidity) StockPoolLiquidity {
	cp := func(p *float64) *float64 {
		if p == nil {
			return nil
		}
		v := *p
		return &v
	}
	s.TvlUSDC = cp(s.TvlUSDC)
	s.BuyDepthUSDC = cp(s.BuyDepthUSDC)
	s.SellDepthUSDC = cp(s.SellDepthUSDC)
	if s.DepthMaxPriceImpactBps != nil {
		v := *s.DepthMaxPriceImpactBps
		s.DepthMaxPriceImpactBps = &v
	}
	return s
}
func (t *StockPriceTracker) UpdateLiquidity(s StockPoolLiquidity) error {
	if _, ok := t.pools[s.Pool]; !ok || s.ObservedUs <= 0 || strings.TrimSpace(s.Source) == "" || (s.DepthMaxPriceImpactBps != nil && *s.DepthMaxPriceImpactBps >= 10000) {
		return errors.New("invalid liquidity snapshot")
	}
	for _, v := range []*float64{s.TvlUSDC, s.BuyDepthUSDC, s.SellDepthUSDC} {
		if v != nil && (!finite(*v) || *v < 0) {
			return errors.New("invalid liquidity value")
		}
	}
	if (s.BuyDepthUSDC != nil || s.SellDepthUSDC != nil) && s.DepthMaxPriceImpactBps == nil {
		return errors.New("depth needs a price impact bound")
	}
	if old, ok := t.liquidity[s.Pool]; ok && (s.Slot < old.Slot || s.ObservedUs < old.ObservedUs) {
		return errors.New("out-of-order liquidity")
	}
	t.liquidity[s.Pool] = cloneLiquidity(s)
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (t *StockPriceTracker) fresh(time int64, slot uint64, m EventMetadata, now int64) bool {
	return time > 0 && time <= now && now-time <= t.config.MaxAgeUs && slot <= m.Slot && m.Slot-slot <= t.config.MaxSlotLag
}

type stockSwap struct {
	pool                string
	dex                 StockDexType
	input, output       string
	amountIn, amountOut uint64
}

func positiveStockU128(s string) bool {
	n, ok := new(big.Int).SetString(s, 10)
	return ok && n.Sign() > 0
}
func executedStockSwap(event DexEvent) (stockSwap, bool) {
	_, status, _ := tradeSource(event)
	if status == nil || *status != TradeParseExecutionConfirmed {
		return stockSwap{}, false
	}
	switch e := event.Data.(type) {
	case *RaydiumCpmmSwapEvent:
		return stockSwap{e.PoolID, StockRaydiumCpmm, e.InputMint, e.OutputMint, e.InputAmount, e.OutputAmount}, e.InputVaultBefore > 0 && e.OutputVaultBefore > 0
	case *RaydiumClmmSwapEvent:
		a, b := e.Amount0, e.Amount1
		if !e.ZeroForOne {
			a, b = b, a
		}
		return stockSwap{e.PoolState, StockRaydiumClmm, e.InputMint, e.OutputMint, a, b}, positiveStockU128(e.SqrtPriceX64)
	case *OrcaWhirlpoolSwapEvent:
		a, b := e.TokenMintA, e.TokenMintB
		if !e.AToB {
			a, b = b, a
		}
		return stockSwap{e.Whirlpool, StockOrcaWhirlpool, a, b, e.InputAmount, e.OutputAmount}, positiveStockU128(e.PostSqrtPrice)
	case *MeteoraDlmmSwapEvent:
		a, b := e.TokenXMint, e.TokenYMint
		if !e.SwapForY {
			a, b = b, a
		}
		return stockSwap{e.Pool, StockMeteoraDlmm, a, b, e.AmountIn, e.AmountOut}, e.EventVersion == 1 || e.EventVersion == 2
	}
	return stockSwap{}, false
}
func (t *StockPriceTracker) Process(event DexEvent, nowUs int64) *StockTokenPriceEvent {
	if len(t.pools) == 0 {
		return nil
	}
	swap, ok := executedStockSwap(event)
	if !ok {
		return nil
	}
	m := event.GetMetadata()
	if swap.amountIn == 0 || swap.amountOut == 0 || !t.fresh(m.GrpcRecvUs, m.Slot, m, nowUs) {
		return nil
	}
	pos := stockPosition{m.Slot, m.TxIndex}
	if old, ok := t.positions[swap.pool]; ok && pos.before(old) {
		return nil
	}
	if swap.pool == t.config.SolUsdcPool && swap.dex == t.config.SolUsdcDex {
		var sol, usdc uint64
		if swap.input == StockWSOL && swap.output == StockUSDC {
			sol, usdc = swap.amountIn, swap.amountOut
		} else if swap.input == StockUSDC && swap.output == StockWSOL {
			sol, usdc = swap.amountOut, swap.amountIn
		} else {
			return nil
		}
		t.reference = &SolUsdcReference{swap.pool, float64(usdc) / float64(sol) * 1000, m}
		t.positions[swap.pool] = pos
		return nil
	}
	tracked, ok := t.pools[swap.pool]
	if !ok || tracked.pool.DexType != swap.dex {
		return nil
	}
	pool := tracked.pool
	var stock, quote uint64
	if swap.input == pool.Mint && swap.output == pool.QuoteMint {
		stock, quote = swap.amountIn, swap.amountOut
	} else if swap.output == pool.Mint && swap.input == pool.QuoteMint {
		stock, quote = swap.amountOut, swap.amountIn
	} else {
		return nil
	}
	price := float64(quote) / float64(stock) * tracked.scale
	if !finite(price) || price <= 0 {
		return nil
	}
	out := &StockTokenPriceEvent{Metadata: m, Mint: pool.Mint, Symbol: t.tokens[pool.Mint].Symbol, Pool: pool.Pool, DexType: pool.DexType, QuoteMint: pool.QuoteMint, PriceInQuote: price, StockAmountRaw: stock, QuoteAmountRaw: quote, Basis: "ReportedSwapAmounts"}
	if r := t.reference; r != nil && (stockPosition{r.Metadata.Slot, r.Metadata.TxIndex}).before(pos) && t.fresh(r.Metadata.GrpcRecvUs, r.Metadata.Slot, m, nowUs) {
		cp := *r
		out.SolUsdcReference = &cp
	}
	if pool.QuoteMint == StockWSOL {
		out.PriceSOL = &price
		if r := out.SolUsdcReference; r != nil {
			v := price * r.UsdcPerSol
			out.PriceUSDC = &v
		}
	} else {
		out.PriceUSDC = &price
		if r := out.SolUsdcReference; r != nil {
			v := price / r.UsdcPerSol
			out.PriceSOL = &v
		}
	}
	if s, ok := t.liquidity[swap.pool]; ok && t.fresh(s.ObservedUs, s.Slot, m, nowUs) {
		cp := cloneLiquidity(s)
		out.Liquidity = &cp
	}
	t.positions[swap.pool] = pos
	return out
}
