// Package config provides configuration settings, default values,
// and icon mapping dictionaries for the Sol Discord Rich Presence system.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultDiscordClientID is the Discord Application ID for Sol.
const DefaultDiscordClientID = "1543849393432698940"

// DefaultGitHubURL is the default repository link displayed on Discord Rich Presence.
const DefaultGitHubURL = "https://github.com/ThinhDost/sol"

// Path display mode constants for directory privacy
const (
	PathModeBasename       = "basename"
	PathModeParentBasename = "parent_basename"
	PathModeRelative       = "relative"
	PathModeHidden         = "hidden"
)

// Config represents the runtime configuration for the Sol daemon.
type Config struct {
	// DiscordClientID is the Discord Application ID registered on Discord Developer Portal.
	DiscordClientID string `json:"client_id"`

	// LargeImageKey is the persistent main presence icon (default: "fubuki").
	LargeImageKey string `json:"large_image_key"`

	// LargeImageText is the hover tooltip for the main presence icon.
	LargeImageText string `json:"large_image_text"`

	// PathDisplayMode controls how directories are sanitized ("basename", "parent_basename", "relative", "hidden").
	PathDisplayMode string `json:"path_display_mode,omitempty"`

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

	largeImage := os.Getenv("SOL_LARGE_IMAGE_KEY")
	if largeImage == "" {
		largeImage = "fubuki"
	}

	largeText := os.Getenv("SOL_LARGE_IMAGE_TEXT")
	if largeText == "" {
		largeText = "Sol Terminal"
	}

	pathMode := os.Getenv("SOL_PATH_MODE")
	if pathMode == "" {
		pathMode = PathModeBasename
	}

	githubURL := os.Getenv("SOL_GITHUB_URL")
	if githubURL == "" {
		githubURL = DefaultGitHubURL
	}

	return &Config{
		DiscordClientID:   clientID,
		LargeImageKey:     largeImage,
		LargeImageText:    largeText,
		PathDisplayMode:   pathMode,
		GitHubURL:         githubURL,
		PipeNameWindows:   `\\.\pipe\sol-ipc`,
		SocketPathUnix:    "/tmp/sol.sock",
		IdleTimeout:       10 * time.Minute,
		RateLimitInterval: 4 * time.Second,
		MaskSecrets:       true,
		AnonymizeHomePath: true,
	}
}

// LoadConfig loads configuration from files (./sol.config.json or ~/.sol/sol.config.json),
// falling back to environment variables and default values.
func LoadConfig() *Config {
	cfg := DefaultConfig()

	homeDir, _ := os.UserHomeDir()
	candidatePaths := []string{
		"sol.config.json",
		filepath.Join(homeDir, ".sol", "sol.config.json"),
	}

	for _, path := range candidatePaths {
		if data, err := os.ReadFile(path); err == nil {
			var fileCfg struct {
				ClientID          string `json:"client_id"`
				LargeImageKey     string `json:"large_image_key"`
				LargeImageText    string `json:"large_image_text"`
				PathDisplayMode   string `json:"path_display_mode"`
				GitHubURL         string `json:"github_url"`
				IdleMinutes       int    `json:"idle_timeout_minutes"`
				RateLimitSeconds  int    `json:"rate_limit_seconds"`
				MaskSecrets       *bool  `json:"mask_secrets"`
				AnonymizeHomePath *bool  `json:"anonymize_home_path"`
			}
			if err := json.Unmarshal(data, &fileCfg); err == nil {
				if fileCfg.ClientID != "" {
					cfg.DiscordClientID = fileCfg.ClientID
				}
				if fileCfg.LargeImageKey != "" {
					cfg.LargeImageKey = fileCfg.LargeImageKey
				}
				if fileCfg.LargeImageText != "" {
					cfg.LargeImageText = fileCfg.LargeImageText
				}
				if fileCfg.PathDisplayMode != "" {
					cfg.PathDisplayMode = fileCfg.PathDisplayMode
				}
				if fileCfg.GitHubURL != "" {
					cfg.GitHubURL = fileCfg.GitHubURL
				}
				if fileCfg.IdleMinutes > 0 {
					cfg.IdleTimeout = time.Duration(fileCfg.IdleMinutes) * time.Minute
				}
				if fileCfg.RateLimitSeconds > 0 {
					cfg.RateLimitInterval = time.Duration(fileCfg.RateLimitSeconds) * time.Second
				}
				if fileCfg.MaskSecrets != nil {
					cfg.MaskSecrets = *fileCfg.MaskSecrets
				}
				if fileCfg.AnonymizeHomePath != nil {
					cfg.AnonymizeHomePath = *fileCfg.AnonymizeHomePath
				}
				break
			}
		}
	}

	return cfg
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
	// Compound / Multi-word Commands (Checked first)
	"docker compose":  {AssetKey: "docker", DisplayName: "Docker Compose"},
	"docker-compose":  {AssetKey: "docker", DisplayName: "Docker Compose"},
	"docker build":    {AssetKey: "docker", DisplayName: "Docker Build"},
	"docker run":      {AssetKey: "docker", DisplayName: "Docker Run"},
	"docker ps":       {AssetKey: "docker", DisplayName: "Docker Containers"},
	"docker exec":     {AssetKey: "docker", DisplayName: "Docker Exec"},
	"cargo build":     {AssetKey: "rust", DisplayName: "Cargo Build (Rust)"},
	"cargo test":      {AssetKey: "rust", DisplayName: "Cargo Test (Rust)"},
	"cargo run":       {AssetKey: "rust", DisplayName: "Cargo Run (Rust)"},
	"cargo check":     {AssetKey: "rust", DisplayName: "Cargo Check (Rust)"},
	"go build":        {AssetKey: "golang", DisplayName: "Go Build"},
	"go test":         {AssetKey: "golang", DisplayName: "Go Test"},
	"go run":          {AssetKey: "golang", DisplayName: "Go Run"},
	"go mod":          {AssetKey: "golang", DisplayName: "Go Modules"},
	"git commit":      {AssetKey: "git", DisplayName: "Git Commit"},
	"git push":        {AssetKey: "git", DisplayName: "Git Push"},
	"git pull":        {AssetKey: "git", DisplayName: "Git Pull"},
	"git status":      {AssetKey: "git", DisplayName: "Git Status"},
	"git diff":        {AssetKey: "git", DisplayName: "Git Diff"},
	"npm run":         {AssetKey: "npm", DisplayName: "NPM Script"},
	"npm test":        {AssetKey: "npm", DisplayName: "NPM Test"},
	"npm install":     {AssetKey: "npm", DisplayName: "NPM Install"},
	"pnpm run":        {AssetKey: "pnpm", DisplayName: "PNPM Script"},
	"yarn run":        {AssetKey: "yarn", DisplayName: "Yarn Script"},
	"bun run":         {AssetKey: "bun", DisplayName: "Bun Run"},

	// Programming Languages & Toolchains
	"go":      {AssetKey: "golang", DisplayName: "Go Toolchain"},
	"cargo":   {AssetKey: "rust", DisplayName: "Cargo (Rust)"},
	"rustc":   {AssetKey: "rust", DisplayName: "Rust Compiler"},
	"python":  {AssetKey: "python", DisplayName: "Python"},
	"python3": {AssetKey: "python", DisplayName: "Python 3"},
	"py":      {AssetKey: "python", DisplayName: "Python"},
	"pytest":  {AssetKey: "python", DisplayName: "Pytest"},
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
	"docker":     {AssetKey: "docker", DisplayName: "Docker Engine"},
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

	// AI & Coding Agents
	"agy":             {AssetKey: "antigravity", DisplayName: "Antigravity AI Agent"},
	"antigravity":     {AssetKey: "antigravity", DisplayName: "Antigravity AI Agent"},
	"agy-cli":         {AssetKey: "antigravity", DisplayName: "Antigravity AI CLI"},
	"antigravity-cli": {AssetKey: "antigravity", DisplayName: "Antigravity AI CLI"},
	"gemini":          {AssetKey: "antigravity", DisplayName: "Gemini CLI"},
	"claude":          {AssetKey: "ai", DisplayName: "Claude Code"},
	"copilot":         {AssetKey: "ai", DisplayName: "GitHub Copilot"},
	"ollama":          {AssetKey: "ai", DisplayName: "Ollama AI"},
}

// ResolveToolAsset inspects a raw command string and returns its corresponding asset info.
// If the command is unrecognized, a default terminal asset is returned.
func ResolveToolAsset(rawCmd string) ToolAssetInfo {
	trimmed := strings.TrimSpace(rawCmd)
	if trimmed == "" {
		return ToolAssetInfo{AssetKey: "terminal", DisplayName: "Terminal"}
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return ToolAssetInfo{AssetKey: "terminal", DisplayName: "Terminal"}
	}

	// 1. Check compound 2-word command (e.g. "docker compose", "cargo build")
	if len(fields) >= 2 {
		compound := strings.ToLower(fields[0] + " " + fields[1])
		compound = strings.TrimSuffix(compound, ".exe")
		if info, found := toolAssetMap[compound]; found {
			return info
		}
	}

	// 2. Check base 1-word command (e.g. "docker", "git")
	baseCmd := strings.ToLower(fields[0])
	baseCmd = strings.TrimSuffix(baseCmd, ".exe")
	baseCmd = strings.TrimSuffix(baseCmd, ".cmd")
	baseCmd = strings.TrimSuffix(baseCmd, ".bat")

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
