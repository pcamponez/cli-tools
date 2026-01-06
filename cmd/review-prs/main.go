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
	URL        string `json:"url"`
	Author     struct {
		Login string `json:"login"`
	} `json:"author"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
}

func main() {
	if auth.HasGhCLI() {
		showReviewPRsWithGh()
		return
	}

	showReviewPRsWithAPI()
}

func showReviewPRsWithGh() {
	// Search for PRs where review is requested
	cmd := exec.Command("gh", "search", "prs",
		"--review-requested", "@me",
		"--state", "open",
		"--json", "number,title,url,author,repository")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, strings.TrimSpace(string(exitErr.Stderr)))
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(prs) == 0 {
		fmt.Println("No PRs awaiting your review")
		return
	}

	fmt.Printf("PRs awaiting your review (%d):\n\n", len(prs))
	for _, pr := range prs {
		fmt.Printf("#%d %s\n", pr.Number, pr.Title)
		fmt.Printf("    by @%s in %s\n", pr.Author.Login, pr.Repository.NameWithOwner)
		fmt.Printf("    %s\n\n", pr.URL)
	}
}

func showReviewPRsWithAPI() {
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

	// Search for PRs where review is requested
	query := fmt.Sprintf("is:pr is:open review-requested:%s", user.Login)
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
			User          struct {
				Login string `json:"login"`
			} `json:"user"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if result.TotalCount == 0 {
		fmt.Println("No PRs awaiting your review")
		return
	}

	fmt.Printf("PRs awaiting your review (%d):\n\n", result.TotalCount)
	for _, pr := range result.Items {
		repo := extractRepoFromURL(pr.RepositoryURL)
		fmt.Printf("#%d %s\n", pr.Number, pr.Title)
		fmt.Printf("    by @%s in %s\n", pr.User.Login, repo)
		fmt.Printf("    %s\n\n", pr.HTMLURL)
	}
}

func extractRepoFromURL(url string) string {
	parts := strings.Split(url, "/repos/")
	if len(parts) > 1 {
		return parts[1]
	}
	return url
}
