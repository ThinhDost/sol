// Package config provides configuration settings, default values,
// and icon mapping dictionaries for the Sol Discord Rich Presence system.
package config

import (
	"os"
	"strings"
	"time"
)

// DefaultDiscordClientID is the Discord Application ID for Sol.
const DefaultDiscordClientID = "1543849393432698940"

// DefaultGitHubURL is the default repository link displayed on Discord Rich Presence.
const DefaultGitHubURL = "https://github.com/ThinhDost/sol"

// Config represents the runtime configuration for the Sol daemon.
type Config struct {
	// DiscordClientID is the Discord Application ID registered on Discord Developer Portal.
	DiscordClientID string `json:"client_id"`

	// GitHubURL is the optional repository link added as an interactive button on your Discord profile.
	GitHubURL string `json:"github_url,omitempty"`

	// PipeNameWindows is the Windows Named Pipe name for Sol IPC.
	PipeNameWindows string `json:"pipe_name_windows,omitempty"`

	// SocketPathUnix is the Unix Domain Socket path for Sol IPC.
	SocketPathUnix string `json:"socket_path_unix,omitempty"`

	// IdleTimeout is the duration of shell inactivity before switching to AFK/Chilling state.
	IdleTimeout time.Duration `json:"idle_timeout,omitempty"`

	// RateLimitInterval is the minimum interval between Discord Rich Presence updates.
	RateLimitInterval time.Duration `json:"rate_limit_interval,omitempty"`

	// MaskSecrets determines whether passwords, tokens, and API keys are redacted.
	MaskSecrets bool `json:"mask_secrets"`

	// AnonymizeHomePath determines whether user home directory is shortened to "~".
	AnonymizeHomePath bool `json:"anonymize_home_path"`
}

// DefaultConfig returns a production-ready default configuration for Sol.
func DefaultConfig() *Config {
	clientID := os.Getenv("SOL_CLIENT_ID")
	if clientID == "" {
		clientID = DefaultDiscordClientID
	}

	githubURL := os.Getenv("SOL_GITHUB_URL")
	if githubURL == "" {
		githubURL = DefaultGitHubURL
	}

	return &Config{
		DiscordClientID:   clientID,
		GitHubURL:         githubURL,
		PipeNameWindows:   `\\.\pipe\sol-ipc`,
		SocketPathUnix:    "/tmp/sol.sock",
		IdleTimeout:       10 * time.Minute,
		RateLimitInterval: 15 * time.Second,
		MaskSecrets:       true,
		AnonymizeHomePath: true,
	}
}

// ToolAssetInfo contains display information and Discord asset keys for recognized commands.
type ToolAssetInfo struct {
	// AssetKey corresponds to the uploaded asset key in Discord Developer Portal.
	AssetKey string
	// DisplayName is a user-friendly label for the tool (e.g., "Cargo / Rust").
	DisplayName string
}

// toolAssetMap maps common CLI commands to their Discord asset keys and friendly names.
var toolAssetMap = map[string]ToolAssetInfo{
	// Programming Languages & Toolchains
	"go":      {AssetKey: "golang", DisplayName: "Go Toolchain"},
	"cargo":   {AssetKey: "rust", DisplayName: "Cargo (Rust)"},
	"rustc":   {AssetKey: "rust", DisplayName: "Rust Compiler"},
	"python":  {AssetKey: "python", DisplayName: "Python"},
	"python3": {AssetKey: "python", DisplayName: "Python 3"},
	"py":      {AssetKey: "python", DisplayName: "Python"},
	"node":    {AssetKey: "nodejs", DisplayName: "Node.js"},
	"npm":     {AssetKey: "npm", DisplayName: "NPM"},
	"yarn":    {AssetKey: "yarn", DisplayName: "Yarn"},
	"pnpm":    {AssetKey: "pnpm", DisplayName: "PNPM"},
	"bun":     {AssetKey: "bun", DisplayName: "Bun Runtime"},
	"deno":    {AssetKey: "deno", DisplayName: "Deno Runtime"},
	"dotnet":  {AssetKey: "dotnet", DisplayName: ".NET SDK"},
	"java":    {AssetKey: "java", DisplayName: "Java (JVM)"},
	"javac":   {AssetKey: "java", DisplayName: "Java Compiler"},
	"gcc":     {AssetKey: "c", DisplayName: "GCC Compiler"},
	"g++":     {AssetKey: "cpp", DisplayName: "G++ Compiler"},
	"clang":   {AssetKey: "clang", DisplayName: "Clang/LLVM"},
	"make":    {AssetKey: "make", DisplayName: "GNU Make"},
	"cmake":   {AssetKey: "cmake", DisplayName: "CMake"},

	// DevOps & Containers
	"docker":     {AssetKey: "docker", DisplayName: "Docker"},
	"docker-compose": {AssetKey: "docker", DisplayName: "Docker Compose"},
	"podman":     {AssetKey: "podman", DisplayName: "Podman"},
	"kubectl":    {AssetKey: "kubernetes", DisplayName: "Kubernetes CLI"},
	"k8s":        {AssetKey: "kubernetes", DisplayName: "Kubernetes"},
	"terraform":  {AssetKey: "terraform", DisplayName: "Terraform"},
	"helm":       {AssetKey: "helm", DisplayName: "Helm"},

	// Version Control
	"git": {AssetKey: "git", DisplayName: "Git VCS"},
	"gh":  {AssetKey: "github", DisplayName: "GitHub CLI"},

	// Editors & Terminal Tools
	"nvim":   {AssetKey: "neovim", DisplayName: "Neovim"},
	"vim":    {AssetKey: "vim", DisplayName: "Vim"},
	"code":   {AssetKey: "vscode", DisplayName: "VS Code"},
	"htop":   {AssetKey: "terminal", DisplayName: "System Monitor"},
	"btop":   {AssetKey: "terminal", DisplayName: "Btop Monitor"},
	"curl":   {AssetKey: "network", DisplayName: "cURL Transfer"},
	"ping":   {AssetKey: "network", DisplayName: "Network Ping"},
	"ssh":    {AssetKey: "ssh", DisplayName: "Secure Shell (SSH)"},
	"ollama": {AssetKey: "ai", DisplayName: "Ollama AI"},
}

// ResolveToolAsset inspects a raw command string and returns its corresponding asset info.
// If the command is unrecognized, a default terminal asset is returned.
func ResolveToolAsset(rawCmd string) ToolAssetInfo {
	trimmed := strings.TrimSpace(rawCmd)
	if trimmed == "" {
		return ToolAssetInfo{AssetKey: "terminal", DisplayName: "Terminal"}
	}

	// Extract the base command name (e.g. "git" from "git commit -m ...")
	fields := strings.Fields(trimmed)
	baseCmd := strings.ToLower(fields[0])

	// Remove common file extension if on Windows (e.g. "cargo.exe" -> "cargo")
	baseCmd = strings.TrimSuffix(baseCmd, ".exe")
	baseCmd = strings.TrimSuffix(baseCmd, ".cmd")
	baseCmd = strings.TrimSuffix(baseCmd, ".bat")

	// Match in the tool asset map
	if info, found := toolAssetMap[baseCmd]; found {
		return info
	}

	// Default fallback
	return ToolAssetInfo{
		AssetKey:    "terminal",
		DisplayName: "Terminal Command",
	}
}

// ResolveShellAsset returns the Discord icon asset key for the specified shell name.
func ResolveShellAsset(shellName string) (assetKey string, displayName string) {
	lower := strings.ToLower(strings.TrimSpace(shellName))
	switch {
	case strings.Contains(lower, "powershell") || lower == "pwsh" || lower == "powershell.exe":
		return "powershell", "PowerShell"
	case strings.Contains(lower, "bash"):
		return "bash", "Bash"
	case strings.Contains(lower, "zsh"):
		return "zsh", "Zsh"
	case strings.Contains(lower, "fish"):
		return "fish", "Fish"
	case strings.Contains(lower, "cmd"):
		return "cmd", "Command Prompt"
	default:
		return "terminal", "Terminal"
	}
}
