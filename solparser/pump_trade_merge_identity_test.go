package solparser

import (
	"encoding/json"
	"testing"
)

func TestNestedPumpTradeIdentity(t *testing.T) {
	for _, kind := range []string{"pump", "swap_buy", "swap_sell"} {
		for _, mismatch := range []string{"venue", "user", "direction", "none"} {
			venue, user := "venueA", "userA"
			v, u := venue, user
			if mismatch == "venue" {
				v = "venueB"
			}
			if mismatch == "user" {
				u = "userB"
			}
			var base, inner DexEvent
			switch kind {
			case "pump":
				base = DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{Mint: venue, User: user, IsBuy: true, IxName: "buy_v3"}}
				inner = DexEvent{Type: EventTypePumpFunTrade, Data: &PumpFunTradeEvent{Mint: v, User: u, IsBuy: mismatch != "direction", IxName: "buy_v3"}}
			case "swap_buy":
				base = DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{Pool: venue, User: user}}
				if mismatch == "direction" {
					inner = DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: v, User: u}}
				} else {
					inner = DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{Pool: v, User: u}}
				}
			case "swap_sell":
				base = DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: venue, User: user}}
				if mismatch == "direction" {
					inner = DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{Pool: v, User: u}}
				} else {
					inner = DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: v, User: u}}
				}
			}
			before, _ := json.Marshal(base)
			j := 0
			out := mergeRpcInstructionEvents([]rpcIndexedEvent{{OuterIdx: 0, Event: base}, {OuterIdx: 0, InnerIdx: &j, Event: inner}})
			want := 2
			if mismatch == "none" {
				want = 1
			}
			if len(out) != want {
				t.Fatalf("%s %s: got %d events want %d", kind, mismatch, len(out), want)
			}
			if mismatch != "none" {
				after, _ := json.Marshal(out[0])
				if string(before) != string(after) {
					t.Fatal("mutated distinct trade", kind, mismatch)
				}
			}
		}
	}
}
