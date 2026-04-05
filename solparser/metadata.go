package solparser

type EventMetadata struct {
	Signature        string `json:"signature"`
	Slot             uint64 `json:"slot"`
	TxIndex          uint64 `json:"tx_index"`
	BlockTimeUs      int64  `json:"block_time_us"`
	GrpcRecvUs       int64  `json:"grpc_recv_us"`
	RecentBlockhash  string `json:"recent_blockhash,omitempty"`
}

func makeMetadata(sig string, slot, tx uint64, blockUs *int64, grpcUs int64, rb string) EventMetadata {
	bt := int64(0)
	if blockUs != nil {
		bt = *blockUs
	}
	return EventMetadata{
		Signature:       sig,
		Slot:            slot,
		TxIndex:         tx,
		BlockTimeUs:     bt,
		GrpcRecvUs:      grpcUs,
		RecentBlockhash: rb,
	}
}
