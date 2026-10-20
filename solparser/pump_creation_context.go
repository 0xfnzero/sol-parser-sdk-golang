package solparser

import "bytes"

// Use execution instructions, not global log hints. No network access.
func fillPumpCreationContext(events []DexEvent, msg *RpcMessage, meta *RpcTransactionMeta) {
	if msg == nil {
		return
	}
	keys := mergeRpcFullAccountKeys(msg.AccountKeys, meta)
	created := make(map[string]bool)
	inspect := func(ix RpcCompiledInstruction) {
		if int(ix.ProgramIDIndex) >= len(keys) || keys[ix.ProgramIDIndex] != PUMPFUN_PROGRAM_ID || len(ix.Data) < 8 || len(ix.Accounts) == 0 {
			return
		}
		if !bytes.Equal(ix.Data[:8], []byte{24, 30, 200, 40, 5, 28, 7, 119}) && !bytes.Equal(ix.Data[:8], []byte{214, 144, 76, 236, 95, 139, 49, 180}) {
			return
		}
		if int(ix.Accounts[0]) < len(keys) {
			created[keys[ix.Accounts[0]]] = true
		}
	}
	for _, ix := range msg.Instructions {
		inspect(ix)
	}
	if meta != nil {
		for _, g := range meta.InnerInstructions {
			for _, ix := range g.Instructions {
				inspect(ix)
			}
		}
	}
	for _, event := range events {
		if trade, ok := event.Data.(*PumpFunTradeEvent); ok {
			trade.IsCreatedBuy = trade.IsBuy && created[trade.Mint]
		}
	}
}
