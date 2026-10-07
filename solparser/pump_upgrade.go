package solparser

import "encoding/binary"

type PumpFunPostCompleteBuyEvent struct {
	Metadata                EventMetadata `json:"metadata"`
	User                    string        `json:"user"`
	Mint                    string        `json:"mint"`
	BondingCurve            string        `json:"bonding_curve"`
	QuoteMint               string        `json:"quote_mint"`
	Timestamp               int64         `json:"timestamp"`
	BaseOut                 uint64        `json:"base_out"`
	QuoteIn                 uint64        `json:"quote_in"`
	FeeBasisPoints          uint64        `json:"fee_basis_points"`
	Fee                     uint64        `json:"fee"`
	CreatorFeeBasisPoints   uint64        `json:"creator_fee_basis_points"`
	CreatorFee              uint64        `json:"creator_fee"`
	BuybackFee              uint64        `json:"buyback_fee"`
	PoolBaseReservesBefore  uint64        `json:"pool_base_reserves_before"`
	PoolQuoteReservesBefore uint64        `json:"pool_quote_reserves_before"`
	PoolBaseReservesAfter   uint64        `json:"pool_base_reserves_after"`
	PoolQuoteReservesAfter  uint64        `json:"pool_quote_reserves_after"`
}

func (e *PumpFunPostCompleteBuyEvent) EventType() EventType       { return EventTypePumpFunPostCompleteBuy }
func (e *PumpFunPostCompleteBuyEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunSweepBondingCurveFeeEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	Timestamp    int64         `json:"timestamp"`
	Mint         string        `json:"mint"`
	BondingCurve string        `json:"bonding_curve"`
	QuoteMint    string        `json:"quote_mint"`
	Recipient    string        `json:"recipient"`
	Amount       uint64        `json:"amount"`
	Bucket       uint8         `json:"bucket"`
}

func (e *PumpFunSweepBondingCurveFeeEvent) EventType() EventType {
	return EventTypePumpFunSweepBondingCurveFee
}
func (e *PumpFunSweepBondingCurveFeeEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpFunCompleteEvent struct {
	Metadata     EventMetadata `json:"metadata"`
	User         string        `json:"user"`
	Mint         string        `json:"mint"`
	BondingCurve string        `json:"bonding_curve"`
	Timestamp    int64         `json:"timestamp"`
	QuoteMint    string        `json:"quote_mint"`
}

func (e *PumpFunCompleteEvent) EventType() EventType       { return EventTypePumpFunComplete }
func (e *PumpFunCompleteEvent) GetMetadata() EventMetadata { return e.Metadata }

type PumpSwapSweepPoolFeeEvent struct {
	Metadata  EventMetadata `json:"metadata"`
	Timestamp int64         `json:"timestamp"`
	Pool      string        `json:"pool"`
	BaseMint  string        `json:"base_mint"`
	QuoteMint string        `json:"quote_mint"`
	Recipient string        `json:"recipient"`
	Payer     string        `json:"payer"`
	Amount    uint64        `json:"amount"`
	Bucket    uint8         `json:"bucket"`
}

func (e *PumpSwapSweepPoolFeeEvent) EventType() EventType       { return EventTypePumpSwapSweepPoolFee }
func (e *PumpSwapSweepPoolFeeEvent) GetMetadata() EventMetadata { return e.Metadata }
func pumpUpgradeEventType(disc uint64, program string) (EventType, bool) {
	if disc == 18146529233607700591 && (program == "" || program == "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P") {
		return EventTypePumpFunPostCompleteBuy, true
	}
	if disc == 3118876958563052404 && (program == "" || program == "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P") {
		return EventTypePumpFunSweepBondingCurveFee, true
	}
	if disc == 619296439455019615 && (program == "" || program == "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P") {
		return EventTypePumpFunComplete, true
	}
	if disc == 11927646055507993730 && (program == "" || program == "pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA") {
		return EventTypePumpSwapSweepPoolFee, true
	}
	return "", false
}
func parsePumpUpgradeEvent(disc uint64, data []byte, meta EventMetadata, program string) DexEvent {
	kind, ok := pumpUpgradeEventType(disc, program)
	if !ok {
		return DexEvent{}
	}
	if kind == EventTypePumpFunPostCompleteBuy {
		if len(data) != 224 {
			return DexEvent{}
		}
		key0, ok := readPubkey(data, 0)
		if !ok {
			return DexEvent{}
		}
		key32, ok := readPubkey(data, 32)
		if !ok {
			return DexEvent{}
		}
		key64, ok := readPubkey(data, 64)
		if !ok {
			return DexEvent{}
		}
		key96, ok := readPubkey(data, 96)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{Type: kind, Data: &PumpFunPostCompleteBuyEvent{Metadata: meta, User: key0, Mint: key32, BondingCurve: key64, QuoteMint: key96, Timestamp: int64(binary.LittleEndian.Uint64(data[128:136])), BaseOut: binary.LittleEndian.Uint64(data[136:144]), QuoteIn: binary.LittleEndian.Uint64(data[144:152]), FeeBasisPoints: binary.LittleEndian.Uint64(data[152:160]), Fee: binary.LittleEndian.Uint64(data[160:168]), CreatorFeeBasisPoints: binary.LittleEndian.Uint64(data[168:176]), CreatorFee: binary.LittleEndian.Uint64(data[176:184]), BuybackFee: binary.LittleEndian.Uint64(data[184:192]), PoolBaseReservesBefore: binary.LittleEndian.Uint64(data[192:200]), PoolQuoteReservesBefore: binary.LittleEndian.Uint64(data[200:208]), PoolBaseReservesAfter: binary.LittleEndian.Uint64(data[208:216]), PoolQuoteReservesAfter: binary.LittleEndian.Uint64(data[216:224])}}
	}
	if kind == EventTypePumpFunSweepBondingCurveFee {
		if len(data) != 145 {
			return DexEvent{}
		}
		key8, ok := readPubkey(data, 8)
		if !ok {
			return DexEvent{}
		}
		key40, ok := readPubkey(data, 40)
		if !ok {
			return DexEvent{}
		}
		key72, ok := readPubkey(data, 72)
		if !ok {
			return DexEvent{}
		}
		key104, ok := readPubkey(data, 104)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{Type: kind, Data: &PumpFunSweepBondingCurveFeeEvent{Metadata: meta, Timestamp: int64(binary.LittleEndian.Uint64(data[0:8])), Mint: key8, BondingCurve: key40, QuoteMint: key72, Recipient: key104, Amount: binary.LittleEndian.Uint64(data[136:144]), Bucket: data[144]}}
	}
	if kind == EventTypePumpFunComplete {
		if len(data) == 104 {
			key := []byte{6, 155, 136, 87, 254, 171, 129, 132, 251, 104, 127, 99, 70, 24, 192, 53, 218, 196, 57, 220, 26, 235, 59, 85, 152, 160, 240, 0, 0, 0, 0, 1}
			data = append(append([]byte(nil), data...), key...)
		}
		if len(data) != 136 {
			return DexEvent{}
		}
		key0, ok := readPubkey(data, 0)
		if !ok {
			return DexEvent{}
		}
		key32, ok := readPubkey(data, 32)
		if !ok {
			return DexEvent{}
		}
		key64, ok := readPubkey(data, 64)
		if !ok {
			return DexEvent{}
		}
		key104, ok := readPubkey(data, 104)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{Type: kind, Data: &PumpFunCompleteEvent{Metadata: meta, User: key0, Mint: key32, BondingCurve: key64, Timestamp: int64(binary.LittleEndian.Uint64(data[96:104])), QuoteMint: key104}}
	}
	if kind == EventTypePumpSwapSweepPoolFee {
		if len(data) != 177 {
			return DexEvent{}
		}
		key8, ok := readPubkey(data, 8)
		if !ok {
			return DexEvent{}
		}
		key40, ok := readPubkey(data, 40)
		if !ok {
			return DexEvent{}
		}
		key72, ok := readPubkey(data, 72)
		if !ok {
			return DexEvent{}
		}
		key104, ok := readPubkey(data, 104)
		if !ok {
			return DexEvent{}
		}
		key136, ok := readPubkey(data, 136)
		if !ok {
			return DexEvent{}
		}
		return DexEvent{Type: kind, Data: &PumpSwapSweepPoolFeeEvent{Metadata: meta, Timestamp: int64(binary.LittleEndian.Uint64(data[0:8])), Pool: key8, BaseMint: key40, QuoteMint: key72, Recipient: key104, Payer: key136, Amount: binary.LittleEndian.Uint64(data[168:176]), Bucket: data[176]}}
	}
	return DexEvent{}
}

// PumpMultiHopIntent contains instruction limits. Executed amounts are per-hop trade events.
type PumpMultiHopIntent struct {
	User, InputAccount, OutputAccount string
	AmountIn, MinAmountOut            uint64
	Hops                              []PumpMultiHopAccounts
}
type PumpMultiHopAccounts struct{ BaseMint, QuoteMint, Venue, BaseVault, QuoteVault string }

func DecodePumpMultiHopIntent(program string, data []byte, accounts []string) *PumpMultiHopIntent {
	if program != PUMPSWAP_PROGRAM_ID || len(data) != 24 || binary.LittleEndian.Uint64(data[:8]) != 10696039120939607083 || len(accounts) < 21 || (len(accounts)-16)%5 != 0 {
		return nil
	}
	amount, minimum := binary.LittleEndian.Uint64(data[8:16]), binary.LittleEndian.Uint64(data[16:24])
	if amount == 0 || minimum == 0 {
		return nil
	}
	result := &PumpMultiHopIntent{User: accounts[0], InputAccount: accounts[1], OutputAccount: accounts[2], AmountIn: amount, MinAmountOut: minimum}
	for i := 16; i < len(accounts); i += 5 {
		result.Hops = append(result.Hops, PumpMultiHopAccounts{accounts[i], accounts[i+1], accounts[i+2], accounts[i+3], accounts[i+4]})
	}
	return result
}
