package solparser

// Program-bound outer instruction intents, never confirmed fills. No RPC.
import (
	"bytes"
	"encoding/binary"
	"encoding/json"
)

const (
	EventTypePumpRouteIntent EventType = "PumpRouteIntent"
	pumpRapid                          = "rapidV5DAKtecYzetwCNoxXWHpezjeg4hGt9BGiVcGW"
	pumpDelta                          = "J7pourVwqP1VtRNBiFNqbdGxwkx7ta2LHAURURRRJqmd"
	routeToken22                       = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
	routeSystem                        = "11111111111111111111111111111111"
	routeATA                           = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
	routeGlobal                        = "4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf"
	routeEvent                         = "Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1"
	routeFees                          = "pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ"
)

var routeMulti = []byte{43, 100, 73, 19, 233, 246, 111, 148}
var routeCurve = []byte{225, 154, 117, 22, 215, 86, 246, 103}

type PumpRouteHop struct {
	BaseMint   string `json:"base_mint"`
	QuoteMint  string `json:"quote_mint"`
	Venue      string `json:"venue"`
	BaseVault  string `json:"base_vault"`
	QuoteVault string `json:"quote_vault"`
}
type PumpRouteIntentEvent struct {
	Metadata         EventMetadata  `json:"metadata"`
	Kind             string         `json:"kind"`
	Program          string         `json:"program"`
	InstructionIndex int            `json:"instruction_index"`
	User             string         `json:"user"`
	InputAccount     string         `json:"input_account"`
	OutputAccount    string         `json:"output_account"`
	AmountIn         *uint64        `json:"amount_in"`
	MinAmountOut     *uint64        `json:"min_amount_out"`
	Hops             []PumpRouteHop `json:"hops"`
	InstructionData  []byte         `json:"instruction_data"`
}

func (e *PumpRouteIntentEvent) GetMetadata() EventMetadata { return e.Metadata }
func prefix(data, disc []byte) bool {
	return len(data) >= len(disc) && bytes.Equal(data[:len(disc)], disc)
}
func MayParsePumpRoute(program string, data []byte) bool {
	return program == PUMPSWAP_PROGRAM_ID && prefix(data, routeMulti) || program == PUMPFUN_PROGRAM_ID && prefix(data, routeCurve) ||
		program == pumpDelta && len(data) == 28 && prefix(data, []byte{122, 28, 94, 147, 180, 45, 240, 97}) ||
		program == pumpRapid && len(data) == 41 && data[0] == 0 && bytes.Equal(data[9:17], []byte{1, 0, 0, 27, 15, 14, 8, 24}) && bytes.Equal(data[17:25], []byte{194, 171, 28, 70, 104, 77, 91, 47}) && bytes.Equal(data[25:33], make([]byte, 8))
}
func routeContext(a []string, expected map[int]string) bool {
	for i, v := range expected {
		if i >= len(a) || a[i] != v {
			return false
		}
	}
	return true
}
func routeTokenProgram(s string) bool { return s == routeToken || s == routeToken22 }
func pumpMultiHopArgs(program string, data []byte, a []string) (uint64, uint64, bool) {
	if program != PUMPSWAP_PROGRAM_ID || len(data) != 24 || !prefix(data, routeMulti) || len(a) < 21 || (len(a)-16)%5 != 0 || !routeContext(a, map[int]string{7: routeToken, 8: routeToken22, 9: routeSystem, 11: PUMPSWAP_PROGRAM_ID, 12: PUMPFUN_PROGRAM_ID}) {
		return 0, 0, false
	}
	amount, minimum := binary.LittleEndian.Uint64(data[8:16]), binary.LittleEndian.Uint64(data[16:24])
	return amount, minimum, amount > 0
}
func allowsPumpRoute(filter EventTypeFilter, program string) bool {
	if filter == nil {
		return true
	}
	if f, ok := filter.(*IncludeOnlyFilter); ok {
		return eventTypeSliceContains(f.IncludeOnly, EventTypePumpRouteIntent) || eventTypeSliceContains(f.IncludeOnly, EventTypePumpFunTrade) || program == PUMPSWAP_PROGRAM_ID && eventTypeSliceContains(f.IncludeOnly, EventTypePumpSwapTrade)
	}
	if f, ok := filter.(*ExcludeFilter); ok {
		if eventTypeSliceContains(f.ExcludeTypes, EventTypePumpRouteIntent) || eventTypeSliceContains(f.ExcludeTypes, EventTypePumpFunTrade) || program == PUMPSWAP_PROGRAM_ID && eventTypeSliceContains(f.ExcludeTypes, EventTypePumpSwapTrade) {
			return false
		}
	}
	return filter.ShouldInclude(EventTypePumpRouteIntent)
}
func ParsePumpRoute(program string, data []byte, a []string, metadata EventMetadata, index int, filter EventTypeFilter) *PumpRouteIntentEvent {
	if !MayParsePumpRoute(program, data) || !allowsPumpRoute(filter, program) {
		return nil
	}
	var amount, minimum *uint64
	kind, fixed, user, input, output := "", 0, 0, 0, 0
	u := func(i int) *uint64 { v := binary.LittleEndian.Uint64(data[i : i+8]); return &v }
	switch program {
	case PUMPSWAP_PROGRAM_ID:
		av, mv, ok := pumpMultiHopArgs(program, data, a)
		if !ok {
			return nil
		}
		amount, minimum = &av, &mv
		kind, fixed, user, input, output = "MultiHopSwap", 16, 0, 1, 2
	case PUMPFUN_PROGRAM_ID:
		if len(data) < 20 {
			return nil
		}
		count := uint64(binary.LittleEndian.Uint32(data[16:20]))
		if count == 0 || uint64(len(data)) != 20+count || uint64(len(a)) != 13+count*5 || !routeContext(a, map[int]string{8: routeToken, 9: routeToken22, 10: routeSystem, 11: routeEvent, 12: PUMPFUN_PROGRAM_ID}) {
			return nil
		}
		kind, fixed, user, input, output = "MultiHopCurveSwap", 13, 3, 5, 6
		amount = u(8)
	case pumpRapid:
		if len(a) != 33 || a[0] != a[19] || a[2] != a[8] || !routeTokenProgram(a[9]) || !routeContext(a, map[int]string{3: routeToken, 4: routeSystem, 5: PUMPFUN_PROGRAM_ID, 6: routeGlobal, 10: routeToken, 11: routeATA, 29: routeFees, 30: routeSystem, 31: routeEvent, 32: PUMPFUN_PROGRAM_ID}) {
			return nil
		}
		kind, user, input, output = "RapidBuy", 0, 21, 20
		amount, minimum = u(1), u(33)
	case pumpDelta:
		if len(a) != 30 || a[0] != a[14] || !routeTokenProgram(a[4]) || !routeTokenProgram(a[5]) || !routeContext(a, map[int]string{1: routeGlobal, 6: routeATA, 24: routeFees, 25: routeSystem, 26: routeEvent, 27: PUMPFUN_PROGRAM_ID}) {
			return nil
		}
		kind, user, input, output = "BuyPumpDelta", 0, 16, 15
	}
	if amount != nil && *amount == 0 {
		return nil
	}
	hops := []PumpRouteHop{}
	if fixed > 0 {
		for i := fixed; i < len(a); i += 5 {
			hops = append(hops, PumpRouteHop{a[i], a[i+1], a[i+2], a[i+3], a[i+4]})
		}
	} else {
		positions := []int{2, 3, 11, 12, 13}
		if program == pumpRapid {
			positions = []int{7, 2, 16, 17, 18}
		}
		hops = append(hops, PumpRouteHop{a[positions[0]], a[positions[1]], a[positions[2]], a[positions[3]], a[positions[4]]})
	}
	return &PumpRouteIntentEvent{metadata, kind, program, index, a[user], a[input], a[output], amount, minimum, hops, append([]byte(nil), data...)}
}

func (e *PumpRouteIntentEvent) EventType() EventType { return EventTypePumpRouteIntent }

// MarshalJSON preserves opaque instruction bytes as an array, matching the route schema.
func (e PumpRouteIntentEvent) MarshalJSON() ([]byte, error) {
	type alias PumpRouteIntentEvent
	data := make([]int, len(e.InstructionData))
	for i, v := range e.InstructionData {
		data[i] = int(v)
	}
	return json.Marshal(struct {
		alias
		InstructionData []int `json:"instruction_data"`
	}{alias(e), data})
}
