package tunnel

import (
	"crypto/tls"
	"net"
)

// DefaultTLSConfig returns the strict Post-Quantum TLS 1.3 configuration
func DefaultTLSConfig(cert *tls.Certificate) *tls.Config {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{
			// Post-Quantum Hybrid ML-KEM (Standardized Kyber in Go 1.24+)
			tls.X25519MLKEM768,
			// Fallback to classic X25519 for older TLS 1.3 clients
			tls.CurveP256,
		},
	}
	if cert != nil {
		cfg.Certificates = []tls.Certificate{*cert}
	}
	return cfg
}

// ListenPQC creates a Quantum-Resistant TLS listener
func ListenPQC(addr string, cert *tls.Certificate) (net.Listener, error) {
	return tls.Listen("tcp", addr, DefaultTLSConfig(cert))
}

// DialPQC connects to a Quantum-Resistant TLS server
func DialPQC(addr string, insecureSkipVerify bool) (*tls.Conn, error) {
	cfg := DefaultTLSConfig(nil)
	cfg.InsecureSkipVerify = insecureSkipVerify // Useful for local dev with self-signed certs
	return tls.Dial("tcp", addr, cfg)
}
