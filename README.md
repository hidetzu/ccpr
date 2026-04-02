# ccpr

[![CI](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hidetzu/ccpr)](https://goreportcard.com/report/github.com/hidetzu/ccpr)
[![Go Reference](https://pkg.go.dev/badge/github.com/hidetzu/ccpr.svg)](https://pkg.go.dev/github.com/hidetzu/ccpr)
[![Release](https://img.shields.io/github/v/release/hidetzu/ccpr)](https://github.com/hidetzu/ccpr/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Turn a CodeCommit pull request into AI-ready review input in one command.

[日本語ドキュメントはこちら](README.ja.md)

```bash
ccpr review <PR_URL> --format json | claude -p "Review this PR"
```

## What ccpr does

- Fetch PR metadata, comments, and diffs from CodeCommit in one shot
- Output as JSON / Patch — ready to pipe into AI tools
- Automate code review with Claude Code

## Before / After

**Without ccpr:**
CodeCommit's CLI is fragmented — gathering PR metadata, comments, and diffs requires multiple API calls and manual assembly. Feeding that to AI tools means even more glue work.

**With ccpr:**
One command gives you everything. JSON output plugs directly into Claude Code or any AI tool for instant review.

## Quick Start

Requires [Go](https://go.dev/dl/) 1.23 or later.

```bash
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
ccpr init
ccpr review <codecommit-pr-url>
```

## AI Integration

### Install from Claude Code

Paste this into Claude Code to install and set up ccpr:

```
Read https://raw.githubusercontent.com/hidetzu/ccpr/main/docs/claude-integration.md and install ccpr
```

### Pipe to Claude

```bash
ccpr review <PR_URL> --format json | claude -p "Review this PR"
```

### Claude Code skill (recommended)

ccpr provides a Claude Code skill for direct PR review integration.

```bash
mkdir -p .claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md .claude/skills/ccpr-review/
```

Then in Claude Code:

```
/ccpr-review <codecommit-pr-url>
```

See [docs/claude-integration.md](docs/claude-integration.md) for more options.

## Use Cases

- **AI code review** — `ccpr review <url> --format json` + Claude Code
- **Create a PR** — `ccpr create --repo <repo> --title "Add feature" --dest main`
- **Post review comments** — `ccpr comment <url> --body-file review.md`
- **Quick PR summary** — `ccpr review <url>` for a human-readable overview
- **CI integration** — pipe JSON/patch output to automated pipelines
- **CLI-based PR browsing** — `ccpr list` + `ccpr review` without opening the console

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

### Create a PR

```bash
ccpr create --repo <repo> --title "Add feature X" --dest main
ccpr create --repo <repo> --title "Add feature X" --dest main --source feature/x
ccpr create --repo <repo> --title "Add feature X" --dest main --description-file desc.md
ccpr create --repo <repo> --title "Add feature X" --dest main --format json
```

### Post a comment

```bash
ccpr comment <codecommit-pr-url> --body "LGTM"
ccpr comment <codecommit-pr-url> --body-file review.md
echo "Looks good" | ccpr comment <codecommit-pr-url> --body -
```

### Setup and diagnostics

```bash
ccpr init                   # Generate config file
ccpr doctor                 # Validate environment and config
```

### Flags

```
--format           Output format: summary (default), json, patch (review only)
--title            PR title (create only, required)
--dest             Destination branch (create only, required)
--source           Source branch (create only, defaults to current branch)
--description      PR description; use - for stdin (create only)
--description-file Path to PR description file (create only)
--body             Comment body; use - for stdin (comment only)
--body-file        Path to comment body file (comment only)
--profile          AWS profile name
--region           AWS region
--config           Path to configuration file
--repo             Repository name
--pr-id            Pull request ID (review/comment only)
--status           PR status filter: open, closed, all (list only)
```

## Configuration

`~/.config/ccpr/config.yaml`

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  your-repo: /path/to/local/clone
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

## JSON Output Contract

ccpr guarantees stable JSON output within v1.x releases. See:

- [JSON Output Reference](docs/json-schema.md) — field definitions for all commands
- [Versioning Policy](docs/versioning.md) — SemVer rules and backward compatibility guarantees

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
