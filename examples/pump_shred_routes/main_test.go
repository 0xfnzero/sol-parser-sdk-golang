package main

import "testing"

func TestSignedReplay(t *testing.T) {
	if runReplay() != 6 {
		t.Fatal("fixture count")
	}
}
