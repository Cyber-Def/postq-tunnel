package client

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/Cyber-Def/postq-tunnel/internal/core"
	"github.com/Cyber-Def/postq-tunnel/internal/version"
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
	v := flag.Bool("version", false, "Print version information")
	
	flag.Parse()
	
	if *v {
		version.PrintBanner("qtun (Agent)")
	}
	
	if *authFull != "" {
		parts := strings.SplitN(*authFull, ":", 2)
		if len(parts) == 2 {
			cfg.BasicAuthUsr = parts[0]
			cfg.BasicAuthPwd = parts[1]
		}
	}
	
	if cfg.Subdomain == "" {
		slog.Error("Subdomain is required (-sub)")
		os.Exit(1)
	}
	return cfg
}

func RunTunnel(ctx context.Context, cfg *Config) error {
	// 1. Dial Edge Server using PQC TLS (tls.X25519MLKEM768)
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	tlsConn, err := tunnel.DialPQC(dialCtx, cfg.ServerAddr, true) // insecureSkipVerify for typical agent-server connection
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

	// Apply handshake timeout
	_ = handshakeStream.SetDeadline(time.Now().Add(10 * time.Second))

	var cidrs []string
	if cfg.AllowCIDRs != "" {
		for _, s := range strings.Split(cfg.AllowCIDRs, ",") {
			cidrs = append(cidrs, strings.TrimSpace(s))
		}
	}

	req := core.HandshakeReq{
		Version:      core.ProtocolVersion,
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
			slog.Error("Server rejected request", "error", resp.Error)
		}
		os.Exit(1) // Exit process immediately if explicitly rejected by server logic (Auth, Domain Taken)
	}
	
	// Reset deadline for normal operation
	_ = handshakeStream.SetDeadline(time.Time{})
	handshakeStream.Close()

	slog.Info("✅ Tunnel Established!", "url", resp.AssignedURL)

	// 4. Infinite Loop: Wait for HTTP streams pushed by Edge Server and pipe them locally
	for {
		stream, err := session.Accept() // Yamux waits until server says "here is an HTTP request"
		if err != nil {
			// Tunnel broke (WiFi disconnect, server restart, etc). Trigger reconnect.
			return err
		}
		
		go proxyLocal(ctx, stream, cfg.LocalTarget)
	}
}

// proxyLocal pipes raw TCP bytes between the incoming server stream and local dev service
func proxyLocal(ctx context.Context, remoteStream net.Conn, localTarget string) {
	defer remoteStream.Close()

	// We connect to the localhost node app or DB
	localConn, err := net.Dial("tcp", localTarget)
	if err != nil {
		slog.Error("Failed to dial local agent service", "target", localTarget, "error", err)
		return
	}
	defer localConn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		localConn.Close()
		remoteStream.Close()
	}()

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
