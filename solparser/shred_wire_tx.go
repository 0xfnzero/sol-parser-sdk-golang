package solparser

import (
	"bytes"
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

// DexEventsFromShredTransactionWire preserves known Pump.fun inputs with unresolved ALT addresses. No RPC.
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
			if i < 0 || i >= len(keys) {
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
	created, mayhem := map[string]bool{}, map[string]bool{}
	// Scan raw create instructions even when creation events are filtered out.
	if filter == nil || EventTypeFilterIncludesPumpfun(filter) {
		for index, ix := range tx.Message.Instructions {
			if ix.ProgramIDIndex < 0 || ix.ProgramIDIndex >= len(keys) || keys[ix.ProgramIDIndex] != PUMPFUN_PROGRAM_ID || len(ix.Accounts) == 0 {
				continue
			}
			mintIndex := ix.Accounts[0]
			if mintIndex < 0 || mintIndex >= len(keys) || keys[mintIndex] == zeroPubkey {
				continue
			}
			var data []byte
			if binaryData != nil {
				data = binaryData[index]
			} else {
				data, _ = base58.Decode(ix.Data)
			}
			if len(data) < 8 || (!bytes.Equal(data[:8], []byte{24, 30, 200, 40, 5, 28, 7, 119}) && !bytes.Equal(data[:8], []byte{214, 144, 76, 236, 95, 139, 49, 180})) {
				continue
			}
			created[keys[mintIndex]] = true
			if bytes.Equal(data[:8], []byte{214, 144, 76, 236, 95, 139, 49, 180}) {
				// Borsh name, symbol, uri, creator, is_mayhem_mode.
				offset := 8
				for i := 0; i < 3 && offset+4 <= len(data); i++ {
					size := int(uint32(data[offset]) | uint32(data[offset+1])<<8 | uint32(data[offset+2])<<16 | uint32(data[offset+3])<<24)
					offset += 4 + size
				}
				if offset+32 < len(data) && data[offset+32] != 0 {
					mayhem[keys[mintIndex]] = true
				}
			}
		}
	}
	var out []DexEvent
	for index, ix := range tx.Message.Instructions {
		pid := ""
		if ix.ProgramIDIndex >= 0 && ix.ProgramIDIndex < len(keys) {
			pid = keys[ix.ProgramIDIndex]
		}
		resolved := pid != ""
		for _, i := range ix.Accounts {
			if i < 0 || i >= len(keys) {
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
		if unknown && shredIncludesPumpTrade(filter) {
			for _, k := range ix.Accounts {
				if k < 0 || k >= len(keys) || keys[k] != PUMPFUN_PROGRAM_ID {
					continue
				}
				raw := &RawInstructionPayload{ProgramID: pid, InstructionIndex: index,
					Accounts:       make([]*string, len(ix.Accounts)),
					AccountIndices: append([]int{}, ix.Accounts...), InstructionData: make([]int, len(data))}
				for i, keyIndex := range ix.Accounts {
					if keyIndex >= 0 && keyIndex < len(keys) {
						key := keys[keyIndex]
						raw.Accounts[i] = &key
					}
				}
				for i, b := range data {
					raw.InstructionData[i] = int(b)
				}
				out = append(out, DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{
					Metadata:     makeMetadata(signature, slot, uint64(txIndex), blockTimeUs, grpcRecvUs, tx.Message.RecentBlockhash),
					ThirdProgram: raw, ParseStatus: TradeParseUnresolved}})
				break
			}
		}
		if unknown {
			continue
		}
		if pid == PUMPSWAP_PROGRAM_ID && len(data) >= 8 && bytes.Equal(data[:8], pumpMultiHopSwap) {
			if shredIncludesPumpTrade(filter) {
				hops := parseShredPumpCurveHops(data, ix.Accounts, keys, index, makeMetadata(signature, slot, uint64(txIndex), blockTimeUs, grpcRecvUs, tx.Message.RecentBlockhash))
				for _, hop := range hops {
					t := hop.Data.(*PumpFunTradeEvent)
					t.IsCreatedBuy = t.IsBuy && created[t.Mint]
				}
				out = append(out, hops...)
			}
			continue
		}
		if !resolved && !bestEffort {
			if pid != PUMPFUN_PROGRAM_ID {
				continue
			}
			if len(data) >= 8 {
				minimum := 0
				switch {
				case bytes.Equal(data[:8], []byte{184, 23, 238, 97, 103, 197, 211, 61}), bytes.Equal(data[:8], []byte{194, 171, 28, 70, 104, 77, 91, 47}):
					minimum = 27
				case bytes.Equal(data[:8], []byte{93, 246, 130, 60, 231, 233, 64, 178}):
					minimum = 26
				}
				if len(ix.Accounts) < minimum {
					continue
				}
			}
		}
		accounts := make([]string, len(ix.Accounts))
		for i, k := range ix.Accounts {
			if k >= 0 && k < len(keys) {
				accounts[i] = keys[k]
			} else {
				accounts[i] = zeroPubkey
			}
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
			if t, ok := e.Data.(*PumpFunTradeEvent); ok && pid == PUMPFUN_PROGRAM_ID {
				t.IsCreatedBuy = t.IsBuy && created[t.Mint]
				t.MayhemMode = mayhem[t.Mint]
				missing := func(i int) bool { return i < len(ix.Accounts) && (ix.Accounts[i] < 0 || ix.Accounts[i] >= len(keys)) }
				if len(t.IxName) >= 3 && (t.IxName[len(t.IxName)-3:] == "_v2" || t.IxName[len(t.IxName)-3:] == "_v3") {
					if missing(2) {
						t.QuoteMint = zeroPubkey
					}
					if missing(3) {
						t.TokenProgram = zeroPubkey
					}
					if missing(4) {
						t.QuoteTokenProgram = zeroPubkey
					}
				} else {
					i := 8
					if t.IxName == "sell" {
						i = 9
					}
					if missing(i) {
						t.TokenProgram = zeroPubkey
					}
				}
				if only, ok := filter.(*IncludeOnlyFilter); ok {
					generic, specific := false, false
					for _, kind := range only.IncludeOnly {
						if kind == EventTypePumpFunTrade {
							generic = true
						}
						if kind == EventTypePumpFunBuy || kind == EventTypePumpFunSell || kind == EventTypePumpFunBuyExactSolIn {
							specific = true
						}
					}
					if generic && !specific {
						e.Type = EventTypePumpFunTrade
					}
				}
			}
			out = append(out, e)
		}
	}
	enrichPumpfunSameTxPostMerge(out)
	for i := range out {
		_, status, _ := tradeSource(out[i])
		if status != nil && *status == "" {
			*status = TradeParseInstructionDecoded
		}
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

func shredIncludesPumpTrade(filter EventTypeFilter) bool {
	if only, ok := filter.(*IncludeOnlyFilter); ok {
		for _, kind := range only.IncludeOnly {
			if kind == EventTypePumpFunTrade {
				return true
			}
		}
		return false
	}
	return filter == nil || filter.ShouldInclude(EventTypePumpFunTrade)
}
