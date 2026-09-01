// Package engine implements the central coordination logic, privacy filters,
// state transitions, and rate-limiting queues for the Sol daemon.
package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sol-rpc/sol/internal/config"
)

// Regex patterns to identify and redact sensitive credentials
var (
	// Matches Authorization headers or tokens: Bearer <token>, token <token>
	authHeaderRegex = regexp.MustCompile(`(?i)(authorization\s*:\s*(bearer|basic|token)\s+)[^\s"']+`)

	// Matches command line flags: --password=<val>, --token=<val>, --api-key=<val>, --secret=<val>
	flagSecretRegex = regexp.MustCompile(`(?i)(--(?:password|passwd|token|api[-_]?key|secret|auth[-_]?token)(?:\s*=\s*|\s+))[^\s"']+`)

	// Matches short secret flags: -p <val>, -t <val>
	shortFlagSecretRegex = regexp.MustCompile(`(?i)(-(?:p|t)(?:\s*=\s*|\s+))[^\s"']+`)

	// Matches GitHub personal access tokens: ghp_..., gho_..., github_pat_...
	githubTokenRegex = regexp.MustCompile(`(?i)(gh[pousr]_[A-Za-z0-9_]{30,}|github_pat_[A-Za-z0-9_]{50,})`)

	// Matches OpenAI / generic sk- API keys
	skKeyRegex = regexp.MustCompile(`(?i)(sk-[A-Za-z0-9_-]{20,})`)

	// Matches Slack bot tokens: xoxb-..., xoxp-...
	slackTokenRegex = regexp.MustCompile(`(?i)(xox[baprs]-[A-Za-z0-9_-]{10,})`)

	// Matches GitLab personal access tokens: glpat-...
	gitlabTokenRegex = regexp.MustCompile(`(?i)(glpat-[A-Za-z0-9_-]{20,})`)
)

// SanitizeCommand strips passwords, access tokens, and secret flags from a command line string.
func SanitizeCommand(cmd string) string {
	cleaned := strings.TrimSpace(cmd)
	if cleaned == "" {
		return ""
	}

	// Redact known token signatures
	cleaned = githubTokenRegex.ReplaceAllString(cleaned, "[REDACTED_GH_TOKEN]")
	cleaned = skKeyRegex.ReplaceAllString(cleaned, "[REDACTED_KEY]")
	cleaned = slackTokenRegex.ReplaceAllString(cleaned, "[REDACTED_SLACK_TOKEN]")
	cleaned = gitlabTokenRegex.ReplaceAllString(cleaned, "[REDACTED_GITLAB_TOKEN]")

	// Redact Authorization headers
	cleaned = authHeaderRegex.ReplaceAllString(cleaned, "${1}[REDACTED_AUTH]")

	// Redact secret flags
	cleaned = flagSecretRegex.ReplaceAllString(cleaned, "${1}[REDACTED]")
	cleaned = shortFlagSecretRegex.ReplaceAllString(cleaned, "${1}[REDACTED]")

	// Discord Rich Presence details has a character limit (128 bytes max recommended)
	if len(cleaned) > 120 {
		cleaned = cleaned[:117] + "..."
	}

	return cleaned
}

// AnonymizePath strips drive letters and full internal machine paths to protect privacy.
// Modes:
// - "basename" (default): Returns only the current folder name (e.g., "Tool" or "~").
// - "parent_basename": Returns "parent/current" (e.g., "reconnaissance/Tool").
// - "relative": Returns path relative to home "~" or stripped of drive letters.
// - "hidden": Returns "Workspace".
func AnonymizePath(rawPath string, mode string) string {
	if rawPath == "" {
		return "Workspace"
	}

	if mode == config.PathModeHidden {
		return "Workspace"
	}

	clean := filepath.Clean(rawPath)
	cleanSlash := filepath.ToSlash(clean)

	// Check if this path matches user's home directory
	homeDir, err := os.UserHomeDir()
	isHome := false
	var homeRel string
	if err == nil && homeDir != "" {
		cleanHome := filepath.Clean(homeDir)
		if strings.EqualFold(clean, cleanHome) {
			isHome = true
			homeRel = ""
		} else if strings.HasPrefix(strings.ToLower(clean), strings.ToLower(cleanHome)+string(filepath.Separator)) {
			rel := clean[len(cleanHome):]
			rel = strings.TrimPrefix(rel, string(filepath.Separator))
			homeRel = filepath.ToSlash(rel)
		}
	}

	if isHome {
		return "~"
	}

	// 1. Mode: Basename (Default & Safest)
	if mode == "" || mode == config.PathModeBasename {
		if homeRel != "" {
			return filepath.Base(homeRel)
		}
		base := filepath.Base(clean)
		// If at volume root (e.g. "D:\" or "/")
		if base == "." || base == "/" || base == "\\" || strings.HasSuffix(clean, ":") || strings.HasSuffix(clean, ":\\") {
			return "Root"
		}
		return base
	}

	// 2. Mode: Parent + Basename (e.g. "reconnaissance/Tool")
	if mode == config.PathModeParentBasename {
		parts := strings.Split(strings.Trim(cleanSlash, "/"), "/")
		// Strip drive letter if present (e.g. "D:")
		if len(parts) > 0 && strings.HasSuffix(parts[0], ":") {
			parts = parts[1:]
		}
		if len(parts) >= 2 {
			return parts[len(parts)-2] + "/" + parts[len(parts)-1]
		} else if len(parts) == 1 && parts[0] != "" {
			return parts[0]
		}
		return "Root"
	}

	// 3. Mode: Relative
	if homeRel != "" {
		return "~/" + homeRel
	}

	// Strip drive letter (e.g. "D:/...")
	vol := filepath.VolumeName(clean)
	if vol != "" {
		cleanSlash = strings.TrimPrefix(cleanSlash, filepath.ToSlash(vol))
		cleanSlash = strings.TrimPrefix(cleanSlash, "/")
	}

	if cleanSlash == "" {
		return "Root"
	}

	// If the path is too long (> 35 chars), abbreviate middle components: .../parent/current
	parts := strings.Split(cleanSlash, "/")
	if len(parts) > 2 && len(cleanSlash) > 35 {
		return ".../" + parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	return cleanSlash
}
