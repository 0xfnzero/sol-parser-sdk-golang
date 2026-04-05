package solparser

import "github.com/mr-tron/base58"

func readPubkey(b []byte, o int) (string, bool) {
	if o+32 > len(b) {
		return "", false
	}
	return base58.Encode(b[o : o+32]), true
}

const zeroPubkey = "11111111111111111111111111111111"
