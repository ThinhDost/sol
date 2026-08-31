//go:build !windows

package ipc

import (
	"fmt"
	"log"
	"net"
	"os"
)

// Start begins listening on the Unix Domain Socket for Sol IPC events.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}

	// Remove stale socket file if it exists
	_ = os.Remove(s.cfg.SocketPathUnix)

	listener, err := net.Listen("unix", s.cfg.SocketPathUnix)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen on unix domain socket %s: %w", s.cfg.SocketPathUnix, err)
	}

	s.listener = listener
	s.running = true
	s.mu.Unlock()

	log.Printf("[IPC] Listening on Unix Domain Socket: %s", s.cfg.SocketPathUnix)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					log.Printf("[IPC] Accept error: %v", err)
					return
				}
			}

			go s.handleConnection(conn)
		}
	}()

	return nil
}

// Stop gracefully shuts down the Unix Domain Socket listener.
func (s *Server) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.mu.Lock()
		s.running = false
		if s.listener != nil {
			err = s.listener.Close()
		}
		_ = os.Remove(s.cfg.SocketPathUnix)
		s.mu.Unlock()
		close(s.events)
	})
	return err
}

// DialClient attempts to connect as a client to the Sol IPC server (for CLI status/ping).
func DialClient(socketPath string) (net.Conn, error) {
	return net.Dial("unix", socketPath)
}
