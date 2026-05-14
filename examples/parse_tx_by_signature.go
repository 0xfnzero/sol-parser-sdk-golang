//go:build ignore

// Parse a PumpFun or PumpSwap transaction from RPC by signature
//
// Fetches a transaction from Solana RPC and parses it using the sol-parser-sdk.
//
// Usage:
//   TX_SIGNATURE=<sig> go run examples/parse_tx_by_signature.go
//   RPC_URL=https://api.mainnet-beta.solana.com TX_SIGNATURE=<sig> go run examples/parse_tx_by_signature.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	solparser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
)

const defaultRPCURL = "https://api.mainnet-beta.solana.com"
const defaultSignature = "3zsihbygW7hoKGtduAyDDFzp4E1eis8gaBzEzzNKr8ma39baffpFcphok9wHFgR3EauDe9vYYsVf4Puh5pZ6UJiS"

type rpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResponse struct {
	Result *txResult `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type txResult struct {
	Slot int64   `json:"slot"`
	Meta *txMeta `json:"meta"`
}

type txMeta struct {
	LogMessages []string `json:"logMessages"`
}

func fetchTransaction(rpcURL, sig string) (*txResult, error) {
	req := rpcRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getTransaction",
		Params: []interface{}{
			sig,
			map[string]interface{}{
				"encoding":                       "jsonParsed",
				"maxSupportedTransactionVersion": 0,
			},
		},
	}
	body, _ := json.Marshal(req)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result rpcResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("RPC error: %s", result.Error.Message)
	}
	return result.Result, nil
}

func main() {
	sig := os.Getenv("TX_SIGNATURE")
	if sig == "" {
		sig = defaultSignature
	}
	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = defaultRPCURL
	}

	fmt.Println("=== Transaction Parser ===\n")
	fmt.Printf("Signature: %s\n", sig)
	fmt.Printf("RPC URL  : %s\n\n", rpcURL)

	fmt.Println("Fetching transaction from RPC...")
	tx, err := fetchTransaction(rpcURL, sig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to fetch: %v\n", err)
		os.Exit(1)
	}
	if tx == nil || tx.Meta == nil {
		fmt.Fprintln(os.Stderr, "Transaction not found. It may be too old or pruned.")
		fmt.Fprintln(os.Stderr, "Use an archive RPC (e.g. Helius, QuickNode) or set RPC_URL.")
		os.Exit(1)
	}

	logs := tx.Meta.LogMessages
	fmt.Printf("Log messages: %d\n\n", len(logs))

	events := solparser.ParseLogsOnly(logs, sig, uint64(tx.Slot), nil)

	if len(events) == 0 {
		fmt.Println("No DEX events found in this transaction.")
		fmt.Println("Try a PumpFun/PumpSwap/Raydium/Orca transaction signature.")
		return
	}

	fmt.Printf("✅ Found %d DEX event(s):\n\n", len(events))
	for i, ev := range events {
		b, _ := json.MarshalIndent(ev, "", "  ")
		fmt.Printf("Event #%d:\n%s\n\n", i+1, string(b))
	}

	fmt.Println("=== Summary ===")
	fmt.Println("✅ sol-parser-sdk successfully parsed the transaction!")
	fmt.Println("   - Direct parsing from RPC (no gRPC streaming needed)")
	fmt.Println("   - All 10 DEX protocols supported")
}
