package solparser

import (
	"fmt"
	"time"

	"github.com/mr-tron/base58"

	pb "sol-parser-sdk-golang/proto"
)

// ParseError RPC 解析错误
type ParseError struct {
	Kind    string
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// RpcClient RPC 客户端接口
type RpcClient interface {
	GetTransaction(signature string, maxSupportedTransactionVersion int) (*RpcTransactionResponse, error)
}

// RpcTransactionResponse RPC 交易响应
type RpcTransactionResponse struct {
	Slot       uint64
	BlockTime  *int64
	Meta       *RpcTransactionMeta
	Transaction *RpcTransaction
}

// RpcTransactionMeta 交易元数据
type RpcTransactionMeta struct {
	Fee                  uint64
	PreBalances          []uint64
	PostBalances         []uint64
	LogMessages          []string
	InnerInstructions    []RpcInnerInstructionGroup
	PreTokenBalances     []RpcTokenBalance
	PostTokenBalances    []RpcTokenBalance
	LoadedAddresses      *RpcLoadedAddresses
	ComputeUnitsConsumed *uint64
}

// RpcInnerInstructionGroup 内部指令组
type RpcInnerInstructionGroup struct {
	Index        uint32
	Instructions []RpcCompiledInstruction
}

// RpcCompiledInstruction 编译指令
type RpcCompiledInstruction struct {
	ProgramIDIndex uint32
	Accounts       []byte
	Data           []byte
}

// RpcTokenBalance Token 余额
type RpcTokenBalance struct {
	AccountIndex  uint32
	Mint          string
	UiTokenAmount RpcUiTokenAmount
}

// RpcUiTokenAmount Token 金额
type RpcUiTokenAmount struct {
	Amount         string
	Decimals       uint32
	UiAmount       float64
	UiAmountString string
}

// RpcLoadedAddresses 加载地址
type RpcLoadedAddresses struct {
	Writable []string
	Readonly []string
}

// RpcTransaction 交易
type RpcTransaction struct {
	Signatures []string
	Message    *RpcMessage
}

// RpcMessage 消息
type RpcMessage struct {
	AccountKeys         []string
	Header              *RpcMessageHeader
	RecentBlockhash     string
	Instructions        []RpcCompiledInstruction
	AddressTableLookups []RpcMessageAddressTableLookup
}

// RpcMessageHeader 消息头
type RpcMessageHeader struct {
	NumRequiredSignatures       uint32
	NumReadonlySignedAccounts   uint32
	NumReadonlyUnsignedAccounts uint32
}

// RpcMessageAddressTableLookup 地址表查找
type RpcMessageAddressTableLookup struct {
	AccountKey      string
	WritableIndexes []byte
	ReadonlyIndexes []byte
}

// parseTransactionFromRpcImpl 内部实现 - 通过 RPC 拉取交易并解析
func parseTransactionFromRpcImpl(
	rpcClient RpcClient,
	signature string,
	filter EventTypeFilter,
) ([]DexEvent, *ParseError) {
	tx, err := rpcClient.GetTransaction(signature, 0)
	if err != nil {
		return nil, &ParseError{
			Kind:    "RpcError",
			Message: fmt.Sprintf("Failed to fetch transaction: %v", err),
		}
	}

	if tx == nil {
		return nil, &ParseError{
			Kind:    "RpcError",
			Message: "Transaction not found or null response (try archive RPC for old txs)",
		}
	}

	grpcRecvUs := time.Now().UnixMicro()
	return parseRpcTransactionImpl(tx, signature, filter, grpcRecvUs)
}

// parseRpcTransactionImpl 内部实现 - 解析已获取的 RPC 交易
func parseRpcTransactionImpl(
	tx *RpcTransactionResponse,
	signature string,
	filter EventTypeFilter,
	grpcRecvUs int64,
) ([]DexEvent, *ParseError) {
	if tx.Transaction == nil || tx.Transaction.Message == nil {
		return nil, &ParseError{
			Kind:    "ConversionError",
			Message: "Transaction message is nil",
		}
	}

	msg := tx.Transaction.Message
	meta := tx.Meta
	if meta == nil {
		meta = &RpcTransactionMeta{}
	}

	slot := tx.Slot
	var blockTimeUs *int64
	if tx.BlockTime != nil {
		us := *tx.BlockTime * 1_000_000
		blockTimeUs = &us
	}

	events := []DexEvent{}

	// 解析外层指令
	for i, ix := range msg.Instructions {
		if ev := parseRpcInstruction(
			ix,
			msg.AccountKeys,
			signature,
			slot,
			uint32(i),
			blockTimeUs,
			grpcRecvUs,
			filter,
		); ev != nil {
			events = append(events, ev)
		}
	}

	// 解析内层指令
	for _, group := range meta.InnerInstructions {
		for _, ix := range group.Instructions {
			if ev := parseRpcInstruction(
				ix,
				msg.AccountKeys,
				signature,
				slot,
				group.Index,
				blockTimeUs,
				grpcRecvUs,
				filter,
			); ev != nil {
				events = append(events, ev)
			}
		}
	}

	// 解析日志
	isCreatedBuy := false
	recentBlockhash := ""
	if msg.RecentBlockhash != "" {
		recentBlockhash = msg.RecentBlockhash
	}

	for _, log := range meta.LogMessages {
		ev := ParseLogOptimized(
			log,
			signature,
			slot,
			0,
			blockTimeUs,
			grpcRecvUs,
			filter,
			isCreatedBuy,
			recentBlockhash,
		)
		if ev != nil {
			// 检查是否是 PumpFun Create 事件
			_, hasPFC := ev["PumpFunCreate"]
			_, hasPFCV2 := ev["PumpFunCreateV2"]
			if hasPFC || hasPFCV2 {
				isCreatedBuy = true
			}
			events = append(events, ev)
		}
	}

	return events, nil
}

// parseRpcInstruction 解析 RPC 指令
func parseRpcInstruction(
	ix RpcCompiledInstruction,
	accountKeys []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
	filter EventTypeFilter,
) DexEvent {
	// 获取程序 ID
	if int(ix.ProgramIDIndex) >= len(accountKeys) {
		return nil
	}
	programId := accountKeys[ix.ProgramIDIndex]

	// 解析指令数据
	data := ix.Data
	if len(data) == 0 {
		return nil
	}

	// 构建账户列表
	accounts := make([]string, len(ix.Accounts))
	for i, accIdx := range ix.Accounts {
		if int(accIdx) < len(accountKeys) {
			accounts[i] = accountKeys[accIdx]
		}
	}

	// 根据程序 ID 路由到相应的解析器
	switch programId {
	case PUMPFUN_PROGRAM_ID:
		if !EventTypeFilterIncludesPumpfun(filter) {
			return nil
		}
		return parsePumpfunInstruction(data, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case PUMPSWAP_PROGRAM_ID:
		if !EventTypeFilterIncludesPumpswap(filter) {
			return nil
		}
		return parsePumpswapInstruction(data, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)

	case METEORA_DAMM_V2_PROGRAM_ID:
		if !EventTypeFilterIncludesMeteoraDammV2(filter) {
			return nil
		}
		return parseMeteoraDammInstruction(data, accounts, signature, slot, txIndex, blockTimeUs, grpcRecvUs)
	}

	return nil
}

// parsePumpfunInstruction 解析 PumpFun 指令
func parsePumpfunInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	// 解析 discriminator (前 8 字节)
	if len(data) < 8 {
		return nil
	}
	disc := [8]byte{}
	copy(disc[:], data[:8])

	// 这里需要根据具体的指令格式解析
	// 暂时返回 nil，需要实现具体的解析逻辑
	_ = disc
	_ = accounts
	_ = signature
	_ = slot
	_ = txIndex
	_ = blockTimeUs
	_ = grpcRecvUs

	return nil
}

// parsePumpswapInstruction 解析 PumpSwap 指令
func parsePumpswapInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return nil
	}
	disc := [8]byte{}
	copy(disc[:], data[:8])

	_ = disc
	_ = accounts
	_ = signature
	_ = slot
	_ = txIndex
	_ = blockTimeUs
	_ = grpcRecvUs

	return nil
}

// parseMeteoraDammInstruction 解析 Meteora DAMM 指令
func parseMeteoraDammInstruction(
	data []byte,
	accounts []string,
	signature string,
	slot uint64,
	txIndex uint32,
	blockTimeUs *int64,
	grpcRecvUs int64,
) DexEvent {
	if len(data) < 8 {
		return nil
	}
	disc := [8]byte{}
	copy(disc[:], data[:8])

	_ = disc
	_ = accounts
	_ = signature
	_ = slot
	_ = txIndex
	_ = blockTimeUs
	_ = grpcRecvUs

	return nil
}

// ConvertRpcToGrpc 将 RPC 格式转换为 gRPC 格式
func ConvertRpcToGrpc(
	rpcTx *RpcTransactionResponse,
) (*pb.TransactionStatusMeta, *pb.Transaction, error) {
	meta := rpcTx.Meta
	if meta == nil {
		return nil, nil, fmt.Errorf("meta is nil")
	}

	// 转换 TransactionStatusMeta
	grpcMeta := &pb.TransactionStatusMeta{
		Fee:              meta.Fee,
		PreBalances:      meta.PreBalances,
		PostBalances:     meta.PostBalances,
		LogMessages:      meta.LogMessages,
		InnerInstructions: make([]*pb.InnerInstructions, len(meta.InnerInstructions)),
		PreTokenBalances: make([]*pb.TokenBalance, len(meta.PreTokenBalances)),
		PostTokenBalances: make([]*pb.TokenBalance, len(meta.PostTokenBalances)),
	}

	// 转换内部指令
	for i, group := range meta.InnerInstructions {
		grpcGroup := &pb.InnerInstructions{
			Index:        group.Index,
			Instructions: make([]*pb.InnerInstruction, len(group.Instructions)),
		}
		for j, ix := range group.Instructions {
			grpcGroup.Instructions[j] = &pb.InnerInstruction{
				ProgramIdIndex: ix.ProgramIDIndex,
				Accounts:       ix.Accounts,
				Data:           ix.Data,
			}
		}
		grpcMeta.InnerInstructions[i] = grpcGroup
	}

	// 转换加载的地址
	if meta.LoadedAddresses != nil {
		grpcMeta.LoadedWritableAddresses = make([][]byte, len(meta.LoadedAddresses.Writable))
		for i, addr := range meta.LoadedAddresses.Writable {
			decoded, err := base58.Decode(addr)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode writable address: %w", err)
			}
			grpcMeta.LoadedWritableAddresses[i] = decoded
		}

		grpcMeta.LoadedReadonlyAddresses = make([][]byte, len(meta.LoadedAddresses.Readonly))
		for i, addr := range meta.LoadedAddresses.Readonly {
			decoded, err := base58.Decode(addr)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode readonly address: %w", err)
			}
			grpcMeta.LoadedReadonlyAddresses[i] = decoded
		}
	}

	// 转换交易
	if rpcTx.Transaction == nil {
		return nil, nil, fmt.Errorf("transaction is nil")
	}

	tx := rpcTx.Transaction
	grpcTx := &pb.Transaction{
		Signatures: make([][]byte, len(tx.Signatures)),
	}

	// 转换签名
	for i, sig := range tx.Signatures {
		decoded, err := base58.Decode(sig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode signature: %w", err)
		}
		grpcTx.Signatures[i] = decoded
	}

	// 转换消息
	if tx.Message != nil {
		msg := tx.Message
		grpcMsg := &pb.Message{
			AccountKeys:         make([][]byte, len(msg.AccountKeys)),
			RecentBlockhash:     make([]byte, 32),
			Instructions:        make([]*pb.CompiledInstruction, len(msg.Instructions)),
			AddressTableLookups: make([]*pb.MessageAddressTableLookup, len(msg.AddressTableLookups)),
		}

		// 转换账户密钥
		for i, key := range msg.AccountKeys {
			decoded, err := base58.Decode(key)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode account key: %w", err)
			}
			grpcMsg.AccountKeys[i] = decoded
		}

		// 转换最近区块哈希
		if msg.RecentBlockhash != "" {
			decoded, err := base58.Decode(msg.RecentBlockhash)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode recent blockhash: %w", err)
			}
			copy(grpcMsg.RecentBlockhash, decoded)
		}

		// 转换指令
		for i, ix := range msg.Instructions {
			grpcMsg.Instructions[i] = &pb.CompiledInstruction{
				ProgramIdIndex: ix.ProgramIDIndex,
				Accounts:       ix.Accounts,
				Data:           ix.Data,
			}
		}

		// 转换地址表查找
		for i, lookup := range msg.AddressTableLookups {
			decoded, err := base58.Decode(lookup.AccountKey)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decode lookup table key: %w", err)
			}
			grpcMsg.AddressTableLookups[i] = &pb.MessageAddressTableLookup{
				AccountKey:      decoded,
				WritableIndexes: lookup.WritableIndexes,
				ReadonlyIndexes: lookup.ReadonlyIndexes,
			}
		}

		// 转换消息头
		if msg.Header != nil {
			grpcMsg.Header = &pb.MessageHeader{
				NumRequiredSignatures:       msg.Header.NumRequiredSignatures,
				NumReadonlySignedAccounts:   msg.Header.NumReadonlySignedAccounts,
				NumReadonlyUnsignedAccounts: msg.Header.NumReadonlyUnsignedAccounts,
			}
		}

		grpcTx.Message = grpcMsg
	}

	return grpcMeta, grpcTx, nil
}
