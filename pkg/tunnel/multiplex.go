package tunnel

import (
	"crypto/tls"
	"github.com/hashicorp/yamux"
)

// ServerSession wraps a PQC TLS connection in a yamux server multiplexer.
// This allows the server to accept multiple logical streams over a single connection.
func ServerSession(conn *tls.Conn) (*yamux.Session, error) {
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = true
	return yamux.Server(conn, cfg)
}

// ClientSession wraps a PQC TLS connection in a yamux client multiplexer.
// This allows the client agent to open streams (e.g. for proxying web traffic or auth).
func ClientSession(conn *tls.Conn) (*yamux.Session, error) {
	cfg := yamux.DefaultConfig()
	cfg.EnableKeepAlive = true
	return yamux.Client(conn, cfg)
}
