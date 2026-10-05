package solparser

import (
	pb "github.com/0xfnzero/sol-parser-sdk-golang/proto"
	"math"
	"strings"
	"testing"
)

func TestBlockMetaAdapter(t *testing.T) {
	ts := int64(1790928211)
	fallback := int64(456)
	e := ParseBlockMetaUpdate(&SubscribeUpdateBlockMeta{Slot: 452558748, Blockhash: "observed", BlockTime: &ts}, 123, &fallback)
	m := e.GetMetadata()
	if e.Type != EventTypeBlockMeta || m.Signature != strings.Repeat("1", 64) || m.Slot != 452558748 || m.BlockTimeUs != 1790928211000000 || m.GrpcRecvUs != 123 || m.RecentBlockhash != "observed" {
		t.Fatal(e, m)
	}
	e = ParseBlockMetaUpdate(&SubscribeUpdateBlockMeta{}, 123, &fallback)
	if e.GetMetadata().BlockTimeUs != 456 {
		t.Fatal("fallback")
	}
	ts = math.MaxInt64
	if ParseBlockMetaUpdate(&SubscribeUpdateBlockMeta{BlockTime: &ts}, 0, nil).GetMetadata().BlockTimeUs != math.MaxInt64 {
		t.Fatal("positive saturation")
	}
	ts = math.MinInt64
	if ParseBlockMetaUpdate(&SubscribeUpdateBlockMeta{BlockTime: &ts}, 0, nil).GetMetadata().BlockTimeUs != math.MinInt64 {
		t.Fatal("negative saturation")
	}
}
func TestLaunchLabAliases(t *testing.T) {
	ids := GetProgramIDsForProtocols([]Protocol{ProtocolLaunchLab, ProtocolStonkFun, ProtocolRaydiumLaunchlab})
	if len(ids) != 1 || ids[0] != RAYDIUM_LAUNCHLAB_PROGRAM_ID {
		t.Fatal(ids)
	}
}
func TestBlockMetaUpdateSubscriptionPreservesFilter(t *testing.T) {
	c := NewYellowstoneGrpc("https://example.invalid")
	c.dexControl = make(chan *pb.SubscribeRequest, 1)
	c.dexFilter = &IncludeOnlyFilter{IncludeOnly: []EventType{EventTypeBlockMeta}}
	if e := c.UpdateSubscription(nil, nil); e != nil {
		t.Fatal(e)
	}
	r := <-c.dexControl
	if _, ok := r.BlocksMeta["block_meta"]; !ok {
		t.Fatal("block metadata filter lost")
	}
}
