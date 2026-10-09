package solparser

// Native Solana transaction decoding, including V1's message-first wire layout.
import (
	"encoding/binary"
	"errors"
	"github.com/mr-tron/base58"
)

type WireHeader struct {
	NumRequiredSignatures       int `json:"numRequiredSignatures"`
	NumReadonlySignedAccounts   int `json:"numReadonlySignedAccounts"`
	NumReadonlyUnsignedAccounts int `json:"numReadonlyUnsignedAccounts"`
}
type WireInstruction struct {
	ProgramIDIndex int    `json:"programIdIndex"`
	Accounts       []int  `json:"accounts"`
	Data           string `json:"data"`
}
type WireLookup struct {
	AccountKey      string `json:"accountKey"`
	WritableIndexes []int  `json:"writableIndexes"`
	ReadonlyIndexes []int  `json:"readonlyIndexes"`
}
type WireConfig struct {
	PriorityFee                 *uint64 `json:"priorityFee,omitempty"`
	ComputeUnitLimit            *uint32 `json:"computeUnitLimit,omitempty"`
	LoadedAccountsDataSizeLimit *uint32 `json:"loadedAccountsDataSizeLimit,omitempty"`
	HeapSize                    *uint32 `json:"heapSize,omitempty"`
}
type WireMessage struct {
	Header              WireHeader        `json:"header"`
	AccountKeys         []string          `json:"accountKeys"`
	RecentBlockhash     string            `json:"recentBlockhash"`
	Instructions        []WireInstruction `json:"instructions"`
	AddressTableLookups []WireLookup      `json:"addressTableLookups"`
	Config              *WireConfig       `json:"config,omitempty"`
}
type NativeWireTransaction struct {
	Signatures []string    `json:"signatures"`
	Message    WireMessage `json:"message"`
	Version    any         `json:"version"`
}
type wireReader struct {
	data []byte
	pos  int
	err  error
}

func (r *wireReader) take(n int) []byte {
	if r.err != nil {
		return make([]byte, n)
	}
	if n < 0 || n > len(r.data)-r.pos {
		r.err = errors.New("truncated wire transaction")
		return make([]byte, n)
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b
}
func (r *wireReader) u8() int { return int(r.take(1)[0]) }
func (r *wireReader) short() int {
	v := 0
	for i := 0; i < 3; i++ {
		b := r.u8()
		if i == 2 && b > 3 {
			r.err = errors.New("invalid short-u16")
			return 0
		}
		v |= (b & 127) << (i * 7)
		if b&128 == 0 {
			if i > 0 && b == 0 {
				r.err = errors.New("noncanonical short-u16")
			}
			return v
		}
	}
	r.err = errors.New("invalid short-u16")
	return 0
}
func (r *wireReader) indices(n int) []int {
	b := r.take(n)
	out := make([]int, n)
	for i, v := range b {
		out[i] = int(v)
	}
	return out
}

// DecodeWireTransaction returns the number of bytes consumed. Set requireComplete
// false when decoding one transaction from a bincode Entry stream.
func DecodeWireTransaction(data []byte, offset int, requireComplete bool) (*NativeWireTransaction, int, error) {
	if offset < 0 || offset >= len(data) {
		return nil, 0, errors.New("invalid wire offset")
	}
	r := wireReader{data: data, pos: offset}
	tx := &NativeWireTransaction{Version: "legacy", Signatures: []string{}}
	m := &tx.Message
	m.Instructions = []WireInstruction{}
	m.AddressTableLookups = []WireLookup{}
	var h []byte
	if data[offset] == 129 {
		r.take(1)
		tx.Version = 1
		h = r.take(3)
		mask := binary.LittleEndian.Uint32(r.take(4))
		m.RecentBlockhash = base58.Encode(r.take(32))
		if mask&^uint32(31) != 0 || (mask&3 != 0 && mask&3 != 3) {
			return nil, 0, errors.New("invalid V1 config mask")
		}
		count, n := r.u8(), r.u8()
		if count > 64 || n > 64 || h[0] > 12 {
			return nil, 0, errors.New("V1 transaction limits")
		}
		m.AccountKeys = make([]string, n)
		for i := range m.AccountKeys {
			m.AccountKeys[i] = base58.Encode(r.take(32))
		}
		m.Config = &WireConfig{}
		if mask&3 != 0 {
			v := binary.LittleEndian.Uint64(r.take(8))
			m.Config.PriorityFee = &v
		}
		if mask&4 != 0 {
			v := binary.LittleEndian.Uint32(r.take(4))
			m.Config.ComputeUnitLimit = &v
		}
		if mask&8 != 0 {
			v := binary.LittleEndian.Uint32(r.take(4))
			m.Config.LoadedAccountsDataSizeLimit = &v
		}
		if mask&16 != 0 {
			v := binary.LittleEndian.Uint32(r.take(4))
			if v < 32768 || v > 262144 || v%1024 != 0 {
				return nil, 0, errors.New("invalid heap size")
			}
			m.Config.HeapSize = &v
		}
		headers := make([][3]int, count)
		for i := range headers {
			headers[i] = [3]int{r.u8(), r.u8(), int(binary.LittleEndian.Uint16(r.take(2)))}
		}
		for _, v := range headers {
			m.Instructions = append(m.Instructions, WireInstruction{v[0], r.indices(v[1]), base58.Encode(r.take(v[2]))})
		}
		for i := 0; i < int(h[0]); i++ {
			tx.Signatures = append(tx.Signatures, base58.Encode(r.take(64)))
		}
		if r.pos-offset > 4096 {
			return nil, 0, errors.New("V1 exceeds 4096 bytes")
		}
	} else {
		n := r.short()
		if n > 127 {
			return nil, 0, errors.New("too many signatures")
		}
		for i := 0; i < n; i++ {
			tx.Signatures = append(tx.Signatures, base58.Encode(r.take(64)))
		}
		first := r.u8()
		if first&128 != 0 {
			if first != 128 {
				return nil, 0, errors.New("unknown wire version")
			}
			tx.Version = 0
			h = r.take(3)
		} else {
			h = append([]byte{byte(first)}, r.take(2)...)
		}
		n = r.short()
		if n > 256 {
			return nil, 0, errors.New("too many keys")
		}
		m.AccountKeys = make([]string, n)
		for i := range m.AccountKeys {
			m.AccountKeys[i] = base58.Encode(r.take(32))
		}
		m.RecentBlockhash = base58.Encode(r.take(32))
		count := r.short()
		for i := 0; i < count && r.err == nil; i++ {
			p := r.u8()
			a := r.indices(r.short())
			d := base58.Encode(r.take(r.short()))
			m.Instructions = append(m.Instructions, WireInstruction{p, a, d})
		}
		if tx.Version == 0 {
			count := r.short()
			for i := 0; i < count && r.err == nil; i++ {
				k := base58.Encode(r.take(32))
				a := r.indices(r.short())
				b := r.indices(r.short())
				m.AddressTableLookups = append(m.AddressTableLookups, WireLookup{k, a, b})
			}
		}
		if len(tx.Signatures) != int(h[0]) {
			return nil, 0, errors.New("signature count mismatch")
		}
	}
	if r.err != nil {
		return nil, 0, r.err
	}
	if int(h[0]) > len(m.AccountKeys) || h[1] >= h[0] || int(h[2]) > len(m.AccountKeys)-int(h[0]) {
		return nil, 0, errors.New("invalid wire header")
	}

	if tx.Version == 1 {
		unique := map[string]bool{}
		for _, key := range m.AccountKeys {
			if unique[key] {
				return nil, 0, errors.New("duplicate V1 accounts")
			}
			unique[key] = true
		}
	}
	accountCount := len(m.AccountKeys)
	for _, l := range m.AddressTableLookups {
		if len(l.WritableIndexes) == 0 && len(l.ReadonlyIndexes) == 0 {
			return nil, 0, errors.New("empty address table lookup")
		}
		accountCount += len(l.WritableIndexes) + len(l.ReadonlyIndexes)
	}
	if accountCount > 256 {
		return nil, 0, errors.New("too many resolved account keys")
	}
	for _, ix := range m.Instructions {
		if ix.ProgramIDIndex == 0 || ix.ProgramIDIndex >= len(m.AccountKeys) {
			return nil, 0, errors.New("invalid program index")
		}
		for _, i := range ix.Accounts {
			if i >= accountCount {
				return nil, 0, errors.New("invalid account index")
			}
		}
	}
	if requireComplete && r.pos != len(data) {
		return nil, 0, errors.New("trailing wire bytes")
	}
	m.Header = WireHeader{int(h[0]), int(h[1]), int(h[2])}
	return tx, r.pos - offset, nil
}
