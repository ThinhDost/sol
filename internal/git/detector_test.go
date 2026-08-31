package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGitContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sol_git_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	headFile := filepath.Join(gitDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/feature/rpc-daemon\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD file: %v", err)
	}

	// Create nested subfolder
	subDir := filepath.Join(tempDir, "src", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	// Test detection from nested subdirectory
	info, err := DetectGitContext(subDir)
	if err != nil {
		t.Fatalf("DetectGitContext failed: %v", err)
	}

	if info.Branch != "feature/rpc-daemon" {
		t.Errorf("expected branch 'feature/rpc-daemon', got '%s'", info.Branch)
	}

	if info.RootDir != tempDir {
		t.Errorf("expected root dir '%s', got '%s'", tempDir, info.RootDir)
	}
}

func TestDetectDetachedHead(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sol_git_detached_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gitDir := filepath.Join(tempDir, ".git")
	_ = os.MkdirAll(gitDir, 0755)
	headFile := filepath.Join(gitDir, "HEAD")
	_ = os.WriteFile(headFile, []byte("a1b2c3d4e5f678901234567890\n"), 0644)

	info, err := DetectGitContext(tempDir)
	if err != nil {
		t.Fatalf("DetectGitContext failed: %v", err)
	}

	if info.Branch != "a1b2c3d" {
		t.Errorf("expected detached short hash 'a1b2c3d', got '%s'", info.Branch)
	}
}
