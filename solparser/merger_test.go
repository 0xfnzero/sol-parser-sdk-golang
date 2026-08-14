package solparser

import (
	"testing"
)

func TestMergeRpcInstructionEvents_OuterBeforeInner(t *testing.T) {
	inner := mergeRpcInstructionEvents([]rpcIndexedEvent{
		{OuterIdx: 0, InnerIdx: intPtr(0), Event: DexEvent{Type: EventTypePumpFunBuy, Data: &PumpFunTradeEvent{Metadata: EventMetadata{}, Mint: "inner"}}},
		{OuterIdx: 0, InnerIdx: nil, Event: DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{Metadata: EventMetadata{}, Mint: "outer", BondingCurve: "bc"}}},
	})
	if len(inner) != 1 {
		t.Fatalf("expected 1 merged, got %d", len(inner))
	}
	tr := inner[0].Data.(*PumpFunTradeEvent)
	if tr.Mint != "inner" || tr.BondingCurve != "bc" {
		t.Fatalf("merge fields: mint=%q bc=%q", tr.Mint, tr.BondingCurve)
	}
}

func TestMergeRpcInstructionEvents_ChainsMultipleInnerEvents(t *testing.T) {
	merged := mergeRpcInstructionEvents([]rpcIndexedEvent{
		{OuterIdx: 0, InnerIdx: nil, Event: DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{Metadata: EventMetadata{}, Mint: "outer"}}},
		{OuterIdx: 0, InnerIdx: intPtr(0), Event: DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{Metadata: EventMetadata{}, Mint: "inner-value"}}},
		{OuterIdx: 0, InnerIdx: intPtr(1), Event: DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{Metadata: EventMetadata{}, BondingCurve: "inner-account"}}},
	})
	if len(merged) != 1 {
		t.Fatalf("expected 1 merged event, got %d", len(merged))
	}
	tr := merged[0].Data.(*PumpFunTradeEvent)
	if tr.Mint != "inner-value" || tr.BondingCurve != "inner-account" {
		t.Fatalf("expected chained inner merge, got mint=%q bonding_curve=%q", tr.Mint, tr.BondingCurve)
	}
}

func dlmmSwapFixture(pool string, amountIn, amountOut uint64) DexEvent {
	return DexEvent{
		Type: EventTypeMeteoraDlmmSwap,
		Data: &MeteoraDlmmSwapEvent{
			Pool: pool, AmountIn: amountIn, AmountOut: amountOut,
		},
	}
}

func TestMergeRpcInstructionEvents_DlmmAggregatorSwaps(t *testing.T) {
	merged := mergeRpcInstructionEvents([]rpcIndexedEvent{
		{OuterIdx: 0, InnerIdx: intPtr(0), StackHeight: uint32Ptr(2), Event: dlmmSwapFixture("pool-1", 1, 0)},
		{OuterIdx: 0, InnerIdx: intPtr(1), StackHeight: uint32Ptr(3), IsDlmmEventCPI: true, Event: dlmmSwapFixture("pool-1", 10, 9)},
		{OuterIdx: 0, InnerIdx: intPtr(2), StackHeight: uint32Ptr(2), Event: dlmmSwapFixture("pool-2", 2, 0)},
		{OuterIdx: 0, InnerIdx: intPtr(3), StackHeight: uint32Ptr(3), IsDlmmEventCPI: true, Event: dlmmSwapFixture("pool-2", 20, 18)},
	})

	if len(merged) != 2 {
		t.Fatalf("expected two swaps, got %d", len(merged))
	}
	first := merged[0].Data.(*MeteoraDlmmSwapEvent)
	second := merged[1].Data.(*MeteoraDlmmSwapEvent)
	if first.AmountIn != 10 || first.AmountOut != 9 || second.AmountIn != 20 || second.AmountOut != 18 {
		t.Fatalf("unexpected merged amounts: first=%+v second=%+v", first, second)
	}
}

func TestMergeRpcInstructionEvents_PreservesUnrelatedInner(t *testing.T) {
	merged := mergeRpcInstructionEvents([]rpcIndexedEvent{
		{OuterIdx: 0, Event: dlmmSwapFixture("pool", 1, 0)},
		{OuterIdx: 0, InnerIdx: intPtr(0), Event: DexEvent{
			Type: EventTypeMeteoraDlmmAddLiquidity,
			Data: &MeteoraDlmmAddLiquidityEvent{},
		}},
	})
	if len(merged) != 2 || merged[0].Type != EventTypeMeteoraDlmmSwap ||
		merged[1].Type != EventTypeMeteoraDlmmAddLiquidity {
		t.Fatalf("unrelated inner event was lost or reordered: %+v", merged)
	}
}

func TestMergeDexEvents_DlmmPositionKeepsInstructionOnlyFields(t *testing.T) {
	base := DexEvent{
		Type: EventTypeMeteoraDlmmCreatePosition,
		Data: &MeteoraDlmmCreatePositionEvent{LowerBinID: -42, Width: 70},
	}
	inner := DexEvent{
		Type: EventTypeMeteoraDlmmCreatePosition,
		Data: &MeteoraDlmmCreatePositionEvent{},
	}
	if !tryMergeDexEvents(&base, inner) {
		t.Fatal("expected DLMM position merge")
	}
	position := base.Data.(*MeteoraDlmmCreatePositionEvent)
	if position.LowerBinID != -42 || position.Width != 70 {
		t.Fatalf("instruction-only fields were lost: %+v", position)
	}
}

func intPtr(i int) *int { return &i }
