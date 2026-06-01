package solparser

import (
	"fmt"
	"math"
	"math/big"
)

type NormalizedTradeSide string

const (
	TradeSideBuy  NormalizedTradeSide = "Buy"
	TradeSideSell NormalizedTradeSide = "Sell"
)

func parseBigIntDecimal(value string) (*big.Int, error) {
	out, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return nil, fmt.Errorf("invalid decimal integer %q", value)
	}
	return out, nil
}

// SqrtPriceX64ToPrice converts a Q64.64 sqrt price into quote-token units per one base token.
func SqrtPriceX64ToPrice(sqrtPriceX64 string, baseDecimals, quoteDecimals int) (float64, error) {
	value, err := parseBigIntDecimal(sqrtPriceX64)
	if err != nil {
		return 0, err
	}
	return SqrtPriceX64ToPriceBig(value, baseDecimals, quoteDecimals), nil
}

// SqrtPriceX64ToPriceBig is the big.Int variant of SqrtPriceX64ToPrice.
func SqrtPriceX64ToPriceBig(sqrtPriceX64 *big.Int, baseDecimals, quoteDecimals int) float64 {
	sqrt := new(big.Float).SetPrec(256).SetInt(sqrtPriceX64)
	q64 := new(big.Float).SetPrec(256).SetFloat64(math.Pow(2, 64))
	sqrt.Quo(sqrt, q64)
	price := new(big.Float).Mul(sqrt, sqrt)
	price.Mul(price, new(big.Float).SetFloat64(math.Pow10(baseDecimals-quoteDecimals)))
	out, _ := price.Float64()
	return out
}

// VaultPriceFromBalances computes quote-token price per one base token from raw vault balances.
func VaultPriceFromBalances(baseRaw, quoteRaw string, baseDecimals, quoteDecimals int) (float64, bool, error) {
	baseInt, err := parseBigIntDecimal(baseRaw)
	if err != nil {
		return 0, false, err
	}
	if baseInt.Sign() == 0 {
		return 0, false, nil
	}
	quoteInt, err := parseBigIntDecimal(quoteRaw)
	if err != nil {
		return 0, false, err
	}
	base := new(big.Float).SetPrec(256).SetInt(baseInt)
	quote := new(big.Float).SetPrec(256).SetInt(quoteInt)
	price := new(big.Float).Quo(quote, base)
	price.Mul(price, new(big.Float).SetFloat64(math.Pow10(baseDecimals-quoteDecimals)))
	out, _ := price.Float64()
	return out, true, nil
}

// NormalizeBuySellFromTokenDelta treats positive watched-token delta as Buy and negative as Sell.
func NormalizeBuySellFromTokenDelta(tokenDelta int64) (NormalizedTradeSide, bool) {
	if tokenDelta > 0 {
		return TradeSideBuy, true
	}
	if tokenDelta < 0 {
		return TradeSideSell, true
	}
	return "", false
}

// NormalizeBuySellFromTokenDeltaString is the decimal-string variant for i128-like deltas.
func NormalizeBuySellFromTokenDeltaString(tokenDelta string) (NormalizedTradeSide, bool, error) {
	value, err := parseBigIntDecimal(tokenDelta)
	if err != nil {
		return "", false, err
	}
	if value.Sign() > 0 {
		return TradeSideBuy, true, nil
	}
	if value.Sign() < 0 {
		return TradeSideSell, true, nil
	}
	return "", false, nil
}

// NormalizeBuySellFromInputMint returns Buy when input is quote, Sell when input is base.
func NormalizeBuySellFromInputMint(inputMint, baseMint, quoteMint string) (NormalizedTradeSide, bool) {
	if baseMint == quoteMint {
		return "", false
	}
	if inputMint == quoteMint {
		return TradeSideBuy, true
	}
	if inputMint == baseMint {
		return TradeSideSell, true
	}
	return "", false
}
