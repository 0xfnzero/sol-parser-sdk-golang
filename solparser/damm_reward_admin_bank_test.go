package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestDammRewardAdminBank(t *testing.T) {
	b, err := os.ReadFile("testdata/damm_reward_admin_20261009.json")
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
	if len(f.Cases) != 37 {
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
				if e.Type != EventTypeMeteoraDammV2WithdrawIneligibleReward && e.Type != EventTypeMeteoraDammV2WithdrawDeadLiquidityReward && e.Type != EventTypeMeteoraDammV2FundReward && e.Type != EventTypeMeteoraDammV2InitializeReward && e.Type != EventTypeMeteoraDammV2UpdateRewardDuration && e.Type != EventTypeMeteoraDammV2UpdateRewardFunder {
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

func TestDammRewardAdministrationLogsAndTruncation(t *testing.T) {
	b, err := os.ReadFile("testdata/damm_reward_admin_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Cases []struct {
			Name string
			Logs []struct {
				Log      string
				Bytes    []byte
				Expected map[string]string
			}
		}
	}
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	for _, c := range f.Cases {
		t.Run(c.Name, func(t *testing.T) {
			for _, item := range c.Logs {
				e := ParseMeteoraDammLog(item.Log, "simulation", 1, 0, nil, 0)
				if string(e.Type) != item.Expected["type"] {
					t.Fatal("missing log event", e.Type)
				}
				encoded, err := json.Marshal(e.Data)
				if err != nil {
					t.Fatal(err)
				}
				decoder := json.NewDecoder(bytes.NewReader(encoded))
				decoder.UseNumber()
				var fields map[string]any
				if err = decoder.Decode(&fields); err != nil {
					t.Fatal(err)
				}
				for k, v := range item.Expected {
					if k != "type" && fmt.Sprint(fields[k]) != v {
						t.Fatal(k, fields[k], v)
					}
				}
				for end := 0; end < len(item.Bytes); end++ {
					log := "Program data: " + base64.StdEncoding.EncodeToString(item.Bytes[:end])
					if ParseMeteoraDammLog(log, "simulation", 1, 0, nil, 0).Data != nil {
						t.Fatal("truncated event retained", end)
					}
				}
			}
		})
	}
}
