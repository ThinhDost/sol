package engine

import (
	"testing"
	"time"

	"github.com/sol-rpc/sol/internal/config"
	"github.com/sol-rpc/sol/internal/ipc"
)

func TestMultiSessionPriorityArbiter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.LargeImageKey = "fubuki"
	sm := NewStateManager(cfg)

	t0 := time.Now()

	// 1. Tab 1 (PID 100) opens and goes idle
	sm.ProcessEvent(&ipc.ShellEvent{
		Event:     ipc.EventIdle,
		Shell:     "powershell",
		Cwd:       "C:/Users/ADMIN",
		Pid:       100,
		Timestamp: t0.Unix(),
	})

	activity := sm.BuildDiscordActivity()
	if activity.Details != "Idle at prompt" {
		t.Errorf("expected Idle at prompt, got %s", activity.Details)
	}

	// 2. Tab 2 (PID 200) opens and runs docker compose
	sm.ProcessEvent(&ipc.ShellEvent{
		Event:     ipc.EventStart,
		Cmd:       "docker compose up -d",
		Shell:     "powershell",
		Cwd:       "C:/Users/ADMIN/Sol",
		Pid:       200,
		Timestamp: t0.Add(1 * time.Minute).Unix(),
	})

	activity = sm.BuildDiscordActivity()
	if activity.Details != "🐳 Docker Compose Execution" {
		t.Errorf("expected Docker Compose Execution, got %s", activity.Details)
	}
	if activity.Assets.SmallImage != "docker" {
		t.Errorf("expected small image 'docker', got '%s'", activity.Assets.SmallImage)
	}

	// 3. Tab 1 (PID 100) sends an idle event (prompt redraw or focus switch)
	sm.ProcessEvent(&ipc.ShellEvent{
		Event:     ipc.EventIdle,
		Shell:     "powershell",
		Cwd:       "C:/Users/ADMIN",
		Pid:       100,
		Timestamp: t0.Add(2 * time.Minute).Unix(),
	})

	// CRITICAL CHECK: Tab 2 is still running, so Discord activity MUST NOT be overwritten by Tab 1's idle!
	activity = sm.BuildDiscordActivity()
	if activity.Details != "🐳 Docker Compose Execution" {
		t.Fatalf("CRITICAL BUG: Running session in Tab 2 was clobbered by Tab 1 idle! Got: %s", activity.Details)
	}

	// 4. Tab 2 completes docker compose and goes idle
	sm.ProcessEvent(&ipc.ShellEvent{
		Event:     ipc.EventIdle,
		Shell:     "powershell",
		Cwd:       "C:/Users/ADMIN/Sol",
		Pid:       200,
		Timestamp: t0.Add(3 * time.Minute).Unix(),
	})

	activity = sm.BuildDiscordActivity()
	if activity.Details != "Idle at prompt" {
		t.Errorf("expected Idle at prompt after completion, got %s", activity.Details)
	}

	// 5. Check continuous session timer (must match t0, not reset)
	if activity.Timestamps.Start != sm.globalSessionStart.Unix() {
		t.Errorf("expected continuous session timer %d, got %d", sm.globalSessionStart.Unix(), activity.Timestamps.Start)
	}
}

func TestSessionExitFallback(t *testing.T) {
	cfg := config.DefaultConfig()
	sm := NewStateManager(cfg)

	// Tab 1 & Tab 2 open
	sm.ProcessEvent(&ipc.ShellEvent{Event: ipc.EventIdle, Pid: 101, Cwd: "C:/Dir1"})
	sm.ProcessEvent(&ipc.ShellEvent{Event: ipc.EventIdle, Pid: 102, Cwd: "C:/Dir2"})

	if len(sm.sessions) != 2 {
		t.Fatalf("expected 2 active sessions, got %d", len(sm.sessions))
	}

	// Tab 2 closes
	sm.ProcessEvent(&ipc.ShellEvent{Event: ipc.EventExit, Pid: 102})
	if len(sm.sessions) != 1 {
		t.Fatalf("expected 1 active session after exit, got %d", len(sm.sessions))
	}

	if _, exists := sm.sessions[101]; !exists {
		t.Errorf("expected Tab 1 (PID 101) to remain active")
	}
}
