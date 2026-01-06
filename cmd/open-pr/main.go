package main

import (
	"fmt"
	"os"
	"os/exec"

	"cli-tools/internal/auth"
	"cli-tools/internal/git"
)

func main() {
	if !git.IsInsideRepo() {
		fmt.Fprintln(os.Stderr, "Error: not inside a git repository")
		os.Exit(1)
	}

	// Use gh CLI if available (handles the web opening itself)
	if auth.HasGhCLI() {
		cmd := exec.Command("gh", "pr", "view", "--web")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// gh will print its own error message (e.g., "no pull request found")
			os.Exit(1)
		}
		return
	}

	// Without gh CLI, we need to find the PR number via API
	fmt.Fprintln(os.Stderr, "Error: gh CLI required for this command")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Install gh CLI:")
	fmt.Fprintln(os.Stderr, "  brew install gh")
	fmt.Fprintln(os.Stderr, "  gh auth login")
	os.Exit(1)
}
