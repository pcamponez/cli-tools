# CLI Tools

A collection of command-line tools to speed up your GitHub workflow. Run commands from any directory inside a git repository to quickly open GitHub pages, manage PRs, and more.

## Installation

### Prerequisites

- **Go 1.21+** - Install via `brew install go`
- **gh CLI** (optional but recommended) - Install via `brew install gh && gh auth login`

### Quick Install

```bash
git clone https://github.com/pedrocamponez/cli-tools
cd cli-tools
make install
```

This installs all commands to `~/go/bin/`. Make sure it's in your PATH:

```bash
# Add to your ~/.zshrc or ~/.bashrc
export PATH="$HOME/go/bin:$PATH"
```

## Commands

### Repository Navigation

| Command | Description |
|---------|-------------|
| `open-repo` | Open the current repository in your browser |
| `open-issues` | Open the issues page |
| `open-actions` | Open the GitHub Actions page |
| `open-file <file[:line]>` | Open a file in GitHub (optionally at a specific line) |
| `open-blame <file[:line]>` | Open the blame view for a file |

**Examples:**

```bash
open-repo                    # Opens github.com/owner/repo
open-file src/main.go        # Opens the file in GitHub
open-file src/main.go:42     # Opens at line 42
open-blame config.yaml:15    # Who changed line 15?
```

### Pull Request Workflow

| Command | Description |
|---------|-------------|
| `create-pr` | Open the PR creation page for your current branch |
| `open-pr` | Open the PR for your current branch |
| `pr-status` | Show the status of your current branch's PR |
| `pr-diff [number]` | Open the PR diff view in browser |
| `pr-checkout <number>` | Checkout a PR locally |
| `my-prs` | List all your open PRs |
| `review-prs` | List PRs awaiting your review |

**Examples:**

```bash
create-pr                    # Opens PR creation with your branch
open-pr                      # Opens your branch's PR
pr-status                    # Shows PR status, checks, reviews
pr-diff                      # Opens diff for current PR
pr-diff 123                  # Opens diff for PR #123
pr-checkout 456              # Checkout PR #456 locally
my-prs                       # What PRs do I have open?
review-prs                   # What PRs need my review?
```

### Issues

| Command | Description |
|---------|-------------|
| `new-issue` | Open the new issue page |
| `issue <number>` | Open a specific issue |
| `my-issues` | List issues assigned to you |

**Examples:**

```bash
new-issue                    # Create a new issue
issue 42                     # Open issue #42
my-issues                    # What's on my plate?
```

## Authentication

Some commands require GitHub authentication (anything that queries the API).

### Option A: Use gh CLI (Recommended)

The easiest way - `gh` handles authentication automatically:

```bash
brew install gh
gh auth login
```

### Option B: Personal Access Token

Set a token as an environment variable:

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
```

Add this to your `~/.zshrc` to make it permanent.

### Multiple GitHub Accounts

If you use SSH host aliases for different GitHub accounts (e.g., work vs personal), the tools automatically detect which token to use based on your repo's remote URL.

**Setup:**

1. Configure SSH aliases in `~/.ssh/config`:
   ```
   # Personal account
   Host github.com
     HostName github.com
     User git
     IdentityFile ~/.ssh/id_personal

   # Work account
   Host github-work
     HostName github.com
     User git
     IdentityFile ~/.ssh/id_work
   ```

2. Set tokens for each alias:
   ```bash
   # Default token (for github.com)
   export GITHUB_TOKEN="ghp_personal_token"

   # Work token (for github-work alias)
   export GITHUB_TOKEN_WORK="ghp_work_token"
   ```

The tools will automatically use `GITHUB_TOKEN_WORK` when you're in a repo cloned via `git@github-work:org/repo.git`.

**Token naming convention:**
- `git@github-rhei:...` → `GITHUB_TOKEN_RHEI`
- `git@github-work:...` → `GITHUB_TOKEN_WORK`
- `git@github.mycompany:...` → `GITHUB_TOKEN_MYCOMPANY`

## Building from Source

```bash
# Build all commands to ./bin/
make build

# Build a specific command
make open-repo

# Install to ~/go/bin/
make install

# Clean up
make clean

# Uninstall
make uninstall
```

## Contributing

Feel free to open issues or submit PRs! Some ideas for new commands:

- `copy-link` - Copy GitHub permalink to clipboard
- `repo-info` - Show repo stats (stars, forks, etc.)
- `run-workflow` - Trigger a GitHub Action

## License

MIT
