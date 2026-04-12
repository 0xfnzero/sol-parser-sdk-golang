package solparser

// Program ID 与 Rust sol-parser-sdk 对齐。
//
// 指令解析、ParseInstructionUnified、RPC 路径使用 **src/instr/program_ids.rs**：
const (
	PUMPFUN_PROGRAM_ID         = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
	PUMPSWAP_PROGRAM_ID        = "pAMMBay6oceH9fJKBRHGP5D4bD4sWpmSwMn52FMfXEA"
	RAYDIUM_CLMM_PROGRAM_ID    = "CAMMCzo5YL8w4VFF8KVHrK22GGUQpMDdHFWF5LCATdCR"
	RAYDIUM_CPMM_PROGRAM_ID    = "CPMMoo8L3F4NbTegBCKVNunggL7H1ZpdTHKxQB5qKP1C"
	RAYDIUM_AMM_V4_PROGRAM_ID  = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	ORCA_WHIRLPOOL_PROGRAM_ID   = "whirLbMiicVdio4qvUfM5KAg6Ct8VwpYzGff3uctyCc"
	METEORA_POOLS_PROGRAM_ID   = "Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB"
	METEORA_DAMM_V2_PROGRAM_ID = "cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG"
	METEORA_DLMM_PROGRAM_ID    = "LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo"
	// BONK / Raydium Launchpad 外层指令（raydium_launchpad.rs）
	BONK_PROGRAM_ID = "DjVE6JNiYqPL2QXyCUUh8rNjHrbz9hXHNYt99MQ59qw1"
)

// Yellowstone / gRPC 账户过滤使用 **src/grpc/program_ids.rs**（与 instr 中 Raydium CLMM、Bonk 可能不同）。
const (
	GrpcRaydiumClmmProgramID   = "CAMMCzo5YL8w4VFF8KVHrK22GGUQtcaMpgYqJPXBDvfE"
	GrpcBonkProgramID          = "BSwp6bEBihVLdqJRKS58NaebUBSDNjN7MdpFwNaR6gn3"
	GrpcMeteoraDammV2ProgramID = "cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG"
	// GrpcPumpSwapFeesProgramID Rust `grpc/program_ids::PUMPSWAP_FEES_PROGRAM_ID`（PumpSwap 费用程序，订阅时可与 PUMPSWAP_PROGRAM_ID 一并加入 account_include）
	GrpcPumpSwapFeesProgramID = "pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ"
)

// Protocol 订阅/过滤用协议枚举（对齐 Rust `get_program_ids_for_protocols` 思路）。
type Protocol string

const (
	ProtocolPumpFun         Protocol = "PumpFun"
	ProtocolPumpSwap        Protocol = "PumpSwap"
	ProtocolPumpSwapFees    Protocol = "PumpSwapFees"
	ProtocolRaydiumClmm     Protocol = "RaydiumClmm"
	ProtocolRaydiumCpmm     Protocol = "RaydiumCpmm"
	ProtocolRaydiumAmmV4    Protocol = "RaydiumAmmV4"
	ProtocolOrcaWhirlpool  Protocol = "OrcaWhirlpool"
	ProtocolMeteoraPools   Protocol = "MeteoraPools"
	ProtocolMeteoraDammV2   Protocol = "MeteoraDammV2"
	ProtocolMeteoraDlmm    Protocol = "MeteoraDlmm"
	ProtocolBonk           Protocol = "Bonk"
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

func programIDsForProtocol(p Protocol) []string {
	switch p {
	case ProtocolPumpFun:
		return []string{PUMPFUN_PROGRAM_ID}
	case ProtocolPumpSwap:
		return []string{PUMPSWAP_PROGRAM_ID, GrpcPumpSwapFeesProgramID}
	case ProtocolPumpSwapFees:
		return []string{GrpcPumpSwapFeesProgramID}
	case ProtocolRaydiumClmm:
		return []string{RAYDIUM_CLMM_PROGRAM_ID, GrpcRaydiumClmmProgramID}
	case ProtocolRaydiumCpmm:
		return []string{RAYDIUM_CPMM_PROGRAM_ID}
	case ProtocolRaydiumAmmV4:
		return []string{RAYDIUM_AMM_V4_PROGRAM_ID}
	case ProtocolOrcaWhirlpool:
		return []string{ORCA_WHIRLPOOL_PROGRAM_ID}
	case ProtocolMeteoraPools:
		return []string{METEORA_POOLS_PROGRAM_ID}
	case ProtocolMeteoraDammV2:
		return []string{METEORA_DAMM_V2_PROGRAM_ID, GrpcMeteoraDammV2ProgramID}
	case ProtocolMeteoraDlmm:
		return []string{METEORA_DLMM_PROGRAM_ID}
	case ProtocolBonk:
		return []string{BONK_PROGRAM_ID, GrpcBonkProgramID}
	default:
		return nil
	}
}
