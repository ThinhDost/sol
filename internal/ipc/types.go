// Package ipc handles local inter-process communication between shell hooks (PowerShell/Bash)
// and the Sol daemon.
package ipc

// ShellEventType defines the type of event emitted by the shell.
type ShellEventType string

const (
	// EventStart indicates a command has begun execution (preexec).
	EventStart ShellEventType = "start"

	// EventIdle indicates a command has finished and the shell prompt is waiting (precmd / prompt).
	EventIdle ShellEventType = "idle"

	// EventExit indicates a shell session/tab is terminating.
	EventExit ShellEventType = "exit"

	// EventPing indicates a keep-alive / healthcheck ping.
	EventPing ShellEventType = "ping"
)

// ShellEvent represents the payload sent from the shell hook to the Sol daemon.
type ShellEvent struct {
	// Event is the lifecycle event type ("start", "idle", "exit", "ping").
	Event ShellEventType `json:"event"`

	// Cmd is the command line string being executed (only populated for "start").
	Cmd string `json:"cmd,omitempty"`

	// Cwd is the current working directory path.
	Cwd string `json:"cwd,omitempty"`

	// Shell is the shell name (e.g., "powershell", "pwsh", "bash", "zsh").
	Shell string `json:"shell,omitempty"`

	// Pid is the process ID of the shell.
	Pid int `json:"pid,omitempty"`

	// Timestamp is the Unix timestamp (in seconds or milliseconds) when the event occurred.
	Timestamp int64 `json:"timestamp,omitempty"`

	// ExitCode is the exit status of the previous command (only populated for "idle").
	ExitCode int `json:"exit_code,omitempty"`
}
