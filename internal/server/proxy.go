package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/Cyber-Def/postq-tunnel/internal/core"
)

// BuildProxy creates a standard ReverseProxy that overrides the Transport.
// Instead of dialing standard TCP sockets, it dials virtual Yamux streams!
func BuildProxy(registry core.TunnelRegistry) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		// Preserve incoming aspects, we simulate an HTTP backend routing
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Extract subdomain from the Host header (e.g. user1.tunnel.io:443 -> user1)
			host := strings.Split(addr, ":")[0]
			subdomain := strings.Split(host, ".")[0]

			stream, err := registry.OpenStream(subdomain)
			if err != nil {
				log.Printf("[Proxy Error] %s: %v", subdomain, err)
				return nil, err
			}
			return stream, nil
		},
	}

	return &httputil.ReverseProxy{
		Director:  director,
		Transport: transport,
	}
}
