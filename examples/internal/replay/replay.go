// Package replay loads saved compiled/base64 transactions without network access.
package replay

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	sdk "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"github.com/mr-tron/base58"
	"os"
)

type Case struct {
	Name      string          `json:"name"`
	Signature string          `json:"signature"`
	RPC       json.RawMessage `json:"rpc"`
	Raw       json.RawMessage `json:"raw"`
	WireRPC   json.RawMessage `json:"wire_rpc"`
}

func Load(defaultPath string) ([]Case, error) {
	path := flag.String("fixture", defaultPath, "saved transaction or cases capture")
	flag.Parse()
	if flag.NArg() != 0 {
		return nil, errors.New("unexpected positional arguments")
	}
	raw, err := os.ReadFile(*path)
	if err != nil {
		return nil, err
	}
	var document struct {
		Cases  []Case          `json:"cases"`
		Result json.RawMessage `json:"result"`
	}
	if err = json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	if document.Cases == nil {
		if document.Result != nil {
			raw = document.Result
		}
		document.Cases = []Case{{RPC: raw}}
	}
	if len(document.Cases) == 0 {
		return nil, errors.New("empty replay")
	}
	for i := range document.Cases {
		c := &document.Cases[i]
		if c.RPC == nil {
			c.RPC = c.WireRPC
		}
		if c.RPC == nil {
			c.RPC = c.Raw
		}
		if c.RPC == nil {
			return nil, errors.New("missing transaction")
		}
	}
	return document.Cases, nil
}

// Decode retains raw JSON integer literals; only instruction bytes are transcoded.
func Decode(raw json.RawMessage) (*sdk.RpcTransactionResponse, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	if result := root["result"]; result != nil {
		return Decode(result)
	}
	body := root["transaction"]
	if len(body) == 0 {
		return nil, errors.New("missing transaction")
	}
	if body[0] == '[' {
		var encoded []string
		if err := json.Unmarshal(body, &encoded); err != nil {
			return nil, err
		}
		if len(encoded) != 2 || encoded[1] != "base64" {
			return nil, errors.New("expected base64 transaction")
		}
		data, err := base64.StdEncoding.DecodeString(encoded[0])
		if err != nil {
			return nil, err
		}
		tx, _, err := sdk.DecodeWireTransaction(data, 0, true)
		if err != nil {
			return nil, err
		}
		body, err = json.Marshal(tx)
		if err != nil {
			return nil, err
		}
	}
	var tx map[string]json.RawMessage
	if err := json.Unmarshal(body, &tx); err != nil {
		return nil, err
	}
	var message map[string]json.RawMessage
	if err := json.Unmarshal(tx["message"], &message); err != nil {
		return nil, err
	}
	convert := func(raw json.RawMessage) (json.RawMessage, error) {
		var instructions []map[string]json.RawMessage
		if err := json.Unmarshal(raw, &instructions); err != nil {
			return nil, err
		}
		for _, ix := range instructions {
			var encoded string
			if err := json.Unmarshal(ix["data"], &encoded); err != nil {
				return nil, err
			}
			data, err := base58.Decode(encoded)
			if err != nil {
				return nil, err
			}
			ix["data"], err = json.Marshal(data)
			if err != nil {
				return nil, err
			}
		}
		return json.Marshal(instructions)
	}
	var err error
	message["instructions"], err = convert(message["instructions"])
	if err != nil {
		return nil, err
	}
	tx["message"], err = json.Marshal(message)
	if err != nil {
		return nil, err
	}
	root["transaction"], err = json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	var meta map[string]json.RawMessage
	if err = json.Unmarshal(root["meta"], &meta); err != nil {
		return nil, err
	}
	var groups []map[string]json.RawMessage
	if len(meta["innerInstructions"]) > 0 {
		if err = json.Unmarshal(meta["innerInstructions"], &groups); err != nil {
			return nil, err
		}
	}
	for _, group := range groups {
		group["instructions"], err = convert(group["instructions"])
		if err != nil {
			return nil, err
		}
	}
	meta["innerInstructions"], err = json.Marshal(groups)
	if err != nil {
		return nil, err
	}
	root["meta"], err = json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	normalized, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	var result sdk.RpcTransactionResponse
	if err = json.Unmarshal(normalized, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
