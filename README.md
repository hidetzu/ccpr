# ccpr

![CI](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml/badge.svg)

Review AWS CodeCommit pull requests from the CLI — optimized for humans and AI.

## Output Examples

### Summary (default)

```
$ ccpr review <codecommit-pr-url>
PR #883: Fix login bug
Author:   example-user
Status:   OPEN
Branch:   feature/login → main
Created:  2026-03-25

Comments: 3 (2 threads)
Files:    7 changed

## Description

Fix session timeout on login page
```

### JSON (for AI tools)

```
$ ccpr review <codecommit-pr-url> --format json
{
  "metadata": {
    "prId": "883",
    "title": "Fix login bug",
    "author": "example-user",
    "authorArn": "arn:aws:iam::123456789012:user/example-user",
    "sourceBranch": "feature/login",
    "destinationBranch": "main",
    "status": "OPEN",
    "creationDate": "2026-03-25T10:30:00Z"
  },
  "comments": [...],
  "diff": "diff --git a/src/login.go ..."
}
```

## Overview

ccpr is a CLI tool that bridges AWS CodeCommit and AI review tools like Claude Code.

It takes a CodeCommit PR URL, fetches metadata, comments, and diffs, and outputs structured data that AI tools can use for code review.

## Install

```bash
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
```

## Setup

Create `~/.config/ccpr/config.yaml`:

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  your-repo: /path/to/local/clone
```

## Usage

### Review a PR

```bash
ccpr review <codecommit-pr-url>                 # Summary (default)
ccpr review <codecommit-pr-url> --format json   # JSON for AI tools
ccpr review <codecommit-pr-url> --format patch  # Diff only
```

### List PRs

```bash
ccpr list --repo <repo>                         # OPEN PRs (default)
ccpr list --repo <repo> --status closed         # CLOSED PRs
ccpr list --repo <repo> --status all            # All PRs
ccpr list --repo <repo> --format json           # JSON output
```

### Version

```bash
ccpr --version
```

### Flags

```
--format     Output format: summary (default), json, patch (review only)
--profile    AWS profile name
--region     AWS region
--config     Path to configuration file
--repo       Repository name
--pr-id      Pull request ID (review only)
--status     PR status filter: open, closed, all (list only)
```

### AWS Profile Resolution

1. `--profile` flag
2. `profile` in config file
3. `AWS_PROFILE` environment variable
4. default

### AWS Region Resolution

1. PR URL (extracted automatically)
2. `--region` flag
3. `region` in config file

## Use Cases

- AI-assisted code review: `ccpr review <url> --format json` → Claude Code
- CLI-based PR browsing: `ccpr list` + `ccpr review` without opening the console
- Quick PR access: `ccpr open <url>` to jump to the PR in browser

## Using with Claude Code

For AWS CodeCommit repositories, you can use `ccpr` to provide PR data to Claude Code.

See [docs/claude-integration.md](docs/claude-integration.md) for setup instructions.

## Development

```bash
make build    # Build binary to bin/ccpr
make test     # Run all tests with -v -race
make lint     # golangci-lint
make vet      # go vet
make clean    # Remove bin/
```

## License

MIT
