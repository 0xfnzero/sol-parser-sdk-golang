package solparser

const EventTypeRawInstruction EventType = "RawInstruction"

// RawInstructionEvent describes only signed outer instruction data, not a CPI or fill.
type RawInstructionEvent struct {
	Metadata         EventMetadata `json:"metadata"`
	ProgramID        string        `json:"program_id"`
	InstructionIndex int           `json:"instruction_index"`
	// Preserve order and duplicate keys. nil means an unresolved ALT address.
	Accounts        []*string `json:"accounts"`
	AccountIndices  []int     `json:"account_indices"`
	InstructionData []int     `json:"instruction_data"`
}

func (e *RawInstructionEvent) GetMetadata() EventMetadata { return e.Metadata }
func (e *RawInstructionEvent) EventType() EventType       { return EventTypeRawInstruction }

var infrastructurePrograms = map[string]bool{
	"11111111111111111111111111111111":             true,
	"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA":  true,
	"TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb":  true,
	"ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL": true,
	"ComputeBudget111111111111111111111111111111":  true,
	"MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr":  true,
	"Memo1UhkJRfHyvLMcVucJwxXeuD728EqVDDwQDxFMNo":  true,
	"AddressLookupTab1e1111111111111111111111111":  true,
	"Vote111111111111111111111111111111111111111":  true,
	"Ed25519SigVerify111111111111111111111111111":  true,
	"KeccakSecp256k11111111111111111111111111111":  true,
	"Secp256r1SigVerify1111111111111111111111111":  true,
}

func isUnknownShredProgram(pid string) bool {
	if pid == "" || infrastructurePrograms[pid] {
		return false
	}
	for _, known := range shredUnknownProgramCandidates {
		if pid == known {
			return false
		}
	}
	return true
}
