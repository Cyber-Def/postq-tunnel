package server

import (
	"errors"
	"net"
	"sync"
	"github.com/hashicorp/yamux"
)

// MaxTunnels is the maximum number of concurrent agent tunnels allowed.
// Protects against resource exhaustion without impacting normal usage.
const MaxTunnels = 100

// Registry implements core.TunnelRegistry in-memory
type Registry struct {
	mu      sync.RWMutex
	tunnels map[string]*yamux.Session
}

func NewRegistry() *Registry {
	return &Registry{
		tunnels: make(map[string]*yamux.Session),
	}
}

func (r *Registry) Register(subdomain string, session *yamux.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.tunnels) >= MaxTunnels {
		return errors.New("server tunnel capacity reached, try again later")
	}
	if _, exists := r.tunnels[subdomain]; exists {
		return errors.New("subdomain is already occupied")
	}
	r.tunnels[subdomain] = session
	return nil
}

func (r *Registry) Unregister(subdomain string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tunnels, subdomain)
}

func (r *Registry) OpenStream(subdomain string) (net.Conn, error) {
	r.mu.RLock()
	session, exists := r.tunnels[subdomain]
	r.mu.RUnlock()

	if !exists {
		return nil, errors.New("tunnel for subdomain not found or offline")
	}
	
	
	// Open a new stream dynamically for each requesting HTTP client
	return session.Open()
}

// TunnelCount returns the number of active Yamux sessions (active agents).
func (r *Registry) TunnelCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tunnels)
}
