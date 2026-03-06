package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"flag"

	"github.com/caddyserver/certmagic"
	"github.com/Cyber-Def/postq-tunnel/internal/core"
	"github.com/Cyber-Def/postq-tunnel/internal/server"
	"github.com/Cyber-Def/postq-tunnel/internal/version"
	"github.com/Cyber-Def/postq-tunnel/pkg/logger"
	"github.com/Cyber-Def/postq-tunnel/pkg/tunnel"
)

func main() {
	v := flag.Bool("version", false, "Print version information")
	flag.Parse()

	if *v {
		version.PrintBanner("qtunnel (Edge Server)")
	}

	logger.InitLogger()
	slog.Info("PostQ-Tunnel Edge Server starting (PQC-Ready)...")
	
	registry := server.NewRegistry()
	proxy := server.BuildProxy(registry)
	
	// 1. Prometheus Metrics tracking
	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "# HELP qtun_active_tunnels The total number of active agent tunnels.\n")
		fmt.Fprintf(w, "# TYPE qtun_active_tunnels gauge\n")
		fmt.Fprintf(w, "qtun_active_tunnels %d\n", registry.TunnelCount())
	})
	
	// Start metrics endpoint quietly on 9090
	go http.ListenAndServe(":9090", metricsMux)

	// 3. Graceful Shutdown Context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Start the PQC control plane for connecting agents
	go startPQCListener(ctx, ":4443", registry)

	// 4. Run Edge Proxy
	domain := os.Getenv("QTUN_DOMAIN")
	if domain == "" {
		slog.Info("No QTUN_DOMAIN set. Starting local HTTP fallback proxy on :8080...")
		
		srv := &http.Server{Addr: ":8080", Handler: proxy}
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("Proxy error", "error", err)
			}
		}()
		
		<-ctx.Done()
		slog.Info("Shutting down Edge Server gracefully...")
		srv.Shutdown(context.Background())
		return
	}

	// CertMagic Automatic HTTPS setup
	certmagic.DefaultACME.Email = os.Getenv("QTUN_EMAIL")
	certmagic.DefaultACME.Agreed = true

	slog.Info("Starting Automatic TLS edge proxy", "domain", domain)
	
	go func() {
		err := certmagic.HTTPS([]string{domain}, proxy)
		if err != nil {
			logger.Fatal("CertMagic fatal error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Edge Proxy Server gracefully. Closing all agent tunnels...")
	time.Sleep(1 * time.Second) // allow final bytes transmission
}

func startPQCListener(ctx context.Context, addr string, registry core.TunnelRegistry) {
	l, err := tunnel.ListenPQC(addr, nil) 
	if err != nil {
		logger.Fatal("Failed to start PQC listener", "error", err)
	}
	slog.Info("PQC Transport running for local agents", "addr", addr)

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		go handleAgentConnection(ctx, conn, registry)
	}
}

func handleAgentConnection(ctx context.Context, conn net.Conn, registry core.TunnelRegistry) {
	// Connection and handshake timeout (10 seconds)
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		conn.Close()
		return
	}

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return
	}

	session, err := tunnel.ServerSession(tlsConn)
	if err != nil {
		conn.Close()
		return
	}
	
	stream, err := session.Accept()
	if err != nil {
		session.Close()
		return
	}
	
	// Stream handshake limit
	stream.SetDeadline(time.Now().Add(5 * time.Second))
	req, err := core.ReadHandshake(stream)
	if err != nil {
		session.Close()
		return
	}
	
	// Reset deadlines for multiplexed streams
	stream.SetDeadline(time.Time{})
	conn.SetDeadline(time.Time{})

	err = registry.Register(req.Subdomain, session)
	if err != nil {
		_ = core.WriteHandshakeResp(stream, core.HandshakeResp{Success: false, Error: err.Error()})
		session.Close()
		return
	}
	
	_ = core.WriteHandshakeResp(stream, core.HandshakeResp{Success: true, AssignedURL: req.Subdomain})
	slog.Info("Subdomain Mounted", "subdomain", req.Subdomain, "active_tunnels", registry.TunnelCount())
	
	go func() {
		<-session.CloseChan()
		registry.Unregister(req.Subdomain)
		slog.Info("Subdomain Unmounted", "subdomain", req.Subdomain, "active_tunnels", registry.TunnelCount())
	}()
}
