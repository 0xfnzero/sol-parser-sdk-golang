package solparser

import (
	"encoding/binary"
	"fmt"
	"testing"
)

func pumpFeesInstructionAccounts(n int) []string {
	accounts := make([]string, n)
	for i := range accounts {
		accounts[i] = fmt.Sprintf("account_%d", i)
	}
	return accounts
}

func pumpFeesUpdateFeeSharesData(disc uint64) []byte {
	data := make([]byte, 8+4+32+2)
	binary.LittleEndian.PutUint64(data[:8], disc)
	binary.LittleEndian.PutUint32(data[8:12], 1)
	copy(data[12:44], bytesRepeat(42, 32))
	binary.LittleEndian.PutUint16(data[44:46], 2500)
	return data
}

func bytesRepeat(v byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = v
	}
	return out
}

func TestParsePumpFeesUpdateFeeSharesV1V2InstructionLayout(t *testing.T) {
	for _, disc := range []uint64{instrPumpFeesUpdateFeeShares, instrPumpFeesUpdateFeeSharesV2} {
		ev := parsePumpFeesInstruction(
			pumpFeesUpdateFeeSharesData(disc),
			pumpFeesInstructionAccounts(8),
			EventMetadata{Signature: "sig", Slot: 1, GrpcRecvUs: 10},
		)
		if ev.Type != EventTypePumpFeesUpdateFeeShares {
			t.Fatalf("expected PumpFeesUpdateFeeShares, got %q", ev.Type)
		}
		data, ok := ev.Data.(*PumpFeesUpdateFeeSharesEvent)
		if !ok {
			t.Fatalf("expected PumpFeesUpdateFeeSharesEvent, got %T", ev.Data)
		}
		if data.Mint != "account_4" || data.SharingConfig != "account_5" || data.Admin != "account_2" ||
			data.BondingCurve != "account_6" || data.PumpCreatorVault != "account_7" {
			t.Fatalf("unexpected update_fee_shares accounts: %+v", data)
		}
		if len(data.NewShareholders) != 1 ||
			data.NewShareholders[0].Address != Base58Encode(bytesRepeat(42, 32)) ||
			data.NewShareholders[0].ShareBps != 2500 {
			t.Fatalf("unexpected update_fee_shares shareholders: %+v", data.NewShareholders)
		}
	}
}

func TestParsePumpFeesResetFeeSharingV1V2InstructionAccountOrder(t *testing.T) {
	for _, disc := range []uint64{instrPumpFeesResetFeeSharingConfig, instrPumpFeesResetFeeSharingConfigV2} {
		data := make([]byte, 8)
		binary.LittleEndian.PutUint64(data, disc)
		ev := parsePumpFeesInstruction(
			data,
			pumpFeesInstructionAccounts(7),
			EventMetadata{Signature: "sig", Slot: 1, GrpcRecvUs: 10},
		)
		if ev.Type != EventTypePumpFeesResetFeeSharingConfig {
			t.Fatalf("expected PumpFeesResetFeeSharingConfig, got %q", ev.Type)
		}
		reset, ok := ev.Data.(*PumpFeesResetFeeSharingConfigEvent)
		if !ok {
			t.Fatalf("expected PumpFeesResetFeeSharingConfigEvent, got %T", ev.Data)
		}
		if reset.NewAdmin != "account_0" || reset.OldAdmin != "account_3" ||
			reset.Mint != "account_5" || reset.SharingConfig != "account_6" {
			t.Fatalf("unexpected reset_fee_sharing accounts: %+v", reset)
		}
	}
}
