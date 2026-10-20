package solparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
)

// RulePubkey accepts either base58 or the Rust Pubkey JSON byte array.
type RulePubkey string

func (p *RulePubkey) UnmarshalJSON(data []byte) error {
	var text string
	if json.Unmarshal(data, &text) == nil {
		if _, err := solana.PublicKeyFromBase58(text); err != nil {
			return err
		}
		*p = RulePubkey(text)
		return nil
	}
	var raw []byte
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 32 {
		return fmt.Errorf("rule public key must contain 32 bytes")
	}
	*p = RulePubkey(solana.PublicKeyFromBytes(raw).String())
	return nil
}

type WrappedAccountConstraint struct {
	Index int
	Key   RulePubkey
}

func (c *WrappedAccountConstraint) UnmarshalJSON(data []byte) error {
	var pair []json.RawMessage
	if err := json.Unmarshal(data, &pair); err != nil {
		return err
	}
	if len(pair) != 2 {
		return fmt.Errorf("expected [account_index, public_key]")
	}
	if err := json.Unmarshal(pair[0], &c.Index); err != nil {
		return err
	}
	return json.Unmarshal(pair[1], &c.Key)
}
func (c WrappedAccountConstraint) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{c.Index, c.Key})
}

// Reviewed rule registry supplied by the caller. Never populated using RPC in
// the hot path. A matching layout is not proof of future program behavior.
type WrappedPumpfunRule struct {
	RuleID            string                     `json:"rule_id"`
	ProgramID         RulePubkey                 `json:"program_id"`
	OuterAccountCount int                        `json:"outer_account_count"`
	DataPattern       []*uint8                   `json:"data_pattern"`
	PumpDiscriminator []uint8                    `json:"pump_discriminator"`
	ArgumentOffsets   []int                      `json:"argument_offsets"`
	AccountMap        []int                      `json:"account_map"`
	FixedAccounts     []WrappedAccountConstraint `json:"fixed_accounts"`
	EqualAccounts     [][2]int                   `json:"equal_accounts"`
}

// ParseWrappedPumpfunInstruction is an opt-in Shred input parser, with no I/O.
// On a miss the caller keeps RawInstruction. Executed amounts stay zero.
func ParseWrappedPumpfunInstruction(raw *RawInstructionEvent, rule *WrappedPumpfunRule, filter EventTypeFilter) DexEvent {
	if raw == nil || rule == nil || raw.ProgramID != string(rule.ProgramID) || len(raw.Accounts) != rule.OuterAccountCount || len(raw.InstructionData) != len(rule.DataPattern) {
		return DexEvent{}
	}
	if filter != nil && !filter.ShouldInclude(EventTypePumpFunTrade) {
		return DexEvent{}
	}
	accounts := make([]string, len(raw.Accounts))
	for i, p := range raw.Accounts {
		if p == nil {
			return DexEvent{}
		}
		accounts[i] = *p
	}
	guarded := false
	for i, p := range rule.DataPattern {
		if p != nil {
			guarded = true
			if int(*p) != raw.InstructionData[i] {
				return DexEvent{}
			}
		}
	}
	if !guarded {
		return DexEvent{}
	}
	valid := func(i int) bool { return i >= 0 && i < len(accounts) }
	for _, c := range rule.FixedAccounts {
		if !valid(c.Index) || accounts[c.Index] != string(c.Key) {
			return DexEvent{}
		}
	}
	for _, pair := range rule.EqualAccounts {
		if !valid(pair[0]) || !valid(pair[1]) || accounts[pair[0]] != accounts[pair[1]] {
			return DexEvent{}
		}
	}
	disc := rule.PumpDiscriminator
	if !bytes.Equal(disc, []byte{184, 23, 238, 97, 103, 197, 211, 61}) && !bytes.Equal(disc, []byte{194, 171, 28, 70, 104, 77, 91, 47}) && !bytes.Equal(disc, []byte{93, 246, 130, 60, 231, 233, 64, 178}) {
		return DexEvent{}
	}
	data := append([]byte{}, disc...)
	if len(rule.ArgumentOffsets) != 2 {
		return DexEvent{}
	}
	for _, offset := range rule.ArgumentOffsets {
		if offset < 0 || offset > len(raw.InstructionData)-8 {
			return DexEvent{}
		}
		for _, b := range raw.InstructionData[offset : offset+8] {
			if b < 0 || b > 255 {
				return DexEvent{}
			}
			data = append(data, byte(b))
		}
	}
	mapped := make([]string, len(rule.AccountMap))
	for i, index := range rule.AccountMap {
		if !valid(index) {
			return DexEvent{}
		}
		mapped[i] = accounts[index]
	}
	m := raw.Metadata
	event := ParseInstructionUnified(data, mapped, m.Signature, m.Slot, uint32(m.TxIndex), &m.BlockTimeUs, m.GrpcRecvUs, nil, PUMPFUN_PROGRAM_ID)
	trade, ok := event.Data.(*PumpFunTradeEvent)
	if !ok {
		return DexEvent{}
	}
	trade.Metadata = m
	trade.SolAmount = 0
	trade.TokenAmount = 0
	trade.QuoteAmount = 0
	trade.ThirdProgram = rawInstructionPayload(raw)
	trade.ThirdProgram.RuleID = rule.RuleID
	trade.ParseStatus = TradeParseInstructionDecoded
	return DexEvent{Type: EventTypePumpFunTrade, Data: trade}
}
