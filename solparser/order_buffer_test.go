package solparser

import "testing"

func orderTestEvent(signature string, slot, txIndex uint64) DexEvent {
	return DexEvent{
		Type: EventTypePumpFunTrade,
		Data: &PumpFunTradeEvent{
			Metadata: EventMetadata{
				Signature: signature,
				Slot:      slot,
				TxIndex:   txIndex,
			},
		},
	}
}

func TestDexOrderDispatcherOrdersBufferedTransactions(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeOrdered
	d := newDexOrderDispatcher(cfg)
	out := make([]DexEvent, 0)
	emit := func(ev DexEvent) { out = append(out, ev) }

	d.pushTransactionEvents([]DexEvent{orderTestEvent("tx2", 1, 2)}, 1, 2, emit)
	d.pushTransactionEvents([]DexEvent{orderTestEvent("tx1", 1, 1)}, 1, 1, emit)
	if len(out) != 0 {
		t.Fatalf("expected no events before next slot, got %d", len(out))
	}
	d.pushTransactionEvents([]DexEvent{orderTestEvent("tx0", 2, 0)}, 2, 0, emit)
	if got := []string{out[0].GetMetadata().Signature, out[1].GetMetadata().Signature}; got[0] != "tx1" || got[1] != "tx2" {
		t.Fatalf("unexpected order: %#v", got)
	}
}

func TestDexOrderDispatcherStreamsWholeTransactionBatch(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeStreamingOrdered
	d := newDexOrderDispatcher(cfg)
	out := make([]DexEvent, 0)
	emit := func(ev DexEvent) { out = append(out, ev) }

	d.pushTransactionEvents([]DexEvent{
		orderTestEvent("tx0-a", 1, 0),
		orderTestEvent("tx0-b", 1, 0),
	}, 1, 0, emit)

	if len(out) != 2 {
		t.Fatalf("expected both same-transaction events, got %d", len(out))
	}
	if out[0].GetMetadata().Signature != "tx0-a" || out[1].GetMetadata().Signature != "tx0-b" {
		t.Fatalf("unexpected batch output: %s, %s", out[0].GetMetadata().Signature, out[1].GetMetadata().Signature)
	}
}
