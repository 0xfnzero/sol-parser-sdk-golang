package solparser

// RPC 路径下对齐 Rust `account_dispatcher::fill_accounts_with_owned_keys` 与 `common_filler::fill_data` 的 **Pump 系子集**。

func buildRpcProgramInvokes(msg *RpcMessage, meta *RpcTransactionMeta) map[string][][2]int32 {
	m := make(map[string][][2]int32)
	if msg == nil {
		return m
	}
	fullKeys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
	for i, ix := range msg.Instructions {
		if int(ix.ProgramIDIndex) < len(fullKeys) {
			pid := fullKeys[ix.ProgramIDIndex]
			m[pid] = append(m[pid], [2]int32{int32(i), -1})
		}
	}
	if meta != nil {
		for _, g := range meta.InnerInstructions {
			for j, ix := range g.Instructions {
				if int(ix.ProgramIDIndex) < len(fullKeys) {
					pid := fullKeys[ix.ProgramIDIndex]
					m[pid] = append(m[pid], [2]int32{int32(g.Index), int32(j)})
				}
			}
		}
	}
	return m
}

func getRpcInstructionData(msg *RpcMessage, meta *RpcTransactionMeta, inv [2]int32) []byte {
	if meta == nil || msg == nil {
		return nil
	}
	if inv[1] >= 0 {
		for _, g := range meta.InnerInstructions {
			if g.Index == uint32(inv[0]) && int(inv[1]) < len(g.Instructions) {
				return g.Instructions[inv[1]].Data
			}
		}
		return nil
	}
	if int(inv[0]) >= 0 && int(inv[0]) < len(msg.Instructions) {
		return msg.Instructions[inv[0]].Data
	}
	return nil
}

// fillRpcDexEventsPump 为 PumpFun Trade 系补账户字段；为 PumpSwap Buy/Sell 填 `is_pump_pool`（Rust `common_filler::fill_data`）。
func fillRpcDexEventsPump(events []DexEvent, msg *RpcMessage, meta *RpcTransactionMeta) {
	if len(events) == 0 || msg == nil {
		return
	}
	invokes := buildRpcProgramInvokes(msg, meta)
	feesInv := invokes[GrpcPumpSwapFeesProgramID]

	for i := range events {
		fillRpcOneEventPump(&events[i], msg, meta, invokes, feesInv)
	}
}

func fillRpcOneEventPump(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, invokes map[string][][2]int32, feesInv [][2]int32) {
	switch ev.Type {
	case EventTypePumpFunTrade, EventTypePumpFunBuy, EventTypePumpFunSell, EventTypePumpFunBuyExactSolIn:
		pumpInv := invokes[PUMPFUN_PROGRAM_ID]
		get := rpcAccountGetter(msg, meta, pumpInv)
		if get == nil {
			return
		}
		tr, ok := ev.Data.(*PumpFunTradeEvent)
		if !ok {
			return
		}
		if tr.User == "" || tr.User == zeroPubkey {
			tr.User = get(6)
		}
		if tr.BondingCurve == "" || tr.BondingCurve == zeroPubkey {
			tr.BondingCurve = get(3)
		}
		if tr.AssociatedBondingCurve == "" || tr.AssociatedBondingCurve == zeroPubkey {
			tr.AssociatedBondingCurve = get(4)
		}
		if tr.CreatorVault == "" || tr.CreatorVault == zeroPubkey {
			if tr.IsBuy {
				tr.CreatorVault = get(9)
			} else {
				tr.CreatorVault = get(8)
			}
		}
		if tr.TokenProgram == "" || tr.TokenProgram == zeroPubkey {
			if tr.IsBuy {
				tr.TokenProgram = get(8)
			} else {
				tr.TokenProgram = get(9)
			}
		}
	case EventTypePumpSwapBuy:
		fillPumpSwapIsPumpPool(ev, msg, meta, feesInv)
	case EventTypePumpSwapSell:
		fillPumpSwapIsPumpPool(ev, msg, meta, feesInv)
	}
}

func fillPumpSwapIsPumpPool(ev *DexEvent, msg *RpcMessage, meta *RpcTransactionMeta, feesInv [][2]int32) {
	if len(feesInv) == 0 {
		return
	}
	last := feesInv[len(feesInv)-1]
	data := getRpcInstructionData(msg, meta, last)
	if len(data) < 10 {
		return
	}
	isPump, ok := readBool(data, 9)
	if !ok {
		return
	}
	switch ev.Type {
	case EventTypePumpSwapBuy:
		if b, ok := ev.Data.(*PumpSwapBuyEvent); ok {
			b.IsPumpPool = isPump
		}
	case EventTypePumpSwapSell:
		if s, ok := ev.Data.(*PumpSwapSellEvent); ok {
			s.IsPumpPool = isPump
		}
	}
}

// rpcAccountGetter 对齐 Rust `find_instruction_invoke` + 账户索引解析：选取账户数最多的一次 invoke。
func rpcAccountGetter(msg *RpcMessage, meta *RpcTransactionMeta, list [][2]int32) func(int) string {
	if len(list) == 0 || msg == nil {
		return nil
	}
	fullKeys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
	best := list[0]
	bestN := -1
	for _, inv := range list {
		n := rpcInvokeAccountLen(msg, meta, inv)
		if n > bestN {
			bestN = n
			best = inv
		}
	}
	if best[1] >= 0 && meta == nil {
		return nil
	}
	return func(i int) string {
		var accounts []byte
		if best[1] >= 0 {
			for _, g := range meta.InnerInstructions {
				if g.Index == uint32(best[0]) && int(best[1]) < len(g.Instructions) {
					accounts = g.Instructions[best[1]].Accounts
					break
				}
			}
		} else if int(best[0]) < len(msg.Instructions) {
			accounts = msg.Instructions[best[0]].Accounts
		}
		if i < 0 || i >= len(accounts) {
			return ""
		}
		idx := int(accounts[i])
		if idx < len(fullKeys) {
			return fullKeys[idx]
		}
		return ""
	}
}

func rpcInvokeAccountLen(msg *RpcMessage, meta *RpcTransactionMeta, inv [2]int32) int {
	if meta == nil {
		if int(inv[0]) < len(msg.Instructions) {
			return len(msg.Instructions[inv[0]].Accounts)
		}
		return 0
	}
	if inv[1] >= 0 {
		for _, g := range meta.InnerInstructions {
			if g.Index == uint32(inv[0]) && int(inv[1]) < len(g.Instructions) {
				return len(g.Instructions[inv[1]].Accounts)
			}
		}
		return 0
	}
	if int(inv[0]) < len(msg.Instructions) {
		return len(msg.Instructions[inv[0]].Accounts)
	}
	return 0
}
