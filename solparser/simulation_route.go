package solparser

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/mr-tron/base58"
)

// AnalyzeSimulationRoutes combines original wire with explicit simulation evidence.
// It performs no RPC. V0 requires caller-resolved result.value.loadedAddresses.
func AnalyzeSimulationRoutes(wire, response []byte, graduatedPools []string) (*TransactionRoute, error) {
	tx, _, err := DecodeWireTransaction(wire, 0, true)
	if err != nil {
		return nil, err
	}
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader(response))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return nil, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("trailing simulation response data")
	}
	result, _ := root["result"].(map[string]any)
	value, _ := result["value"].(map[string]any)
	_, hasErr := value["err"]
	inner, hasInner := value["innerInstructions"]
	if root["error"] != nil || value == nil || !hasErr || !hasInner {
		return nil, fmt.Errorf("simulation response missing execution metadata")
	}
	keys := append([]string{}, tx.Message.AccountKeys...)
	loaded, supplied := value["loadedAddresses"]
	if (!supplied || loaded == nil) && len(tx.Message.AddressTableLookups) != 0 {
		return nil, fmt.Errorf("V0 ALT addresses unavailable; provide loadedAddresses")
	}
	if supplied && loaded != nil {
		addresses, ok := loaded.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid simulation loaded addresses")
		}
		for _, side := range []string{"writable", "readonly"} {
			expected := 0
			for _, lookup := range tx.Message.AddressTableLookups {
				if side == "writable" {
					expected += len(lookup.WritableIndexes)
				} else {
					expected += len(lookup.ReadonlyIndexes)
				}
			}
			items, ok := addresses[side].([]any)
			if !ok || len(items) != expected {
				return nil, fmt.Errorf("simulation loaded address count does not match wire lookups")
			}
			for _, item := range items {
				key, ok := item.(string)
				bytes, err := base58.Decode(key)
				if !ok || err != nil || len(bytes) != 32 {
					return nil, fmt.Errorf("invalid simulation loaded public key")
				}
				keys = append(keys, key)
			}
		}
	}
	index := func(v any) (int, error) {
		key, ok := v.(string)
		if ok {
			for i, k := range keys {
				if k == key {
					return i, nil
				}
			}
		}
		return 0, fmt.Errorf("simulation account absent from transaction")
	}
	validIndex := func(v any) (int, error) {
		n, err := simulationIndex(v)
		if err != nil || n >= uint64(len(keys)) {
			return 0, fmt.Errorf("invalid simulation account index")
		}
		return int(n), nil
	}
	groups := []any{}
	if inner != nil {
		items, ok := inner.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid simulation inner instructions")
		}
		seen := map[uint64]bool{}
		for _, item := range items {
			group, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid simulation instruction group")
			}
			n, e := simulationIndex(group["index"])
			instructions, ok := group["instructions"].([]any)
			if e != nil || n >= uint64(len(tx.Message.Instructions)) || seen[n] || !ok {
				return nil, fmt.Errorf("invalid simulation instruction group")
			}
			seen[n] = true
			converted := []any{}
			for _, raw := range instructions {
				ix, ok := raw.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid simulation instruction")
				}
				if ix["stackHeight"] != nil {
					h, e := simulationIndex(ix["stackHeight"])
					if e != nil || h < 2 {
						return nil, fmt.Errorf("invalid simulation stack height")
					}
				}
				pid := 0
				accountIndex := index
				if p, exists := ix["programIdIndex"]; exists {
					pid, e = validIndex(p)
					accountIndex = validIndex
					if _, exists := ix["parsed"]; exists {
						return nil, fmt.Errorf("invalid compiled simulation instruction")
					}
				} else {
					pid, e = index(ix["programId"])
				}
				if e != nil {
					return nil, e
				}
				var accounts []any
				var data string
				if parsedRaw, exists := ix["parsed"]; exists {
					parsed, _ := parsedRaw.(map[string]any)
					info, _ := parsed["info"].(map[string]any)
					if info == nil {
						return nil, fmt.Errorf("unsupported parsed simulation instruction")
					}
					setupAccounts, setupData, err := simulationAccountSetup(keys[pid], parsed["type"], info)
					if err != nil {
						return nil, err
					}
					if setupData != nil {
						indices := []int{}
						for _, a := range setupAccounts {
							i, err := index(a)
							if err != nil {
								return nil, err
							}
							indices = append(indices, i)
						}
						converted = append(converted, map[string]any{"programIdIndex": pid, "accounts": indices, "data": base58.Encode(setupData), "stackHeight": ix["stackHeight"]})
						continue
					}
					if keys[pid] != routeToken && keys[pid] != routeToken2022 {
						return nil, fmt.Errorf("unsupported parsed simulation program")
					}
					authority := info["authority"]
					if authority == nil {
						authority = info["multisigAuthority"]
					}
					signers := []any{}
					if s, exists := info["signers"]; exists {
						var ok bool
						signers, ok = s.([]any)
						if !ok {
							return nil, fmt.Errorf("invalid simulation signers")
						}
					}
					var encoded []byte
					switch parsed["type"] {
					case "transfer":
						amount, err := simulationUnsigned(info["amount"], 64)
						if err != nil {
							return nil, err
						}
						encoded = make([]byte, 9)
						encoded[0] = 3
						binary.LittleEndian.PutUint64(encoded[1:], amount)
						accounts = []any{info["source"], info["destination"], authority}
					case "transferChecked", "transferCheckedWithFee":
						tokenAmount, _ := info["tokenAmount"].(map[string]any)
						amount, err := simulationUnsigned(tokenAmount["amount"], 64)
						if err != nil {
							return nil, err
						}
						decimals, err := simulationUnsigned(tokenAmount["decimals"], 8)
						if err != nil {
							return nil, err
						}
						if parsed["type"] == "transferCheckedWithFee" {
							if keys[pid] != routeToken2022 {
								return nil, fmt.Errorf("transfer fee requires Token-2022")
							}
							fee, err := simulationUnsigned(info["feeAmount"], 64)
							if err != nil {
								return nil, err
							}
							encoded = make([]byte, 19)
							encoded[0], encoded[1], encoded[10] = 26, 1, byte(decimals)
							binary.LittleEndian.PutUint64(encoded[2:], amount)
							binary.LittleEndian.PutUint64(encoded[11:], fee)
						} else {
							encoded = make([]byte, 10)
							encoded[0], encoded[9] = 12, byte(decimals)
							binary.LittleEndian.PutUint64(encoded[1:], amount)
						}
						accounts = []any{info["source"], info["mint"], info["destination"], authority}
					default:
						return nil, fmt.Errorf("unsupported parsed simulation instruction")
					}
					accounts = append(accounts, signers...)
					data = base58.Encode(encoded)
				} else {
					var ok bool
					accounts, ok = ix["accounts"].([]any)
					if !ok {
						return nil, fmt.Errorf("invalid raw simulation instruction")
					}
					data, ok = ix["data"].(string)
					if !ok {
						return nil, fmt.Errorf("invalid raw simulation instruction")
					}
					if _, err := base58.Decode(data); err != nil {
						return nil, err
					}
				}
				indices := []int{}
				for _, a := range accounts {
					i, err := accountIndex(a)
					if err != nil {
						return nil, err
					}
					indices = append(indices, i)
				}
				converted = append(converted, map[string]any{"programIdIndex": pid, "accounts": indices, "data": data, "stackHeight": ix["stackHeight"]})
			}
			groups = append(groups, map[string]any{"index": n, "instructions": converted})
		}
	}
	value["innerInstructions"] = groups
	normalized, err := json.Marshal(map[string]any{"transaction": tx, "meta": value})
	if err != nil {
		return nil, err
	}
	return AnalyzeRPCTransactionRoutes(normalized, graduatedPools)
}

// Decode known ATA setup CPI without dropping positions or inventing transfers.
func simulationAccountSetup(program string, kind any, info map[string]any) ([]any, []byte, error) {
	pubkey := func(v any) ([]byte, error) {
		key, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid simulation public key")
		}
		data, err := base58.Decode(key)
		if err != nil || len(data) != 32 {
			return nil, fmt.Errorf("invalid simulation public key")
		}
		return data, nil
	}
	if program == zeroPubkey && kind == "createAccount" {
		lamports, err := simulationUnsigned(info["lamports"], 64)
		if err != nil {
			return nil, nil, err
		}
		space, err := simulationUnsigned(info["space"], 64)
		if err != nil {
			return nil, nil, err
		}
		owner, err := pubkey(info["owner"])
		if err != nil {
			return nil, nil, err
		}
		data := make([]byte, 52)
		binary.LittleEndian.PutUint64(data[4:], lamports)
		binary.LittleEndian.PutUint64(data[12:], space)
		copy(data[20:], owner)
		return []any{info["source"], info["newAccount"]}, data, nil
	}
	if program == zeroPubkey && (kind == "transfer" || kind == "allocate") {
		field, tag, accounts := "lamports", uint32(2), []any{info["source"], info["destination"]}
		if kind == "allocate" {
			field, tag, accounts = "space", 8, []any{info["account"]}
		}
		n, err := simulationUnsigned(info[field], 64)
		if err != nil {
			return nil, nil, err
		}
		data := make([]byte, 12)
		binary.LittleEndian.PutUint32(data, tag)
		binary.LittleEndian.PutUint64(data[4:], n)
		return accounts, data, nil
	}
	if program == zeroPubkey && kind == "assign" {
		owner, err := pubkey(info["owner"])
		if err != nil {
			return nil, nil, err
		}
		data := make([]byte, 36)
		binary.LittleEndian.PutUint32(data, 1)
		copy(data[4:], owner)
		return []any{info["account"]}, data, nil
	}
	if program != routeToken && program != routeToken2022 {
		return nil, nil, nil
	}
	switch kind {
	case "mintTo", "mintToChecked", "burn", "burnChecked":
		minting := kind == "mintTo" || kind == "mintToChecked"
		checked := kind == "mintToChecked" || kind == "burnChecked"
		authorityKey, multisigKey := "authority", "multisigAuthority"
		if minting {
			authorityKey, multisigKey = "mintAuthority", "multisigMintAuthority"
		}
		authority := info[authorityKey]
		if authority == nil {
			authority = info[multisigKey]
		}
		signers := []any{}
		if raw, exists := info["signers"]; exists {
			var ok bool
			signers, ok = raw.([]any)
			if !ok {
				return nil, nil, fmt.Errorf("invalid simulation signers")
			}
		}
		amountInfo := info
		if checked {
			var ok bool
			amountInfo, ok = info["tokenAmount"].(map[string]any)
			if !ok {
				return nil, nil, fmt.Errorf("invalid simulation token amount")
			}
		}
		amount, err := simulationUnsigned(amountInfo["amount"], 64)
		if err != nil {
			return nil, nil, err
		}
		data := make([]byte, 9)
		if minting {
			data[0] = 7
		} else {
			data[0] = 8
		}
		binary.LittleEndian.PutUint64(data[1:], amount)
		if checked {
			decimals, err := simulationUnsigned(amountInfo["decimals"], 8)
			if err != nil {
				return nil, nil, err
			}
			data[0] += 7
			data = append(data, byte(decimals))
		}
		accounts := []any{info["account"], info["mint"], authority}
		if minting {
			accounts[0], accounts[1] = accounts[1], accounts[0]
		}
		return append(accounts, signers...), data, nil
	case "initializeImmutableOwner":
		return []any{info["account"]}, []byte{22}, nil
	case "initializeAccount3":
		owner, err := pubkey(info["owner"])
		if err != nil {
			return nil, nil, err
		}
		return []any{info["account"], info["mint"]}, append([]byte{18}, owner...), nil
	case "getAccountDataSize":
		extensions, ok := info["extensionTypes"].([]any)
		if !ok {
			return nil, nil, fmt.Errorf("invalid simulation extension types")
		}
		data := make([]byte, 1+len(extensions)*2)
		data[0] = 21
		for i, extension := range extensions {
			if extension != "immutableOwner" {
				return nil, nil, fmt.Errorf("unsupported simulation extension type")
			}
			binary.LittleEndian.PutUint16(data[1+2*i:], 7)
		}
		return []any{info["mint"]}, data, nil
	}
	return nil, nil, nil
}

func simulationUnsigned(v any, bits int) (uint64, error) {
	var s string
	switch n := v.(type) {
	case json.Number:
		s = string(n)
	case string:
		s = n
	default:
		return 0, fmt.Errorf("invalid simulation integer")
	}
	if s == "" {
		return 0, fmt.Errorf("invalid simulation integer")
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid simulation integer")
		}
	}
	n, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		return 0, fmt.Errorf("simulation integer outside range")
	}
	return n, nil
}

func simulationIndex(v any) (uint64, error) {
	if _, ok := v.(json.Number); !ok {
		return 0, fmt.Errorf("invalid simulation index")
	}
	return simulationUnsigned(v, 32)
}
