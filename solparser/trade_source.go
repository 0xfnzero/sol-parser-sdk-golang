package solparser

import "bytes"

func tradeSource(event DexEvent) (**RawInstructionPayload, *TradeParseStatus, string) {
	switch e := event.Data.(type) {
	case *PumpFunTradeEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, PUMPFUN_PROGRAM_ID
		}
	case *PumpSwapBuyEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, PUMPSWAP_PROGRAM_ID
		}
	case *PumpSwapSellEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, PUMPSWAP_PROGRAM_ID
		}
	case *RaydiumLaunchlabTradeEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, RAYDIUM_LAUNCHLAB_PROGRAM_ID
		}
	case *RaydiumClmmSwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, RAYDIUM_CLMM_PROGRAM_ID
		}
	case *RaydiumCpmmSwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, RAYDIUM_CPMM_PROGRAM_ID
		}
	case *RaydiumAmmV4SwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, RAYDIUM_AMM_V4_PROGRAM_ID
		}
	case *OrcaWhirlpoolSwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, ORCA_WHIRLPOOL_PROGRAM_ID
		}
	case *MeteoraPoolsSwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, METEORA_POOLS_PROGRAM_ID
		}
	case *MeteoraDammV2SwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, METEORA_DAMM_V2_PROGRAM_ID
		}
	case *MeteoraDlmmSwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, METEORA_DLMM_PROGRAM_ID
		}
	case *MeteoraDbcSwapEvent:
		if e != nil {
			return &e.ThirdProgram, &e.ParseStatus, METEORA_DBC_PROGRAM_ID
		}
	}
	return nil, nil, ""
}
func markTradeExecutionSource(event DexEvent, index int, message *RpcMessage, keys []string, meta *RpcTransactionMeta) {
	source, status, native := tradeSource(event)
	if status == nil {
		return
	}
	*status = TradeParseExecutionConfirmed
	if index < 0 || index >= len(message.Instructions) {
		return
	}
	ix := message.Instructions[index]
	if int(ix.ProgramIDIndex) >= len(keys) {
		return
	}
	program := keys[ix.ProgramIDIndex]
	switch e := event.Data.(type) {
	case *RaydiumClmmSwapEvent:
		a, b := executionMintPair(index, message, keys, meta, native, e.PoolState, 2, 11, 12)
		if a != "" {
			e.InputMint, e.OutputMint = a, b
		}
	case *OrcaWhirlpoolSwapEvent:
		a, b := executionMintPair(index, message, keys, meta, native, e.Whirlpool, 4, 5, 6)
		if a != "" {
			e.TokenMintA, e.TokenMintB = a, b
		}
	}
	if program == native {
		return
	}
	raw := &RawInstructionEvent{ProgramID: program, InstructionIndex: index, Accounts: make([]*string, len(ix.Accounts))}
	for i, k := range ix.Accounts {
		raw.AccountIndices = append(raw.AccountIndices, int(k))
		if int(k) < len(keys) {
			value := keys[k]
			raw.Accounts[i] = &value
		}
	}
	for _, b := range ix.Data {
		raw.InstructionData = append(raw.InstructionData, int(b))
	}
	*source = rawInstructionPayload(raw)
}

func executionMintPair(index int, message *RpcMessage, keys []string, meta *RpcTransactionMeta, program, pool string, poolIndex, aIndex, bIndex int) (string, string) {
	candidates := []RpcCompiledInstruction{message.Instructions[index]}
	for _, group := range meta.InnerInstructions {
		if int(group.Index) == index {
			candidates = append(candidates, group.Instructions...)
		}
	}
	pairs := map[[2]string]bool{}
	for _, ix := range candidates {
		if int(ix.ProgramIDIndex) >= len(keys) || keys[ix.ProgramIDIndex] != program || len(ix.Data) < 8 || !bytes.Equal(ix.Data[:8], []byte{43, 4, 237, 11, 26, 201, 30, 98}) {
			continue
		}
		get := func(i int) string {
			if i < len(ix.Accounts) && int(ix.Accounts[i]) < len(keys) {
				return keys[ix.Accounts[i]]
			}
			return ""
		}
		if get(poolIndex) != pool {
			continue
		}
		a, b := get(aIndex), get(bIndex)
		if a != "" && b != "" {
			pairs[[2]string{a, b}] = true
		}
	}
	if len(pairs) == 1 {
		for pair := range pairs {
			return pair[0], pair[1]
		}
	}
	return "", ""
}

func logOuterIndices(logs []string, message *RpcMessage, keys []string) []int {
	positions := map[string][]int{}
	counts, used := map[string]int{}, map[string]int{}
	for i, ix := range message.Instructions {
		if int(ix.ProgramIDIndex) < len(keys) {
			p := keys[ix.ProgramIDIndex]
			positions[p] = append(positions[p], i)
		}
	}
	for _, line := range logs {
		if p, depth, ok := ParseInvokeInfo(line); ok && depth == 1 {
			counts[p]++
		}
	}
	result, current := make([]int, len(logs)), -1
	for i, line := range logs {
		if p, depth, ok := ParseInvokeInfo(line); ok && depth == 1 {
			current = -1
			if counts[p] == len(positions[p]) {
				current = positions[p][used[p]]
			}
			used[p]++
		}
		result[i] = current
	}
	return result
}

// TradeParseStatus describes evidence, not whether every executed amount is known.
type TradeParseStatus string

const (
	TradeParseUnresolved         TradeParseStatus = "Unresolved"
	TradeParseInstructionDecoded TradeParseStatus = "InstructionDecoded"
	TradeParseExecutionConfirmed TradeParseStatus = "ExecutionConfirmed"
)

type RawInstructionPayload struct {
	ProgramID        string    `json:"program_id"`
	InstructionIndex int       `json:"instruction_index"`
	Accounts         []*string `json:"accounts"`
	AccountIndices   []int     `json:"account_indices"`
	InstructionData  []int     `json:"instruction_data"`
	RuleID           string    `json:"rule_id,omitempty"`
}

func rawInstructionPayload(raw *RawInstructionEvent) *RawInstructionPayload {
	accounts := make([]*string, len(raw.Accounts))
	for i, key := range raw.Accounts {
		if key != nil {
			value := *key
			accounts[i] = &value
		}
	}
	return &RawInstructionPayload{ProgramID: raw.ProgramID, InstructionIndex: raw.InstructionIndex,
		Accounts: accounts, AccountIndices: append([]int{}, raw.AccountIndices...), InstructionData: append([]int{}, raw.InstructionData...)}
}
