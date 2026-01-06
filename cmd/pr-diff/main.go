package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"cli-tools/internal/auth"
	"cli-tools/internal/browser"
	"cli-tools/internal/git"
	"cli-tools/internal/github"
)

func main() {
	if !git.IsInsideRepo() {
		fmt.Fprintln(os.Stderr, "Error: not inside a git repository")
		os.Exit(1)
	}

	var prNum int
	var err error

	// Check if PR number was provided as argument
	if len(os.Args) >= 2 {
		prNum, err = strconv.Atoi(os.Args[1])
		if err != nil || prNum <= 0 {
			fmt.Fprintln(os.Stderr, "Error: invalid PR number")
			os.Exit(1)
		}
	} else {
		// Try to get PR for current branch
		prNum, err = getCurrentPRNumber()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if prNum == 0 {
			fmt.Fprintln(os.Stderr, "Error: no PR found for current branch")
			fmt.Fprintln(os.Stderr, "Usage: pr-diff [pr-number]")
			os.Exit(1)
		}
	}

	url, err := github.BuildURL(fmt.Sprintf("/pull/%d/files", prNum))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := browser.Open(url); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening browser: %v\n", err)
		os.Exit(1)
	}
}

func getCurrentPRNumber() (int, error) {
	if auth.HasGhCLI() {
		cmd := exec.Command("gh", "pr", "view", "--json", "number")
		out, err := cmd.Output()
		if err != nil {
			return 0, nil // No PR exists
		}

		var result struct {
			Number int `json:"number"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			return 0, fmt.Errorf("failed to parse PR info: %w", err)
		}
		return result.Number, nil
	}

	return 0, fmt.Errorf("gh CLI required to detect current PR. Install with: brew install gh")
}
