package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestDlmmRewardsBank(t *testing.T) {
	b, err := os.ReadFile("testdata/dlmm_rewards_20261008.json")
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
	if len(f.Cases) != 7 {
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
				if e.Type != EventTypeMeteoraDlmmClaimReward {
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

func TestDlmmRewardClaimLogPrecisionAndBoundaries(t *testing.T) {
	payload := make([]byte, 116)
	for i := 0; i < 96; i++ {
		payload[i] = byte(i + 1)
	}
	binary.LittleEndian.PutUint64(payload[96:104], 1)
	binary.LittleEndian.PutUint64(payload[104:112], ^uint64(0))
	disc := []byte{27, 143, 244, 33, 80, 43, 110, 146}
	body := append(append([]byte{}, disc...), payload...)
	for n := 0; n < len(body); n++ {
		log := "Program data: " + base64.StdEncoding.EncodeToString(body[:n])
		if e := ParseMeteoraDlmmLog(log, "simulation", 0, 0, nil, 0); e.Type != "" {
			t.Fatalf("truncated %d", n)
		}
	}
	log := "Program data: " + base64.StdEncoding.EncodeToString(body)
	filter := EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDlmmClaimReward})
	ev := ParseLogOptimizedWithProgramID(log, "simulation", 0, 0, nil, 0, filter, false, "", METEORA_DLMM_PROGRAM_ID)
	data, ok := ev.Data.(*MeteoraDlmmClaimRewardEvent)
	if !ok || data.RewardIndex != 1 || data.TotalReward != ^uint64(0) {
		t.Fatal("u64 precision", ev)
	}
	if ev := ParseLogOptimized(log, "simulation", 0, 0, nil, 0, filter, false, ""); ev.Type != EventTypeMeteoraDlmmClaimReward {
		t.Fatal("unscoped", ev)
	}
	exclude := EventTypeFilterExclude([]EventType{EventTypeMeteoraDlmmClaimReward})
	if ev := ParseLogOptimizedWithProgramID(log, "simulation", 0, 0, nil, 0, exclude, false, "", METEORA_DLMM_PROGRAM_ID); ev.Type != "" {
		t.Fatal("exclude", ev)
	}
}
