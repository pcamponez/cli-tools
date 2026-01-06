package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cli-tools/internal/auth"
)

type Issue struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	State      string `json:"state"`
	URL        string `json:"url"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

func main() {
	if auth.HasGhCLI() {
		showIssuesWithGh()
		return
	}

	showIssuesWithAPI()
}

func showIssuesWithGh() {
	// Search for issues assigned to the current user
	cmd := exec.Command("gh", "search", "issues",
		"--assignee", "@me",
		"--state", "open",
		"--json", "number,title,state,url,repository")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, strings.TrimSpace(string(exitErr.Stderr)))
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	var issues []Issue
	if err := json.Unmarshal(out, &issues); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(issues) == 0 {
		fmt.Println("No open issues assigned to you")
		return
	}

	fmt.Printf("Issues assigned to you (%d):\n\n", len(issues))
	for _, issue := range issues {
		fmt.Printf("#%d %s\n", issue.Number, issue.Title)
		fmt.Printf("    %s\n", issue.Repository.NameWithOwner)
		fmt.Printf("    %s\n\n", issue.URL)
	}
}

func showIssuesWithAPI() {
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

	// Search for issues assigned to the user
	query := fmt.Sprintf("is:issue is:open assignee:%s", user.Login)
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
		fmt.Println("No open issues assigned to you")
		return
	}

	fmt.Printf("Issues assigned to you (%d):\n\n", result.TotalCount)
	for _, issue := range result.Items {
		repo := extractRepoFromURL(issue.RepositoryURL)
		fmt.Printf("#%d %s\n", issue.Number, issue.Title)
		fmt.Printf("    %s\n", repo)
		fmt.Printf("    %s\n\n", issue.HTMLURL)
	}
}

func extractRepoFromURL(url string) string {
	parts := strings.Split(url, "/repos/")
	if len(parts) > 1 {
		return parts[1]
	}
	return url
}
