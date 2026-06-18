package solparser

import (
	"encoding/base64"
	"strings"
)

const programDataPrefix = "Program data: "

func decodeProgramDataLine(log string) []byte {
	i := strings.Index(log, programDataPrefix)
	if i < 0 {
		return nil
	}
	trimmed := strings.TrimSpace(log[i+len(programDataPrefix):])
	out, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil || len(out) < 8 {
		return nil
	}
	return out
}
