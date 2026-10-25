package solparser

import (
	"bytes"
	"crypto/sha256"
)

func expectedPumpMethod(tx *RpcTransactionResponse, legacy string) string {
	keys := append([]string{}, tx.Transaction.Message.AccountKeys...)
	if tx.Meta.LoadedAddresses != nil {
		keys = append(keys, tx.Meta.LoadedAddresses.Writable...)
		keys = append(keys, tx.Meta.LoadedAddresses.Readonly...)
	}
	instructions := append([]RpcCompiledInstruction{}, tx.Transaction.Message.Instructions...)
	for _, g := range tx.Meta.InnerInstructions {
		instructions = append(instructions, g.Instructions...)
	}
	for _, method := range []string{legacy + "_v3", legacy + "_v2"} {
		disc := sha256.Sum256([]byte("global:" + method))
		for _, ix := range instructions {
			if int(ix.ProgramIDIndex) < len(keys) && keys[ix.ProgramIDIndex] == PUMPFUN_PROGRAM_ID && len(ix.Data) >= 8 && bytes.Equal(ix.Data[:8], disc[:8]) {
				return method
			}
		}
	}
	return legacy
}
