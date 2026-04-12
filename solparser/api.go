package solparser

// ParseTransactionEvents 对齐 Rust `parse_transaction_events` - 解析完整交易并返回所有 DEX 事件
func ParseTransactionEvents(logs []string, signature string, slot uint64, blockTimeUs *int64) []DexEvent {
	return ParseLogsOnly(logs, signature, slot, blockTimeUs)
}

// ParseLogsOnly 对齐 Rust `parse_logs_only`
func ParseLogsOnly(logs []string, signature string, slot uint64, blockTimeUs *int64) []DexEvent {
	var out []DexEvent
	for _, log := range logs {
		if ev := ParseLogUnified(log, signature, slot, blockTimeUs); ev.Type != "" {
			out = append(out, ev)
		}
	}
	return out
}

// ParseLogsStreaming 对齐 Rust `parse_logs_streaming` - 流式解析，每解析出一个事件立即回调
func ParseLogsStreaming(logs []string, signature string, slot uint64, blockTimeUs *int64, callback func(DexEvent)) {
	for _, log := range logs {
		if ev := ParseLogUnified(log, signature, slot, blockTimeUs); ev.Type != "" {
			callback(ev)
		}
	}
}

// ParseTransactionEventsStreaming 对齐 Rust `parse_transaction_events_streaming`
func ParseTransactionEventsStreaming(logs []string, signature string, slot uint64, blockTimeUs *int64, callback func(DexEvent)) {
	ParseLogsStreaming(logs, signature, slot, blockTimeUs, callback)
}

// EventListener 接口对齐 Rust `EventListener` trait
type EventListener interface {
	OnDexEvent(event DexEvent)
}

// ParseTransactionWithListener 对齐 Rust `parse_transaction_with_listener`
func ParseTransactionWithListener(logs []string, signature string, slot uint64, blockTimeUs *int64, listener EventListener) {
	events := ParseLogsOnly(logs, signature, slot, blockTimeUs)
	for _, ev := range events {
		listener.OnDexEvent(ev)
	}
}

// StreamingEventListener 接口对齐 Rust `StreamingEventListener` trait
type StreamingEventListener interface {
	OnDexEventStreaming(event DexEvent)
}

// ParseTransactionWithStreamingListener 对齐 Rust `parse_transaction_with_streaming_listener`
func ParseTransactionWithStreamingListener(logs []string, signature string, slot uint64, blockTimeUs *int64, listener StreamingEventListener) {
	ParseLogsStreaming(logs, signature, slot, blockTimeUs, func(ev DexEvent) {
		listener.OnDexEventStreaming(ev)
	})
}

// ParseLog 对齐 Rust `parse_log` - 带完整 gRPC 元数据字段的日志解析
func ParseLog(log, signature string, slot, txIndex uint64, blockTimeUs *int64, grpcRecvUs int64, isCreatedBuy bool, recentBlockhash string) DexEvent {
	return ParseLogOptimized(log, signature, slot, txIndex, blockTimeUs, grpcRecvUs, nil, isCreatedBuy, recentBlockhash)
}

// WarmupParser 对齐 Rust `warmup_parser`
func WarmupParser() {
	_ = decodeProgramDataLine("Program data: AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
}

// ParseTransactionFromRpc 对齐 Rust `parse_transaction_from_rpc` - 通过 RPC 拉取交易并解析
func ParseTransactionFromRpc(rpcClient RpcClient, signature string, filter EventTypeFilter) ([]DexEvent, *ParseError) {
	return parseTransactionFromRpcImpl(rpcClient, signature, filter)
}

// ParseRpcTransaction 对齐 Rust `parse_rpc_transaction` / `parse_instructions_enhanced` 核心行为：
// 全量账户表（含 ALT LoadedAddresses）、外层 8B + 内层 16B 指令解析、事件合并（merger）、Pump 系账户与 fill_data（is_pump_pool）、再解析 Program logs。
func ParseRpcTransaction(tx *RpcTransactionResponse, signature string, filter EventTypeFilter, grpcRecvUs int64) ([]DexEvent, *ParseError) {
	return parseRpcTransactionImpl(tx, signature, filter, grpcRecvUs)
}

// 注意：以下函数在 accounts.go 和 instructions.go 中实现
// 这里只保留文档注释，实际实现已在其他文件中

// ParseAccountUnified 对齐 Rust `parse_account_unified` - 统一的账户解析入口
// 实现位于 accounts.go

// ParseTokenAccount 对齐 Rust `parse_token_account` - 解析 Token 账户
// 实现位于 accounts.go

// ParseNonceAccount 对齐 Rust `parse_nonce_account` - 解析 Nonce 账户
// 实现位于 accounts.go

// IsNonceAccount 对齐 Rust `is_nonce_account` - 检测是否为 Nonce 账户
// 实现位于 accounts.go

// ParsePumpswapGlobalConfig 对齐 Rust `parse_pumpswap_global_config` - 解析 PumpSwap Global Config
// 实现位于 accounts.go

// ParsePumpswapPool 对齐 Rust `parse_pumpswap_pool` - 解析 PumpSwap Pool
// 实现位于 accounts.go

// ParseInstructionUnified 对齐 Rust `parse_instruction_unified` - 统一的指令解析入口
// 实现位于 instructions.go

// ParsePumpfunInstruction 对齐 Rust `pump::parse_instruction`（外层 Create / CreateV2、CPI Migrate；不解析 Buy/Sell 外层指令）。
// 实现位于 instructions.go

// ParsePumpswapInstruction 对齐 Rust `pump_amm::parse_instruction`（**指令** discriminator，与 Program log 的 Event disc 不同）。
// 实现位于 instructions.go

// ParseMeteoraDammInstruction 对齐 Rust `meteora_damm::parse_instruction`（CPI discriminator 在指令数据 [8..16)）。
// ParseMeteoraDammCpiInstruction 为同上逻辑的纯载荷入口，实现位于 meteora_extra.go / instructions.go
