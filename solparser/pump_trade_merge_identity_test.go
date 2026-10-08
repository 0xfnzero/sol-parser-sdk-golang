package solparser

import (
	"encoding/json"
	"fmt"
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

func TestRepeatedPumpTradesPairWithTheirOwnCPI(t *testing.T) {
	for _, known := range []bool{true, false} {
		event := func(amount uint64, executed bool) DexEvent {
			t := &PumpFunTradeEvent{Mint: "mintA", User: "userA", IsBuy: true, IxName: "buy_v3"}
			if executed {
				t.TokenAmount = amount
			} else {
				t.Amount = amount
			}
			return DexEvent{Type: EventTypePumpFunTrade, Data: t}
		}
		height := func(n uint32) *uint32 {
			if !known {
				return nil
			}
			return &n
		}
		a, b, c := 0, 1, 2
		out := mergeRpcInstructionEvents([]rpcIndexedEvent{{OuterIdx: 0, StackHeight: height(1), Event: event(100, false)}, {OuterIdx: 0, InnerIdx: &a, StackHeight: height(2), IsEventCPI: true, Event: event(90, true)}, {OuterIdx: 0, InnerIdx: &b, StackHeight: height(2), Event: event(200, false)}, {OuterIdx: 0, InnerIdx: &c, StackHeight: height(3), IsEventCPI: true, Event: event(180, true)}})
		if len(out) != 2 {
			t.Fatalf("known=%v got %d trades", known, len(out))
		}
		for i, want := range [][2]uint64{{100, 90}, {200, 180}} {
			got := out[i].Data.(*PumpFunTradeEvent)
			if got.Amount != want[0] || got.TokenAmount != want[1] {
				t.Fatal("incorrect pairing", known, i, got.Amount, got.TokenAmount)
			}
		}
	}
}

func TestRepeatedPumpSwapDedupOccurrences(t *testing.T) {
	for _, buy := range []bool{true, false} {
		logs, instructions := []DexEvent{}, []DexEvent{}
		for _, n := range []uint64{100, 200} {
			if buy {
				logs = append(logs, DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{Pool: "pool", User: "user", QuoteAmountIn: n - 10}})
				instructions = append(instructions, DexEvent{Type: EventTypePumpSwapBuy, Data: &PumpSwapBuyEvent{Pool: "pool", User: "user", MaxQuoteAmountIn: n, UserBaseTokenAccount: fmt.Sprint(n)}})
			} else {
				logs = append(logs, DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: "pool", User: "user", QuoteAmountOut: n - 10}})
				instructions = append(instructions, DexEvent{Type: EventTypePumpSwapSell, Data: &PumpSwapSellEvent{Pool: "pool", User: "user", MinQuoteAmountOut: n, UserBaseTokenAccount: fmt.Sprint(n)}})
			}
		}
		if len(DedupeLogInstructionEvents(nil, instructions)) != 2 {
			t.Fatal("collapsed repeated instructions")
		}
		out := DedupeLogInstructionEvents(logs, instructions)
		if len(out) != 2 {
			t.Fatal("wrong occurrence count")
		}
		for i, n := range []uint64{100, 200} {
			if buy {
				e := out[i].Data.(*PumpSwapBuyEvent)
				if e.UserBaseTokenAccount != fmt.Sprint(n) || e.QuoteAmountIn != n-10 {
					t.Fatal("wrong buy pairing")
				}
			} else {
				e := out[i].Data.(*PumpSwapSellEvent)
				if e.UserBaseTokenAccount != fmt.Sprint(n) || e.QuoteAmountOut != n-10 {
					t.Fatal("wrong sell pairing")
				}
			}
		}
	}
}
