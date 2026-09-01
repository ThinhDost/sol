package engine

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sol-rpc/sol/internal/config"
	"github.com/sol-rpc/sol/internal/discord"
	"github.com/sol-rpc/sol/internal/git"
	"github.com/sol-rpc/sol/internal/ipc"
)

// SessionState represents the terminal activity state.
type SessionState string

const (
	// StateRunning indicates a command is actively executing.
	StateRunning SessionState = "RUNNING"

	// StateIdle indicates the shell is waiting for user input at the prompt.
	StateIdle SessionState = "IDLE"

	// StateAFK indicates all shells have been idle longer than the configured timeout.
	StateAFK SessionState = "AFK"
)

// TerminalSession tracks the state of a single terminal window or tab (keyed by PID).
type TerminalSession struct {
	PID          int          `json:"pid"`
	Shell        string       `json:"shell"`
	State        SessionState `json:"state"`
	CurrentCmd   string       `json:"cmd"`
	Cwd          string       `json:"cwd"`
	LastActive   time.Time    `json:"last_active"`
	CommandStart time.Time    `json:"command_start"`
}

// StateManager tracks multiple concurrent terminal sessions and arbitrates active Rich Presence.
type StateManager struct {
	mu                 sync.Mutex
	cfg                *config.Config
	sessions           map[int]*TerminalSession
	globalSessionStart time.Time
	lastGlobalActive   time.Time
}

// NewStateManager creates a new multi-session state manager.
func NewStateManager(cfg *config.Config) *StateManager {
	now := time.Now()
	return &StateManager{
		cfg:                cfg,
		sessions:           make(map[int]*TerminalSession),
		globalSessionStart: now,
		lastGlobalActive:   now,
	}
}

// ProcessEvent updates session registry based on an incoming shell event.
func (sm *StateManager) ProcessEvent(event *ipc.ShellEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	sm.lastGlobalActive = now

	// If this is the first session, or session was reset, record global work start time
	if len(sm.sessions) == 0 || sm.globalSessionStart.IsZero() {
		sm.globalSessionStart = now
	}

	pid := event.Pid
	if pid <= 0 {
		// Fallback for legacy hooks without PID
		pid = 1
	}

	if event.Event == ipc.EventExit {
		delete(sm.sessions, pid)
		return
	}

	sess, exists := sm.sessions[pid]
	if !exists {
		shellName := event.Shell
		if shellName == "" {
			shellName = "powershell"
		}
		sess = &TerminalSession{
			PID:          pid,
			Shell:        shellName,
			State:        StateIdle,
			Cwd:          event.Cwd,
			LastActive:   now,
			CommandStart: now,
		}
		sm.sessions[pid] = sess
	}

	if event.Shell != "" {
		sess.Shell = event.Shell
	}
	if event.Cwd != "" {
		sess.Cwd = event.Cwd
	}
	sess.LastActive = now

	switch event.Event {
	case ipc.EventStart:
		sess.State = StateRunning
		sess.CurrentCmd = event.Cmd
		sess.CommandStart = now

	case ipc.EventIdle:
		sess.State = StateIdle
		sess.CurrentCmd = ""

	case ipc.EventPing:
		// Keepalive
	}
}

// CleanupDeadProcesses scans registered PIDs and removes processes that no longer exist in OS.
// Returns true if state changed.
func (sm *StateManager) CleanupDeadProcesses() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	changed := false
	for pid := range sm.sessions {
		// If PID is 1 (fallback) skip OS check, otherwise verify liveness
		if pid > 1 && !isProcessAlive(pid) {
			delete(sm.sessions, pid)
			changed = true
		}
	}

	// If all terminal sessions are closed and idle timeout elapsed, reset session start
	if len(sm.sessions) == 0 && time.Since(sm.lastGlobalActive) > sm.cfg.IdleTimeout {
		if !sm.globalSessionStart.IsZero() {
			sm.globalSessionStart = time.Time{}
			changed = true
		}
	}

	return changed
}

// CheckIdleTimeout returns true if all sessions are idle or no active session for longer than timeout.
func (sm *StateManager) CheckIdleTimeout() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return time.Since(sm.lastGlobalActive) > sm.cfg.IdleTimeout
}

// resolveActiveSession determines which terminal session should dominate the Discord display.
// Priority 1: Any session currently in StateRunning (latest command start wins).
// Priority 2: Any session in StateIdle (latest user activity wins).
func (sm *StateManager) resolveActiveSession() (*TerminalSession, SessionState) {
	if len(sm.sessions) == 0 {
		if time.Since(sm.lastGlobalActive) > sm.cfg.IdleTimeout {
			return nil, StateAFK
		}
		return nil, StateIdle
	}

	// 1. Check for running commands across all tabs/windows
	var latestRunning *TerminalSession
	for _, sess := range sm.sessions {
		if sess.State == StateRunning {
			if latestRunning == nil || sess.CommandStart.After(latestRunning.CommandStart) {
				latestRunning = sess
			}
		}
	}
	if latestRunning != nil {
		return latestRunning, StateRunning
	}

	// 2. If all tabs are idle, pick the one where user last interacted
	var latestIdle *TerminalSession
	for _, sess := range sm.sessions {
		if latestIdle == nil || sess.LastActive.After(latestIdle.LastActive) {
			latestIdle = sess
		}
	}

	if time.Since(sm.lastGlobalActive) > sm.cfg.IdleTimeout {
		return latestIdle, StateAFK
	}

	return latestIdle, StateIdle
}

// BuildDiscordActivity constructs the Rich Presence activity object.
// - Timestamps.Start represents the continuous total work elapsed timer.
// - LargeImage is always the configured LargeImageKey ("fubuki").
// - SmallImage dynamically reflects the active tool / command / shell.
func (sm *StateManager) BuildDiscordActivity() *discord.Activity {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	activeSess, overallState := sm.resolveActiveSession()

	largeKey := sm.cfg.LargeImageKey
	if largeKey == "" {
		largeKey = "fubuki"
	}
	largeText := sm.cfg.LargeImageText
	if largeText == "" {
		largeText = "Sol Terminal"
	}

	// Continuous total session work elapsed timer
	startUnix := sm.globalSessionStart.Unix()
	if sm.globalSessionStart.IsZero() {
		startUnix = time.Now().Unix()
	}

	// Buttons (e.g. GitHub Repository link)
	var buttons []*discord.ActivityButton
	if sm.cfg.GitHubURL != "" {
		buttons = append(buttons, &discord.ActivityButton{
			Label: "View on GitHub",
			URL:   sm.cfg.GitHubURL,
		})
	}

	// Resolve shell asset
	shellName := "powershell"
	cwd := ""
	if activeSess != nil {
		shellName = activeSess.Shell
		cwd = activeSess.Cwd
	}
	shellAsset, shellDisplay := config.ResolveShellAsset(shellName)

	// Resolve Git context
	var gitCtx string
	if cwd != "" {
		if gitInfo, err := git.DetectGitContext(cwd); err == nil {
			gitCtx = fmt.Sprintf("📁 %s [git: %s]", gitInfo.RepoName, gitInfo.Branch)
		} else {
			dirName := filepath.Base(cwd)
			if sm.cfg.AnonymizeHomePath {
				dirName = AnonymizePath(cwd)
			}
			gitCtx = fmt.Sprintf("📁 %s", dirName)
		}
	} else {
		gitCtx = "Terminal Session"
	}

	switch overallState {
	case StateRunning:
		sanitizedCmd := SanitizeCommand(activeSess.CurrentCmd)
		toolAsset := config.ResolveToolAsset(sanitizedCmd)

		details := fmt.Sprintf("Running: %s", sanitizedCmd)
		fields := strings.Fields(sanitizedCmd)
		if len(fields) > 0 {
			base := strings.ToLower(fields[0])
			base = strings.TrimSuffix(base, ".exe")
			base = strings.TrimSuffix(base, ".cmd")

			switch base {
			case "agy", "antigravity", "agy-cli", "antigravity-cli":
				if len(fields) == 1 {
					details = "✨ Coding with Antigravity AI"
				} else {
					prompt := strings.Join(fields[1:], " ")
					details = fmt.Sprintf("🤖 AGY: %s", prompt)
				}
			case "claude":
				details = "🤖 Prompting Claude Code"
			case "copilot":
				details = "🤖 Copilot CLI Session"
			case "ollama":
				details = "🧠 Running Ollama Local LLM"
			case "docker":
				if len(fields) >= 2 && strings.ToLower(fields[1]) == "compose" {
					details = "🐳 Docker Compose Execution"
				}
			}
		}

		return &discord.Activity{
			Details: details,
			State:   gitCtx,
			Timestamps: &discord.ActivityTimestamps{
				Start: startUnix,
			},
			Assets: &discord.ActivityAssets{
				LargeImage: largeKey,
				LargeText:  largeText,
				SmallImage: toolAsset.AssetKey,
				SmallText:  toolAsset.DisplayName,
			},
			Buttons: buttons,
		}

	case StateAFK:
		return &discord.Activity{
			Details: "Taking a break / AFK",
			State:   "💤 Idle in terminal",
			Timestamps: &discord.ActivityTimestamps{
				Start: startUnix,
			},
			Assets: &discord.ActivityAssets{
				LargeImage: largeKey,
				LargeText:  largeText,
				SmallImage: "sleep",
				SmallText:  "AFK / Away",
			},
			Buttons: buttons,
		}

	default: // StateIdle
		return &discord.Activity{
			Details: "Idle at prompt",
			State:   gitCtx,
			Timestamps: &discord.ActivityTimestamps{
				Start: startUnix,
			},
			Assets: &discord.ActivityAssets{
				LargeImage: largeKey,
				LargeText:  largeText,
				SmallImage: shellAsset,
				SmallText:  fmt.Sprintf("%s (Idle)", shellDisplay),
			},
			Buttons: buttons,
		}
	}
}
