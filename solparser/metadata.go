package solparser

import (
	"encoding/json"
	"errors"
	"strconv"
)

type EventMetadata struct {
	Signature       string `json:"signature"`
	Slot            uint64 `json:"slot"`
	TxIndex         uint64 `json:"tx_index"`
	BlockTimeUs     int64  `json:"block_time_us"`
	GrpcRecvUs      int64  `json:"grpc_recv_us"`
	RecentBlockhash string `json:"recent_blockhash,omitempty"`
}

func makeMetadata(sig string, slot, tx uint64, blockUs *int64, grpcUs int64, rb string) EventMetadata {
	bt := int64(0)
	if blockUs != nil {
		bt = *blockUs
	}
	return EventMetadata{
		Signature:       sig,
		Slot:            slot,
		TxIndex:         tx,
		BlockTimeUs:     bt,
		GrpcRecvUs:      grpcUs,
		RecentBlockhash: rb,
	}
}

// MarshalJSON uses decimal strings for exact 64-bit metadata, matching Node and Python.
func (m EventMetadata) MarshalJSON() ([]byte, error) {
	type wire struct {
		Signature       string `json:"signature"`
		Slot            string `json:"slot"`
		TxIndex         string `json:"tx_index"`
		BlockTimeUs     string `json:"block_time_us"`
		GrpcRecvUs      string `json:"grpc_recv_us"`
		RecentBlockhash string `json:"recent_blockhash,omitempty"`
	}
	return json.Marshal(wire{m.Signature, strconv.FormatUint(m.Slot, 10), strconv.FormatUint(m.TxIndex, 10), strconv.FormatInt(m.BlockTimeUs, 10), strconv.FormatInt(m.GrpcRecvUs, 10), m.RecentBlockhash})
}

// UnmarshalJSON accepts exact JSON integer literals and decimal strings. Floats/null fail.
func (m *EventMetadata) UnmarshalJSON(data []byte) error {
	var raw struct {
		Signature       string          `json:"signature"`
		Slot            json.RawMessage `json:"slot"`
		TxIndex         json.RawMessage `json:"tx_index"`
		BlockTimeUs     json.RawMessage `json:"block_time_us"`
		GrpcRecvUs      json.RawMessage `json:"grpc_recv_us"`
		RecentBlockhash string          `json:"recent_blockhash"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	integer := func(value json.RawMessage) (string, error) {
		if len(value) == 0 {
			return "0", nil
		}
		if value[0] == '"' {
			var text string
			if err := json.Unmarshal(value, &text); err != nil {
				return "", err
			}
			if len(text) == 0 {
				return "", errors.New("invalid metadata integer")
			}
			for i, c := range text {
				if c == '-' && i == 0 && len(text) > 1 {
					continue
				}
				if c < '0' || c > '9' {
					return "", errors.New("invalid metadata integer")
				}
			}
			return text, nil
		}
		return string(value), nil
	}
	unsigned := func(value json.RawMessage) (uint64, error) {
		text, e := integer(value)
		if e != nil {
			return 0, e
		}
		return strconv.ParseUint(text, 10, 64)
	}
	signed := func(value json.RawMessage) (int64, error) {
		text, e := integer(value)
		if e != nil {
			return 0, e
		}
		return strconv.ParseInt(text, 10, 64)
	}
	next := EventMetadata{Signature: raw.Signature, RecentBlockhash: raw.RecentBlockhash}
	var err error
	if next.Slot, err = unsigned(raw.Slot); err != nil {
		return err
	}
	if next.TxIndex, err = unsigned(raw.TxIndex); err != nil {
		return err
	}
	if next.BlockTimeUs, err = signed(raw.BlockTimeUs); err != nil {
		return err
	}
	if next.GrpcRecvUs, err = signed(raw.GrpcRecvUs); err != nil {
		return err
	}
	*m = next
	return nil
}
