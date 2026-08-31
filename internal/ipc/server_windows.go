//go:build windows

package ipc

import (
	"fmt"
	"log"
	"net"

	"github.com/Microsoft/go-winio"
)

// Start begins listening on the Windows Named Pipe for Sol IPC events.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}

	listener, err := winio.ListenPipe(s.cfg.PipeNameWindows, nil)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to create named pipe listener %s: %w", s.cfg.PipeNameWindows, err)
	}

	s.listener = listener
	s.running = true
	s.mu.Unlock()

	log.Printf("[IPC] Listening on Windows Named Pipe: %s", s.cfg.PipeNameWindows)

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

// Stop gracefully shuts down the Windows Named Pipe listener.
func (s *Server) Stop() error {
	var err error
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.mu.Lock()
		s.running = false
		if s.listener != nil {
			err = s.listener.Close()
		}
		s.mu.Unlock()
		close(s.events)
	})
	return err
}

// DialClient attempts to connect as a client to the Sol IPC server (for CLI status/ping).
func DialClient(pipePath string) (net.Conn, error) {
	return winio.DialPipe(pipePath, nil)
}
