// Package main provides the Sol CLI management tool for checking daemon status
// and sending test notifications to the local IPC socket.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/sol-rpc/sol/internal/config"
	"github.com/sol-rpc/sol/internal/ipc"
)

const Version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status", "ping":
		checkStatus()

	case "notify":
		notifyFlags := flag.NewFlagSet("notify", flag.ExitOnError)
		event := notifyFlags.String("event", "start", "Event type (start, idle, exit, ping)")
		cmd := notifyFlags.String("cmd", "", "Command string")
		cwd := notifyFlags.String("cwd", "", "Current working directory")
		shell := notifyFlags.String("shell", "cli", "Shell name")
		_ = notifyFlags.Parse(os.Args[2:])

		sendNotification(*event, *cmd, *cwd, *shell)

	case "version", "--version", "-v":
		fmt.Printf("Sol CLI v%s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf("Sol CLI v%s - Discord Rich Presence for Terminal\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  sol status               Check if Sol daemon is running")
	fmt.Println("  sol ping                 Ping the Sol daemon")
	fmt.Println("  sol notify [options]     Send an event to Sol daemon")
	fmt.Println("  sol version              Print version")
	fmt.Println("\nOptions for 'notify':")
	fmt.Println("  --event <type>           start, idle, exit, ping (default: start)")
	fmt.Println("  --cmd <string>           Command string")
	fmt.Println("  --cwd <path>             Working directory")
	fmt.Println("  --shell <name>           Shell name (powershell, bash, zsh)")
}

func getSocketPath() string {
	cfg := config.DefaultConfig()
	if runtime.GOOS == "windows" {
		return cfg.PipeNameWindows
	}
	return cfg.SocketPathUnix
}

func checkStatus() {
	target := getSocketPath()
	conn, err := ipc.DialClient(target)
	if err != nil {
		fmt.Printf("❌ Sol daemon is NOT running (could not connect to %s)\n", target)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Printf("✅ Sol daemon is RUNNING and listening on %s\n", target)
}

func sendNotification(eventType, cmd, cwd, shell string) {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	event := ipc.ShellEvent{
		Event:     ipc.ShellEventType(eventType),
		Cmd:       cmd,
		Cwd:       cwd,
		Shell:     shell,
		Pid:       os.Getpid(),
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling event: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')

	target := getSocketPath()
	var conn net.Conn
	conn, err = ipc.DialClient(target)
	if err != nil {
		// Fail silently or notify
		fmt.Fprintf(os.Stderr, "Error connecting to Sol daemon: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	_, _ = conn.Write(data)
}
