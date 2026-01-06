package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRemoteURL returns the URL of the origin remote
func GetRemoteURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCurrentBranch returns the name of the current branch
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRelativePath returns the path of a file relative to the repo root
func GetRelativePath(filePath string) (string, error) {
	root, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// GetDefaultBranch returns the default branch (main or master)
func GetDefaultBranch() (string, error) {
	// Try to get the default branch from remote HEAD
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	out, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		// Remove "origin/" prefix
		return strings.TrimPrefix(branch, "origin/"), nil
	}

	// Fallback: check if main or master exists
	for _, branch := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
		if err := cmd.Run(); err == nil {
			return branch, nil
		}
	}

	return "main", nil // Default fallback
}

// IsInsideRepo checks if the current directory is inside a git repository
func IsInsideRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}
