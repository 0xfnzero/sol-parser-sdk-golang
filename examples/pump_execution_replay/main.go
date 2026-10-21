// Offline captured execution replay. Optional signature fetch uses SOLANA_RPC_URL.
// Inspect third_program and parse_status. Execution metadata is unavailable to
// Shred; one outer wrapper can produce several fills with different event counts.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"github.com/mr-tron/base58"
	"io"
	"net/http"
	"os"
	"time"
)

// RPC instruction data is base58; the SDK's []byte JSON representation is base64.
// UseNumber preserves u64 token balances while translating instruction bytes.
func decode(raw json.RawMessage) (*sdk.RpcTransactionResponse, error) {
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&value); err != nil {
		return nil, err
	}
	var walk func(any) error
	walk = func(v any) error {
		switch v := v.(type) {
		case map[string]any:
			if data, ok := v["data"].(string); ok {
				decoded, err := base58.Decode(data)
				if err != nil {
					return err
				}
				v["data"] = decoded
			}
			for _, child := range v {
				if err := walk(child); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range v {
				if err := walk(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(value); err != nil {
		return nil, err
	}
	b, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var tx sdk.RpcTransactionResponse
	err = json.Unmarshal(b, &tx)
	return &tx, err
}

func run() error {
	var fixture struct {
		Cases []struct{ Raw json.RawMessage }
	}
	if len(os.Args) > 1 {
		url := os.Getenv("SOLANA_RPC_URL")
		if url == "" {
			url = "https://api.mainnet-beta.solana.com"
		}
		body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "getTransaction", "params": []any{os.Args[1], map[string]any{"encoding": "json", "maxSupportedTransactionVersion": 1}}})
		client := http.Client{Timeout: 30 * time.Second}
		res, err := client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		var response struct {
			Result json.RawMessage
			Error  any
		}
		if err = json.Unmarshal(b, &response); err != nil {
			return err
		}
		if response.Error != nil || len(response.Result) == 0 || string(response.Result) == "null" {
			return fmt.Errorf("RPC transaction unavailable: %s", b)
		}
		fixture.Cases = append(fixture.Cases, struct{ Raw json.RawMessage }{response.Result})
	} else {
		b, err := os.ReadFile("solparser/testdata/pump_execution_20261019.json")
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, &fixture); err != nil {
			return err
		}
		b, err = os.ReadFile("solparser/testdata/wrapped_pumpfun_reported.json")
		if err != nil {
			return err
		}
		var reported struct {
			Cases []struct{ Raw json.RawMessage }
		}
		if err = json.Unmarshal(b, &reported); err != nil {
			return err
		}
		fixture.Cases = append(fixture.Cases, reported.Cases...)
	}
	for _, c := range fixture.Cases {
		tx, err := decode(c.Raw)
		if err != nil {
			return err
		}
		logs := tx.Meta.LogMessages
		for _, mode := range []string{"full", "no_logs", "truncated_logs"} {
			tx.Meta.LogMessages = logs
			if mode == "no_logs" {
				tx.Meta.LogMessages = nil
			}
			if mode == "truncated_logs" {
				tx.Meta.LogMessages = logs[:len(logs)/2]
			}
			meta, transaction, err := sdk.ConvertRpcToGrpc(tx)
			if err != nil {
				return err
			}
			events, parseErr := sdk.ParseSubscribeTransaction(tx.Slot, &sdk.SubscribeUpdateTransactionInfo{Transaction: transaction, Meta: meta}, nil, 0)
			if parseErr != nil {
				return parseErr
			}
			out, err := json.Marshal(map[string]any{"source": "execution_replay", "mode": mode, "event_count": len(events), "slot": tx.Slot, "signature": tx.Transaction.Signatures[0], "events": events})
			if err != nil {
				return err
			}
			fmt.Println(string(out))
		}
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
