package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cli-tools/internal/auth"
	"cli-tools/internal/git"
	"cli-tools/internal/github"
)

type PRInfo struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	URL       string `json:"url"`
	Mergeable string `json:"mergeable"`
	Reviews   struct {
		TotalCount int `json:"totalCount"`
	} `json:"reviews"`
	ReviewDecision   string `json:"reviewDecision"`
	StatusCheckRollup struct {
		Contexts []struct {
			State      string `json:"state"`
			Conclusion string `json:"conclusion"`
			Name       string `json:"name"`
		} `json:"contexts"`
	} `json:"statusCheckRollup"`
}

func main() {
	if !git.IsInsideRepo() {
		fmt.Fprintln(os.Stderr, "Error: not inside a git repository")
		os.Exit(1)
	}

	if auth.HasGhCLI() {
		showPRStatusWithGh()
		return
	}

	// Fallback without gh CLI
	showPRStatusWithAPI()
}

func showPRStatusWithGh() {
	// Get detailed PR info
	cmd := exec.Command("gh", "pr", "view", "--json",
		"number,title,state,url,mergeable,reviews,reviewDecision,statusCheckRollup")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, strings.TrimSpace(string(exitErr.Stderr)))
		} else {
			fmt.Fprintln(os.Stderr, "No PR found for current branch")
		}
		os.Exit(1)
	}

	var pr PRInfo
	if err := json.Unmarshal(out, &pr); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing PR info: %v\n", err)
		os.Exit(1)
	}

	// Display PR status
	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Printf("State: %s\n", pr.State)
	fmt.Printf("URL: %s\n", pr.URL)
	fmt.Println()

	// Mergeable status
	fmt.Printf("Mergeable: %s\n", formatMergeable(pr.Mergeable))

	// Review status
	fmt.Printf("Reviews: %d (%s)\n", pr.Reviews.TotalCount, formatReviewDecision(pr.ReviewDecision))

	// CI status
	if len(pr.StatusCheckRollup.Contexts) > 0 {
		fmt.Println()
		fmt.Println("Checks:")
		for _, check := range pr.StatusCheckRollup.Contexts {
			status := check.Conclusion
			if status == "" {
				status = check.State
			}
			fmt.Printf("  %s %s\n", formatCheckStatus(status), check.Name)
		}
	}
}

func showPRStatusWithAPI() {
	// Get current branch
	branch, err := git.GetCurrentBranch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ownerRepo, err := github.GetOwnerRepo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Search for PR
	endpoint := fmt.Sprintf("/repos/%s/pulls?head=%s&state=open", ownerRepo, branch)
	data, err := auth.APIRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "")
		fmt.Print(auth.AuthSetupMessage())
		os.Exit(1)
	}

	var prs []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		URL    string `json:"html_url"`
	}
	if err := json.Unmarshal(data, &prs); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	if len(prs) == 0 {
		fmt.Println("No open PR found for current branch")
		os.Exit(0)
	}

	pr := prs[0]
	fmt.Printf("PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Printf("State: %s\n", pr.State)
	fmt.Printf("URL: %s\n", pr.URL)
	fmt.Println()
	fmt.Println("(Install gh CLI for more detailed status)")
}

func formatMergeable(s string) string {
	switch s {
	case "MERGEABLE":
		return "Yes"
	case "CONFLICTING":
		return "No (conflicts)"
	case "UNKNOWN":
		return "Checking..."
	default:
		return s
	}
}

func formatReviewDecision(s string) string {
	switch s {
	case "APPROVED":
		return "approved"
	case "CHANGES_REQUESTED":
		return "changes requested"
	case "REVIEW_REQUIRED":
		return "review required"
	default:
		return "pending"
	}
}

func formatCheckStatus(s string) string {
	switch strings.ToUpper(s) {
	case "SUCCESS":
		return "[OK]"
	case "FAILURE":
		return "[FAIL]"
	case "PENDING", "IN_PROGRESS", "QUEUED":
		return "[...]"
	case "SKIPPED":
		return "[SKIP]"
	default:
		return "[" + s + "]"
	}
}
