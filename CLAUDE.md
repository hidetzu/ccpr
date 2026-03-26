# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ccpr is a Go CLI tool that takes an AWS CodeCommit PR URL, fetches metadata/comments/diffs, and outputs AI-ready JSON for code review. It is in early MVP stage — no Go source files exist yet.

## Development Rules

- Follow spec-driven development
- Do not implement features before updating docs
- Always update:
  - docs/use-cases.md
  - docs/requirements.md
  - docs/spec.md
    before writing implementation code

## MVP Scope

Focus only on:

- PR URL parsing
- Fetching PR metadata
- Fetching comments
- Generating git-based diff
- Outputting review JSON

Out of scope (for now):

- Line-level comments
- TUI
- Multi-provider support (GitHub/GitLab)
- Complex approval workflows

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

## Build & Development Commands

```bash
go build ./...          # Build all packages
go test ./... -v -race  # Run all tests with race detection
go test ./path/to/pkg -run TestName -v  # Run a single test
go vet ./...            # Static analysis
```

Lint is run via golangci-lint v2.11.4 (see CI config):

```bash
golangci-lint run
```

## CI

GitHub Actions runs on push/PR to main: build, test (`-race`), vet, and golangci-lint. Go version is read from `go.mod`.

## CLI Interface (planned)

```bash
ccpr review <codecommit-pr-url> --json
```
