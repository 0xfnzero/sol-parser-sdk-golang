package shredstream

import (
	"encoding/binary"
	"testing"
)

func TestDecodeEntriesBincode_EmptyVec(t *testing.T) {
	// bincode Vec<Entry> length 0: little-endian u64
	data := make([]byte, 8)
	txs, err := DecodeEntriesBincode(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 0 {
		t.Fatalf("expected 0 txs, got %d", len(txs))
	}
}

func TestDecodeEntriesBincode_OneEntryZeroTx(t *testing.T) {
	// One Entry: num_hashes=0, hash=32 zeros, tx_count=0
	var buf []byte
	n := make([]byte, 8)
	binary.LittleEndian.PutUint64(n, 1) // vec len 1
	buf = append(buf, n...)
	binary.LittleEndian.PutUint64(n, 0) // num_hashes
	buf = append(buf, n...)
	buf = append(buf, make([]byte, 32)...) // hash
	binary.LittleEndian.PutUint64(n, 0) // tx_count
	buf = append(buf, n...)

	txs, err := DecodeEntriesBincode(buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 0 {
		t.Fatalf("expected 0 txs, got %d", len(txs))
	}
}

func TestDecodeEntriesBincode_TruncatedVec(t *testing.T) {
	_, err := DecodeEntriesBincode([]byte{1, 0, 0, 0}) // too short for u64 count
	if err == nil {
		t.Fatal("expected error")
	}
}
