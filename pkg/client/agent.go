package client

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/Cyber-Def/postq-tunnel/internal/core"
	"github.com/Cyber-Def/postq-tunnel/pkg/tunnel"
)

type Config struct {
	ServerAddr   string
	Token        string
	Subdomain    string
	LocalTarget  string
	AllowCIDRs   string
	BasicAuthUsr string
	BasicAuthPwd string
}

func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.ServerAddr, "server", "tunnel.local:4443", "PQC Edge Server address")
	flag.StringVar(&cfg.Token, "token", "", "Authentication token for the server (optional)")
	flag.StringVar(&cfg.Subdomain, "sub", "", "Requested subdomain (e.g., 'myapp')")
	flag.StringVar(&cfg.LocalTarget, "local", "localhost:8080", "Local service to expose")
	flag.StringVar(&cfg.AllowCIDRs, "allow-ip", "", "Comma-separated CIDRs to whitelist (e.g., '192.168.1.0/24')")
	authFull := flag.String("auth", "", "Basic Auth in format 'user:pass'")
	
	flag.Parse()
	
	if *authFull != "" {
		parts := strings.SplitN(*authFull, ":", 2)
		if len(parts) == 2 {
			cfg.BasicAuthUsr = parts[0]
			cfg.BasicAuthPwd = parts[1]
		}
	}
	
	if cfg.Subdomain == "" {
		log.Fatal("Fatal: Subdomain is required (-sub)")
	}
	return cfg
}

func RunTunnel(cfg *Config) error {
	// 1. Dial Edge Server using PQC TLS (tls.X25519MLKEM768)
	tlsConn, err := tunnel.DialPQC(cfg.ServerAddr, true) // insecureSkipVerify for typical agent-server connection
	if err != nil {
		return err
	}
	defer tlsConn.Close()

	// 2. Wrap the Quantum-secure connection in Yamux client multiplexer
	session, err := tunnel.ClientSession(tlsConn)
	if err != nil {
		return err
	}
	defer session.Close()

	// 3. Open Handshake technical stream and push configs to Edge Server
	handshakeStream, err := session.Open()
	if err != nil {
		return err
	}

	var cidrs []string
	if cfg.AllowCIDRs != "" {
		for _, s := range strings.Split(cfg.AllowCIDRs, ",") {
			cidrs = append(cidrs, strings.TrimSpace(s))
		}
	}

	req := core.HandshakeReq{
		Subdomain:    cfg.Subdomain,
		Token:        cfg.Token,
		AllowCIDRs:   cidrs,
		BasicAuthUsr: cfg.BasicAuthUsr,
		BasicAuthPwd: cfg.BasicAuthPwd,
	}

	if err := core.WriteHandshake(handshakeStream, req); err != nil {
		return err
	}

	resp, err := core.ReadHandshakeResp(handshakeStream)
	if err != nil || !resp.Success {
		if err == nil {
			log.Printf("Server rejected request: %s", resp.Error)
		}
		os.Exit(1) // Exit process immediately if explicitly rejected by server logic (Auth, Domain Taken)
	}

	log.Printf("✅ Tunnel Established! Public URL mounted at subdomain '%s'", resp.AssignedURL)

	// 4. Infinite Loop: Wait for HTTP streams pushed by Edge Server and pipe them locally
	for {
		stream, err := session.Accept() // Yamux waits until server says "here is an HTTP request"
		if err != nil {
			// Tunnel broke (WiFi disconnect, server restart, etc). Trigger reconnect.
			return err
		}
		
		go proxyLocal(stream, cfg.LocalTarget)
	}
}

// proxyLocal pipes raw TCP bytes between the incoming server stream and local dev service
func proxyLocal(remoteStream net.Conn, localTarget string) {
	defer remoteStream.Close()

	// We connect to the localhost node app or DB
	localConn, err := net.Dial("tcp", localTarget)
	if err != nil {
		log.Printf("Failed to dial local agent service %s: %v", localTarget, err)
		return
	}
	defer localConn.Close()

	// Bidirectional byte streaming
	errc := make(chan error, 2)
	go func() {
		_, err := io.Copy(localConn, remoteStream)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(remoteStream, localConn)
		errc <- err
	}()
	<-errc
}
