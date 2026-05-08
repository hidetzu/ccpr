# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ccpr is a local helper CLI that replaces `gh` for CodeCommit-based review workflows with Claude Code.

It takes a CodeCommit PR URL, fetches metadata/comments/diffs, and outputs structured data for AI-assisted code review. This is not an official CodeCommit integration — it exists to bridge the gap between CodeCommit and AI review tools.

## Intended Workflow

```
ccpr review <codecommit-pr-url>                 # Summary for humans
ccpr review <codecommit-pr-url> --format json   # JSON for Claude Code / AI tools
ccpr review <codecommit-pr-url> --format patch  # Diff only
```

The primary use case: developer runs `ccpr review <url> --format json`, passes the output to Claude Code, and Claude generates review comments from it. Do not search for an official Claude Code ↔ CodeCommit integration — this repository is that bridge.

## Development Rules

- Follow spec-driven development
- Do not implement features before updating docs
- Always update:
  - docs/use-cases.md
  - docs/requirements.md
  - docs/spec.md
  before writing implementation code
- Do not introduce breaking changes to JSON output (see docs/versioning.md)
  - Do not rename, remove, or change the type of existing JSON fields
  - New fields must be backward-compatible additions only
  - If golden tests in `cmd/ccpr/testdata/` need updating, flag for explicit review

## Scope

MVP implemented:

- PR URL parsing
- Fetching PR metadata and comments
- Generating git-based diff (merge-base strategy)
- Output: summary (default), JSON (`--format json`), patch (`--format patch`)
- AWS profile resolution (`--profile` > config > env > default)
- Config file with repo mappings
- PR listing (`ccpr list`)
- Open PR in browser (`ccpr open`)
- PR creation (`ccpr create`)
- Post comment to PR (`ccpr comment`)
- Config initialization (`ccpr init`)
- Environment validation (`ccpr doctor`)

Out of scope (permanent non-goals):

- Line-level review comments — `gh` CLI does not expose these either, and ccpr targets feature parity with `gh` for CodeCommit. PR-level comments via `ccpr comment` cover the intended workflow.
- Multi-provider support (GitHub / GitLab) — ccpr exists specifically as a `gh`-style CLI for CodeCommit. GitHub already has `gh` and GitLab has `glab`; expanding ccpr to other providers contradicts its reason for existing.
- TUI — ccpr's role is to emit machine-readable output (JSON / patch) for AI tools. An interactive TUI is a different product and is not aligned with this design.
- Complex approval workflows (multi-reviewer rules, re-request flows, etc.) — out of step with the thin-CLI principle and not part of the `gh`-style baseline ccpr targets.

## Architecture Principles

- Use AWS SDK for CodeCommit interactions
- Use local Git for diff generation (not AWS API)
- CLI should be thin and composable
- Output must be machine-readable (JSON/patch) for AI tools

## Implementation Order

1. Define interfaces
2. Write tests (where applicable)
3. Implement minimal functionality
4. Expand incrementally

Do not write large implementations in a single step.

## Input Model

The primary input is a CodeCommit PR URL.

All commands should accept:

- Full PR URL (preferred)
- Optional flags for repo/region/pr-id (secondary)

## AWS Profile Resolution

Priority order:
1. `--profile` flag (explicit)
2. `profile` field in config file (`~/.config/ccpr/config.yaml`)
3. `AWS_PROFILE` environment variable
4. default

## Build & Development Commands

```bash
make build              # Build binary to bin/ccpr
make build-mcp          # Build MCP server binary to bin/ccpr-mcp
make test               # Run all tests with -v -race
make lint               # golangci-lint v2.11.4
make vet                # go vet
make clean              # Remove bin/
go test ./path/to/pkg -run TestName -v  # Run a single test
```

## CI

GitHub Actions runs on push/PR to main: build, test (`-race`), vet, and golangci-lint. Go version is read from `go.mod`.
