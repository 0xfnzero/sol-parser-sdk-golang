package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mr-tron/base58"
	"os"
	"testing"
)

func TestDbcV1MigrationBank(t *testing.T) {
	b, err := os.ReadFile("testdata/dbc_v1_migration_20261008.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name     string
			Raw      json.RawMessage
			Expected []map[string]string
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Cases) != 2 {
		t.Fatal("scenario count")
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			tx := pumpUpgradeRPC(t, c.Raw)
			events, err := ParseRpcTransaction(tx, "simulation", nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			if tx.Meta.Err != nil {
				if len(events) != 0 {
					t.Fatal("rolled-back events retained")
				}
				return
			}
			var lp []map[string]any
			for _, e := range events {
				if e.Type != EventTypeMeteoraPoolsPoolCreated {
					continue
				}
				b, err := json.Marshal(e.Data)
				if err != nil {
					t.Fatal(err)
				}
				d := json.NewDecoder(bytes.NewReader(b))
				d.UseNumber()
				var fields map[string]any
				if err = d.Decode(&fields); err != nil {
					t.Fatal(err)
				}
				fields["type"] = string(e.Type)
				lp = append(lp, fields)
			}
			if len(lp) != len(c.Expected) {
				t.Fatal("LP occurrences", len(lp), len(c.Expected))
			}
			for i, w := range c.Expected {
				for k, v := range w {
					if fmt.Sprint(lp[i][k]) != v {
						t.Fatal(k, lp[i][k], v)
					}
				}
			}
		})
	}
}

func TestDammV1Config2InstructionWithoutLogs(t *testing.T) {
	b, e := os.ReadFile("testdata/damm_v1_config2_instruction_20261008.json")
	if e != nil {
		t.Fatal(e)
	}
	var f struct {
		Data     string
		Accounts []string
		Expected map[string]string
	}
	if e = json.Unmarshal(b, &f); e != nil {
		t.Fatal(e)
	}
	data, err := base58.Decode(f.Data)
	if err != nil {
		t.Fatal(err)
	}
	invoke := func(d []byte) DexEvent {
		return ParseMeteoraPoolsInstruction(d, f.Accounts, "simulation", 0, 0, nil, 0)
	}
	event := invoke(data)
	if event.Data == nil {
		t.Fatal("missing event")
	}
	body, _ := json.Marshal(event.Data)
	var fields map[string]any
	json.Unmarshal(body, &fields)
	for k, v := range f.Expected {
		if k != "type" && fmt.Sprint(fields[k]) != v {
			t.Fatal(k, fields[k], v)
		}
	}
	if invoke(data[:24]).Data != nil {
		t.Fatal("truncated option")
	}
	bad := append([]byte(nil), data...)
	bad[24] = 2
	if invoke(bad).Data != nil {
		t.Fatal("invalid option")
	}
	bad[24] = 1
	if invoke(bad[:32]).Data != nil {
		t.Fatal("truncated Some")
	}
	obsolete := append([]byte{95, 180, 10, 172, 84, 174, 232, 40}, make([]byte, 49)...)
	if invoke(obsolete).Data != nil {
		t.Fatal("obsolete discriminator")
	}
}
