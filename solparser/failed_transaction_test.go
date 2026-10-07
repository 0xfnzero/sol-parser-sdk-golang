package solparser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
	"github.com/mr-tron/base58"
)

func TestFailedMainnetTransactionsDoNotEmitRolledBackEvents(t *testing.T) {
	paths, err := filepath.Glob("testdata/pumpfun_failed_mainnet/*.json")
	if err != nil || len(paths) != 8 {
		t.Fatalf("failed fixtures: %d, %v", len(paths), err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var raw struct {
				Slot uint64
				Meta struct {
					Err             any
					LogMessages     []string
					LoadedAddresses RpcLoadedAddresses
				}
				Transaction struct {
					Signatures []string
					Message    struct {
						AccountKeys     []string
						RecentBlockhash string
						Instructions    []struct {
							ProgramIDIndex uint32
							Accounts       []int
							Data           string
						}
					}
				}
			}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatal(err)
			}
			if raw.Meta.Err == nil {
				t.Fatal("expected actual mainnet failure")
			}
			msg := &RpcMessage{AccountKeys: raw.Transaction.Message.AccountKeys, RecentBlockhash: raw.Transaction.Message.RecentBlockhash}
			for _, ix := range raw.Transaction.Message.Instructions {
				var decoded []byte
				var err error
				if ix.Data != "" {
					decoded, err = base58.Decode(ix.Data)
				}
				if err != nil {
					t.Fatal(err)
				}
				accounts := make([]byte, len(ix.Accounts))
				for i, a := range ix.Accounts {
					accounts[i] = byte(a)
				}
				msg.Instructions = append(msg.Instructions, RpcCompiledInstruction{ProgramIDIndex: ix.ProgramIDIndex, Accounts: accounts, Data: decoded})
			}
			tx := &RpcTransactionResponse{Slot: raw.Slot, Meta: &RpcTransactionMeta{Err: raw.Meta.Err, LogMessages: raw.Meta.LogMessages, LoadedAddresses: &raw.Meta.LoadedAddresses}, Transaction: &RpcTransaction{Message: msg, Signatures: raw.Transaction.Signatures}}
			events, pe := ParseRpcTransaction(tx, raw.Transaction.Signatures[0], nil, 0)
			if pe != nil || len(events) != 0 {
				t.Fatalf("rolled-back events: %d, %v", len(events), pe)
			}
			grpcMeta, _, err := ConvertRpcToGrpc(tx)
			if err != nil || grpcMeta.Err == nil {
				t.Fatalf("RPC conversion lost failed status: %v", err)
			}
			// The same instruction/log data would emit events without its failed status.
			tx.Meta.Err = nil
			events, pe = ParseRpcTransaction(tx, raw.Transaction.Signatures[0], nil, 0)
			if pe != nil || len(events) == 0 {
				t.Fatalf("invalid regression trigger: %d, %v", len(events), pe)
			}
		})
	}
}

func TestGeyserFailedStatusSurvivesAdapter(t *testing.T) {
	for _, payload := range [][]byte{nil, {1}} {
		info := &SubscribeUpdateTransactionInfo{
			Transaction: &pb.Transaction{Message: &pb.Message{}},
			Meta:        &pb.TransactionStatusMeta{Err: &pb.TransactionError{Err: payload}},
		}
		rpc, err := SubscribeUpdateInfoToRpc(1, info)
		if err != nil || rpc.Meta.Err == nil {
			t.Fatalf("lost failed status: %v", err)
		}
		events, pe := ParseSubscribeTransaction(1, info, nil, 0)
		if pe != nil || len(events) != 0 {
			t.Fatalf("failed gRPC events: %d, %v", len(events), pe)
		}
	}
}
