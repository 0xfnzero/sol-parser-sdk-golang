package solparser

import (
	"encoding/base64"
	"encoding/binary"
	"testing"
)

func pushPubkeyBytes(b *[]byte, fill byte) string {
	key := make([]byte, 32)
	for i := range key {
		key[i] = fill
	}
	*b = append(*b, key...)
	s, _ := readPubkey(key, 0)
	return s
}

func programDataLogForTest(disc uint64, payload []byte) string {
	buf := make([]byte, 8+len(payload))
	binary.LittleEndian.PutUint64(buf[:8], disc)
	copy(buf[8:], payload)
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func currentSwap2Payload(t *testing.T) ([]byte, string) {
	t.Helper()
	data := make([]byte, 0, 180)
	pool := pushPubkeyBytes(&data, 1)
	data = append(data, 1, 2, 1) // trade_direction, collect_fee_mode, has_referral
	putU64 := func(v uint64) {
		var tmp [8]byte
		binary.LittleEndian.PutUint64(tmp[:], v)
		data = append(data, tmp[:]...)
	}
	putU64(1000) // amount_0
	putU64(900)  // amount_1
	data = append(data, 0) // exact-in
	putU64(1000)
	putU64(990)
	putU64(5)
	putU64(880)
	var u128 [16]byte
	binary.LittleEndian.PutUint64(u128[:8], 123456)
	data = append(data, u128[:]...)
	putU64(3) // claiming
	putU64(2) // protocol
	putU64(1) // compounding
	putU64(4)
	putU64(10)
	putU64(11)
	putU64(870)
	putU64(1_725_000_000)
	putU64(50_000)
	putU64(60_000)
	if len(data) != 180 {
		t.Fatalf("swap2 payload len = %d, want 180", len(data))
	}
	return data, pool
}

func TestParseDammSwap2CurrentLayoutAndModes(t *testing.T) {
	payload, pool := currentSwap2Payload(t)
	meta := EventMetadata{Slot: compoundingFeeLayoutActivationSlot}
	ev := parseDammSwap2(payload, meta)
	if ev.Type != EventTypeMeteoraDammV2Swap {
		t.Fatalf("type = %q", ev.Type)
	}
	s := ev.Data.(*MeteoraDammV2SwapEvent)
	if s.Pool != pool || s.AmountIn != 1000 || s.MinimumAmountOut != 900 {
		t.Fatalf("unexpected swap2 core fields: %+v", s)
	}
	if s.ClaimingFee != 3 || s.CompoundingFee != 1 || s.LpFee != 4 || s.PartnerFee != 1 {
		t.Fatalf("fee fields: claiming=%d compounding=%d lp=%d partner=%d",
			s.ClaimingFee, s.CompoundingFee, s.LpFee, s.PartnerFee)
	}
	if s.ReserveAAmount != 50_000 || s.ReserveBAmount != 60_000 {
		t.Fatalf("reserves = %d/%d", s.ReserveAAmount, s.ReserveBAmount)
	}

	legacy := parseDammSwap2(payload, EventMetadata{Slot: compoundingFeeLayoutActivationSlot - 1})
	ls := legacy.Data.(*MeteoraDammV2SwapEvent)
	if ls.LpFee != 3 || ls.PartnerFee != 1 || ls.ClaimingFee != 0 || ls.CompoundingFee != 0 {
		t.Fatalf("legacy fees: %+v", ls)
	}

	const swapModeOffset = 32 + 1 + 1 + 1 + 8 + 8
	for _, tc := range []struct {
		mode       uint8
		wantIn     uint64
		wantMinOut uint64
	}{
		{0, 1000, 900},
		{1, 1000, 900},
		{2, 900, 1000},
	} {
		p := append([]byte(nil), payload...)
		p[swapModeOffset] = tc.mode
		got := parseDammSwap2(p, meta).Data.(*MeteoraDammV2SwapEvent)
		if got.AmountIn != tc.wantIn || got.MinimumAmountOut != tc.wantMinOut {
			t.Fatalf("mode %d: got in/out %d/%d want %d/%d", tc.mode, got.AmountIn, got.MinimumAmountOut, tc.wantIn, tc.wantMinOut)
		}
	}
	bad := append([]byte(nil), payload...)
	bad[swapModeOffset] = 3
	if parseDammSwap2(bad, meta).Type != "" {
		t.Fatalf("invalid swap mode should fail")
	}
}

func TestParseDammLiquidityChangeAddRemove(t *testing.T) {
	for _, tc := range []struct {
		changeType uint8
		wantType   EventType
	}{
		{0, EventTypeMeteoraDammV2AddLiquidity},
		{1, EventTypeMeteoraDammV2RemoveLiquidity},
	} {
		data := make([]byte, 0, 177)
		pool := pushPubkeyBytes(&data, 2)
		pos := pushPubkeyBytes(&data, 3)
		owner := pushPubkeyBytes(&data, 4)
		putU64 := func(v uint64) {
			var tmp [8]byte
			binary.LittleEndian.PutUint64(tmp[:], v)
			data = append(data, tmp[:]...)
		}
		putU64(101)
		putU64(202)
		putU64(111)
		putU64(222)
		putU64(1001)
		putU64(2002)
		var u128 [16]byte
		binary.LittleEndian.PutUint64(u128[:8], 303)
		data = append(data, u128[:]...)
		putU64(404)
		putU64(505)
		data = append(data, tc.changeType)
		log := programDataLogForTest(discDammLiquidityChange, data)
		ev := ParseLogOptimizedWithProgramID(log, "sig", 1, 0, nil, 1,
			EventTypeFilterIncludeOnly([]EventType{tc.wantType}), false, "", METEORA_DAMM_V2_PROGRAM_ID)
		if ev.Type != tc.wantType {
			t.Fatalf("change_type %d: got %q want %q", tc.changeType, ev.Type, tc.wantType)
		}
		switch d := ev.Data.(type) {
		case *MeteoraDammV2AddLiquidityEvent:
			if d.Pool != pool || d.Position != pos || d.Owner != owner || d.ReserveAAmount != 1001 {
				t.Fatalf("add fields: %+v", d)
			}
		case *MeteoraDammV2RemoveLiquidityEvent:
			if d.Pool != pool || d.Position != pos || d.Owner != owner || d.ReserveBAmount != 2002 {
				t.Fatalf("remove fields: %+v", d)
			}
		default:
			t.Fatalf("unexpected data %T", ev.Data)
		}
		reject := EventTypeMeteoraDammV2RemoveLiquidity
		if tc.changeType == 1 {
			reject = EventTypeMeteoraDammV2AddLiquidity
		}
		if ParseLogOptimizedWithProgramID(log, "sig", 1, 0, nil, 1,
			EventTypeFilterIncludeOnly([]EventType{reject}), false, "", METEORA_DAMM_V2_PROGRAM_ID).Type != "" {
			t.Fatalf("non-matching filter must reject change_type %d", tc.changeType)
		}
	}
}

func TestParseDammNewEventsDelegateConfig(t *testing.T) {
	withDel := make([]byte, 0, 101)
	pos := pushPubkeyBytes(&withDel, 11)
	owner := pushPubkeyBytes(&withDel, 12)
	var p32 [4]byte
	binary.LittleEndian.PutUint32(p32[:], 0x00ff)
	withDel = append(withDel, p32[:]...)
	withDel = append(withDel, 1)
	del := pushPubkeyBytes(&withDel, 13)
	ev := parseDammUpdateDelegatePermission(withDel, EventMetadata{})
	if ev.Type != EventTypeMeteoraDammV2UpdateDelegatePermission {
		t.Fatalf("type=%q", ev.Type)
	}
	u := ev.Data.(*MeteoraDammV2UpdateDelegatePermissionEvent)
	if u.Position != pos || u.Owner != owner || u.Permission != 0x00ff || u.Delegate == nil || *u.Delegate != del {
		t.Fatalf("delegate event: %+v", u)
	}

	cleared := make([]byte, 0, 69)
	pushPubkeyBytes(&cleared, 21)
	pushPubkeyBytes(&cleared, 22)
	cleared = append(cleared, 0, 0, 0, 0, 0)
	clearedEv := parseDammUpdateDelegatePermission(cleared, EventMetadata{})
	if clearedEv.Data.(*MeteoraDammV2UpdateDelegatePermissionEvent).Delegate != nil {
		t.Fatalf("cleared delegate should be nil")
	}

	dead := make([]byte, 0, 72)
	pool := pushPubkeyBytes(&dead, 31)
	mint := pushPubkeyBytes(&dead, 32)
	var amt [8]byte
	binary.LittleEndian.PutUint64(amt[:], 777)
	dead = append(dead, amt[:]...)
	w := parseDammWithdrawDeadLiquidityReward(dead, EventMetadata{}).Data.(*MeteoraDammV2WithdrawDeadLiquidityRewardEvent)
	if w.Pool != pool || w.RewardMint != mint || w.Amount != 777 {
		t.Fatalf("withdraw: %+v", w)
	}

	cfg := make([]byte, 0, 200)
	cfg = append(cfg, dammBytesRepeat(7, 27)...)
	var u16b [2]byte
	binary.LittleEndian.PutUint16(u16b[:], 250)
	cfg = append(cfg, u16b[:]...)
	cfg = append(cfg, 0, 0) // padding, no dynamic fee
	vault := pushPubkeyBytes(&cfg, 41)
	auth := pushPubkeyBytes(&cfg, 42)
	cfg = append(cfg, 1)
	var u128 [16]byte
	binary.LittleEndian.PutUint64(u128[:8], 11)
	cfg = append(cfg, u128[:]...)
	binary.LittleEndian.PutUint64(u128[:8], 22)
	cfg = append(cfg, u128[:]...)
	cfg = append(cfg, 0)
	var u64b [8]byte
	binary.LittleEndian.PutUint64(u64b[:], 9)
	cfg = append(cfg, u64b[:]...)
	config := pushPubkeyBytes(&cfg, 43)
	binary.LittleEndian.PutUint64(u128[:8], 0xabcd)
	binary.LittleEndian.PutUint64(u128[8:], 0)
	cfg = append(cfg, u128[:]...)
	c := parseDammCreateConfig(cfg, EventMetadata{}).Data.(*MeteoraDammV2CreateConfigEvent)
	if c.CompoundingFeeBps != 250 || c.DynamicFee != nil || c.VaultConfigKey != vault ||
		c.PoolCreatorAuthority != auth || c.Config != config || c.Permission != "43981" {
		t.Fatalf("create config: %+v", c)
	}

	dyn := make([]byte, 0, 88)
	config2 := pushPubkeyBytes(&dyn, 51)
	auth2 := pushPubkeyBytes(&dyn, 52)
	binary.LittleEndian.PutUint64(u64b[:], 3)
	dyn = append(dyn, u64b[:]...)
	binary.LittleEndian.PutUint64(u128[:8], 99)
	binary.LittleEndian.PutUint64(u128[8:], 0)
	dyn = append(dyn, u128[:]...)
	d := parseDammCreateDynamicConfig(dyn, EventMetadata{}).Data.(*MeteoraDammV2CreateDynamicConfigEvent)
	if d.Config != config2 || d.PoolCreatorAuthority != auth2 || d.Index != 3 || d.Permission != "99" {
		t.Fatalf("dynamic config: %+v", d)
	}
}

func dammBytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func TestDammNewEventTypeFilterIncludes(t *testing.T) {
	for _, et := range []EventType{
		EventTypeMeteoraDammV2UpdateDelegatePermission,
		EventTypeMeteoraDammV2WithdrawDeadLiquidityReward,
		EventTypeMeteoraDammV2CreateConfig,
		EventTypeMeteoraDammV2CreateDynamicConfig,
	} {
		if !EventTypeFilterIncludesMeteoraDammV2(EventTypeFilterIncludeOnly([]EventType{et})) {
			t.Fatalf("filter should include %s", et)
		}
	}
}
