package solparser

import (
	"encoding/base64"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestPrefixPrefilterEncodingSemantics(t *testing.T) {
	body := make([]byte, 80)
	binary.LittleEndian.PutUint64(body, discPumpFeesUpdateAdmin)
	encoded := base64.StdEncoding.EncodeToString(body)
	included := EventTypeFilterIncludeOnly([]EventType{EventTypePumpFeesUpdateAdmin})
	excluded := EventTypeFilterIncludeOnly([]EventType{EventTypeRaydiumCpmmSwap})
	parse := func(payload string, filter EventTypeFilter) DexEvent {
		return ParseLogOptimizedWithProgramID("Program data: "+payload, "offline", 1, 0, nil, 0, filter, false, "", PUMP_FEES_PROGRAM_ID)
	}
	for _, payload := range []string{encoded, " " + encoded + " ", encoded[:4] + "\n" + encoded[4:], encoded[:4] + "!" + encoded[4:], encoded[:len(encoded)-1], encoded + "!", "AA=="} {
		decoded := decodeProgramDataLine("Program data: " + payload)
		expected := DexEvent{}
		if decoded != nil {
			expected = parse(base64.StdEncoding.EncodeToString(decoded), included)
		}
		if !reflect.DeepEqual(parse(payload, included), expected) {
			t.Fatalf("encoding semantics changed: %q", payload)
		}
		if parse(payload, excluded).Type != "" {
			t.Fatalf("excluded payload emitted: %q", payload)
		}
	}
	for _, size := range []int{512, 4096} {
		body := make([]byte, size)
		binary.LittleEndian.PutUint64(body, discPumpFeesUpdateAdmin)
		payload := base64.StdEncoding.EncodeToString(body)
		log := "Program data: " + payload
		allocations := testing.AllocsPerRun(100, func() {
			ParseLogOptimizedWithProgramID(log, "offline", 1, 0, nil, 0, excluded, false, "", PUMP_FEES_PROGRAM_ID)
		})
		if allocations != 0 {
			t.Fatalf("excluded %d bytes allocates: %v", size, allocations)
		}
	}
}
