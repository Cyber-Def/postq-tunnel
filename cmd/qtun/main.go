package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cyber-Def/postq-tunnel/pkg/client"
	"github.com/Cyber-Def/postq-tunnel/pkg/logger"
)

func main() {
	logger.InitLogger()
	cfg := client.ParseFlags()
	
	slog.Info("Starting qtun agent", "target", cfg.LocalTarget)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	// Implement auto-reconnect mechanism with exponential backoff algorithm
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		if ctx.Err() != nil {
			slog.Info("Context cancelled, shutting down gracefully.")
			break
		}

		err := client.RunTunnel(ctx, cfg)
		if err != nil {
			slog.Error("Tunnel transport dropped/failed", "error", err)
		} else {
			slog.Info("Tunnel gracefully closed by manual termination.")
			break
		}

		slog.Info("Reconnecting to Edge Server...", "backoff", backoff)
		
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			slog.Info("Context cancelled during backoff, shutting down.")
			return
		}
		
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}
