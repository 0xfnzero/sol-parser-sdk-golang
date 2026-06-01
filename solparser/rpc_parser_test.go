package solparser

import "testing"

func TestParseRpcTransactionUsesBlockTransactionIndexForLogEvents(t *testing.T) {
	tx := &RpcTransactionResponse{
		Slot:             9,
		TransactionIndex: 42,
		Meta: &RpcTransactionMeta{
			LogMessages: []string{pumpFeesUpdateAdminLogForTest()},
		},
		Transaction: &RpcTransaction{
			Message: &RpcMessage{},
		},
	}

	events, parseErr := parseRpcTransactionImpl(tx, "sig", nil, 100)
	if parseErr != nil {
		t.Fatalf("unexpected parse error: %v", parseErr)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if got := events[0].GetMetadata().TxIndex; got != 42 {
		t.Fatalf("tx_index: got %d want 42", got)
	}
}
