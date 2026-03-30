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

## UC-05 Initialize configuration

- Actor: Developer (first-time setup)
- Trigger: Developer installs ccpr and wants to create a config file
- Precondition: ccpr is installed
- Main flow:
  1. Developer runs `ccpr init`
  2. ccpr detects AWS profile from `AWS_PROFILE` env or defaults to empty
  3. ccpr detects region from AWS CLI config or defaults to empty
  4. ccpr writes `~/.config/ccpr/config.yaml` with detected values
  5. ccpr prints the generated config path and values
- Alternative flow:
  1. Developer runs `ccpr init --profile my-profile --region us-east-1`
  2. ccpr uses the explicitly provided values instead of auto-detection
- Success:
  - Config file is created and ready to use
- Failure:
  - Config file already exists → error with message (use `--force` to overwrite)
  - Cannot create config directory → error with path

## UC-06 Review a PR via Claude Code skill

- Actor: Developer using Claude Code
- Trigger: A CodeCommit PR URL is available and developer wants AI review directly in Claude Code
- Precondition: ccpr installed and configured, SKILL.md copied from `examples/claude/ccpr-review/` to user's `.claude/skills/ccpr-review/`
- Main flow:
  1. Developer invokes `/ccpr-review <PR_URL>` in Claude Code
  2. The skill runs `ccpr review <PR_URL> --format json`
  3. Claude reviews the PR output focusing on: correctness, security, performance, readability
  4. Claude returns a structured review: summary, risks, suggested changes
- Success:
  - Developer gets AI-generated review without manual piping or CLAUDE.md setup
- Note:
  - The skill is a thin wrapper — ccpr handles data fetching, the skill defines review behavior
  - Users can customize review focus by copying the skill to their own `.claude/skills/`
  - Copy `examples/claude/ccpr-review/SKILL.md` to `.claude/skills/ccpr-review/` (project) or `~/.claude/skills/ccpr-review/` (global)
