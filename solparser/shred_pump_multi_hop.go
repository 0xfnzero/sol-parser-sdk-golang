package solparser

import (
	"bytes"
	solana "github.com/gagliardetto/solana-go"
)

var pumpMultiHopSwap = []byte{43, 100, 73, 19, 233, 246, 111, 148}

// A fixed Pump account alone does not identify a curve hop (all-AMM routes have it too).
// Validate the canonical venue PDA and expose route inputs without estimated hop fills.
func parseShredPumpCurveHops(data []byte, indices []int, keys []string, index int, metadata EventMetadata) []DexEvent {
	if len(data) != 24 || !bytes.Equal(data[:8], pumpMultiHopSwap) || len(indices) < 21 || (len(indices)-16)%5 != 0 || bytes.Equal(data[8:16], make([]byte, 8)) || bytes.Equal(data[16:24], make([]byte, 8)) {
		return nil
	}
	get := func(i int) string {
		k := indices[i]
		if k >= 0 && k < len(keys) {
			return keys[k]
		}
		return ""
	}
	if get(11) != PUMPSWAP_PROGRAM_ID || get(12) != PUMPFUN_PROGRAM_ID || get(0) == "" {
		return nil
	}
	hops := (len(indices) - 16) / 5
	connected := func(buy bool) bool {
		if hops < 2 {
			return false
		}
		fromOffset, toOffset := 1, 0
		if buy {
			fromOffset, toOffset = 0, 1
		}
		for i := 0; i < hops-1; i++ {
			from := get(16 + i*5 + fromOffset)
			if from == "" || from == zeroPubkey || from != get(16+(i+1)*5+toOffset) {
				return false
			}
		}
		return true
	}
	buy, sell := connected(true), connected(false)
	known := buy != sell
	var out []DexEvent
	for h := 0; h < hops; h++ {
		start := 16 + h*5
		mint, quote, curve := get(start), get(start+1), get(start+2)
		if mint == "" || quote == "" || curve == "" || mint == zeroPubkey || quote == zeroPubkey || mint == quote {
			continue
		}
		pk, err := solana.PublicKeyFromBase58(mint)
		if err != nil {
			continue
		}
		expected, _, err := solana.FindProgramAddress([][]byte{[]byte("bonding-curve"), pk.Bytes()}, solana.MustPublicKeyFromBase58(PUMPFUN_PROGRAM_ID))
		if err != nil || curve != expected.String() {
			continue
		}
		raw := &RawInstructionPayload{ProgramID: PUMPSWAP_PROGRAM_ID, InstructionIndex: index, Accounts: make([]*string, len(indices)), AccountIndices: append([]int{}, indices...), InstructionData: make([]int, len(data))}
		for i, k := range indices {
			if k >= 0 && k < len(keys) {
				v := keys[k]
				raw.Accounts[i] = &v
			}
		}
		for i, b := range data {
			raw.InstructionData[i] = int(b)
		}
		status := TradeParseUnresolved
		if known {
			status = TradeParseInstructionDecoded
		}
		out = append(out, DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{
			Metadata: metadata, IxName: "multi_hop_swap", Mint: mint, QuoteMint: quote, User: get(0), BondingCurve: curve,
			AssociatedBondingCurve: get(start + 3), AssociatedQuoteBondingCurve: get(start + 4), Global: get(13), FeeConfig: get(14), EventAuthority: get(15), Program: PUMPFUN_PROGRAM_ID, SystemProgram: get(9), IsBuy: known && buy, ParseStatus: status, ThirdProgram: raw}})
	}
	return out
}
