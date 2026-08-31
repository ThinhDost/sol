// Package git provides zero-process Git repository and branch detection
// by directly reading filesystem metadata without spawning `git.exe` processes.
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GitInfo contains extracted Git context information.
type GitInfo struct {
	// RepoName is the base name of the repository root folder.
	RepoName string
	// Branch is the current branch name, tag, or short commit hash.
	Branch string
	// RootDir is the absolute path to the repository root.
	RootDir string
}

// DetectGitContext searches upwards from `startDir` to locate the git repository root
// and parses the current branch name from `.git/HEAD`.
func DetectGitContext(startDir string) (*GitInfo, error) {
	if startDir == "" {
		return nil, fmt.Errorf("empty directory")
	}

	cleanDir := filepath.Clean(startDir)
	curr := cleanDir

	for {
		gitPath := filepath.Join(curr, ".git")
		fi, err := os.Stat(gitPath)
		if err == nil {
			// Found git root
			var headFilePath string
			if fi.IsDir() {
				// Standard Git repository
				headFilePath = filepath.Join(gitPath, "HEAD")
			} else {
				// Worktree or submodule where .git is a file containing "gitdir: <path>"
				headFilePath, err = resolveGitdirFile(gitPath)
				if err != nil {
					return nil, err
				}
			}

			branch, err := parseHeadFile(headFilePath)
			if err != nil {
				return nil, err
			}

			return &GitInfo{
				RepoName: filepath.Base(curr),
				Branch:   branch,
				RootDir:  curr,
			}, nil
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			// Reached filesystem root without finding .git
			break
		}
		curr = parent
	}

	return nil, fmt.Errorf("not a git repository")
}

// resolveGitdirFile reads a `.git` file (used in worktrees/submodules) and resolves the HEAD path.
func resolveGitdirFile(gitFilePath string) (string, error) {
	content, err := os.ReadFile(gitFilePath)
	if err != nil {
		return "", err
	}

	str := strings.TrimSpace(string(content))
	if strings.HasPrefix(str, "gitdir:") {
		dir := strings.TrimSpace(strings.TrimPrefix(str, "gitdir:"))
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(filepath.Dir(gitFilePath), dir)
		}
		return filepath.Join(dir, "HEAD"), nil
	}

	return "", fmt.Errorf("invalid .git file format")
}

// parseHeadFile reads the HEAD file and returns the clean branch name or short commit hash.
func parseHeadFile(headPath string) (string, error) {
	data, err := os.ReadFile(headPath)
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/"), nil
	}
	if strings.HasPrefix(content, "ref: refs/tags/") {
		return strings.TrimPrefix(content, "ref: refs/tags/"), nil
	}

	// Detached HEAD commit hash (truncate to 7 chars)
	if len(content) >= 7 {
		return content[:7], nil
	}

	return content, nil
}
