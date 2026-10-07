//go:build ignore

// Replay finalized mainnet RPC responses through instruction, log, RPC and
// reconstructed Yellowstone protobuf paths. This never submits transactions.
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
	sp "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"github.com/mr-tron/base58"
	"os"
	"path/filepath"
)

type ix struct {
	ProgramIDIndex uint32
	Accounts       []int
	Data           string
	StackHeight    *uint32
}
type group struct {
	Index        uint32
	Instructions []ix
}
type raw struct {
	Slot      uint64
	BlockTime *int64
	Meta      struct {
		Err                       json.RawMessage
		Fee                       uint64
		PreBalances, PostBalances []uint64
		LogMessages               []string
		InnerInstructions         []group
		LoadedAddresses           struct{ Writable, Readonly []string }
	}
	Transaction struct {
		Signatures []string
		Message    struct {
			AccountKeys     []string
			RecentBlockhash string
			Header          sp.RpcMessageHeader
			Instructions    []ix
		}
	}
}

func dec(s string) []byte {
	if s == "" {
		return nil
	}
	b, e := base58.Decode(s)
	if e != nil {
		panic(e)
	}
	return b
}
func indices(a []int) []byte {
	b := make([]byte, len(a))
	for i, v := range a {
		if v < 0 || v > 255 {
			panic("invalid account index")
		}
		b[i] = byte(v)
	}
	return b
}
func rpcix(i ix) sp.RpcCompiledInstruction {
	return sp.RpcCompiledInstruction{ProgramIDIndex: i.ProgramIDIndex, Accounts: indices(i.Accounts), Data: dec(i.Data), StackHeight: i.StackHeight}
}
func creates(evs []sp.DexEvent) []sp.DexEvent {
	out := []sp.DexEvent{}
	for _, e := range evs {
		if e.Type == sp.EventTypePumpFunCreate || e.Type == sp.EventTypePumpFunCreateV2 {
			out = append(out, e)
		}
	}
	return out
}
func main() {
	if len(os.Args) != 2 {
		panic("usage: go run replay.go transactions-directory")
	}
	paths, e := filepath.Glob(filepath.Join(os.Args[1], "*.json"))
	if e != nil {
		panic(e)
	}
	reports := []map[string]any{}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		var r raw
		if e = json.Unmarshal(b, &r); e != nil {
			panic(e)
		}
		sig := r.Transaction.Signatures[0]
		m := r.Transaction.Message
		tx := &sp.RpcTransactionResponse{Slot: r.Slot, BlockTime: r.BlockTime, Transaction: &sp.RpcTransaction{Signatures: r.Transaction.Signatures, Message: &sp.RpcMessage{AccountKeys: m.AccountKeys, Header: &m.Header, RecentBlockhash: m.RecentBlockhash}}, Meta: &sp.RpcTransactionMeta{Fee: r.Meta.Fee, PreBalances: r.Meta.PreBalances, PostBalances: r.Meta.PostBalances, LogMessages: r.Meta.LogMessages, LoadedAddresses: &sp.RpcLoadedAddresses{Writable: r.Meta.LoadedAddresses.Writable, Readonly: r.Meta.LoadedAddresses.Readonly}}}
		pm := &pb.Message{RecentBlockhash: dec(m.RecentBlockhash), Header: &pb.MessageHeader{NumRequiredSignatures: m.Header.NumRequiredSignatures, NumReadonlySignedAccounts: m.Header.NumReadonlySignedAccounts, NumReadonlyUnsignedAccounts: m.Header.NumReadonlyUnsignedAccounts}}
		for _, k := range m.AccountKeys {
			pm.AccountKeys = append(pm.AccountKeys, dec(k))
		}
		for _, i := range m.Instructions {
			tx.Transaction.Message.Instructions = append(tx.Transaction.Message.Instructions, rpcix(i))
			pm.Instructions = append(pm.Instructions, &pb.CompiledInstruction{ProgramIdIndex: i.ProgramIDIndex, Accounts: indices(i.Accounts), Data: dec(i.Data)})
		}
		meta := &pb.TransactionStatusMeta{Fee: r.Meta.Fee, PreBalances: r.Meta.PreBalances, PostBalances: r.Meta.PostBalances, LogMessages: r.Meta.LogMessages}
		for _, k := range r.Meta.LoadedAddresses.Writable {
			meta.LoadedWritableAddresses = append(meta.LoadedWritableAddresses, dec(k))
		}
		for _, k := range r.Meta.LoadedAddresses.Readonly {
			meta.LoadedReadonlyAddresses = append(meta.LoadedReadonlyAddresses, dec(k))
		}
		allix := append([]ix{}, m.Instructions...)
		for _, g := range r.Meta.InnerInstructions {
			rg := sp.RpcInnerInstructionGroup{Index: g.Index}
			pg := &pb.InnerInstructions{Index: g.Index}
			for _, i := range g.Instructions {
				rg.Instructions = append(rg.Instructions, rpcix(i))
				pg.Instructions = append(pg.Instructions, &pb.InnerInstruction{ProgramIdIndex: i.ProgramIDIndex, Accounts: indices(i.Accounts), Data: dec(i.Data), StackHeight: i.StackHeight})
			}
			tx.Meta.InnerInstructions = append(tx.Meta.InnerInstructions, rg)
			meta.InnerInstructions = append(meta.InnerInstructions, pg)
			allix = append(allix, g.Instructions...)
		}
		info := &sp.SubscribeUpdateTransactionInfo{Signature: dec(sig), Transaction: &pb.Transaction{Message: pm}, Meta: meta}
		for _, s := range r.Transaction.Signatures {
			info.Transaction.Signatures = append(info.Transaction.Signatures, dec(s))
		}
		keys := append(append(append([]string{}, m.AccountKeys...), r.Meta.LoadedAddresses.Writable...), r.Meta.LoadedAddresses.Readonly...)
		instruction := []map[string]any{}
		for _, i := range allix {
			if int(i.ProgramIDIndex) >= len(keys) || keys[i.ProgramIDIndex] != sp.PUMPFUN_PROGRAM_ID {
				continue
			}
			data := dec(i.Data)
			if len(data) < 8 {
				continue
			}
			d := binary.LittleEndian.Uint64(data[:8])
			if d != binary.LittleEndian.Uint64([]byte{24, 30, 200, 40, 5, 28, 7, 119}) && d != binary.LittleEndian.Uint64([]byte{214, 144, 76, 236, 95, 139, 49, 180}) {
				continue
			}
			a := []string{}
			for _, v := range i.Accounts {
				a = append(a, keys[v])
			}
			ev := sp.ParsePumpfunInstruction(data, a, sig, r.Slot, 0, nil, 0)
			instruction = append(instruction, map[string]any{"account_count": len(a), "accounts": a, "event": ev})
		}
		if len(instruction) == 0 {
			continue
		}
		logs := creates(sp.ParseLogsOnly(r.Meta.LogMessages, sig, r.Slot, nil))
		full, pe := sp.ParseRpcTransaction(tx, sig, nil, 0)
		grpc, ge := sp.ParseSubscribeTransaction(r.Slot, info, nil, 0)
		report := map[string]any{"signature": sig, "slot": r.Slot, "block_time": r.BlockTime, "transaction_error": r.Meta.Err, "instructions": instruction, "logs": logs, "rpc": creates(full), "grpc_reconstructed": creates(grpc), "rpc_error": pe, "grpc_error": ge, "loaded_account_count": len(keys) - len(m.AccountKeys)}
		reports = append(reports, report)
	}
	out, e := json.MarshalIndent(reports, "", "  ")
	if e != nil {
		panic(e)
	}
	fmt.Println(string(out))
}
