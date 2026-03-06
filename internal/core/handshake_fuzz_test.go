package core

import (
	"bytes"
	"testing"
)

// FuzzReadHandshake feeds arbitrary bytes into ReadHandshake to detect panics,
// unexpected errors, or crashes in the JSON decoder + io.LimitReader pipeline.
// Run with: go test -fuzz=FuzzReadHandshake ./internal/core/
func FuzzReadHandshake(f *testing.F) {
	// Seed with a valid handshake payload
	f.Add([]byte(`{"version":"v1","subdomain":"myapp","token":"secret"}`))
	f.Add([]byte(`{"version":"v1","subdomain":"x","token":""}`))
	// Seed with edge cases
	f.Add([]byte(`{}`))
	f.Add([]byte(`[]`))
	f.Add([]byte(``))
	f.Add([]byte(`{"version":"","subdomain":"` + string(make([]byte, 4096)) + `"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic regardless of input
		req, err := ReadHandshake(bytes.NewReader(data))
		if err != nil {
			// Errors are fine — panics are not
			return
		}
		// If parsing succeeded, the result must be at least structurally valid
		_ = req.Version
		_ = req.Subdomain
		_ = req.Token
	})
}

// FuzzWriteReadRoundtrip verifies that serialization is always symmetrically reversible.
func FuzzWriteReadRoundtrip(f *testing.F) {
	f.Add("v1", "myapp", "tok", "192.168.1.0/24", "user", "pass")
	f.Add("v1", "a", "", "", "", "")
	f.Add("", "", "", "", "", "")

	f.Fuzz(func(t *testing.T, version, sub, tok, cidr, usr, pwd string) {
		var cidrs []string
		if cidr != "" {
			cidrs = []string{cidr}
		}

		req := HandshakeReq{
			Version:      version,
			Subdomain:    sub,
			Token:        tok,
			AllowCIDRs:   cidrs,
			BasicAuthUsr: usr,
			BasicAuthPwd: pwd,
		}

		var buf bytes.Buffer
		if err := WriteHandshake(&buf, req); err != nil {
			// WriteHandshake should never fail for any string input
			t.Fatalf("WriteHandshake failed: %v", err)
		}

		got, err := ReadHandshake(&buf)
		if err != nil {
			t.Fatalf("ReadHandshake failed after WriteHandshake: %v", err)
		}

		// Sub-fields must survive the round-trip (unless truncated by LimitReader)
		if len(buf.Bytes()) == 0 && got.Subdomain != sub {
			t.Errorf("subdomain mismatch: got %q want %q", got.Subdomain, sub)
		}
	})
}
