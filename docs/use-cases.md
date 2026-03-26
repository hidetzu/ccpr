# Use Cases

## UC-01 Review a PR with Claude Code

- Actor: Developer
- Trigger: A CodeCommit PR URL is available
- Precondition: AWS credentials are configured, target repository is cloned locally
- Main flow:
  1. Developer gives PR URL to Claude Code
  2. Claude runs `ccpr review <url> --json`
  3. ccpr parses the URL to extract region, repository, and PR ID
  4. ccpr fetches PR metadata and comments via AWS SDK
  5. ccpr generates a diff using local Git
  6. ccpr outputs structured JSON combining metadata, comments, and diff
  7. Claude uses the JSON to generate review feedback
- Success:
  - Developer gets AI-generated review suggestions based on complete PR context
- Failure:
  - Invalid URL format → clear error message with expected format
  - AWS credentials missing/invalid → error with guidance
  - Repository not found → error with repository name

## UC-02 Get PR diff only

- Actor: Developer / Script
- Trigger: Developer needs just the diff from a CodeCommit PR
- Precondition: AWS credentials are configured, target repository is cloned locally
- Main flow:
  1. Developer runs `ccpr review <url> --patch`
  2. ccpr outputs the diff in unified patch format
- Success:
  - Patch output is written to stdout, suitable for piping to other tools

## UC-03 Review a PR with explicit parameters

- Actor: Developer
- Trigger: Developer has repo name, region, and PR ID but not the full URL
- Main flow:
  1. Developer runs `ccpr review --repo <repo> --region <region> --pr-id <id> --json`
  2. ccpr proceeds as in UC-01 step 4 onward
- Success:
  - Same as UC-01
