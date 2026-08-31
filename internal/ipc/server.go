package ipc

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"

	"github.com/sol-rpc/sol/internal/config"
)

// Server listens for incoming shell events from local hooks.
type Server struct {
	cfg      *config.Config
	listener net.Listener
	events   chan *ShellEvent
	mu       sync.Mutex
	running  bool
	stopOnce sync.Once
	stopChan chan struct{}
}

// NewServer creates a new local IPC server instance.
func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:      cfg,
		events:   make(chan *ShellEvent, 64),
		stopChan: make(chan struct{}),
	}
}

// Events returns the receive-only channel for incoming shell events.
func (s *Server) Events() <-chan *ShellEvent {
	return s.events
}

// handleConnection reads line-delimited JSON events from a connected shell client.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			var event ShellEvent
			if err := json.Unmarshal(line, &event); err == nil {
				select {
				case s.events <- &event:
				case <-s.stopChan:
					return
				default:
					// Drop if event queue is overloaded to guarantee non-blocking behavior
					log.Printf("[IPC] Warning: event queue full, dropping event")
				}
			}
		}

		if err != nil {
			if err != io.EOF {
				// Connection ended or error
			}
			return
		}
	}
}
