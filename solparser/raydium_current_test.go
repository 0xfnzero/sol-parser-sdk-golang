package solparser

import (
	"encoding/binary"
	"testing"
)

const (
	currentCpmmMainnetSignature  = "4v27ccyrAgpCdCHLvjvn8smFn4Fb4HGcRVTSt952eNcF5jg5niA5bKRLPoGrzxXZdZULEZujgA5TXdESNbwmFYE8"
	currentAmmV4MainnetSignature = "2iHYs4AHC5nutcbBxpA5aptBYTGaDUYBgamohfetDnAPiPBW5NkguxgnjVF5886Jy8MZ19UXdeZyPKq9C5wqAki4"
)

func TestCurrentCpmmSwapEventMainnetFixture(t *testing.T) {
	// Mainnet slot 438881024, block time 2026-08-13T03:13:23+08:00.
	log := "Program data: QMbN6CYIceIWNdNqQlI8Wx1EIXxFK7oiJvUFtAHi+7rKiF7En2cJzzm+3/AAAAAAhg0mMwQAAACn/AwAAAAAAHrPOQAAAAAAAAAAAAAAAAAAAAAAAAAAAAEGm4hX/quBhPtof2NGGMA12sQ53BrrO1WYoPAAAAAAAbw1n9V4Uh7ztRZFBlw3IE5nZpsKp/d96y1JFGAScORfUAgAAAAAAAAAAAAAAAAAAAE="
	ev := ParseLogOptimizedWithProgramID(
		log, currentCpmmMainnetSignature, 438881024, 0, nil, 1, nil, false, "", RAYDIUM_CPMM_PROGRAM_ID,
	)
	if ev.Type != EventTypeRaydiumCpmmSwap {
		t.Fatalf("expected CPMM swap, got %q", ev.Type)
	}
	swap := ev.Data.(*RaydiumCpmmSwapEvent)
	if swap.PoolID != "2VhaFEYL1exY86u8tTisfRjNwtCkX8bNbupHvSKzJuQJ" ||
		swap.InputAmount != 851111 || swap.OutputAmount != 3788666 || !swap.BaseInput {
		t.Fatalf("unexpected current CPMM event: %+v", swap)
	}
}

func TestCurrentAmmV4RayLogMainnetFixture(t *testing.T) {
	// Mainnet slot 438881026, block time 2026-08-13T03:13:24+08:00.
	log := "Program log: ray_log: A2kEi30yGgAAAAAAAAAAAAACAAAAAAAAAP7r/DBUGgAA0qbjFeLumzn51XOmVTgAAPOsjRkAAAAA"
	ev := ParseLogOptimizedWithProgramID(
		log, currentAmmV4MainnetSignature, 438881026, 0, nil, 1, nil, false, "", RAYDIUM_AMM_V4_PROGRAM_ID,
	)
	if ev.Type != EventTypeRaydiumAmmV4Swap {
		t.Fatalf("expected AMM V4 swap, got %q", ev.Type)
	}
	swap := ev.Data.(*RaydiumAmmV4SwapEvent)
	if swap.AmountIn != 28804156949609 || swap.MinimumAmountOut != 0 || swap.AmountOut != 428715251 {
		t.Fatalf("unexpected current AMM V4 ray_log: %+v", swap)
	}
}

func TestCurrentAmmV4SwapSupports17AccountLayout(t *testing.T) {
	data := make([]byte, 17)
	data[0] = 9
	binary.LittleEndian.PutUint64(data[1:], 28804156949609)
	binary.LittleEndian.PutUint64(data[9:], 0)
	accounts := make([]string, 17)
	for i := range accounts {
		accounts[i] = string(rune('A' + i))
	}
	ev := ParseRaydiumAmmV4Instruction(data, accounts, "sig", 1, 0, nil, 1)
	swap := ev.Data.(*RaydiumAmmV4SwapEvent)
	if swap.PoolCoinTokenAccount != accounts[4] || swap.PoolPcTokenAccount != accounts[5] ||
		swap.UserSourceOwner != accounts[16] {
		t.Fatalf("17-account shift was not applied: %+v", swap)
	}
}

func TestCurrentCpmmInstructionUsesPoolAccountThree(t *testing.T) {
	data := make([]byte, 24)
	binary.LittleEndian.PutUint64(data, discCpmmSwapIn)
	accounts := []string{"payer", "authority", "config", "pool"}
	ev := ParseRaydiumCpmmInstruction(data, accounts, "sig", 1, 0, nil, 1)
	swap := ev.Data.(*RaydiumCpmmSwapEvent)
	if swap.PoolID != "pool" {
		t.Fatalf("expected pool account 3, got %q", swap.PoolID)
	}
}

func TestCurrentRaydiumSwapDedupePreservesOccurrences(t *testing.T) {
	cpmm := func(amount uint64) DexEvent {
		return DexEvent{Type: EventTypeRaydiumCpmmSwap, Data: &RaydiumCpmmSwapEvent{
			PoolID: "pool", InputAmount: amount, BaseInput: true,
		}}
	}
	amm := func(amount uint64) DexEvent {
		return DexEvent{Type: EventTypeRaydiumAmmV4Swap, Data: &RaydiumAmmV4SwapEvent{
			AmountIn: amount,
		}}
	}
	out := DedupeLogInstructionEvents(
		[]DexEvent{cpmm(10), cpmm(20), amm(30), amm(30)},
		[]DexEvent{cpmm(0), cpmm(0), amm(30), amm(30)},
	)
	if len(out) != 4 {
		t.Fatalf("expected four distinct swap occurrences, got %d", len(out))
	}
}
