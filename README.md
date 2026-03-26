# ccpr

![CI](https://github.com/yourname/ccpr/actions/workflows/ci.yml/badge.svg)

Turn a CodeCommit PR URL into an AI-ready review in one command.

## Overview

ccpr is a minimal CLI tool for reviewing AWS CodeCommit pull requests using AI tools like Claude Code.

It takes a CodeCommit PR URL as input, fetches metadata, comments, and diffs, and outputs a structured JSON or patch that AI tools can use for review.

## Features (MVP)

- Parse CodeCommit PR URL
- Fetch PR metadata
- Fetch existing comments
- Generate git-based diff (patch)
- Output AI-ready JSON for review

## Example

```bash
ccpr review <codecommit-pr-url> --json
```

## Why

AWS CodeCommit does not provide a developer-friendly CLI experience like `gh` for GitHub.

ccpr fills that gap by providing a simple, URL-driven workflow optimized for AI-assisted code review.

## Status

🚧 Work in progress (MVP)

## License

MIT
