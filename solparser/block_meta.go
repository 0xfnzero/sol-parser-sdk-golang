package solparser

import (
	"math"
	"strings"
)

// BlockMetaEvent carries current block time/blockhash without RPC.
type BlockMetaEvent struct {
	Metadata EventMetadata `json:"metadata"`
}

func (e *BlockMetaEvent) GetMetadata() EventMetadata { return e.Metadata }
func ParseBlockMetaUpdate(meta *SubscribeUpdateBlockMeta, grpcRecvUs int64, fallbackBlockUs *int64) DexEvent {
	if meta == nil {
		return DexEvent{}
	}
	blockUs := int64(0)
	if fallbackBlockUs != nil {
		blockUs = *fallbackBlockUs
	}
	if meta.BlockTime != nil {
		t := *meta.BlockTime
		if t > math.MaxInt64/1000000 {
			blockUs = math.MaxInt64
		} else if t < math.MinInt64/1000000 {
			blockUs = math.MinInt64
		} else {
			blockUs = t * 1000000
		}
	}
	return DexEvent{Type: EventTypeBlockMeta, Data: &BlockMetaEvent{EventMetadata{Signature: strings.Repeat("1", 64), Slot: meta.Slot, BlockTimeUs: blockUs, GrpcRecvUs: grpcRecvUs, RecentBlockhash: meta.Blockhash}}}
}
func (e *BlockMetaEvent) EventType() EventType { return EventTypeBlockMeta }
