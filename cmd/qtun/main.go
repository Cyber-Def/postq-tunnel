package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Cyber-Def/postq-tunnel/pkg/client"
)

func main() {
	cfg := client.ParseFlags()
	
	log.Printf("Starting qtun agent. Targeting local port: %s", cfg.LocalTarget)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	
	// Implement auto-reconnect mechanism with exponential backoff algorithm
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		if ctx.Err() != nil {
			log.Println("Context cancelled, shutting down gracefully.")
			break
		}

		err := client.RunTunnel(ctx, cfg)
		if err != nil {
			log.Printf("Tunnel transport dropped/failed: %v", err)
		} else {
			log.Println("Tunnel gracefully closed by manual termination.")
			break
		}

		log.Printf("Reconnecting to Edge Server in %v...", backoff)
		
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			log.Println("Context cancelled during backoff, shutting down.")
			return
		}
		
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}
