package main

import (
	"encoding/base64"
	solparser "github.com/0xfnzero/sol-parser-sdk-golang/solparser"
	"testing"
)

func TestSnapshotVersionsAndClosure(t *testing.T) {
	old := savedAccount{"key", "owner", base64.StdEncoding.EncodeToString([]byte("one")), "100", "1"}
	e := &solparser.RawAccountSnapshotEvent{Metadata: solparser.EventMetadata{Slot: 100}, WriteVersion: 1, Account: solparser.AccountData{Pubkey: "key", Owner: "owner", Data: []byte("one"), Lamports: 1}}
	if got, changed, err := applyRawSnapshot(old, e); err != nil || changed || got != old {
		t.Fatal(got, changed, err)
	}
	e.Metadata.Slot = 99
	e.WriteVersion = 999
	e.Account.Data = []byte("old")
	if _, changed, err := applyRawSnapshot(old, e); err != nil || changed {
		t.Fatal(changed, err)
	}
	e.Metadata.Slot = 100
	e.WriteVersion = 1
	if _, _, err := applyRawSnapshot(old, e); err == nil {
		t.Fatal("conflict accepted")
	}
	e.Metadata.Slot = 101
	e.Account.Lamports = 0
	if got, changed, err := applyRawSnapshot(old, e); err != nil || !changed || got.Data != "" {
		t.Fatal(got, changed, err)
	}
}
