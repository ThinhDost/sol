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

	// StateAFK indicates the shell has been idle longer than the configured timeout.
	StateAFK SessionState = "AFK"
)

// StateManager tracks the active session state and formats Discord activity objects.
type StateManager struct {
	mu           sync.Mutex
	cfg          *config.Config
	currentState SessionState
	lastActive   time.Time
	currentCmd   string
	currentCwd   string
	currentShell string
	startTime    time.Time
}

// NewStateManager creates a new state manager.
func NewStateManager(cfg *config.Config) *StateManager {
	now := time.Now()
	return &StateManager{
		cfg:          cfg,
		currentState: StateIdle,
		lastActive:   now,
		startTime:    now,
		currentShell: "powershell",
	}
}

// ProcessEvent updates the session state based on an incoming shell event.
func (sm *StateManager) ProcessEvent(event *ipc.ShellEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	sm.lastActive = now

	if event.Shell != "" {
		sm.currentShell = event.Shell
	}
	if event.Cwd != "" {
		sm.currentCwd = event.Cwd
	}

	switch event.Event {
	case ipc.EventStart:
		sm.currentState = StateRunning
		sm.currentCmd = event.Cmd
		sm.startTime = now
	case ipc.EventIdle:
		if sm.currentState == StateRunning {
			sm.currentState = StateIdle
			sm.currentCmd = ""
			sm.startTime = now
		}
	case ipc.EventExit:
		sm.currentState = StateIdle
		sm.currentCmd = ""
	}
}

// CheckIdleTimeout transitions the state to AFK if no activity has occurred within the idle timeout.
func (sm *StateManager) CheckIdleTimeout() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentState == StateIdle && time.Since(sm.lastActive) > sm.cfg.IdleTimeout {
		sm.currentState = StateAFK
		return true
	}
	return false
}

// BuildDiscordActivity constructs a rich presence activity payload based on the current state.
func (sm *StateManager) BuildDiscordActivity() *discord.Activity {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	shellAsset, shellDisplay := config.ResolveShellAsset(sm.currentShell)
	startUnix := sm.startTime.Unix()

	// Try to detect Git repository context
	var gitCtx string
	if sm.currentCwd != "" {
		if gitInfo, err := git.DetectGitContext(sm.currentCwd); err == nil {
			gitCtx = fmt.Sprintf("📁 %s [git: %s]", gitInfo.RepoName, gitInfo.Branch)
		} else {
			dirName := filepath.Base(sm.currentCwd)
			if sm.cfg.AnonymizeHomePath {
				dirName = AnonymizePath(sm.currentCwd)
			}
			gitCtx = fmt.Sprintf("📁 %s", dirName)
		}
	} else {
		gitCtx = "Terminal Session"
	}

	var buttons []*discord.ActivityButton
	if sm.cfg.GitHubURL != "" {
		buttons = append(buttons, &discord.ActivityButton{
			Label: "View on GitHub",
			URL:   sm.cfg.GitHubURL,
		})
	}

	largeKey := sm.cfg.LargeImageKey
	if largeKey == "" {
		largeKey = "fubuki"
	}
	largeText := sm.cfg.LargeImageText
	if largeText == "" {
		largeText = "Sol Terminal"
	}

	switch sm.currentState {
	case StateRunning:
		sanitizedCmd := SanitizeCommand(sm.currentCmd)
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
