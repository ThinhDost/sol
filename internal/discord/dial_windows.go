//go:build windows

package discord

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// dialDiscordSocket attempts to connect to the Discord Named Pipe on Windows.
// It iterates through \\.\pipe\discord-ipc-0 to \\.\pipe\discord-ipc-9.
func dialDiscordSocket() (net.Conn, error) {
	for i := 0; i < 10; i++ {
		pipePath := fmt.Sprintf(`\\.\pipe\discord-ipc-%d`, i)
		timeout := 200 * time.Millisecond
		conn, err := winio.DialPipe(pipePath, &timeout)
		if err == nil {
			return conn, nil
		}
	}
	return nil, fmt.Errorf("could not connect to any Discord IPC named pipe (discord-ipc-0 to 9)")
}
