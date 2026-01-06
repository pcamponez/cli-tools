package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
		fmt.Fprintln(os.Stderr, "Usage: open-blame <file[:line]>")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  open-blame main.go")
		fmt.Fprintln(os.Stderr, "  open-blame main.go:42")
		os.Exit(1)
	}

	filePath, line := parseFileArg(os.Args[1])

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: file not found: %s\n", filePath)
		os.Exit(1)
	}

	url, err := github.BuildBlameURL(filePath, line, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := browser.Open(url); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening browser: %v\n", err)
		os.Exit(1)
	}
}

// parseFileArg parses "file:line" format, returns file path and line number
func parseFileArg(arg string) (string, int) {
	// Check for file:line format
	if idx := strings.LastIndex(arg, ":"); idx != -1 {
		lineStr := arg[idx+1:]
		if line, err := strconv.Atoi(lineStr); err == nil && line > 0 {
			return arg[:idx], line
		}
	}
	return arg, 0
}
