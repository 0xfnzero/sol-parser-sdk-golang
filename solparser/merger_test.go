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

func intPtr(i int) *int { return &i }
