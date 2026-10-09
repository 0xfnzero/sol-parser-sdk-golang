package solparser

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestOrderedRejectsClosedSlotLateBatches(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeOrdered
	d := newDexOrderDispatcher(cfg)
	var out []DexEvent
	emit := func(e DexEvent) { out = append(out, e) }
	for _, p := range [][2]uint64{{10, 2}, {11, 0}, {10, 1}, {10, 9}, {12, 0}} {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("tx", p[0], p[1])}, p[0], p[1], emit)
	}
	d.flushAll(emit)
	if len(out) != 3 || d.orderedLateTransactions != 2 {
		t.Fatalf("bad late handling: output=%d drops=%d", len(out), d.orderedLateTransactions)
	}
	for i, p := range [][2]uint64{{10, 2}, {11, 0}, {12, 0}} {
		m := out[i].GetMetadata()
		if m.Slot != p[0] || m.TxIndex != p[1] {
			t.Fatalf("bad order: %+v", m)
		}
	}
}
func TestOrderedTimeoutWatermarkPreservesWholeBatch(t *testing.T) {
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeOrdered
	d := newDexOrderDispatcher(cfg)
	var out []DexEvent
	emit := func(e DexEvent) { out = append(out, e) }
	d.pushTransactionEvents([]DexEvent{orderTestEvent("a", 42, 2), orderTestEvent("b", 42, 2)}, 42, 2, emit)
	d.lastFlush = d.lastFlush.Add(-2 * d.timeout)
	d.flushDue(emit)
	for _, idx := range []uint64{0, 1, 2, 3} {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("next", 42, idx)}, 42, idx, emit)
	}
	d.flushAll(emit)
	if len(out) != 3 || out[0].GetMetadata().Signature != "a" || out[1].GetMetadata().Signature != "b" || out[2].GetMetadata().TxIndex != 3 || d.orderedLateTransactions != 3 {
		t.Fatalf("bad timeout watermark: output=%+v drops=%d", out, d.orderedLateTransactions)
	}
}

func TestOrderedLateReplayBoundsDiagnostics(t *testing.T) {
	var diagnostics bytes.Buffer
	original := log.Writer()
	log.SetOutput(&diagnostics)
	defer log.SetOutput(original)
	cfg := DefaultClientConfig()
	cfg.OrderMode = OrderModeOrdered
	d := newDexOrderDispatcher(cfg)
	var out []DexEvent
	emit := func(e DexEvent) { out = append(out, e) }
	d.pushTransactionEvents([]DexEvent{orderTestEvent("tx", 10, 2)}, 10, 2, emit)
	d.flushAll(emit)
	for i := 0; i < 10000; i++ {
		d.pushTransactionEvents([]DexEvent{orderTestEvent("late", 10, 1)}, 10, 1, emit)
	}
	if d.orderedLateTransactions != 10000 || len(out) != 1 || strings.Count(diagnostics.String(), "Ordered continuity break") != 20 {
		t.Fatalf("drops=%d emitted=%d logs=%d", d.orderedLateTransactions, len(out), strings.Count(diagnostics.String(), "Ordered continuity break"))
	}
	d.pushTransactionEvents([]DexEvent{orderTestEvent("next", 10, 3)}, 10, 3, emit)
	d.flushAll(emit)
	if len(out) != 2 {
		t.Fatal("valid progress lost")
	}
}
