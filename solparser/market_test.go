package solparser

import "testing"

func TestMarketHelpers(t *testing.T) {
	price, err := SqrtPriceX64ToPrice("18446744073709551616", 6, 6)
	if err != nil {
		t.Fatalf("unexpected sqrt price error: %v", err)
	}
	if price != 1 {
		t.Fatalf("expected identity price, got %f", price)
	}

	vaultPrice, ok, err := VaultPriceFromBalances("1000000000", "2000000", 9, 6)
	if err != nil || !ok {
		t.Fatalf("unexpected vault price result price=%f ok=%v err=%v", vaultPrice, ok, err)
	}
	if vaultPrice != 2 {
		t.Fatalf("expected vault price 2, got %f", vaultPrice)
	}

	side, ok := NormalizeBuySellFromInputMint("USDC", "SOL", "USDC")
	if !ok || side != TradeSideBuy {
		t.Fatalf("expected buy side, got side=%q ok=%v", side, ok)
	}
}
