package github

import (
	"fmt"
	"regexp"
	"strings"

	"cli-tools/internal/git"
)

// RepoInfo contains parsed GitHub repository information
type RepoInfo struct {
	Owner    string
	Repo     string
	Host     string // SSH host alias (e.g., "github.com", "github-rhei")
	BaseURL  string // Full HTTPS URL to the repo
}

// GetRepoInfo parses the git remote and returns GitHub repo information
func GetRepoInfo() (*RepoInfo, error) {
	remoteURL, err := git.GetRemoteURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	return ParseRemoteURL(remoteURL)
}

// ParseRemoteURL converts a git remote URL to RepoInfo
// Supports:
//   - git@github.com:owner/repo.git
//   - git@github-rhei:owner/repo.git (custom SSH alias)
//   - https://github.com/owner/repo.git
//   - https://github.com/owner/repo
func ParseRemoteURL(remoteURL string) (*RepoInfo, error) {
	info := &RepoInfo{}

	// SSH format: git@host:owner/repo.git
	sshRegex := regexp.MustCompile(`^git@([^:]+):([^/]+)/(.+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(remoteURL); matches != nil {
		info.Host = matches[1]
		info.Owner = matches[2]
		info.Repo = matches[3]
		info.BaseURL = fmt.Sprintf("https://github.com/%s/%s", info.Owner, info.Repo)
		return info, nil
	}

	// HTTPS format: https://github.com/owner/repo.git
	httpsRegex := regexp.MustCompile(`^https?://([^/]+)/([^/]+)/(.+?)(?:\.git)?$`)
	if matches := httpsRegex.FindStringSubmatch(remoteURL); matches != nil {
		info.Host = matches[1]
		info.Owner = matches[2]
		info.Repo = matches[3]
		info.BaseURL = fmt.Sprintf("https://%s/%s/%s", info.Host, info.Owner, info.Repo)
		return info, nil
	}

	return nil, fmt.Errorf("unable to parse remote URL: %s", remoteURL)
}

// GetRepoURL returns the HTTPS URL for the current repository
func GetRepoURL() (string, error) {
	info, err := GetRepoInfo()
	if err != nil {
		return "", err
	}
	return info.BaseURL, nil
}

// GetOwnerRepo returns "owner/repo" for the current repository
func GetOwnerRepo() (string, error) {
	info, err := GetRepoInfo()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", info.Owner, info.Repo), nil
}

// BuildURL constructs a GitHub URL by appending a path to the repo URL
func BuildURL(path string) (string, error) {
	baseURL, err := GetRepoURL()
	if err != nil {
		return "", err
	}
	return baseURL + path, nil
}

// BuildFileURL constructs a URL to view a file on GitHub
func BuildFileURL(filePath string, line int, branch string) (string, error) {
	baseURL, err := GetRepoURL()
	if err != nil {
		return "", err
	}

	relPath, err := git.GetRelativePath(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	// Use current branch if not specified
	if branch == "" {
		branch, err = git.GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	url := fmt.Sprintf("%s/blob/%s/%s", baseURL, branch, relPath)
	if line > 0 {
		url += fmt.Sprintf("#L%d", line)
	}
	return url, nil
}

// BuildBlameURL constructs a URL to view blame for a file on GitHub
func BuildBlameURL(filePath string, line int, branch string) (string, error) {
	baseURL, err := GetRepoURL()
	if err != nil {
		return "", err
	}

	relPath, err := git.GetRelativePath(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	// Use current branch if not specified
	if branch == "" {
		branch, err = git.GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	url := fmt.Sprintf("%s/blame/%s/%s", baseURL, branch, relPath)
	if line > 0 {
		url += fmt.Sprintf("#L%d", line)
	}
	return url, nil
}

// BuildCompareURL constructs a URL to create a PR (compare view)
func BuildCompareURL(branch string) (string, error) {
	baseURL, err := GetRepoURL()
	if err != nil {
		return "", err
	}

	if branch == "" {
		var err error
		branch, err = git.GetCurrentBranch()
		if err != nil {
			return "", fmt.Errorf("failed to get current branch: %w", err)
		}
	}

	return fmt.Sprintf("%s/compare/%s?expand=1", baseURL, branch), nil
}

// GetSSHHostAlias extracts the SSH host alias from the remote URL
// Returns empty string for HTTPS URLs
func GetSSHHostAlias() (string, error) {
	info, err := GetRepoInfo()
	if err != nil {
		return "", err
	}

	// Only return non-standard hosts (not github.com)
	if info.Host != "github.com" && !strings.Contains(info.Host, "github.com") {
		return info.Host, nil
	}
	return "", nil
}
