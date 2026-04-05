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
	if len(trimmed) > 2700 {
		return nil
	}
	out := make([]byte, 2048)
	n, err := base64.StdEncoding.Decode(out, []byte(trimmed))
	if err != nil || n < 8 || n > 2048 {
		return nil
	}
	return out[:n]
}
