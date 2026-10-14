package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestPumpDepth2SignedBankReplay(t *testing.T) {
	b, err := os.ReadFile("testdata/pump_depth2_20261014.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name     string
			Raw      json.RawMessage
			Expected map[string]any
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			tx := pumpUpgradeRPC(t, c.Raw)
			events, err := ParseRpcTransaction(tx, "signed-bank", nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, event := range events {
				if event.Type != EventTypePumpFunCreate {
					continue
				}
				count++
				b, err := json.Marshal(event.Data)
				if err != nil {
					t.Fatal(err)
				}
				d := json.NewDecoder(bytes.NewReader(b))
				d.UseNumber()
				var body map[string]any
				if err = d.Decode(&body); err != nil {
					t.Fatal(err)
				}
				for k, v := range c.Expected {
					if fmt.Sprint(body[k]) != fmt.Sprint(v) {
						t.Fatalf("%s: %v != %v", k, body[k], v)
					}
				}
			}
			if count != 1 {
				t.Fatal("create occurrences", count)
			}
		})
	}
}
