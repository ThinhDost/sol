//go:build !windows

package engine

import (
	"syscall"
)

// isProcessAlive queries Unix kernel to check if the given PID is actively running.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
