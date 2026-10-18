package main

import "testing"

func TestPumpOfflineBank(t *testing.T) {
	if runBank("") != 20 {
		t.Fatal("matrix size")
	}
}
