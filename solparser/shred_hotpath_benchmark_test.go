package solparser

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func BenchmarkShredWireRoutes(b *testing.B) {
	raw, err := os.ReadFile("../examples/fixtures/pump_shred_routes.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixtures []struct {
		Wire     string
		Slot     uint64
		Writable []string `json:"loaded_writable"`
		Readonly []string `json:"loaded_readonly"`
		Expected []json.RawMessage
	}
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		b.Fatal(err)
	}
	wires := make([][]byte, len(fixtures))
	for i, f := range fixtures {
		wires[i], err = base64.StdEncoding.DecodeString(f.Wire)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, f := range fixtures {
			events, e := DexEventsFromShredTransactionWireWithLoadedAddresses(wires[j], f.Writable, f.Readonly, "", f.Slot, 888, nil, 999, nil)
			if e != nil || len(events) != len(f.Expected) {
				b.Fatal("event mismatch", e)
			}
		}
	}
}
