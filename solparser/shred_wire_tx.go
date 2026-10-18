package solparser

import (
	"fmt"
	"github.com/mr-tron/base58"
)

var shredUnknownProgramCandidates = []string{
	PUMPFUN_PROGRAM_ID,
	PUMPSWAP_PROGRAM_ID,
	PUMP_FEES_PROGRAM_ID,
	RAYDIUM_LAUNCHLAB_PROGRAM_ID,
	RAYDIUM_CPMM_PROGRAM_ID,
	RAYDIUM_CLMM_PROGRAM_ID,
	RAYDIUM_AMM_V4_PROGRAM_ID,
	ORCA_WHIRLPOOL_PROGRAM_ID,
	METEORA_POOLS_PROGRAM_ID,
	METEORA_DAMM_V2_PROGRAM_ID,
	METEORA_DLMM_PROGRAM_ID,
}

// DexEventsFromShredTransactionWire parses only fully resolved static instructions. No RPC.
func DexEventsFromShredTransactionWire(raw []byte, signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64, filter EventTypeFilter) []DexEvent {
	var binaryData [][]byte
	tx, _, err := decodeWireTransaction(raw, 0, true, &binaryData)
	if err != nil {
		return nil
	}
	return parseShredKeys(tx, tx.Message.AccountKeys, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, false, binaryData)
}

// Explicit legacy compatibility; route adapters still require all keys.
func DexEventsFromShredTransactionWireBestEffort(raw []byte, signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64, filter EventTypeFilter) []DexEvent {
	var binaryData [][]byte
	tx, _, err := decodeWireTransaction(raw, 0, true, &binaryData)
	if err != nil {
		return nil
	}
	return parseShredKeys(tx, tx.Message.AccountKeys, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, true, binaryData)
}

func DexEventsFromShredTransactionWireWithLoadedAddresses(raw []byte, writable, readonly []string, signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64, filter EventTypeFilter) ([]DexEvent, error) {
	var binaryData [][]byte
	tx, _, err := decodeWireTransaction(raw, 0, true, &binaryData)
	if err != nil {
		return nil, err
	}
	return parseShredLoaded(tx, writable, readonly, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, binaryData)
}

// ShredAddressLookupResolver must read a slot-aware prewarmed cache, never RPC.
type ShredAddressLookupResolver func(*NativeWireTransaction, uint64) ([]string, []string, error)

func DexEventsFromShredTransactionWireWithResolver(raw []byte, signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64, filter EventTypeFilter, resolver ShredAddressLookupResolver) ([]DexEvent, error) {
	tx, _, err := DecodeWireTransaction(raw, 0, true)
	if err != nil {
		return nil, err
	}
	if resolver == nil || len(tx.Message.AddressTableLookups) == 0 {
		return parseShredKeys(tx, tx.Message.AccountKeys, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, false, nil), nil
	}
	w, r, err := resolver(tx, slot)
	if err != nil {
		return nil, err
	}
	return parseShredLoaded(tx, w, r, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, nil)
}
func parseShredLoaded(tx *NativeWireTransaction, writable, readonly []string, signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64, filter EventTypeFilter, binaryData [][]byte) ([]DexEvent, error) {
	nw, nr := 0, 0
	for _, l := range tx.Message.AddressTableLookups {
		nw += len(l.WritableIndexes)
		nr += len(l.ReadonlyIndexes)
	}
	if nw != len(writable) || nr != len(readonly) {
		return nil, fmt.Errorf("ALT address counts do not match transaction")
	}
	keys := append(append(append([]string(nil), tx.Message.AccountKeys...), writable...), readonly...)
	for _, ix := range tx.Message.Instructions {
		if ix.ProgramIDIndex >= len(keys) {
			return nil, fmt.Errorf("unresolved program index")
		}
		for _, i := range ix.Accounts {
			if i >= len(keys) {
				return nil, fmt.Errorf("unresolved account index")
			}
		}
	}
	return parseShredKeys(tx, keys, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, false, binaryData), nil
}
func parseShredKeys(tx *NativeWireTransaction, keys []string, signature string, slot uint64, txIndex uint32, blockTimeUs *int64, grpcRecvUs int64, filter EventTypeFilter, bestEffort bool, binaryData [][]byte) []DexEvent {
	if parseInstructionUnifiedPreFilterRust(filter) {
		return nil
	}
	if signature == "" {
		signature = tx.Signatures[0]
	}
	var out []DexEvent
	for index, ix := range tx.Message.Instructions {
		pid := ""
		if ix.ProgramIDIndex < len(keys) {
			pid = keys[ix.ProgramIDIndex]
		}
		resolved := pid != ""
		for _, i := range ix.Accounts {
			if i >= len(keys) {
				resolved = false
				break
			}
		}
		var data []byte
		if binaryData != nil {
			data = binaryData[index]
		} else {
			var err error
			data, err = base58.Decode(ix.Data)
			if err != nil {
				continue
			}
		}
		unknown := isUnknownShredProgram(pid)
		explicitRaw := false
		if f, ok := filter.(*IncludeOnlyFilter); ok {
			explicitRaw = eventTypeSliceContains(f.IncludeOnly, EventTypeRawInstruction)
		}
		if pid != "" && (unknown || explicitRaw) {
			accounts := make([]*string, len(ix.Accounts))
			indices := make([]int, len(ix.Accounts))
			for i, k := range ix.Accounts {
				indices[i] = k
				if k < len(keys) {
					key := keys[k]
					accounts[i] = &key
				}
			}
			payload := make([]int, len(data))
			for i, b := range data {
				payload[i] = int(b)
			}
			dexes := rawInstructionDexes(pid, accounts)
			if rawInstructionMatchesFilter(dexes, filter) {
				out = append(out, DexEvent{Type: EventTypeRawInstruction, Data: &RawInstructionEvent{
					Dexes:     dexes,
					Metadata:  makeMetadata(signature, slot, uint64(txIndex), blockTimeUs, grpcRecvUs, tx.Message.RecentBlockhash),
					ProgramID: pid, InstructionIndex: index, Accounts: accounts, AccountIndices: indices, InstructionData: payload,
				}})
			}
		}
		if unknown {
			continue
		}
		if !resolved && !bestEffort {
			continue
		}
		route := MayParsePumpRoute(pid, data)
		if route && (!resolved || !allowsPumpRoute(filter, pid)) {
			continue
		}
		accounts := make([]string, len(ix.Accounts))
		for i, k := range ix.Accounts {
			if k < len(keys) {
				accounts[i] = keys[k]
			} else {
				accounts[i] = zeroPubkey
			}
		}
		if route {
			meta := makeMetadata(signature, slot, uint64(txIndex), blockTimeUs, grpcRecvUs, tx.Message.RecentBlockhash)
			if e := ParsePumpRoute(pid, data, accounts, meta, index, filter); e != nil {
				out = append(out, DexEvent{Type: EventTypePumpRouteIntent, Data: e})
			}
			continue
		}
		if pid == "" {
			for _, candidate := range shredUnknownProgramCandidates {
				e := ParseInstructionUnified(data, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, candidate)
				if e.Type != "" {
					out = append(out, e)
					break
				}
			}
			continue
		}
		e := ParseInstructionUnified(data, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, pid)
		if e.Type != "" {
			out = append(out, e)
		}
	}
	enrichPumpfunSameTxPostMerge(out)
	for i := range out {
		out[i].SetRecentBlockhash(tx.Message.RecentBlockhash)
	}
	return out
}

// ResolveShredAddressesFromCache keeps runtime writable-then-readonly ordering.
func ResolveShredAddressesFromCache(tx *NativeWireTransaction, tables map[string][]string) ([]string, []string, error) {
	var w, r []string
	for _, l := range tx.Message.AddressTableLookups {
		table, ok := tables[l.AccountKey]
		if !ok {
			return nil, nil, fmt.Errorf("ALT cache miss: %s", l.AccountKey)
		}
		for _, i := range l.WritableIndexes {
			if i < 0 || i >= len(table) {
				return nil, nil, fmt.Errorf("ALT index out of range")
			}
			w = append(w, table[i])
		}
		for _, i := range l.ReadonlyIndexes {
			if i < 0 || i >= len(table) {
				return nil, nil, fmt.Errorf("ALT index out of range")
			}
			r = append(r, table[i])
		}
	}
	return w, r, nil
}
