# Use Cases

## UC-01 Review a PR with Claude Code

- Actor: Developer
- Trigger: A CodeCommit PR URL is available
- Precondition: AWS credentials are configured, target repository is cloned locally, repo mapping configured in `~/.config/ccpr/config.yaml`
- Main flow:
  1. Developer runs `ccpr review <url> --json` (directly or via Claude Code)
  2. ccpr parses the URL to extract region, repository, and PR ID
  3. ccpr resolves AWS profile (--profile > config > AWS_PROFILE > default)
  4. ccpr fetches PR metadata and comments via AWS SDK
  5. ccpr generates a diff using local Git (merge-base strategy)
  6. ccpr outputs structured JSON combining metadata, comments, and diff
  7. Developer passes the JSON to Claude Code for review
  8. Claude generates review feedback from the JSON
- Success:
  - Developer gets AI-generated review suggestions based on complete PR context
- Failure:
  - Invalid URL format → clear error message with expected format
  - AWS credentials missing/invalid → error with guidance
  - Repository not mapped in config → error with repository name
- Note:
  - ccpr is a local helper CLI, not an official CodeCommit integration
  - Do not search for or expect an official Claude Code ↔ CodeCommit provider

## UC-02 Quick PR summary

- Actor: Developer
- Trigger: Developer wants a quick overview of a CodeCommit PR
- Main flow:
  1. Developer runs `ccpr review <url>` (no flags)
  2. ccpr outputs a human-readable summary: title, author, status, branches, comment count, files changed
- Success:
  - Developer quickly understands the PR scope without opening the console

## UC-03 Get PR diff only

- Actor: Developer / Script
- Trigger: Developer needs just the diff from a CodeCommit PR
- Precondition: AWS credentials are configured, target repository is cloned locally
- Main flow:
  1. Developer runs `ccpr review <url> --patch`
  2. ccpr outputs the diff in unified patch format
- Success:
  - Patch output is written to stdout, suitable for piping to other tools

## UC-04 Review a PR with explicit parameters

- Actor: Developer
- Trigger: Developer has repo name, region, and PR ID but not the full URL
- Main flow:
  1. Developer runs `ccpr review --repo <repo> --region <region> --pr-id <id> --json`
  2. ccpr proceeds as in UC-01 step 3 onward
- Success:
  - Same as UC-01
