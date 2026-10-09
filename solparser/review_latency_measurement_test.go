package solparser

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestMeasureBankRouteLatency(t *testing.T) {
	output := os.Getenv("SDK_LATENCY_OUTPUT")
	if output == "" {
		t.Skip("set SDK_LATENCY_OUTPUT for an offline latency measurement")
	}
	data, err := os.ReadFile("testdata/signed_identical_alt_bank_20261009.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name, Wire string
			RPC        struct{ Meta map[string]any }
		}
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	paths := []map[string]any{}
	for _, c := range corpus.Cases[:2] {
		wire, err := base64.StdEncoding.DecodeString(c.Wire)
		if err != nil {
			t.Fatal(err)
		}
		response, err := json.Marshal(map[string]any{"result": map[string]any{"value": c.RPC.Meta}})
		if err != nil {
			t.Fatal(err)
		}
		for _, routeMode := range []bool{false, true} {
			operation := func() {
				if routeMode {
					route, err := AnalyzeSimulationRoutes(wire, response, nil)
					if err != nil || !route.Succeeded || len(route.Legs) != 2 {
						t.Fatalf("route failed: %v", err)
					}
				} else {
					tx, size, err := DecodeWireTransaction(wire, 0, true)
					if err != nil || size != len(wire) || len(tx.Signatures) != 2 {
						t.Fatalf("decode failed: %v", err)
					}
				}
			}
			rounds := []map[string]any{}
			for round := 0; round < 3; round++ {
				for i := 0; i < 200; i++ {
					operation()
				}
				samples := make([]float64, 1000)
				for i := range samples {
					start := time.Now()
					operation()
					samples[i] = float64(time.Since(start).Nanoseconds()) / 1000
				}
				sort.Float64s(samples)
				rounds = append(rounds, map[string]any{"p50": samples[499], "p95": samples[949], "p99": samples[989], "maximum": samples[999], "samples_us": samples})
			}
			label := "Go Parser decode "
			if routeMode {
				label = "Go Parser decode + response JSON + route "
			}
			paths = append(paths, map[string]any{"name": label + c.Name, "unit": "microseconds", "warmup_per_round": 200, "measured_per_round": 1000, "rounds": rounds})
		}
	}
	result := map[string]any{"paths": paths, "measured_rpc_calls": 0, "environment": map[string]any{"go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH, "gomaxprocs": runtime.GOMAXPROCS(0)}, "scope": "Warm actual SDK wire decode + route, metadata JSON decode included for route API; signature verification/transport/bank/landing excluded; fixture setup outside timing."}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, encoded, 0644); err != nil {
		t.Fatal(err)
	}
}
