package ipc

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/sol-rpc/sol/internal/config"
)

func TestIPCServerClient(t *testing.T) {
	cfg := config.DefaultConfig()

	// Use unique pipe / socket name for testing
	if runtime.GOOS == "windows" {
		cfg.PipeNameWindows = fmt.Sprintf(`\\.\pipe\sol-test-%d`, time.Now().UnixNano())
	} else {
		cfg.SocketPathUnix = fmt.Sprintf("/tmp/sol-test-%d.sock", time.Now().UnixNano())
		defer os.Remove(cfg.SocketPathUnix)
	}

	server := NewServer(cfg)
	if err := server.Start(); err != nil {
		t.Fatalf("failed to start IPC server: %v", err)
	}
	defer server.Stop()

	// Wait briefly for listener to be active
	time.Sleep(50 * time.Millisecond)

	var target string
	if runtime.GOOS == "windows" {
		target = cfg.PipeNameWindows
	} else {
		target = cfg.SocketPathUnix
	}

	conn, err := DialClient(target)
	if err != nil {
		t.Fatalf("failed to dial IPC server: %v", err)
	}
	defer conn.Close()

	testEvent := ShellEvent{
		Event:     EventStart,
		Cmd:       "cargo build --release",
		Cwd:       "C:/Projects/sol",
		Shell:     "powershell",
		Pid:       1234,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}
	data = append(data, '\n')

	if _, err := conn.Write(data); err != nil {
		t.Fatalf("failed to write event to IPC: %v", err)
	}

	select {
	case received := <-server.Events():
		if received.Event != EventStart {
			t.Errorf("expected event %s, got %s", EventStart, received.Event)
		}
		if received.Cmd != testEvent.Cmd {
			t.Errorf("expected cmd %s, got %s", testEvent.Cmd, received.Cmd)
		}
		if received.Shell != testEvent.Shell {
			t.Errorf("expected shell %s, got %s", testEvent.Shell, received.Shell)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for IPC event")
	}
}
