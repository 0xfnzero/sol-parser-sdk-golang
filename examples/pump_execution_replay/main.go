// Offline captured execution replay. Optional signature fetch uses SOLANA_RPC_URL.
// Inspect third_program and parse_status. Execution metadata is unavailable to
// Shred; one outer wrapper can produce several fills with different event counts.
package main

import (
	"bytes"
	"encoding/base64"
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

type replayCase struct {
	Raw     json.RawMessage
	WireRPC json.RawMessage `json:"wire_rpc"`
}

func run() error {
	var fixture struct {
		Cases []replayCase
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
		wireBody, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 2, "method": "getTransaction", "params": []any{os.Args[1], map[string]any{"encoding": "base64", "maxSupportedTransactionVersion": 1}}})
		wireRes, err := client.Post(url, "application/json", bytes.NewReader(wireBody))
		if err != nil {
			return err
		}
		defer wireRes.Body.Close()
		var wireResponse struct {
			Result json.RawMessage
			Error  any
		}
		if err = json.NewDecoder(wireRes.Body).Decode(&wireResponse); err != nil {
			return err
		}
		if wireResponse.Error != nil || len(wireResponse.Result) == 0 || string(wireResponse.Result) == "null" {
			return fmt.Errorf("wire transaction unavailable")
		}
		fixture.Cases = append(fixture.Cases, replayCase{Raw: response.Result, WireRPC: wireResponse.Result})
	} else {
		for _, name := range []string{"pump_create_alt_20261026.json", "pump_multi_hop_20261025.json", "wrapped_pumpfun_reported.json"} {
			b, err := os.ReadFile("solparser/testdata/" + name)
			if err != nil {
				return err
			}
			var part struct{ Cases []replayCase }
			if err = json.Unmarshal(b, &part); err != nil {
				return err
			}
			fixture.Cases = append(fixture.Cases, part.Cases...)
		}
	}
	for _, c := range fixture.Cases {
		tx, err := decode(c.Raw)
		if err != nil {
			return err
		}
		if len(c.WireRPC) > 0 {
			var captured struct{ Transaction []json.RawMessage }
			if err = json.Unmarshal(c.WireRPC, &captured); err != nil {
				return err
			}
			var encoded string
			if err = json.Unmarshal(captured.Transaction[0], &encoded); err != nil {
				return err
			}
			wire, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return err
			}
			filter := &sdk.IncludeOnlyFilter{IncludeOnly: []sdk.EventType{sdk.EventTypePumpFunCreate, sdk.EventTypePumpFunTrade}}
			static := sdk.DexEventsFromShredTransactionWire(wire, "", tx.Slot, 888, nil, 123456, filter)
			loaded := tx.Meta.LoadedAddresses
			full, err := sdk.DexEventsFromShredTransactionWireWithLoadedAddresses(wire, loaded.Writable, loaded.Readonly, "", tx.Slot, 888, nil, 123456, filter)
			if err != nil {
				return err
			}
			for _, v := range []struct {
				Mode   string
				Events []sdk.DexEvent
			}{{"static_keys", static}, {"resolved_alt", full}} {
				out, err := json.Marshal(map[string]any{"source": "shred", "mode": v.Mode, "signature": tx.Transaction.Signatures[0], "events": v.Events})
				if err != nil {
					return err
				}
				fmt.Println(string(out))
			}
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
