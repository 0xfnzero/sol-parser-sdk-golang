package solparser

import (
	"bytes"
	"github.com/mr-tron/base58"
	"testing"
)

func TestStonkFunRegistryAtomicReplay(t *testing.T) {
	pk := func(n byte) string { return base58.Encode(bytes.Repeat([]byte{n}, 32)) }
	entry := StonkFunGraduatedPool{pk(1), pk(2), pk(3), pk(4), StonkFunStandardPlatformConfig, base58.Encode(bytes.Repeat([]byte{5}, 64)), 1}
	r, err := NewStonkFunPoolRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	if n, e := r.ObserveMigrations([]StonkFunGraduatedPool{entry}, false); e != nil || n != 0 {
		t.Fatal("failed tx changed registry")
	}
	if n, e := r.ObserveMigrations([]StonkFunGraduatedPool{entry}, true); e != nil || n != 1 {
		t.Fatal(e)
	}
	if n, e := r.ObserveMigrations([]StonkFunGraduatedPool{entry}, true); e != nil || n != 0 {
		t.Fatal("replay not idempotent")
	}
	fresh, conflict := entry, entry
	fresh.Pool = pk(6)
	conflict.MigrationSlot = 2
	if _, e := r.ObserveMigrations([]StonkFunGraduatedPool{fresh, conflict}, true); e == nil {
		t.Fatal("conflict accepted")
	}
	if len(r.Pools()) != 1 {
		t.Fatal("batch was not atomic")
	}
}
func TestDynamicTickSnapshotValidation(t *testing.T) {
	d := make([]byte, 148)
	copy(d, []byte{17, 216, 246, 142, 225, 199, 218, 56})
	a := &AccountData{Pubkey: base58.Encode(bytes.Repeat([]byte{1}, 32)), Owner: ORCA_WHIRLPOOL_PROGRAM_ID, Data: d, Lamports: 1}
	ev := ParseLiquidityAccount(a, EventMetadata{})
	if ev.Type != EventTypeAccountLiquiditySnapshot {
		t.Fatal("missing snapshot")
	}
	d[44] = 1
	if ParseLiquidityAccount(a, EventMetadata{}).Type != "" {
		t.Fatal("inconsistent tick bitmap accepted")
	}
	d[60] = 1
	if ParseLiquidityAccount(a, EventMetadata{}).Type != "" {
		t.Fatal("truncated initialized tick accepted")
	}
	snapshot := ev.Data.(*LiquidityAccountSnapshotEvent)
	if snapshot.Data[44] != 0 {
		t.Fatal("snapshot aliases stream buffer")
	}
}
func TestRawClosedAccountSnapshot(t *testing.T) {
	update := &SubscribeUpdateAccount{Slot: 42, IsStartup: true, Account: &SubscribeUpdateAccountInfo{Pubkey: bytes.Repeat([]byte{1}, 32), Owner: bytes.Repeat([]byte{2}, 32), WriteVersion: 99, Data: []byte{}, Lamports: 0}}
	ev := parseAccountDexEvent(update, nil, 1, nil, true)
	raw, ok := ev.Data.(*RawAccountSnapshotEvent)
	if !ok || raw.Metadata.Slot != 42 || raw.WriteVersion != 99 || !raw.IsStartup || raw.Account.Lamports != 0 {
		t.Fatal("closed account provenance lost")
	}
}
