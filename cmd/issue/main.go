package main

import (
	"fmt"
	"os"
	"strconv"

	"cli-tools/internal/browser"
	"cli-tools/internal/git"
	"cli-tools/internal/github"
)

func main() {
	if !git.IsInsideRepo() {
		fmt.Fprintln(os.Stderr, "Error: not inside a git repository")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: issue <number>")
		os.Exit(1)
	}

	// Validate that the argument is a number
	issueNum, err := strconv.Atoi(os.Args[1])
	if err != nil || issueNum <= 0 {
		fmt.Fprintln(os.Stderr, "Error: issue number must be a positive integer")
		os.Exit(1)
	}

	url, err := github.BuildURL(fmt.Sprintf("/issues/%d", issueNum))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := browser.Open(url); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening browser: %v\n", err)
		os.Exit(1)
	}
}
