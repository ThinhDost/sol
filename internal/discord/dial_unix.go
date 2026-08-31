//go:build !windows

package discord

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// dialDiscordSocket attempts to connect to the Discord Unix Domain Socket on Linux / macOS.
func dialDiscordSocket() (net.Conn, error) {
	possibleDirs := []string{
		os.Getenv("XDG_RUNTIME_DIR"),
		os.Getenv("TMPDIR"),
		os.Getenv("TMP"),
		os.Getenv("TEMP"),
		"/tmp",
	}

	for _, dir := range possibleDirs {
		if dir == "" {
			continue
		}
		for i := 0; i < 10; i++ {
			socketPath := filepath.Join(dir, fmt.Sprintf("discord-ipc-%d", i))
			conn, err := net.DialTimeout("unix", socketPath, 200*time.Millisecond)
			if err == nil {
				return conn, nil
			}
		}
	}
	return nil, fmt.Errorf("could not connect to any Discord Unix domain socket")
}
