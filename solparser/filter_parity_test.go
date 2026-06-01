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
	if !EventTypeFilterIncludesPumpFees(filter) {
		t.Fatalf("PumpFees route should stay enabled for exclude filters")
	}
	if !EventTypeFilterIncludesRaydiumCpmm(EventTypeFilterExclude([]EventType{EventTypeRaydiumCpmmSwap})) {
		t.Fatalf("partial CPMM excludes should keep CPMM route enabled")
	}
	if EventTypeFilterIncludesRaydiumCpmm(EventTypeFilterExclude([]EventType{
		EventTypeRaydiumCpmmSwap,
		EventTypeRaydiumCpmmDeposit,
		EventTypeRaydiumCpmmWithdraw,
		EventTypeRaydiumCpmmInitialize,
	})) {
		t.Fatalf("excluding every CPMM event should skip CPMM route")
	}
	if EventTypeFilterIncludesRaydiumLaunchlab(EventTypeFilterExclude([]EventType{
		EventTypeRaydiumLaunchlabTrade,
		EventTypeRaydiumLaunchlabPoolCreate,
		EventTypeRaydiumLaunchlabMigrateAmm,
	})) {
		t.Fatalf("excluding every LaunchLab event should skip LaunchLab route")
	}
	if EventTypeFilterIncludesPumpfun(EventTypeFilterIncludeOnly([]EventType{EventTypeAccountPumpFunGlobal})) {
		t.Fatalf("PumpFun account-only filters should not enable instruction routes")
	}
	if EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeAccountPumpFunGlobal}) {
		t.Fatalf("PumpFun account-only filters should not pass instruction prefilter")
	}
	if EventTypeFilterIncludesRaydiumClmm(EventTypeFilterIncludeOnly([]EventType{EventTypeAccountRaydiumClmmPoolState})) {
		t.Fatalf("Raydium CLMM account-only filters should not enable instruction routes")
	}
	if EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeAccountRaydiumClmmPoolState}) {
		t.Fatalf("Raydium CLMM account-only filters should not pass instruction prefilter")
	}
	if EventTypeFilterIncludesPumpfun(EventTypeFilterIncludeOnly([]EventType{EventTypePumpFeesUpdateAdmin})) {
		t.Fatalf("PumpFees filters should not enable PumpFun route")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypePumpFeesUpdateAdmin}) {
		t.Fatalf("PumpFees filters should pass instruction prefilter")
	}
	if !EventTypeFilterIncludesPumpswap(EventTypeFilterIncludeOnly([]EventType{EventTypePumpSwapTrade})) {
		t.Fatalf("PumpSwapTrade should enable PumpSwap route")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypePumpSwapTrade}) {
		t.Fatalf("PumpSwapTrade should pass instruction prefilter")
	}
	if !EventTypeFilterAllowsInstructionParsing([]EventType{EventTypeMeteoraDammV2InitializePool}) {
		t.Fatalf("Meteora DAMM initialize pool should pass instruction prefilter")
	}
	if !EventTypeFilterIncludesMeteoraPools(EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraPoolsSwap})) {
		t.Fatalf("Meteora Pools swap should enable Meteora Pools route helper")
	}
	if EventTypeFilterIncludesMeteoraPools(EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunTrade})) {
		t.Fatalf("PumpFun filters should not enable Meteora Pools route helper")
	}
	if !EventTypeFilterIncludesMeteoraDlmm(EventTypeFilterIncludeOnly([]EventType{EventTypeMeteoraDlmmSwap})) {
		t.Fatalf("Meteora DLMM swap should enable Meteora DLMM route helper")
	}
	if EventTypeFilterIncludesMeteoraDlmm(EventTypeFilterIncludeOnly([]EventType{EventTypePumpFunTrade})) {
		t.Fatalf("PumpFun filters should not enable Meteora DLMM route helper")
	}
	if !EventTypeFilterIncludesRaydiumAmmV4(EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumAmmV4Deposit})) {
		t.Fatalf("Raydium AMM V4 deposit should enable AMM V4 route")
	}
	if !EventTypeFilterIncludeOnly([]EventType{EventTypePumpSwapTrade}).ShouldInclude(EventTypePumpSwapBuy) {
		t.Fatalf("PumpSwapTrade should include concrete buy events")
	}
	if EventTypeFilterExclude([]EventType{EventTypePumpSwapTrade}).ShouldInclude(EventTypePumpSwapSell) {
		t.Fatalf("excluding PumpSwapTrade should drop concrete sell events")
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
