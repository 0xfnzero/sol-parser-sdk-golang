package solparser

import (
	"testing"
	"time"
)

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

func TestStreamingTimeoutRetainsProgressAndBoundsPendingState(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeStreamingOrdered
	d := newDexOrderDispatcher(cfg)
	var out []DexEvent
	emit := func(e DexEvent) { out = append(out, e) }
	for _, index := range []uint64{0, 2, 2} {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("tx", 42, index)}, 42, index, emit)
	}
	d.lastFlush = time.Now().Add(-2 * d.timeout)
	d.flushDue(emit)
	for _, index := range []uint64{0, 1, 2, 3} {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("tx", 42, index)}, 42, index, emit)
	}
	if len(out) != 3 || out[0].GetMetadata().TxIndex != 0 || out[1].GetMetadata().TxIndex != 2 || out[2].GetMetadata().TxIndex != 3 {
		t.Fatalf("timeout/replay output: %#v", out)
	}
	if len(d.watermarks) != 1 || d.watermarks[42] != 4 || len(d.streamingPending) != 0 || len(d.slots) != 0 {
		t.Fatal("unexpected retained state")
	}
	d.pushTransactionEvents([]DexEvent{orderTestEvent("next", 43, 0)}, 43, 0, emit)
	d.pushTransactionEvents([]DexEvent{orderTestEvent("old", 42, 0)}, 42, 0, emit)
	if len(out) != 4 || out[3].GetMetadata().Signature != "next" || len(d.watermarks) != 1 {
		t.Fatal("old slot replay accepted")
	}
	d.pushTransactionEvents([]DexEvent{orderTestEvent("a", 43, 2), orderTestEvent("b", 43, 2)}, 43, 2, emit)
	d.pushTransactionEvents([]DexEvent{orderTestEvent("duplicate", 43, 2)}, 43, 2, emit)
	d.pushTransactionEvents([]DexEvent{orderTestEvent("gap", 43, 1)}, 43, 1, emit)
	if len(out) != 7 || out[4].GetMetadata().Signature != "gap" || out[5].GetMetadata().Signature != "a" || out[6].GetMetadata().Signature != "b" {
		t.Fatal("buffered batch order/uniqueness changed")
	}
	if len(d.streamingPending) != 0 || len(d.slots) != 0 {
		t.Fatal("pending batch state leaked")
	}
	d.flushAll(emit)
	d.pushTransactionEvents([]DexEvent{orderTestEvent("replay", 43, 0)}, 43, 0, emit)
	if len(out) != 7 {
		t.Fatal("explicit flush forgot streaming progress")
	}
}

func TestStreamingRepeatedPendingAndOldIndexesRemainBounded(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeStreamingOrdered
	d := newDexOrderDispatcher(cfg)
	var out []DexEvent
	emit := func(e DexEvent) { out = append(out, e) }
	for i := 0; i < 10000; i++ {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("pending", 42, 2)}, 42, 2, emit)
	}
	if len(d.slots[42]) != 1 || len(d.streamingPending) != 1 {
		t.Fatal("duplicate input grew pending state")
	}
	d.flushAll(emit)
	if len(out) != 1 || len(d.slots) != 0 || len(d.streamingPending) != 0 {
		t.Fatal("flush did not release/reset pending state")
	}
	d.pushTransactionEvents([]DexEvent{orderTestEvent("max", 42, ^uint64(0))}, 42, ^uint64(0), emit)
	d.flushAll(emit)
	if len(out) != 1 {
		t.Fatal("max index cannot advance a u64 watermark")
	}
	d.pushTransactionEvents([]DexEvent{orderTestEvent("next", 43, 0)}, 43, 0, emit)
	for i := 0; i < 10000; i++ {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("old", 42, 2)}, 42, 2, emit)
	}
	if len(out) != 2 || len(d.watermarks) != 1 || len(d.slots) != 0 {
		t.Fatal("old input grew state or replayed events")
	}
}
