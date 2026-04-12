// Copyright (c) ShredStream / shredstream-sdk-go contributors (MIT).
// Decoding logic adapted from https://github.com/shredstream/shredstream-sdk-go/blob/main/decoder.go
// to decode gRPC `Entry.entries`: bincode `Vec<solana_entry::entry::Entry>` with per-tx wire layout.

package shredstream

import (
	"encoding/binary"
	"fmt"

	"github.com/mr-tron/base58"

	pb "sol-parser-sdk-golang/shredstream/pb"
)

// DecodedTransaction 表示从 `entries` 负载中解出的一笔线格式交易（签名 + 原始字节）。
// 与 shredstream-sdk-go 的 Transaction 语义一致，便于对接解析或转发。
type DecodedTransaction struct {
	Signatures [][]byte
	Raw        []byte
}

// Signature 返回第一枚签名的 Base58（与常见 Solana 展示一致）；无签名时为空串。
func (tx *DecodedTransaction) Signature() string {
	if len(tx.Signatures) == 0 {
		return ""
	}
	return base58.Encode(tx.Signatures[0])
}

// DecodeGRPCEntry 解码 SubscribeEntries 流中的整条 `pb.Entry`：返回 slot 与负载内全部交易。
func DecodeGRPCEntry(e *pb.Entry) (slot uint64, txs []DecodedTransaction, err error) {
	if e == nil {
		return 0, nil, fmt.Errorf("shredstream: nil entry")
	}
	txs, err = DecodeEntriesBincode(e.GetEntries())
	if err != nil {
		return 0, nil, err
	}
	return e.GetSlot(), txs, nil
}

// DecodeEntriesBincode 解码 `Entry.entries` 字节：Rust 侧
// `bincode::deserialize::<Vec<solana_entry::entry::Entry>>` 与 solana-streamer 用法一致。
// 内部按 Solana Entry 字段顺序（num_hashes、hash、transactions 线格式）解析，与
// github.com/shredstream/shredstream-sdk-go 的 BatchDecoder 布局对齐。
func DecodeEntriesBincode(entriesBytes []byte) ([]DecodedTransaction, error) {
	if len(entriesBytes) == 0 {
		return nil, nil
	}
	dec := newBatchDecoder()
	txs, err := dec.push(entriesBytes)
	if err != nil {
		return nil, err
	}
	if dec.expectedCount != nil && dec.entriesYielded < *dec.expectedCount {
		return nil, fmt.Errorf("shredstream: truncated entries payload (got %d/%d entries)", dec.entriesYielded, *dec.expectedCount)
	}
	return txs, nil
}

type batchDecoder struct {
	buffer         []byte
	expectedCount  *uint64
	entriesYielded uint64
	cursor         int
}

func newBatchDecoder() *batchDecoder {
	return &batchDecoder{}
}

func (d *batchDecoder) push(payload []byte) ([]DecodedTransaction, error) {
	d.buffer = append(d.buffer, payload...)

	if d.expectedCount == nil {
		if len(d.buffer) < 8 {
			return nil, fmt.Errorf("shredstream: entries payload too short for vec length")
		}
		count := binary.LittleEndian.Uint64(d.buffer[0:8])
		if count > 100_000 {
			return nil, fmt.Errorf("shredstream: corrupt entry_count %d exceeds limit", count)
		}
		d.expectedCount = &count
		d.cursor = 8
	}

	var txs []DecodedTransaction

	for d.entriesYielded < *d.expectedCount {
		entryTxs, err := d.tryDecodeEntry()
		if err != nil {
			return txs, err
		}
		if entryTxs == nil {
			return nil, fmt.Errorf("shredstream: incomplete entry at offset %d", d.cursor)
		}
		txs = append(txs, entryTxs...)
		d.entriesYielded++
	}

	return txs, nil
}

func (d *batchDecoder) tryDecodeEntry() ([]DecodedTransaction, error) {
	pos := d.cursor
	buf := d.buffer

	if pos+48 > len(buf) {
		return nil, nil
	}

	pos += 8 + 32
	txCount := binary.LittleEndian.Uint64(buf[pos:])
	pos += 8

	txs := make([]DecodedTransaction, 0, txCount)

	for i := uint64(0); i < txCount; i++ {
		txStart := pos

		txLen, sigs := parseTransaction(buf, pos)
		if txLen < 0 {
			return nil, fmt.Errorf("shredstream: truncated transaction %d in entry", i)
		}

		txEnd := pos + txLen
		txRaw := make([]byte, txLen)
		copy(txRaw, buf[txStart:txEnd])

		txs = append(txs, DecodedTransaction{
			Signatures: sigs,
			Raw:        txRaw,
		})

		pos = txEnd
	}

	d.cursor = pos
	return txs, nil
}

func parseTransaction(buf []byte, pos int) (int, [][]byte) {
	start := pos

	if pos >= len(buf) {
		return -1, nil
	}
	sigCount, n := decodeCompactU16(buf, pos)
	if n < 0 {
		return -1, nil
	}
	pos += n

	sigsEnd := pos + sigCount*64
	if sigsEnd > len(buf) {
		return -1, nil
	}

	sigs := make([][]byte, sigCount)
	for i := 0; i < sigCount; i++ {
		sig := make([]byte, 64)
		copy(sig, buf[pos:pos+64])
		sigs[i] = sig
		pos += 64
	}

	if pos >= len(buf) {
		return -1, nil
	}
	msgFirst := buf[pos]
	isV0 := msgFirst >= 0x80

	if isV0 {
		pos++
	}

	pos += 3
	if pos > len(buf) {
		return -1, nil
	}

	if pos >= len(buf) {
		return -1, nil
	}
	acctCount, n := decodeCompactU16(buf, pos)
	if n < 0 {
		return -1, nil
	}
	pos += n
	pos += acctCount * 32
	if pos > len(buf) {
		return -1, nil
	}

	pos += 32
	if pos > len(buf) {
		return -1, nil
	}

	if pos >= len(buf) {
		return -1, nil
	}
	ixCount, n := decodeCompactU16(buf, pos)
	if n < 0 {
		return -1, nil
	}
	pos += n

	for ix := 0; ix < ixCount; ix++ {
		pos++
		if pos > len(buf) {
			return -1, nil
		}

		if pos >= len(buf) {
			return -1, nil
		}
		acctLen, n := decodeCompactU16(buf, pos)
		if n < 0 {
			return -1, nil
		}
		pos += n
		pos += acctLen
		if pos > len(buf) {
			return -1, nil
		}

		if pos >= len(buf) {
			return -1, nil
		}
		dataLen, n := decodeCompactU16(buf, pos)
		if n < 0 {
			return -1, nil
		}
		pos += n
		pos += dataLen
		if pos > len(buf) {
			return -1, nil
		}
	}

	if isV0 {
		if pos >= len(buf) {
			return -1, nil
		}
		atlCount, n := decodeCompactU16(buf, pos)
		if n < 0 {
			return -1, nil
		}
		pos += n

		for atl := 0; atl < atlCount; atl++ {
			pos += 32
			if pos > len(buf) {
				return -1, nil
			}

			if pos >= len(buf) {
				return -1, nil
			}
			wLen, n := decodeCompactU16(buf, pos)
			if n < 0 {
				return -1, nil
			}
			pos += n
			pos += wLen
			if pos > len(buf) {
				return -1, nil
			}

			if pos >= len(buf) {
				return -1, nil
			}
			rLen, n := decodeCompactU16(buf, pos)
			if n < 0 {
				return -1, nil
			}
			pos += n
			pos += rLen
			if pos > len(buf) {
				return -1, nil
			}
		}
	}

	return pos - start, sigs
}

func decodeCompactU16(buf []byte, pos int) (int, int) {
	if pos >= len(buf) {
		return 0, -1
	}

	b0 := buf[pos]
	if b0 < 0x80 {
		return int(b0), 1
	}

	if pos+1 >= len(buf) {
		return 0, -1
	}
	b1 := buf[pos+1]
	if b1 < 0x80 {
		return int(b0&0x7F) | (int(b1) << 7), 2
	}

	if pos+2 >= len(buf) {
		return 0, -1
	}
	b2 := buf[pos+2]
	return int(b0&0x7F) | (int(b1&0x7F) << 7) | (int(b2) << 14), 3
}
