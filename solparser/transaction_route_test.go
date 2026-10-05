package solparser

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestNativeRouteRust077Golden(t *testing.T) {
	data, err := os.ReadFile("testdata/stonkfun_routes_0_7_7.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name        string
			Transaction json.RawMessage
			Expected    json.RawMessage
		}
	}
	if err = json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, c := range corpus.Cases {
		t.Run(c.Name, func(t *testing.T) {
			route, err := AnalyzeRPCTransactionRoutes(c.Transaction, nil)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := json.Marshal(route)
			if err != nil {
				t.Fatal(err)
			}
			var want, got any
			json.Unmarshal(c.Expected, &want)
			json.Unmarshal(actual, &got)
			if !reflect.DeepEqual(want, got) {
				t.Errorf("Rust route mismatch\nwant %s\ngot %s", c.Expected, actual)
			}
			var tx struct {
				Transaction []string
				Result      *struct{ Transaction []string }
			}
			json.Unmarshal(c.Transaction, &tx)
			encoded := tx.Transaction
			if tx.Result != nil {
				encoded = tx.Result.Transaction
			}
			raw, err := base64.StdEncoding.DecodeString(encoded[0])
			if err != nil {
				t.Fatal(err)
			}
			decoded, n, err := DecodeWireTransaction(raw, 0, true)
			if err != nil {
				t.Fatal(err)
			}
			if n != len(raw) || decoded.Signatures[0] != route.Signature {
				t.Fatal("wire identity mismatch")
			}
			if _, _, err = DecodeWireTransaction(raw[:len(raw)-1], 0, true); err == nil {
				t.Fatal("accepted truncated transaction")
			}
			if _, _, err = DecodeWireTransaction(append(raw, 0), 0, true); err == nil {
				t.Fatal("accepted trailing bytes")
			}
		})
	}
}
