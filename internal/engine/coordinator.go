package engine

import (
	"log"
	"sync"
	"time"

	"github.com/sol-rpc/sol/internal/config"
	"github.com/sol-rpc/sol/internal/discord"
	"github.com/sol-rpc/sol/internal/ipc"
)

// Coordinator wires together the IPC server, state manager, rate limiter, and Discord client.
type Coordinator struct {
	cfg          *config.Config
	ipcServer    *ipc.Server
	discordCli   *discord.Client
	stateManager *StateManager
	rateLimiter  *RateLimiter
	stopChan     chan struct{}
	stopOnce     sync.Once
}

// NewCoordinator initializes the core Sol coordinator engine.
func NewCoordinator(cfg *config.Config) *Coordinator {
	discordCli := discord.NewClient(cfg.DiscordClientID)
	stateManager := NewStateManager(cfg)

	c := &Coordinator{
		cfg:          cfg,
		ipcServer:    ipc.NewServer(cfg),
		discordCli:   discordCli,
		stateManager: stateManager,
		stopChan:     make(chan struct{}),
	}

	// Rate limiter with dispatch callback to Discord client
	c.rateLimiter = NewRateLimiter(cfg.RateLimitInterval, func(activity *discord.Activity) error {
		if !c.discordCli.IsConnected() {
			if err := c.discordCli.Connect(); err != nil {
				// Discord desktop client not open or pipe busy
				return err
			}
			log.Printf("[Discord] Connected to Discord IPC successfully")
		}

		if err := c.discordCli.SetActivity(activity); err != nil {
			log.Printf("[Discord] SetActivity error: %v", err)
			return err
		}
		return nil
	})

	return c
}

// Start begins the IPC server and starts the main event coordination loop.
func (c *Coordinator) Start() error {
	if err := c.ipcServer.Start(); err != nil {
		return err
	}

	// Try initial connection to Discord
	go func() {
		if err := c.discordCli.Connect(); err != nil {
			log.Printf("[Discord] Initial connection failed: %v (will retry automatically)", err)
		} else {
			log.Printf("[Discord] Connected to Discord IPC successfully!")
			// Send initial idle presence
			c.rateLimiter.Submit(c.stateManager.BuildDiscordActivity())
		}
	}()

	go c.runLoop()
	return nil
}

// runLoop processes shell events and background timers.
func (c *Coordinator) runLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return

		case event, ok := <-c.ipcServer.Events():
			if !ok {
				return
			}
			c.handleShellEvent(event)

		case <-ticker.C:
			c.handleTick()
		}
	}
}

// handleShellEvent processes a single event received from a shell hook.
func (c *Coordinator) handleShellEvent(event *ipc.ShellEvent) {
	c.stateManager.ProcessEvent(event)
	activity := c.stateManager.BuildDiscordActivity()
	c.rateLimiter.Submit(activity)
}

// handleTick checks idle timeouts and attempts reconnection if Discord was closed.
func (c *Coordinator) handleTick() {
	if c.stateManager.CheckIdleTimeout() {
		activity := c.stateManager.BuildDiscordActivity()
		c.rateLimiter.Submit(activity)
	}

	// Reconnect if disconnected
	if !c.discordCli.IsConnected() {
		_ = c.discordCli.Connect()
	}
}

// Stop gracefully shuts down the coordinator and all subcomponents.
func (c *Coordinator) Stop() {
	c.stopOnce.Do(func() {
		close(c.stopChan)
		c.rateLimiter.Stop()
		_ = c.ipcServer.Stop()
		_ = c.discordCli.ClearActivity()
		_ = c.discordCli.Close()
		log.Printf("[Sol] Coordinator stopped gracefully")
	})
}
