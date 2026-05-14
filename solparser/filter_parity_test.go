package solparser

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"testing"

	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
)

func pumpFeesUpdateAdminLogForTest() string {
	buf := make([]byte, 8+8+32+32)
	binary.LittleEndian.PutUint64(buf[:8], discPumpFeesUpdateAdmin)
	return "Program data: " + base64.StdEncoding.EncodeToString(buf)
}

func TestParseLogOptimizedAppliesEventTypeFilter(t *testing.T) {
	log := pumpFeesUpdateAdminLogForTest()

	ev := ParseLogOptimized(log, "sig", 1, 0, nil, 1, nil, false, "")
	if ev.Type != EventTypePumpFeesUpdateAdmin {
		t.Fatalf("expected unfiltered UpdateAdmin, got %q", ev.Type)
	}

	ev = ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpFeesUpdateAdmin}),
		false,
		"",
	)
	if ev.Type != EventTypePumpFeesUpdateAdmin {
		t.Fatalf("expected include-only UpdateAdmin, got %q", ev.Type)
	}

	ev = ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunCreate}),
		false,
		"",
	)
	if ev.Type != "" {
		t.Fatalf("expected non-matching include-only filter to drop event, got %q", ev.Type)
	}

	ev = ParseLogOptimized(
		log,
		"sig",
		1,
		0,
		nil,
		1,
		EventTypeFilterExclude([]EventType{EventTypePumpFeesUpdateAdmin}),
		false,
		"",
	)
	if ev.Type != "" {
		t.Fatalf("expected exclude filter to drop event, got %q", ev.Type)
	}
}

func TestProtocolHelperExcludeMatchesRustSemantics(t *testing.T) {
	filter := EventTypeFilterExclude([]EventType{EventTypePumpFeesUpdateAdmin})
	if EventTypeFilterIncludesPumpFees(filter) {
		t.Fatalf("PumpFees route should be disabled when any PumpFees type is excluded")
	}
}

func TestSubscribeRequestBuilderIncludesAccountFilters(t *testing.T) {
	client := NewYellowstoneGrpc("127.0.0.1:10000")
	txFilter := TransactionFilterForProtocols([]Protocol{ProtocolPumpFun})
	accFilter := AccountFilterForProtocols([]Protocol{ProtocolPumpSwap})
	accFilter.Filters = append(accFilter.Filters, AccountFilterMemcmp(32, []byte{1, 2, 3}))

	req := client.buildSubscribeRequestMulti([]TransactionFilter{txFilter}, []AccountFilter{accFilter})
	if req.Commitment == nil || *req.Commitment != pb.CommitmentLevel_PROCESSED {
		t.Fatalf("expected processed commitment, got %v", req.Commitment)
	}
	if got := req.Transactions["tx_0"].AccountInclude; len(got) != 1 || got[0] != PUMPFUN_PROGRAM_ID {
		t.Fatalf("unexpected transaction account_include: %#v", got)
	}

	accountReq := req.Accounts["acc_0"]
	if accountReq == nil {
		t.Fatalf("expected account filter acc_0")
	}
	if len(accountReq.Owner) == 0 || accountReq.Owner[0] != PUMPSWAP_PROGRAM_ID {
		t.Fatalf("unexpected account owners: %#v", accountReq.Owner)
	}
	if len(accountReq.Filters) != 1 {
		t.Fatalf("expected one account memcmp filter, got %d", len(accountReq.Filters))
	}
	memcmp := accountReq.Filters[0].GetMemcmp()
	if memcmp == nil || memcmp.Offset != 32 || !bytes.Equal(memcmp.GetBytes(), []byte{1, 2, 3}) {
		t.Fatalf("unexpected memcmp filter: %#v", memcmp)
	}
}
