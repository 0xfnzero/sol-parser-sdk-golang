package solparser

import (
	"encoding/binary"
	"testing"
)

func testAccount(owner string, data []byte) *AccountData {
	return &AccountData{
		Pubkey:     "account",
		Executable: false,
		Lamports:   1,
		Owner:      owner,
		RentEpoch:  0,
		Data:       data,
	}
}

func testPk(seed byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = seed
	}
	return out
}

func appendU16(out []byte, v uint16) []byte {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	return append(out, b[:]...)
}

func appendU32(out []byte, v uint32) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	return append(out, b[:]...)
}

func appendU64(out []byte, v uint64) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return append(out, b[:]...)
}

func clmmAmmConfigAccountData() []byte {
	out := append([]byte{}, clmmAmmConfigDisc...)
	out = append(out, 9)
	out = appendU16(out, 7)
	out = append(out, testPk(1)...)
	out = appendU32(out, 111)
	out = appendU32(out, 222)
	out = appendU16(out, 64)
	out = appendU32(out, 333)
	out = appendU32(out, 444)
	out = append(out, testPk(2)...)
	for _, v := range []uint64{1, 2, 3} {
		out = appendU64(out, v)
	}
	return out
}

func cpmmPoolStateAccountData() []byte {
	out := append([]byte{}, clmmPoolStateDisc...)
	for seed := byte(1); seed <= 10; seed++ {
		out = append(out, testPk(seed)...)
	}
	out = append(out, 11, 1, 9, 6, 6)
	for _, v := range []uint64{100, 1, 2, 3, 4, 123456, 99} {
		out = appendU64(out, v)
	}
	out = append(out, 2, 1)
	out = append(out, make([]byte, 6)...)
	out = appendU64(out, 5)
	out = appendU64(out, 6)
	for v := uint64(0); v < 28; v++ {
		out = appendU64(out, v)
	}
	return out
}

func orcaFeeTierAccountData() []byte {
	out := append([]byte{}, orcaFeeTierDisc...)
	out = append(out, testPk(9)...)
	out = appendU16(out, 128)
	out = appendU16(out, 500)
	return out
}

func TestRaydiumClmmAccountParserReadsAmmConfig(t *testing.T) {
	ev := ParseAccountUnified(
		testAccount(RAYDIUM_CLMM_PROGRAM_ID, clmmAmmConfigAccountData()),
		EventMetadata{Signature: "sig", Slot: 1},
		EventTypeFilterIncludeOnly([]EventType{EventTypeAccountRaydiumClmmAmmConfig}),
	)
	if ev.Type != EventTypeAccountRaydiumClmmAmmConfig {
		t.Fatalf("expected CLMM AMM config account, got %q", ev.Type)
	}
	cfg := ev.Data.(*RaydiumClmmAmmConfigAccountEvent).AmmConfig
	if cfg.Owner != Base58Encode(testPk(1)) || cfg.TickSpacing != 64 || len(cfg.Padding) != 3 || cfg.Padding[2] != 3 {
		t.Fatalf("unexpected CLMM config: %+v", cfg)
	}
}

func TestRaydiumCpmmAccountParserScopesSharedDiscriminators(t *testing.T) {
	account := testAccount(RAYDIUM_CPMM_PROGRAM_ID, cpmmPoolStateAccountData())
	ev := ParseAccountUnified(
		account,
		EventMetadata{Signature: "sig", Slot: 1},
		EventTypeFilterIncludeOnly([]EventType{EventTypeAccountRaydiumCpmmPoolState}),
	)
	if ev.Type != EventTypeAccountRaydiumCpmmPoolState {
		t.Fatalf("expected CPMM pool state account, got %q", ev.Type)
	}
	state := ev.Data.(*RaydiumCpmmPoolStateAccountEvent).PoolState
	if state.AuthBump != 11 || state.LpSupply != 100 || !state.EnableCreatorFee {
		t.Fatalf("unexpected CPMM pool state: %+v", state)
	}

	ev = ParseAccountUnified(
		account,
		EventMetadata{Signature: "sig", Slot: 1},
		EventTypeFilterIncludeOnly([]EventType{EventTypeAccountRaydiumClmmPoolState}),
	)
	if ev.Type != "" {
		t.Fatalf("CPMM account should not satisfy CLMM account filter, got %q", ev.Type)
	}
}

func TestOrcaWhirlpoolAccountParserReadsFeeTier(t *testing.T) {
	ev := ParseAccountUnified(
		testAccount(ORCA_WHIRLPOOL_PROGRAM_ID, orcaFeeTierAccountData()),
		EventMetadata{Signature: "sig", Slot: 1},
		EventTypeFilterIncludeOnly([]EventType{EventTypeAccountOrcaFeeTier}),
	)
	if ev.Type != EventTypeAccountOrcaFeeTier {
		t.Fatalf("expected Orca fee tier account, got %q", ev.Type)
	}
	feeTier := ev.Data.(*OrcaFeeTierAccountEvent).FeeTier
	if feeTier.WhirlpoolsConfig != Base58Encode(testPk(9)) || feeTier.TickSpacing != 128 || feeTier.DefaultFeeRate != 500 {
		t.Fatalf("unexpected Orca fee tier: %+v", feeTier)
	}
}
