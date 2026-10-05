package solparser

import (
	"encoding/json"
	"math"
	"testing"
)

func TestMetadataDecimalJSON(t *testing.T) {
	m := EventMetadata{Slot: math.MaxUint64, TxIndex: math.MaxUint64, BlockTimeUs: math.MinInt64, GrpcRecvUs: math.MaxInt64}
	data, e := json.Marshal(m)
	if e != nil {
		t.Fatal(e)
	}
	var object map[string]any
	if e = json.Unmarshal(data, &object); e != nil {
		t.Fatal(e)
	}
	if object["slot"] != "18446744073709551615" || object["block_time_us"] != "-9223372036854775808" {
		t.Fatal(string(data))
	}
	var round EventMetadata
	if e = json.Unmarshal(data, &round); e != nil || round != m {
		t.Fatal(round, e)
	}
	for _, invalid := range []string{`{"slot":18446744073709551616}`, `{"slot":1.5}`, `{"slot":true}`, `{"slot":null}`, `{"slot":-1}`, `{"block_time_us":9223372036854775808}`} {
		if json.Unmarshal([]byte(invalid), &round) == nil {
			t.Fatal("invalid metadata accepted", invalid)
		}
	}
	if e = json.Unmarshal([]byte(`{"slot":18446744073709551615,"block_time_us":-9223372036854775808}`), &round); e != nil {
		t.Fatal(e)
	}
}
