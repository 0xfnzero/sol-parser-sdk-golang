package solparser

import (
	"fmt"

	"github.com/mr-tron/base58"

	pb "sol-parser-sdk-golang/proto"
)

// SubscribeUpdateInfoToRpc 将 gRPC 回调中的 SubscribeUpdateTransactionInfo（与 Yellowstone 一致）
// 转为 RpcTransactionResponse，供 ParseRpcTransaction 走指令 + 日志全量解析。
func SubscribeUpdateInfoToRpc(slot uint64, info *SubscribeUpdateTransactionInfo) (*RpcTransactionResponse, error) {
	if info == nil {
		return nil, fmt.Errorf("geyser: nil transaction info")
	}
	if info.Transaction == nil || info.Transaction.Message == nil {
		return nil, fmt.Errorf("geyser: nil transaction message")
	}

	rpcTx, err := pbTransactionToRpc(info.Transaction)
	if err != nil {
		return nil, err
	}
	meta := pbMetaToRpc(info.Meta)

	return &RpcTransactionResponse{
		Slot:        slot,
		BlockTime:   nil,
		Meta:        meta,
		Transaction: rpcTx,
	}, nil
}

func pbTransactionToRpc(tx *pb.Transaction) (*RpcTransaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("nil transaction")
	}
	out := &RpcTransaction{
		Signatures: make([]string, len(tx.Signatures)),
	}
	for i, s := range tx.Signatures {
		out.Signatures[i] = base58.Encode(s)
	}
	if tx.Message != nil {
		out.Message = pbMessageToRpc(tx.Message)
	}
	return out, nil
}

func pbMessageToRpc(msg *pb.Message) *RpcMessage {
	m := &RpcMessage{
		AccountKeys:         make([]string, len(msg.AccountKeys)),
		Instructions:        make([]RpcCompiledInstruction, len(msg.Instructions)),
		AddressTableLookups: make([]RpcMessageAddressTableLookup, len(msg.AddressTableLookups)),
	}
	for i, k := range msg.AccountKeys {
		m.AccountKeys[i] = base58.Encode(k)
	}
	if len(msg.RecentBlockhash) > 0 {
		m.RecentBlockhash = base58.Encode(msg.RecentBlockhash)
	}
	if msg.Header != nil {
		m.Header = &RpcMessageHeader{
			NumRequiredSignatures:       msg.Header.NumRequiredSignatures,
			NumReadonlySignedAccounts:   msg.Header.NumReadonlySignedAccounts,
			NumReadonlyUnsignedAccounts: msg.Header.NumReadonlyUnsignedAccounts,
		}
	}
	for i, ix := range msg.Instructions {
		m.Instructions[i] = RpcCompiledInstruction{
			ProgramIDIndex: ix.ProgramIdIndex,
			Accounts:       ix.Accounts,
			Data:           ix.Data,
		}
	}
	for i, l := range msg.AddressTableLookups {
		m.AddressTableLookups[i] = RpcMessageAddressTableLookup{
			AccountKey:      base58.Encode(l.AccountKey),
			WritableIndexes: l.WritableIndexes,
			ReadonlyIndexes: l.ReadonlyIndexes,
		}
	}
	return m
}

func pbMetaToRpc(meta *pb.TransactionStatusMeta) *RpcTransactionMeta {
	if meta == nil {
		return &RpcTransactionMeta{}
	}
	out := &RpcTransactionMeta{
		Fee:               meta.Fee,
		PreBalances:       meta.PreBalances,
		PostBalances:      meta.PostBalances,
		LogMessages:       meta.LogMessages,
		InnerInstructions: make([]RpcInnerInstructionGroup, len(meta.InnerInstructions)),
	}
	for i, g := range meta.InnerInstructions {
		grp := RpcInnerInstructionGroup{Index: g.Index}
		for _, ix := range g.Instructions {
			grp.Instructions = append(grp.Instructions, RpcCompiledInstruction{
				ProgramIDIndex: ix.ProgramIdIndex,
				Accounts:       ix.Accounts,
				Data:           ix.Data,
			})
		}
		out.InnerInstructions[i] = grp
	}
	if len(meta.LoadedWritableAddresses)+len(meta.LoadedReadonlyAddresses) > 0 {
		out.LoadedAddresses = &RpcLoadedAddresses{}
		for _, w := range meta.LoadedWritableAddresses {
			out.LoadedAddresses.Writable = append(out.LoadedAddresses.Writable, base58.Encode(w))
		}
		for _, r := range meta.LoadedReadonlyAddresses {
			out.LoadedAddresses.Readonly = append(out.LoadedAddresses.Readonly, base58.Encode(r))
		}
	}
	if meta.ComputeUnitsConsumed != nil {
		v := meta.GetComputeUnitsConsumed()
		out.ComputeUnitsConsumed = &v
	}
	return out
}

// ParseSubscribeTransaction 解析 Geyser 订阅到的单笔交易（指令层账户 + 日志 Program data），
// 并对同一 signature 重复的 PumpSwap Buy/Sell 做字段合并（补全 mint 等）。
// metadata.tx_index 使用 SubscribeUpdateTransactionInfo.index（slot 内交易序号）。
func ParseSubscribeTransaction(
	slot uint64,
	info *SubscribeUpdateTransactionInfo,
	filter EventTypeFilter,
	grpcRecvUs int64,
) ([]DexEvent, *ParseError) {
	rpcTx, err := SubscribeUpdateInfoToRpc(slot, info)
	if err != nil {
		return nil, &ParseError{Kind: "ConversionError", Message: err.Error()}
	}
	sig := base58.Encode(info.Signature)
	if len(info.Transaction.Signatures) > 0 {
		sig = base58.Encode(info.Transaction.Signatures[0])
	}
	events, perr := parseRpcTransactionImpl(rpcTx, sig, filter, grpcRecvUs)
	if perr != nil {
		return events, perr
	}
	ApplyMetadataTxIndex(events, info.Index)
	return events, nil
}
