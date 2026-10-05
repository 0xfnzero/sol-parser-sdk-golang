package solparser

import "testing"

func TestNormalizeGRPCEndpoint(t *testing.T) {
	for _, c := range []struct {
		url, target string
		config, tls bool
	}{
		{"node.example:443", "node.example:443", true, true},
		{"https://node.example:443", "node.example:443", false, true},
		{"https://node.example/", "node.example:443", true, true},
		{"http://localhost:10000", "localhost:10000", true, false},
		{"http://[::1]:10000", "[::1]:10000", true, false},
	} {
		target, tls, e := normalizeGRPCEndpoint(c.url, c.config)
		if e != nil || target != c.target || tls != c.tls {
			t.Fatal(c, target, tls, e)
		}
		if tls && tlsConfigForGRPCEndpoint(target).ServerName == "" {
			t.Fatal("SNI missing")
		}
	}
	for _, s := range []string{"ftp://node.example", "https://user:secret@node.example", "https://node.example/path", "https://node.example?token=secret", "https://"} {
		if _, _, e := normalizeGRPCEndpoint(s, true); e == nil {
			t.Fatal("invalid endpoint accepted")
		}
	}
}
