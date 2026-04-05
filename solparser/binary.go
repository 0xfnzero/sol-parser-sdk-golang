package solparser

import (
	"encoding/binary"
	"math/big"
)

func disc8(b ...byte) uint64 {
	var a [8]byte
	copy(a[:], b)
	return binary.LittleEndian.Uint64(a[:])
}

func readU8(b []byte, o int) (byte, bool) {
	if o >= len(b) {
		return 0, false
	}
	return b[o], true
}

func readU16LE(b []byte, o int) (uint16, bool) {
	if o+2 > len(b) {
		return 0, false
	}
	return binary.LittleEndian.Uint16(b[o:]), true
}

func readU32LE(b []byte, o int) (uint32, bool) {
	if o+4 > len(b) {
		return 0, false
	}
	return binary.LittleEndian.Uint32(b[o:]), true
}

func readI32LE(b []byte, o int) (int32, bool) {
	if o+4 > len(b) {
		return 0, false
	}
	v := binary.LittleEndian.Uint32(b[o:])
	return int32(v), true
}

func readU64LE(b []byte, o int) (uint64, bool) {
	if o+8 > len(b) {
		return 0, false
	}
	return binary.LittleEndian.Uint64(b[o:]), true
}

func readI64LE(b []byte, o int) (int64, bool) {
	u, ok := readU64LE(b, o)
	if !ok {
		return 0, false
	}
	return int64(u), true
}

func readU128LE(b []byte, o int) ([16]byte, bool) {
	var out [16]byte
	if o+16 > len(b) {
		return out, false
	}
	copy(out[:], b[o:o+16])
	return out, true
}

func readBool(b []byte, o int) (bool, bool) {
	v, ok := readU8(b, o)
	if !ok {
		return false, false
	}
	return v == 1, true
}

func readDiscU64(b []byte) (uint64, bool) {
	return readU64LE(b, 0)
}

func readBorshString(b []byte, o int) (s string, next int, ok bool) {
	l, ok := readU32LE(b, o)
	if !ok || int(o)+4+int(l) > len(b) {
		return "", o, false
	}
	return string(b[o+4 : o+4+int(l)]), o + 4 + int(l), true
}

// u128LEDecimalString 将 16 字节小端 u128 转为十进制字符串，与 TS `bigint` 经 JSON 序列化后的形态一致。
func u128LEDecimalString(u [16]byte) string {
	be := make([]byte, 16)
	for i := 0; i < 16; i++ {
		be[15-i] = u[i]
	}
	return new(big.Int).SetBytes(be).String()
}
