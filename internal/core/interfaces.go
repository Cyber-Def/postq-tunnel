package core

import (
	"net/http"
	"net"
	"github.com/hashicorp/yamux"
)

// TunnelRegistry manages active tunnels via subdomains.
type TunnelRegistry interface {
	// Register adds a new multiplexed session for a given subdomain.
	Register(subdomain string, session *yamux.Session) error
	
	// Unregister removes the active tunnel session.
	Unregister(subdomain string)
	
	// OpenStream opens a new logical connection in the session to route HTTP requests.
	OpenStream(subdomain string) (net.Conn, error)
}

// Authenticator validates tunnel creation requests based on predefined team/user tokens.
type Authenticator interface {
	Authenticate(token string) (bool, error)
}

// ProxyMiddleware represents a chainable HTTP middleware handler (e.g., Auth, Rate-Limit).
type ProxyMiddleware interface {
	Handle(next http.Handler) http.Handler
}
