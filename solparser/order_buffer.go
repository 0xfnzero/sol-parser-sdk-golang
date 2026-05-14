package solparser

import (
	"context"
	"sort"
	"sync"
	"time"
)

type dexEventBatch struct {
	slot    uint64
	txIndex uint64
	seq     uint64
	events  []DexEvent
}

type dexOrderDispatcher struct {
	mode          OrderMode
	timeout       time.Duration
	microBatchWin time.Duration

	mu          sync.Mutex
	slots       map[uint64][]dexEventBatch
	watermarks  map[uint64]uint64
	microBatch  []dexEventBatch
	microStart  time.Time
	lastFlush   time.Time
	currentSlot uint64
	seq         uint64

	stopOnce sync.Once
	stopCh   chan struct{}
}

func newDexOrderDispatcher(config ClientConfig) *dexOrderDispatcher {
	timeoutMs := config.OrderTimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 100
	}
	microBatchUs := config.MicroBatchUs
	if microBatchUs <= 0 {
		microBatchUs = 100
	}
	return &dexOrderDispatcher{
		mode:          config.OrderMode,
		timeout:       time.Duration(timeoutMs) * time.Millisecond,
		microBatchWin: time.Duration(microBatchUs) * time.Microsecond,
		slots:         make(map[uint64][]dexEventBatch),
		watermarks:    make(map[uint64]uint64),
		lastFlush:     time.Now(),
		stopCh:        make(chan struct{}),
	}
}

func (d *dexOrderDispatcher) start(ctx context.Context, emit func(DexEvent)) {
	if d.mode == OrderModeUnordered {
		return
	}
	interval := d.timeout / 2
	if d.mode == OrderModeMicroBatch {
		interval = d.microBatchWin
	}
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-d.stopCh:
				return
			case <-ticker.C:
				d.flushDue(emit)
			}
		}
	}()
}

func (d *dexOrderDispatcher) stop() {
	d.stopOnce.Do(func() {
		close(d.stopCh)
	})
}

func (d *dexOrderDispatcher) pushTransactionEvents(events []DexEvent, fallbackSlot, fallbackTxIndex uint64, emit func(DexEvent)) {
	if len(events) == 0 {
		return
	}
	meta := events[0].GetMetadata()
	slot := fallbackSlot
	if meta.Slot != 0 {
		slot = meta.Slot
	}
	txIndex := fallbackTxIndex
	if meta.TxIndex != 0 {
		txIndex = meta.TxIndex
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	batch := dexEventBatch{slot: slot, txIndex: txIndex, seq: d.seq, events: events}
	d.seq++

	switch d.mode {
	case OrderModeUnordered:
		emitBatch(batch, emit)
	case OrderModeOrdered:
		d.pushOrderedLocked(batch, emit)
	case OrderModeStreamingOrdered:
		d.pushStreamingLocked(batch, emit)
	case OrderModeMicroBatch:
		d.pushMicroBatchLocked(batch, emit)
	default:
		emitBatch(batch, emit)
	}
}

func (d *dexOrderDispatcher) flushDue(emit func(DexEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	if (d.mode == OrderModeOrdered || d.mode == OrderModeStreamingOrdered) && now.Sub(d.lastFlush) > d.timeout {
		d.flushAllSlotsLocked(emit)
	}
	if d.mode == OrderModeMicroBatch && len(d.microBatch) > 0 && now.Sub(d.microStart) >= d.microBatchWin {
		d.flushMicroBatchLocked(emit)
	}
}

func (d *dexOrderDispatcher) flushAll(emit func(DexEvent)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.flushAllSlotsLocked(emit)
	d.flushMicroBatchLocked(emit)
}

func (d *dexOrderDispatcher) pushOrderedLocked(batch dexEventBatch, emit func(DexEvent)) {
	if batch.slot > d.currentSlot && d.currentSlot > 0 {
		d.flushBeforeLocked(batch.slot, emit)
	}
	if batch.slot > d.currentSlot {
		d.currentSlot = batch.slot
	}
	d.slots[batch.slot] = append(d.slots[batch.slot], batch)
}

func (d *dexOrderDispatcher) pushStreamingLocked(batch dexEventBatch, emit func(DexEvent)) {
	if batch.slot > d.currentSlot && d.currentSlot > 0 {
		d.flushBeforeLocked(batch.slot, emit)
		for slot := range d.watermarks {
			if slot < batch.slot {
				delete(d.watermarks, slot)
			}
		}
	}
	if batch.slot > d.currentSlot {
		d.currentSlot = batch.slot
	}

	expected := d.watermarks[batch.slot]
	switch {
	case batch.txIndex == expected:
		emitBatch(batch, emit)
		watermark := expected + 1
		buffered := d.slots[batch.slot]
		sortBatches(buffered)
		for {
			pos := -1
			for i := range buffered {
				if buffered[i].txIndex == watermark {
					pos = i
					break
				}
			}
			if pos < 0 {
				break
			}
			emitBatch(buffered[pos], emit)
			buffered = append(buffered[:pos], buffered[pos+1:]...)
			watermark++
		}
		if len(buffered) == 0 {
			delete(d.slots, batch.slot)
		} else {
			d.slots[batch.slot] = buffered
		}
		d.watermarks[batch.slot] = watermark
		d.lastFlush = time.Now()
	case batch.txIndex > expected:
		d.slots[batch.slot] = append(d.slots[batch.slot], batch)
	}
}

func (d *dexOrderDispatcher) pushMicroBatchLocked(batch dexEventBatch, emit func(DexEvent)) {
	if len(d.microBatch) == 0 {
		d.microStart = time.Now()
	}
	d.microBatch = append(d.microBatch, batch)
	if time.Since(d.microStart) >= d.microBatchWin {
		d.flushMicroBatchLocked(emit)
	}
}

func (d *dexOrderDispatcher) flushBeforeLocked(slot uint64, emit func(DexEvent)) {
	keys := make([]uint64, 0, len(d.slots))
	for s := range d.slots {
		if s < slot {
			keys = append(keys, s)
		}
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, s := range keys {
		batches := d.slots[s]
		sortBatches(batches)
		for _, batch := range batches {
			emitBatch(batch, emit)
		}
		delete(d.slots, s)
		delete(d.watermarks, s)
	}
	d.lastFlush = time.Now()
}

func (d *dexOrderDispatcher) flushAllSlotsLocked(emit func(DexEvent)) {
	keys := make([]uint64, 0, len(d.slots))
	for s := range d.slots {
		keys = append(keys, s)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, s := range keys {
		batches := d.slots[s]
		sortBatches(batches)
		for _, batch := range batches {
			emitBatch(batch, emit)
		}
		delete(d.slots, s)
		delete(d.watermarks, s)
	}
	d.lastFlush = time.Now()
}

func (d *dexOrderDispatcher) flushMicroBatchLocked(emit func(DexEvent)) {
	if len(d.microBatch) == 0 {
		return
	}
	sortBatches(d.microBatch)
	for _, batch := range d.microBatch {
		emitBatch(batch, emit)
	}
	d.microBatch = nil
	d.microStart = time.Time{}
	d.lastFlush = time.Now()
}

func sortBatches(batches []dexEventBatch) {
	sort.SliceStable(batches, func(i, j int) bool {
		if batches[i].slot != batches[j].slot {
			return batches[i].slot < batches[j].slot
		}
		if batches[i].txIndex != batches[j].txIndex {
			return batches[i].txIndex < batches[j].txIndex
		}
		return batches[i].seq < batches[j].seq
	})
}

func emitBatch(batch dexEventBatch, emit func(DexEvent)) {
	for _, event := range batch.events {
		emit(event)
	}
}
