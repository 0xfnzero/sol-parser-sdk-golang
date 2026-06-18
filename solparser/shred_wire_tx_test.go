package solparser

import (
	"encoding/binary"
	"testing"

	solana "github.com/gagliardetto/solana-go"
)

func shredWireTestPubkey(seed byte) solana.PublicKey {
	b := make([]byte, 32)
	for i := range b {
		b[i] = seed
	}
	return solana.PublicKeyFromBytes(b)
}

func shredWireTestRawTx(t *testing.T, keys []solana.PublicKey, ixs ...solana.CompiledInstruction) []byte {
	t.Helper()
	tx := solana.Transaction{
		Message: solana.Message{
			AccountKeys:     keys,
			RecentBlockhash: solana.Hash{},
			Instructions:    ixs,
		},
	}
	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal tx: %v", err)
	}
	return raw
}

func shredWireClmmSwapInstruction() []byte {
	data := make([]byte, 41)
	binary.LittleEndian.PutUint64(data[:8], instrClmmSwap)
	binary.LittleEndian.PutUint64(data[24:32], 123)
	data[40] = 1
	return data
}

func TestDexEventsFromShredTransactionWireDefaultsAltLoadedAccounts(t *testing.T) {
	pool := shredWireTestPubkey(1)
	raw := shredWireTestRawTx(
		t,
		[]solana.PublicKey{solana.MustPublicKeyFromBase58(RAYDIUM_CLMM_PROGRAM_ID), pool},
		solana.CompiledInstruction{
			ProgramIDIndex: 0,
			Accounts:       []uint16{99, 99, 1},
			Data:           solana.Base58(shredWireClmmSwapInstruction()),
		},
	)

	events := DexEventsFromShredTransactionWire(raw, "sig", 1, 0, nil, 10, nil)
	if len(events) != 1 || events[0].Type != EventTypeRaydiumClmmSwap {
		t.Fatalf("expected one RaydiumClmmSwap event, got %+v", events)
	}
	swap, ok := events[0].Data.(*RaydiumClmmSwapEvent)
	if !ok {
		t.Fatalf("expected RaydiumClmmSwapEvent, got %T", events[0].Data)
	}
	if swap.PoolState != pool.String() || swap.Sender != zeroPubkey {
		t.Fatalf("unexpected CLMM ALT fallback accounts: %+v", swap)
	}
}

func TestDexEventsFromShredTransactionWireFallbacksAltLoadedProgramID(t *testing.T) {
	global := shredWireTestPubkey(2)
	mint := shredWireTestPubkey(3)
	raw := shredWireTestRawTx(
		t,
		[]solana.PublicKey{global, mint},
		solana.CompiledInstruction{
			ProgramIDIndex: 99,
			Accounts:       []uint16{0, 1},
			Data:           solana.Base58(pumpfunV2Instruction(instrPumpOuterBuyV2, 123, 456)),
		},
	)

	events := DexEventsFromShredTransactionWire(raw, "sig", 1, 0, nil, 10, &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypePumpFunBuy}})
	if len(events) != 1 || events[0].Type != EventTypePumpFunBuy {
		t.Fatalf("expected one PumpFunBuy event, got %+v", events)
	}
	trade, ok := events[0].Data.(*PumpFunTradeEvent)
	if !ok {
		t.Fatalf("expected PumpFunTradeEvent, got %T", events[0].Data)
	}
	if trade.Mint != mint.String() || trade.TokenAmount != 123 || trade.SolAmount != 456 {
		t.Fatalf("unexpected PumpFun ALT program fallback event: %+v", trade)
	}
}

func TestDexEventsFromShredTransactionWireSkipsAccountOnlyFilter(t *testing.T) {
	raw := shredWireTestRawTx(
		t,
		[]solana.PublicKey{solana.MustPublicKeyFromBase58(RAYDIUM_CLMM_PROGRAM_ID), shredWireTestPubkey(1)},
		solana.CompiledInstruction{
			ProgramIDIndex: 0,
			Accounts:       []uint16{1},
			Data:           solana.Base58(shredWireClmmSwapInstruction()),
		},
	)

	events := DexEventsFromShredTransactionWire(
		raw,
		"sig",
		1,
		0,
		nil,
		10,
		&IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeAccountRaydiumClmmPoolState}},
	)
	if len(events) != 0 {
		t.Fatalf("expected account-only Shred filter to skip instruction parsing, got %+v", events)
	}
}
