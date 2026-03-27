# ccpr

![CI](https://github.com/hidetzu/ccpr/actions/workflows/ci.yml/badge.svg)

Turn a CodeCommit PR URL into an AI-ready review in one command.

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
ccpr review <codecommit-pr-url>          # Summary (default)
ccpr review <codecommit-pr-url> --json   # JSON for AI tools
ccpr review <codecommit-pr-url> --patch  # Diff only
```

### List PRs

```bash
ccpr list --repo <repo>                  # OPEN PRs (default)
ccpr list --repo <repo> --status closed  # CLOSED PRs
ccpr list --repo <repo> --status all     # All PRs
ccpr list --repo <repo> --json           # JSON output
```

### Version

```bash
ccpr --version
```

### Flags

```
--json       Output as JSON
--patch      Output diff only (mutually exclusive with --json)
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
