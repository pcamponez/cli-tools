package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"cli-tools/internal/github"
)

// HasGhCLI checks if the gh CLI is installed and authenticated
func HasGhCLI() bool {
	cmd := exec.Command("gh", "auth", "status")
	err := cmd.Run()
	return err == nil
}

// GetToken returns the appropriate GitHub token for the current repository
// It first checks for a host-specific token (e.g., GITHUB_TOKEN_RHEI),
// then falls back to GITHUB_TOKEN
func GetToken() (string, error) {
	// Try to get host-specific token first
	hostAlias, err := github.GetSSHHostAlias()
	if err == nil && hostAlias != "" {
		// Convert host alias to env var name
		// e.g., "github-rhei" -> "GITHUB_TOKEN_RHEI"
		envName := hostAliasToEnvVar(hostAlias)
		if token := os.Getenv(envName); token != "" {
			return token, nil
		}
	}

	// Fall back to default GITHUB_TOKEN
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no GitHub token found. Set GITHUB_TOKEN or use 'gh auth login'")
}

// hostAliasToEnvVar converts an SSH host alias to an environment variable name
// e.g., "github-rhei" -> "GITHUB_TOKEN_RHEI"
func hostAliasToEnvVar(hostAlias string) string {
	// Remove common prefixes
	name := strings.TrimPrefix(hostAlias, "github-")
	name = strings.TrimPrefix(name, "github.")

	// Convert to uppercase and replace special chars
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")

	return "GITHUB_TOKEN_" + name
}

// APIRequest makes an authenticated request to the GitHub API
// It prefers using gh CLI if available, otherwise uses token auth
func APIRequest(method, endpoint string, body interface{}) ([]byte, error) {
	if HasGhCLI() {
		return ghAPIRequest(method, endpoint, body)
	}
	return tokenAPIRequest(method, endpoint, body)
}

// ghAPIRequest uses the gh CLI to make API requests
func ghAPIRequest(method, endpoint string, body interface{}) ([]byte, error) {
	args := []string{"api", "-X", method, endpoint}

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		args = append(args, "-f", string(jsonBody))
	}

	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh api error: %s", string(exitErr.Stderr))
		}
		return nil, err
	}
	return out, nil
}

// tokenAPIRequest makes a direct HTTP request using token auth
func tokenAPIRequest(method, endpoint string, body interface{}) ([]byte, error) {
	token, err := GetToken()
	if err != nil {
		return nil, err
	}

	url := endpoint
	if !strings.HasPrefix(endpoint, "https://") {
		url = "https://api.github.com" + endpoint
	}

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetCurrentPR returns the PR number for the current branch, or 0 if none exists
func GetCurrentPR() (int, error) {
	if HasGhCLI() {
		cmd := exec.Command("gh", "pr", "view", "--json", "number", "-q", ".number")
		out, err := cmd.Output()
		if err != nil {
			return 0, nil // No PR for this branch
		}
		var num int
		fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &num)
		return num, nil
	}

	// Fallback: use API directly
	// This is more complex, so for now return 0 if gh is not available
	return 0, fmt.Errorf("gh CLI required for PR detection without manual PR number")
}

// RunGhCommand runs a gh CLI command and returns the output
func RunGhCommand(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("%s", string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// AuthSetupMessage returns a helpful message for users who need to set up auth
func AuthSetupMessage() string {
	return `Authentication required. Choose one option:

Option A: Use gh CLI (recommended)
  brew install gh
  gh auth login

Option B: Set a personal access token
  export GITHUB_TOKEN="ghp_xxxx"

For multiple GitHub accounts with SSH aliases:
  export GITHUB_TOKEN_RHEI="ghp_work_token"  # for git@github-rhei
`
}
