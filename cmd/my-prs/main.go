package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cli-tools/internal/auth"
)

type PR struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	State      string `json:"state"`
	URL        string `json:"url"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	CreatedAt string `json:"createdAt"`
}

func main() {
	if auth.HasGhCLI() {
		showPRsWithGh()
		return
	}

	showPRsWithAPI()
}

func showPRsWithGh() {
	// Use gh CLI to search for PRs authored by the current user
	cmd := exec.Command("gh", "pr", "list",
		"--author", "@me",
		"--state", "open",
		"--json", "number,title,state,url,repository,createdAt")
	out, err := cmd.Output()
	if err != nil {
		// Try without --author flag (some gh versions don't support it in list)
		cmd = exec.Command("gh", "search", "prs",
			"--author", "@me",
			"--state", "open",
			"--json", "number,title,state,url,repository")
		out, err = cmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(prs) == 0 {
		fmt.Println("No open PRs found")
		return
	}

	fmt.Printf("Your open PRs (%d):\n\n", len(prs))
	for _, pr := range prs {
		repo := pr.Repository.NameWithOwner
		if repo == "" {
			repo = "current repo"
		}
		fmt.Printf("#%d %s\n", pr.Number, pr.Title)
		fmt.Printf("    %s\n", repo)
		fmt.Printf("    %s\n\n", pr.URL)
	}
}

func showPRsWithAPI() {
	// Get current user
	userData, err := auth.APIRequest("GET", "/user", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Print(auth.AuthSetupMessage())
		os.Exit(1)
	}

	var user struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(userData, &user); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing user: %v\n", err)
		os.Exit(1)
	}

	// Search for open PRs by the user
	query := fmt.Sprintf("is:pr is:open author:%s", user.Login)
	endpoint := fmt.Sprintf("/search/issues?q=%s&sort=updated&per_page=20", strings.ReplaceAll(query, " ", "+"))

	data, err := auth.APIRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result struct {
		TotalCount int `json:"total_count"`
		Items      []struct {
			Number        int    `json:"number"`
			Title         string `json:"title"`
			HTMLURL       string `json:"html_url"`
			RepositoryURL string `json:"repository_url"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.TotalCount == 0 {
		fmt.Println("No open PRs found")
		return
	}

	fmt.Printf("Your open PRs (%d):\n\n", result.TotalCount)
	for _, pr := range result.Items {
		// Extract repo name from URL
		repo := extractRepoFromURL(pr.RepositoryURL)
		fmt.Printf("#%d %s\n", pr.Number, pr.Title)
		fmt.Printf("    %s\n", repo)
		fmt.Printf("    %s\n\n", pr.HTMLURL)
	}
}

func extractRepoFromURL(url string) string {
	// https://api.github.com/repos/owner/repo -> owner/repo
	parts := strings.Split(url, "/repos/")
	if len(parts) > 1 {
		return parts[1]
	}
	return url
}
