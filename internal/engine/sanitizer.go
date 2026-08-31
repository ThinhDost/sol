// Package engine implements the central coordination logic, privacy filters,
// state transitions, and rate-limiting queues for the Sol daemon.
package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// AnonymizePath replaces the user's home directory path with '~' and normalizes directory separators.
func AnonymizePath(rawPath string) string {
	if rawPath == "" {
		return ""
	}

	clean := filepath.Clean(rawPath)
	homeDir, err := os.UserHomeDir()
	if err == nil && homeDir != "" {
		cleanHome := filepath.Clean(homeDir)
		if strings.HasPrefix(strings.ToLower(clean), strings.ToLower(cleanHome)) {
			rel := clean[len(cleanHome):]
			rel = strings.TrimPrefix(rel, string(filepath.Separator))
			if rel == "" {
				return "~"
			}
			return "~/" + filepath.ToSlash(rel)
		}
	}

	return filepath.ToSlash(clean)
}
