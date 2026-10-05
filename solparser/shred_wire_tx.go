package solparser

import (
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

// DexEventsFromShredTransactionWire 从 Shred / gRPC 负载中的**线格式交易字节**解析外层编译指令并调用
// ParseInstructionUnified（与 sol-parser-sdk-ts `dexEventsFromShredWasmTx` 一致）。
//
// 限制：仅使用消息中的**静态**账户表。V0 交易若指令账户索引指向 ALT 加载地址，
// 会以默认 pubkey 占位继续 best-effort 解析。program id 必须属于静态账户表；
// 非法 program/account 索引在 wire 解码阶段拒绝，不使用 discriminator 猜测程序身份。
func DexEventsFromShredTransactionWire(
	raw []byte,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
	filter EventTypeFilter,
) []DexEvent {
	if parseInstructionUnifiedPreFilterRust(filter) {
		return nil
	}

	tx, _, err := DecodeWireTransaction(raw, 0, true)
	if err != nil {
		return nil
	}
	msg := tx.Message
	keys := msg.AccountKeys
	if len(keys) == 0 {
		return nil
	}

	var out []DexEvent
	for _, ix := range msg.Instructions {
		pid := ""
		if int(ix.ProgramIDIndex) < len(keys) {
			pid = keys[ix.ProgramIDIndex]
		}

		accStrs := make([]string, 0, len(ix.Accounts))
		for _, ai := range ix.Accounts {
			if int(ai) >= len(keys) {
				accStrs = append(accStrs, zeroPubkey)
				continue
			}
			accStrs = append(accStrs, keys[ai])
		}

		data, err := base58.Decode(ix.Data)
		if err != nil {
			continue
		}
		if pid == "" {
			for _, candidate := range shredUnknownProgramCandidates {
				ev := ParseInstructionUnified(data, accStrs, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, candidate)
				if ev.Type != "" {
					out = append(out, ev)
					break
				}
			}
			continue
		}
		ev := ParseInstructionUnified(data, accStrs, signature, slot, txIndex, blockTimeUs, grpcRecvUs, filter, pid)
		if ev.Type != "" {
			out = append(out, ev)
		}
	}
	enrichPumpfunSameTxPostMerge(out)
	return out
}
