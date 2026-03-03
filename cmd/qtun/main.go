package main

import (
	"log"
	"time"

	"github.com/Cyber-Def/postq-tunnel/pkg/client"
)

func main() {
	cfg := client.ParseFlags()
	
	log.Printf("Starting qtun agent. Targeting local port: %s", cfg.LocalTarget)
	
	// Implement auto-reconnect mechanism with exponential backoff algorithm
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		err := client.RunTunnel(cfg)
		if err != nil {
			log.Printf("Tunnel transport dropped/failed: %v", err)
		} else {
			log.Println("Tunnel gracefully closed by manual termination.")
			break
		}

		log.Printf("Reconnecting to Edge Server in %v...", backoff)
		time.Sleep(backoff)
		
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}
