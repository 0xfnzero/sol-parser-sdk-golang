package solparser

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// 回归：Program log 的 8 字节 disc（binary.go）须与 inner 16 字节的后 8 字节及 magic 前缀一致（Rust pump_amm_inner / pump_inner）。

func TestPumpSwapInnerDiscMatchesLogDiscPlusMagic(t *testing.T) {
	magic := []byte{228, 69, 165, 46, 81, 203, 154, 29}
	var buf8 [8]byte
	binary.LittleEndian.PutUint64(buf8[:], discPSBuy)
	full := append(magic, buf8[:]...)
	if !bytes.Equal(full, pumpswapInnerBuy) {
		t.Fatalf("buy inner disc mismatch")
	}
	binary.LittleEndian.PutUint64(buf8[:], discPSSell)
	full = append(magic, buf8[:]...)
	if !bytes.Equal(full, pumpswapInnerSell) {
		t.Fatalf("sell inner disc mismatch")
	}
}

func TestPumpfunInnerTradeDiscMatchesLogDiscPlusSuffix(t *testing.T) {
	var d16 [16]byte
	binary.LittleEndian.PutUint64(d16[:8], discPumpTrade)
	copy(d16[8:], []byte{155, 167, 108, 32, 122, 76, 173, 64})
	if !bytes.Equal(d16[:], pumpfunInnerTradeEvent) {
		t.Fatalf("pumpfun inner trade disc mismatch")
	}
}
