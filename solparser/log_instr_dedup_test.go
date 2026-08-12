package solparser

import "testing"

func clmmDedupFixture(pool string, zeroForOne bool, amount0 uint64) DexEvent {
	return DexEvent{
		Type: EventTypeRaydiumClmmSwap,
		Data: &RaydiumClmmSwapEvent{PoolState: pool, ZeroForOne: zeroForOne, Amount0: amount0},
	}
}

func TestClmmLogInstructionDedupIgnoresPlaceholderDirection(t *testing.T) {
	out := DedupeLogInstructionEvents(
		[]DexEvent{clmmDedupFixture("pool", false, 123)},
		[]DexEvent{clmmDedupFixture("pool", true, 0)},
	)
	if len(out) != 1 {
		t.Fatalf("expected one CLMM swap, got %d", len(out))
	}
	swap := out[0].Data.(*RaydiumClmmSwapEvent)
	if swap.Amount0 != 123 || swap.ZeroForOne {
		t.Fatalf("log values were not preserved: %+v", swap)
	}
}

func TestClmmSamePoolOccurrencesAreRetained(t *testing.T) {
	out := DedupeLogInstructionEvents(
		[]DexEvent{clmmDedupFixture("pool", false, 1), clmmDedupFixture("pool", true, 2)},
		[]DexEvent{clmmDedupFixture("pool", true, 0), clmmDedupFixture("pool", false, 0)},
	)
	if len(out) != 2 {
		t.Fatalf("expected two CLMM swaps, got %d", len(out))
	}
}
