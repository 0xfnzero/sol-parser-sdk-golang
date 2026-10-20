package solparser

import "strings"

// Execution logs can omit the version; prefer the actual CPI IDL method name.
func preferPumpInstructionName(current, incoming string) bool {
	versioned := func(s string) bool { return strings.HasSuffix(s, "_v2") || strings.HasSuffix(s, "_v3") }
	return incoming != "" && (!versioned(current) || versioned(incoming))
}
