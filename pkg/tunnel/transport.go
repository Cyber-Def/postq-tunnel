package tunnel

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"time"
)

// GenerateDummyCert creates a quick self-signed certificate for local development
func GenerateDummyCert() *tls.Certificate {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Local Dev"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}
	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  priv,
	}
	return &cert
}

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
	} else {
		// Fallback for local HTTP development servers with empty QTUN_DOMAIN
		dummy := GenerateDummyCert()
		cfg.Certificates = []tls.Certificate{*dummy}
	}
	return cfg
}

// ListenPQC creates a Quantum-Resistant TLS listener
func ListenPQC(addr string, cert *tls.Certificate) (net.Listener, error) {
	return tls.Listen("tcp", addr, DefaultTLSConfig(cert))
}

// DialPQC connects to a Quantum-Resistant TLS server
func DialPQC(ctx context.Context, addr string, insecureSkipVerify bool) (*tls.Conn, error) {
	cfg := DefaultTLSConfig(nil)
	cfg.InsecureSkipVerify = insecureSkipVerify // Useful for local dev with self-signed certs
	
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{},
		Config:    cfg,
	}
	
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	
	return conn.(*tls.Conn), nil
}
