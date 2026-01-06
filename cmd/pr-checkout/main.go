package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"cli-tools/internal/auth"
	"cli-tools/internal/git"
)

func main() {
	if !git.IsInsideRepo() {
		fmt.Fprintln(os.Stderr, "Error: not inside a git repository")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: pr-checkout <pr-number>")
		os.Exit(1)
	}

	prNum, err := strconv.Atoi(os.Args[1])
	if err != nil || prNum <= 0 {
		fmt.Fprintln(os.Stderr, "Error: PR number must be a positive integer")
		os.Exit(1)
	}

	if !auth.HasGhCLI() {
		fmt.Fprintln(os.Stderr, "Error: gh CLI required for this command")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Install gh CLI:")
		fmt.Fprintln(os.Stderr, "  brew install gh")
		fmt.Fprintln(os.Stderr, "  gh auth login")
		os.Exit(1)
	}

	cmd := exec.Command("gh", "pr", "checkout", os.Args[1])
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
