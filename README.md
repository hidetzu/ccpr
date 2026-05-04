# ccpr

[![CI](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml/badge.svg)](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hidetzu/ccpr)](https://goreportcard.com/report/github.com/hidetzu/ccpr)
[![Go Reference](https://pkg.go.dev/badge/github.com/hidetzu/ccpr.svg)](https://pkg.go.dev/github.com/hidetzu/ccpr)
[![Release](https://img.shields.io/github/v/release/hidetzu/ccpr)](https://github.com/hidetzu/ccpr/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

## Turn CodeCommit PRs into AI-ready input — in one command

**PR → AI review, fully automated.**

[日本語ドキュメントはこちら](README.ja.md)

---

## If you use CodeCommit and want AI-powered reviews:

- Gathering diffs, comments, and metadata requires multiple API calls
- AWS CLI is fragmented — no single command for PR data
- Copy-pasting into AI tools is a waste of time

ccpr solves all of this.

---

## The simplest way to use it

```bash
ccpr review <PR_URL> --format json | claude -p "Review this PR"
```

---

## Before / After

### Without ccpr (painful)

```bash
aws codecommit get-pull-request --pull-request-id 123
aws codecommit get-comments-for-pull-request --pull-request-id 123
aws codecommit get-differences --repository-name my-repo \
  --before-commit-specifier main --after-commit-specifier feature/x

# + manually assemble and paste into AI...
```

### With ccpr (1 command)

```bash
ccpr review <PR_URL> --format json | claude -p "Review this PR"
```

Fetches PR metadata, comments, and diff in one shot — ready for AI.

---

## Quick Start (30 seconds)

Requires [Go](https://go.dev/dl/) 1.25 or later.

```bash
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
ccpr init
ccpr review <codecommit-pr-url>
```

---

## Use Cases

- AI code review (Claude / GPT)
- PR summary generation
- Automated review in CI
- CLI-only PR workflow (list / review / create)
- Auto-post review comments

---

## Claude Code Integration

### Pipe to Claude (quickest)

```bash
ccpr review <PR_URL> --format json | claude -p "Review this PR"
```

### Install from Claude Code

Paste this into Claude Code to install and set up ccpr:

```
Read https://raw.githubusercontent.com/hidetzu/ccpr/main/docs/claude-integration.md and install ccpr
```

### Claude Code skill (recommended)

```bash
mkdir -p .claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md .claude/skills/ccpr-review/
```

```
/ccpr-review <codecommit-pr-url>
```

See [docs/claude-integration.md](docs/claude-integration.md) for full setup guide.

### MCP server

ccpr also provides an experimental MCP server binary for Claude Code and other MCP clients. Currently exposed tools:

```text
ccpr_list      List PRs for a repository                     (read-only)
ccpr_review    Fetch PR metadata, comments, and diff         (read-only)
ccpr_comment   Post a comment to a PR                        (write-side)
```

Build it locally:

```bash
make build-mcp
```

Register it with Claude Code:

```bash
claude mcp add ccpr -- /path/to/ccpr/bin/ccpr-mcp
```

Then ask Claude Code to list or review CodeCommit PRs:

- `ccpr_list` accepts `repo`, `status`, `region`, `profile`, and `config`, and returns the same PR summary schema as `ccpr list --format json` (wrapped under `pullRequests`).
- `ccpr_review` accepts either `url` or `repo` + `prId`, plus optional `region`, `profile`, and `config`, and returns the same `{metadata, comments, diff}` payload as `ccpr review --format json`.
- `ccpr_comment` accepts `body` and either `url` or `repo` + `prId`, plus optional `region`, `profile`, and `config`, and returns the same `{commentId, pullRequestId, authorArn, creationDate}` payload as `ccpr comment --format json`. **Write-side**: each successful call posts a real comment to CodeCommit, so MCP hosts should prompt before invocation.

---

## What ccpr can do

- Fetch PR data (summary / json / patch)
- List PRs
- Create PRs
- Post comments
- Open PRs in browser
- Full CodeCommit workflow from CLI

---

## Detailed Reference

<details>
<summary>Commands</summary>

```bash
# Review
ccpr review <PR_URL>                 # Summary (default)
ccpr review <PR_URL> --format json   # JSON for AI tools
ccpr review <PR_URL> --format patch  # Diff only

# List
ccpr list --repo <repo>                         # OPEN PRs (default)
ccpr list --repo <repo> --status closed         # CLOSED PRs
ccpr list --repo <repo> --format json           # JSON output

# Create
ccpr create --repo <repo> --title "Add feature X" --dest main
ccpr create --repo <repo> --title "Add feature X" --dest main --source feature/x

# Comment
ccpr comment <PR_URL> --body "LGTM"
ccpr comment <PR_URL> --body-file review.md

# Other
ccpr open <PR_URL>          # Open PR in browser
ccpr init                   # Generate config file
ccpr doctor                 # Validate environment and config
```

</details>

<details>
<summary>Output examples</summary>

#### Summary (default)

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

#### JSON (for AI tools)

```json
{
  "metadata": {
    "prId": "883",
    "title": "Fix login bug",
    "author": "example-user",
    "sourceBranch": "feature/login",
    "destinationBranch": "main",
    "status": "OPEN",
    "creationDate": "2026-03-25T10:30:00Z"
  },
  "comments": [...],
  "diff": "diff --git a/src/login.go ..."
}
```

</details>

<details>
<summary>JSON output contract</summary>

ccpr guarantees stable JSON output within v1.x releases.

- [JSON Output Reference](docs/json-schema.md) — field definitions for all commands
- [Versioning Policy](docs/versioning.md) — SemVer rules and backward compatibility guarantees

</details>

---

## Configuration

`~/.config/ccpr/config.yaml`

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  your-repo: /path/to/local/clone
```

**AWS Profile Resolution:** `--profile` flag > config file > `AWS_PROFILE` env > default

**AWS Region Resolution:** PR URL (auto) > `--region` flag > config file > `AWS_REGION` env > `AWS_DEFAULT_REGION` env

---

## Troubleshooting

- Run `ccpr doctor` first
- Check AWS credentials: `aws sts get-caller-identity --profile your-aws-profile`
- Using SSO? Run `aws sso login --profile your-aws-profile` first
- `no local path mapping` → add repo to `repoMappings` in config.yaml
- `region is required` → set via `--region` flag, config file, or `AWS_REGION` / `AWS_DEFAULT_REGION` env
- Empty diff → run `git fetch origin` and retry

---

## Development

```bash
make build    # Build binary to bin/ccpr
make build-mcp # Build MCP server binary to bin/ccpr-mcp
make test     # Run all tests with -v -race
make lint     # golangci-lint
make vet      # go vet
make clean    # Remove bin/
```

## License

MIT
