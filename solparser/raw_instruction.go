package solparser

import (
	"encoding/json"
	"fmt"
)

const EventTypeRawInstruction EventType = "RawInstruction"

// RawInstructionEvent describes only signed outer instruction data, not a CPI or fill.
type RawInstructionEvent struct {
	Dexes            []RawInstructionDex `json:"dexes"`
	Metadata         EventMetadata       `json:"metadata"`
	ProgramID        string              `json:"program_id"`
	InstructionIndex int                 `json:"instruction_index"`
	// Preserve order and duplicate keys. nil means an unresolved ALT address.
	Accounts        []*string `json:"accounts"`
	AccountIndices  []int     `json:"account_indices"`
	InstructionData []int     `json:"instruction_data"`
}

func (e *RawInstructionEvent) GetMetadata() EventMetadata { return e.Metadata }
func (e *RawInstructionEvent) EventType() EventType       { return EventTypeRawInstruction }

var infrastructurePrograms = map[string]bool{
	"11111111111111111111111111111111":             true,
	"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA":  true,
	"TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb":  true,
	"ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL": true,
	"ComputeBudget111111111111111111111111111111":  true,
	"MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr":  true,
	"Memo1UhkJRfHyvLMcVucJwxXeuD728EqVDDwQDxFMNo":  true,
	"AddressLookupTab1e1111111111111111111111111":  true,
	"Vote111111111111111111111111111111111111111":  true,
	"Ed25519SigVerify111111111111111111111111111":  true,
	"KeccakSecp256k11111111111111111111111111111":  true,
	"Secp256r1SigVerify1111111111111111111111111":  true,
}

func isUnknownShredProgram(pid string) bool {
	if pid == "" || infrastructurePrograms[pid] {
		return false
	}
	for _, known := range shredUnknownProgramCandidates {
		if pid == known {
			return false
		}
	}
	return true
}

// RawInstructionDexType is a typed string enum with stable JSON values.
type RawInstructionDexType string

const (
	RawInstructionDexTypePumpFun          RawInstructionDexType = "pumpfun"
	RawInstructionDexTypePumpSwap         RawInstructionDexType = "pumpswap"
	RawInstructionDexTypeRaydiumLaunchlab RawInstructionDexType = "raydium_launchlab"
	RawInstructionDexTypeRaydiumCpmm      RawInstructionDexType = "raydium_cpmm"
	RawInstructionDexTypeRaydiumClmm      RawInstructionDexType = "raydium_clmm"
	RawInstructionDexTypeRaydiumAmmV4     RawInstructionDexType = "raydium_amm_v4"
	RawInstructionDexTypeOrcaWhirlpool    RawInstructionDexType = "orca_whirlpool"
	RawInstructionDexTypeMeteoraPools     RawInstructionDexType = "meteora_pools"
	RawInstructionDexTypeMeteoraDammV2    RawInstructionDexType = "meteora_damm_v2"
	RawInstructionDexTypeMeteoraDlmm      RawInstructionDexType = "meteora_dlmm"
	RawInstructionDexTypeMeteoraDbc       RawInstructionDexType = "meteora_dbc"
)

// RawInstructionDex is a program reference, not an executed CPI or known pool.
type RawInstructionDex struct {
	DexType   RawInstructionDexType `json:"dex_type"`
	ProgramID string                `json:"program_id"`
	Pool      *string               `json:"pool"`
}

var rawDexPrograms = []struct {
	name    RawInstructionDexType
	program string
	event   EventType
}{
	{RawInstructionDexTypePumpFun, PUMPFUN_PROGRAM_ID, EventTypePumpFunTrade},
	{RawInstructionDexTypePumpSwap, PUMPSWAP_PROGRAM_ID, EventTypePumpSwapTrade},
	{RawInstructionDexTypeRaydiumLaunchlab, RAYDIUM_LAUNCHLAB_PROGRAM_ID, EventTypeRaydiumLaunchlabTrade},
	{RawInstructionDexTypeRaydiumCpmm, RAYDIUM_CPMM_PROGRAM_ID, EventTypeRaydiumCpmmSwap},
	{RawInstructionDexTypeRaydiumClmm, RAYDIUM_CLMM_PROGRAM_ID, EventTypeRaydiumClmmSwap},
	{RawInstructionDexTypeRaydiumAmmV4, RAYDIUM_AMM_V4_PROGRAM_ID, EventTypeRaydiumAmmV4Swap},
	{RawInstructionDexTypeOrcaWhirlpool, ORCA_WHIRLPOOL_PROGRAM_ID, EventTypeOrcaWhirlpoolSwap},
	{RawInstructionDexTypeMeteoraPools, METEORA_POOLS_PROGRAM_ID, EventTypeMeteoraPoolsSwap},
	{RawInstructionDexTypeMeteoraDammV2, METEORA_DAMM_V2_PROGRAM_ID, EventTypeMeteoraDammV2Swap},
	{RawInstructionDexTypeMeteoraDlmm, METEORA_DLMM_PROGRAM_ID, EventTypeMeteoraDlmmSwap},
	{RawInstructionDexTypeMeteoraDbc, METEORA_DBC_PROGRAM_ID, EventTypeMeteoraDbcSwap},
}

func rawInstructionDexes(program string, accounts []*string) []RawInstructionDex {
	out := []RawInstructionDex{}
	for _, dex := range rawDexPrograms {
		found := program == dex.program
		for _, account := range accounts {
			if account != nil && *account == dex.program {
				found = true
			}
		}
		if found {
			out = append(out, RawInstructionDex{DexType: dex.name, ProgramID: dex.program})
		}
	}
	return out
}
func rawInstructionMatchesFilter(dexes []RawInstructionDex, filter EventTypeFilter) bool {
	if filter == nil {
		return true
	}
	if f, ok := filter.(*ExcludeFilter); ok && eventTypeSliceContains(f.ExcludeTypes, EventTypeRawInstruction) {
		return false
	}
	included, limited := filter.(*IncludeOnlyFilter)
	if limited && eventTypeSliceContains(included.IncludeOnly, EventTypeRawInstruction) {
		return true
	}
	if len(dexes) == 0 {
		return !limited && filter.ShouldInclude(EventTypeRawInstruction)
	}
	for _, dex := range dexes {
		for _, known := range rawDexPrograms {
			if dex.DexType == known.name && (!limited || eventTypeSliceContains(included.IncludeOnly, known.event)) && filter.ShouldInclude(known.event) {
				return true
			}
		}
	}
	return false
}

func (d *RawInstructionDexType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	candidate := RawInstructionDexType(value)
	switch candidate {
	case RawInstructionDexTypePumpFun, RawInstructionDexTypePumpSwap, RawInstructionDexTypeRaydiumLaunchlab, RawInstructionDexTypeRaydiumCpmm, RawInstructionDexTypeRaydiumClmm, RawInstructionDexTypeRaydiumAmmV4, RawInstructionDexTypeOrcaWhirlpool, RawInstructionDexTypeMeteoraPools, RawInstructionDexTypeMeteoraDammV2, RawInstructionDexTypeMeteoraDlmm, RawInstructionDexTypeMeteoraDbc:
		*d = candidate
		return nil
	default:
		return fmt.Errorf("unsupported raw instruction DEX type: %q", value)
	}
}
