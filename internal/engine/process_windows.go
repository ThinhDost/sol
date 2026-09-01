//go:build windows

package engine

import (
	"golang.org/x/sys/windows"
)

// isProcessAlive queries Windows kernel to check if the given PID is actively running.
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(handle)

	var exitCode uint32
	if err := windows.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false
	}

	// 259 is STILL_ACTIVE in Win32 API
	return exitCode == 259
}
