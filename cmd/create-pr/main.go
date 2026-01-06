package main

import (
	"fmt"
	"os"

	"cli-tools/internal/browser"
	"cli-tools/internal/git"
	"cli-tools/internal/github"
)

func main() {
	if !git.IsInsideRepo() {
		fmt.Fprintln(os.Stderr, "Error: not inside a git repository")
		os.Exit(1)
	}

	// Get current branch
	branch, err := git.GetCurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current branch: %v\n", err)
		os.Exit(1)
	}

	// Check if on default branch
	defaultBranch, _ := git.GetDefaultBranch()
	if branch == defaultBranch {
		fmt.Fprintf(os.Stderr, "Error: cannot create PR from %s branch\n", defaultBranch)
		os.Exit(1)
	}

	url, err := github.BuildCompareURL(branch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := browser.Open(url); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening browser: %v\n", err)
		os.Exit(1)
	}
}
