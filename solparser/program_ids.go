package solparser

// Program ID 与 Rust sol-parser-sdk 对齐。
//
// 指令解析、ParseInstructionUnified、RPC 路径使用 **src/instr/program_ids.rs**：
const (
	PUMPFUN_PROGRAM_ID         = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
	PUMPSWAP_PROGRAM_ID        = "pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"
	PUMP_FEES_PROGRAM_ID       = "pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ"
	RAYDIUM_CLMM_PROGRAM_ID    = "CAMMCzo5YL8w4VFF8KVHrK22GGUsp5VTaW7grrKgrWqK"
	RAYDIUM_CPMM_PROGRAM_ID    = "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C"
	RAYDIUM_AMM_V4_PROGRAM_ID  = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	ORCA_WHIRLPOOL_PROGRAM_ID  = "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"
	METEORA_POOLS_PROGRAM_ID   = "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB"
	METEORA_DAMM_V2_PROGRAM_ID = "cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG"
	METEORA_DLMM_PROGRAM_ID    = "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"
	METEORA_DBC_PROGRAM_ID     = "dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN"
	// Raydium LaunchLab 外层指令（raydium_launchlab.rs）。
	RAYDIUM_LAUNCHLAB_PROGRAM_ID = "LanMV9sAd7wArD4vJFi2qDdfnVhFxYSUg6eADduJ3uj"
)

var (
	GrpcRaydiumClmmProgramID   = RAYDIUM_CLMM_PROGRAM_ID
	GrpcMeteoraDammV2ProgramID = METEORA_DAMM_V2_PROGRAM_ID
	GrpcPumpSwapFeesProgramID  = PUMP_FEES_PROGRAM_ID
)

// Protocol 订阅/过滤用协议枚举（对齐 Rust `get_program_ids_for_protocols` 思路）。
type Protocol string

const (
	ProtocolPumpFun          Protocol = "PumpFun"
	ProtocolPumpSwap         Protocol = "PumpSwap"
	ProtocolPumpFees         Protocol = "PumpFees"
	ProtocolPumpSwapFees     Protocol = "PumpSwapFees" // backward-compatible alias
	ProtocolRaydiumLaunchlab Protocol = "RaydiumLaunchlab"
	ProtocolRaydiumClmm      Protocol = "RaydiumClmm"
	ProtocolRaydiumCpmm      Protocol = "RaydiumCpmm"
	ProtocolRaydiumAmmV4     Protocol = "RaydiumAmmV4"
	ProtocolOrcaWhirlpool    Protocol = "OrcaWhirlpool"
	ProtocolMeteoraPools     Protocol = "MeteoraPools"
	ProtocolMeteoraDammV2    Protocol = "MeteoraDammV2"
	ProtocolMeteoraDlmm      Protocol = "MeteoraDlmm"
	ProtocolMeteoraDbc       Protocol = "MeteoraDbc"
)

// GetProgramIDsForProtocols 返回给定协议对应的链上 Program ID 列表（去重、保序），便于 Yellowstone 订阅 account_include。
func GetProgramIDsForProtocols(protocols []Protocol) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, p := range protocols {
		for _, id := range programIDsForProtocol(p) {
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

// TransactionFilterForProtocols 从协议列表创建交易过滤器。
func TransactionFilterForProtocols(protocols []Protocol) TransactionFilter {
	return TransactionFilter{
		AccountInclude:  GetProgramIDsForProtocols(protocols),
		AccountExclude:  []string{},
		AccountRequired: []string{},
	}
}

// AccountFilterForProtocols 从协议列表创建 owner 账户过滤器。
func AccountFilterForProtocols(protocols []Protocol) AccountFilter {
	return AccountFilterFromProgramOwners(GetProgramIDsForProtocols(protocols))
}

func programIDsForProtocol(p Protocol) []string {
	switch p {
	case ProtocolPumpFun:
		return []string{PUMPFUN_PROGRAM_ID}
	case ProtocolPumpSwap:
		return []string{PUMPSWAP_PROGRAM_ID}
	case ProtocolPumpFees, ProtocolPumpSwapFees:
		return []string{PUMP_FEES_PROGRAM_ID}
	case ProtocolRaydiumClmm:
		return []string{RAYDIUM_CLMM_PROGRAM_ID}
	case ProtocolRaydiumCpmm:
		return []string{RAYDIUM_CPMM_PROGRAM_ID}
	case ProtocolRaydiumAmmV4:
		return []string{RAYDIUM_AMM_V4_PROGRAM_ID}
	case ProtocolOrcaWhirlpool:
		return []string{ORCA_WHIRLPOOL_PROGRAM_ID}
	case ProtocolMeteoraPools:
		return []string{METEORA_POOLS_PROGRAM_ID}
	case ProtocolMeteoraDammV2:
		return []string{METEORA_DAMM_V2_PROGRAM_ID}
	case ProtocolMeteoraDlmm:
		return []string{METEORA_DLMM_PROGRAM_ID}
	case ProtocolMeteoraDbc:
		return []string{METEORA_DBC_PROGRAM_ID}
	case ProtocolRaydiumLaunchlab:
		return []string{RAYDIUM_LAUNCHLAB_PROGRAM_ID}
	default:
		return nil
	}
}
