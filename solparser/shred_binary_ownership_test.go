package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/mr-tron/base58"
	"os"
	"reflect"
	"testing"
)

// Binary parsing must retain the public decoder/resolver schema and ownership:
// packet-buffer reuse cannot alter an already decoded instruction/event.
func TestShredBinaryDecodeParityAndOwnership(t *testing.T) {
	raw, e := os.ReadFile("../examples/fixtures/pump_shred_routes.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixtures []struct{ Wire string }
	if e = json.Unmarshal(raw, &fixtures); e != nil {
		t.Fatal(e)
	}
	for _, f := range fixtures {
		wire, e := base64.StdEncoding.DecodeString(f.Wire)
		if e != nil {
			t.Fatal(e)
		}
		public, n, e := DecodeWireTransaction(wire, 0, true)
		if e != nil {
			t.Fatal(e)
		}
		var binaryData [][]byte
		decoded, m, e := decodeWireTransaction(wire, 0, true, &binaryData)
		if e != nil || m != n {
			t.Fatal("binary decode mismatch", e)
		}
		copyData := make([][]byte, len(binaryData))
		for i, data := range binaryData {
			copyData[i] = bytes.Clone(data)
			decoded.Message.Instructions[i].Data = base58.Encode(data)
		}
		if !reflect.DeepEqual(decoded, public) {
			t.Fatal("public schema changed")
		}
		for _, end := range []int{0, 1, 63, 64, len(wire) / 2, len(wire) - 1} {
			_, _, oldErr := DecodeWireTransaction(wire[:end], 0, true)
			var out [][]byte
			_, _, newErr := decodeWireTransaction(wire[:end], 0, true, &out)
			if oldErr == nil || newErr == nil || oldErr.Error() != newErr.Error() {
				t.Fatal("truncation accepted or error changed", end)
			}
		}
		clear(wire)
		if !reflect.DeepEqual(binaryData, copyData) {
			t.Fatal("binary instruction retained input buffer")
		}
	}
}
