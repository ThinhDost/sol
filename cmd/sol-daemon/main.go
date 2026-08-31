// Package main is the entry point for the Sol Discord Rich Presence background daemon.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sol-rpc/sol/internal/config"
	"github.com/sol-rpc/sol/internal/engine"
)

const Version = "1.0.0"

func main() {
	clientID := flag.String("client-id", "", "Discord Application Client ID")
	githubURL := flag.String("github-url", "", "GitHub repository link to display on Discord profile")
	idleMinutes := flag.Int("idle-timeout", 10, "Minutes of inactivity before switching to AFK state")
	rateLimitSec := flag.Int("rate-limit", 15, "Minimum seconds between Discord Rich Presence updates")
	showVersion := flag.Bool("version", false, "Print version information and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Sol Daemon v%s\n", Version)
		return
	}

	cfg := config.DefaultConfig()
	if *clientID != "" {
		cfg.DiscordClientID = *clientID
	}
	if *githubURL != "" {
		cfg.GitHubURL = *githubURL
	}
	if *idleMinutes > 0 {
		cfg.IdleTimeout = time.Duration(*idleMinutes) * time.Minute
	}
	if *rateLimitSec > 0 {
		cfg.RateLimitInterval = time.Duration(*rateLimitSec) * time.Second
	}

	log.Printf("==================================================")
	log.Printf(" Sol Daemon v%s Starting...", Version)
	log.Printf(" Discord Client ID : %s", cfg.DiscordClientID)
	if cfg.GitHubURL != "" {
		log.Printf(" GitHub Repository  : %s", cfg.GitHubURL)
	}
	log.Printf(" Idle Timeout       : %v", cfg.IdleTimeout)
	log.Printf(" Rate Limit Interval: %v", cfg.RateLimitInterval)
	log.Printf("==================================================")

	coordinator := engine.NewCoordinator(cfg)
	if err := coordinator.Start(); err != nil {
		log.Fatalf("[FATAL] Failed to start coordinator: %v", err)
	}

	// Handle graceful shutdown on OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[Sol] Received shutdown signal: %v. Cleaning up...", sig)
	coordinator.Stop()
	log.Printf("[Sol] Exiting cleanly. Goodbye!")
}
