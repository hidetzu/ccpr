# Use Cases

## UC-01 Review a PR with Claude Code

- Actor: Developer
- Trigger: A CodeCommit PR URL is available
- Precondition: AWS credentials are configured, target repository is cloned locally, repo mapping configured in `~/.config/ccpr/config.yaml`
- Main flow:
  1. Developer runs `ccpr review <url> --format json` (directly or via Claude Code)
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
  1. Developer runs `ccpr review <url> --format patch`
  2. ccpr outputs the diff in unified patch format
- Success:
  - Patch output is written to stdout, suitable for piping to other tools

## UC-04 Review a PR with explicit parameters

- Actor: Developer
- Trigger: Developer has repo name, region, and PR ID but not the full URL
- Main flow:
  1. Developer runs `ccpr review --repo <repo> --region <region> --pr-id <id> --format json`
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

## UC-06 Validate environment and config

- Actor: Developer
- Trigger: Developer wants to verify ccpr setup is working correctly
- Precondition: ccpr is installed
- Main flow:
  1. Developer runs `ccpr doctor`
  2. ccpr checks config file existence and validity
  3. ccpr validates AWS credentials via STS GetCallerIdentity
  4. ccpr checks each repoMappings entry (path exists, is a git repo)
  5. ccpr prints a checklist of results with pass/fail markers
- Success:
  - All checks pass — developer is confident the setup works
- Failure:
  - Each failed check includes a suggestion for how to fix it

## UC-07 Post a comment to a PR

- Actor: Developer
- Trigger: Developer wants to post a review comment to a CodeCommit PR from the CLI
- Precondition: AWS credentials are configured, config file exists
- Main flow:
  1. Developer runs `ccpr comment <PR_URL> --body "comment text"`
  2. ccpr parses URL, resolves config, fetches PR metadata for commit IDs
  3. ccpr calls PostCommentForPullRequest API
  4. ccpr prints comment ID and timestamp on success
- Alternative flows:
  - `--body-file review.md` reads body from file
  - `--body -` reads body from stdin
  - `--repo`, `--pr-id` flags instead of URL
- Success:
  - Comment is posted and confirmation is printed
- Failure:
  - Missing body → error with usage hint
  - AWS API failure → error with context

## UC-08 Review a PR via Claude Code skill

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

## UC-09 Create a CodeCommit PR

- Actor: Developer
- Trigger: Developer wants to create a CodeCommit PR from the CLI
- Precondition: AWS credentials are configured, config file exists with repo mapping
- Main flow:
  1. Developer runs `ccpr create --repo my-repo --title "Add feature X" --dest main`
  2. ccpr resolves source branch from current Git HEAD (or `--source` flag)
  3. ccpr resolves AWS profile and region via standard priority chain
  4. ccpr calls CreatePullRequest API
  5. ccpr prints PR ID, title, branches, and console URL
- Alternative flows:
  - `--description "text"` sets PR description inline
  - `--description -` reads description from stdin
  - `--description-file desc.md` reads description from file
  - `--source feature/fix` specifies source branch explicitly
  - `--format json` outputs machine-readable result for downstream commands
- Success:
  - PR is created and confirmation with console URL is printed
- Failure:
  - Missing required flags (--repo, --title, --dest) → error with usage hint
  - AWS API failure → error with context
  - Source branch same as destination → error with message

## UC-10 Open a PR in the browser

- Actor: Developer
- Trigger: Developer wants to quickly open a CodeCommit PR in the AWS console
- Precondition: Config file exists with region configured
- Main flow:
  1. Developer runs `ccpr open <PR_URL>`
  2. ccpr parses URL to extract region, repository, and PR ID
  3. ccpr opens the CodeCommit console URL in the default browser
- Alternative flows:
  - `--repo my-repo --pr-id 42` flags instead of URL
- Success:
  - PR page opens in the browser
- Failure:
  - Missing region → error with guidance
  - Browser cannot be opened → prints URL to stdout as fallback

## UC-11 List PRs via MCP

- Actor: Developer using Claude Code or another MCP client
- Trigger: Developer wants the AI assistant to inspect available CodeCommit PRs without manually running `ccpr list`
- Precondition: `ccpr-mcp` is built and registered with the MCP client, AWS credentials are configured, config file exists with region/profile as needed
- Main flow:
  1. Developer asks the MCP client to list pull requests for a repository
  2. The client invokes the `ccpr_list` MCP tool with `repo` and optional `status`, `region`, `profile`, `config`
  3. `ccpr-mcp` resolves config, AWS profile, and AWS region using the same rules as `ccpr list`
  4. `ccpr-mcp` calls CodeCommit through the shared list use case
  5. `ccpr-mcp` returns `{ "pullRequests": [...] }`, where each array item uses the same PR summary schema as `ccpr list --format json`
- Success:
  - Claude Code can retrieve the PR list directly as a tool result and use it in the review workflow
- Failure:
  - Missing repo → tool call returns a validation error
  - Missing region or invalid AWS credentials → tool call returns the same guidance as the CLI path
- Note:
  - This first MCP integration is read-only. Write-side MCP tools and MCP tools for create/comment/open remain out of scope for this use case.

## UC-12 Review PR via MCP

- Actor: Developer using Claude Code or another MCP client
- Trigger: Developer wants the AI assistant to fetch a PR's metadata, comments, and diff for review without manually running `ccpr review --format json`
- Precondition: `ccpr-mcp` is built and registered with the MCP client, AWS credentials are configured, config file maps the repo to a local Git path, region/profile are resolvable
- Main flow:
  1. Developer asks the MCP client to review a specific PR
  2. The client invokes the `ccpr_review` MCP tool with either a `url` or `repo` + `prId` (plus optional `region`, `profile`, `config`)
  3. `ccpr-mcp` resolves config, AWS profile, AWS region, and the local repo path using the same rules as `ccpr review`
  4. `ccpr-mcp` fetches PR metadata and comments through CodeCommit and generates a unified diff via local Git through the shared review use case
  5. `ccpr-mcp` returns the same `{metadata, comments, diff}` object as `ccpr review --format json`
- Success:
  - Claude Code receives the full review payload as a single structured tool result and proceeds to generate review feedback
- Failure:
  - Missing both `url` and `repo` + `prId` → tool call returns a validation error
  - Missing region, missing repo mapping, invalid AWS credentials, or local Git diff failure → tool call returns the same guidance as the CLI path
- Note:
  - Only the structured (JSON-equivalent) output shape is exposed via MCP. The `summary` and `patch` formats remain CLI-only.

## UC-13 Post a PR Comment via MCP

- Actor: Developer using Claude Code or another MCP client
- Trigger: After reviewing a PR with `ccpr_review`, the AI assistant has feedback ready and the developer wants to post it back to the PR without dropping to the shell
- Precondition: `ccpr-mcp` is built and registered with the MCP client, AWS credentials are configured, region/profile are resolvable, and the MCP host is configured to prompt the user before invoking write-side tools
- Main flow:
  1. Developer instructs the MCP client to post a review comment on a specific PR
  2. The client invokes the `ccpr_comment` MCP tool with `body` plus either a `url` or `repo` + `prId` (and optional `region`, `profile`, `config`)
  3. The MCP host prompts the user to approve the call (because the tool is write-side)
  4. `ccpr-mcp` resolves config, AWS profile, and AWS region using the same rules as `ccpr comment`
  5. `ccpr-mcp` fetches PR metadata to obtain the source/destination commits, then calls CodeCommit `PostComment` through the shared comment use case
  6. `ccpr-mcp` returns the same `{commentId, pullRequestId, authorArn, creationDate}` object as `ccpr comment --format json`
- Success:
  - The comment is posted to the PR and Claude Code receives the resulting comment metadata
- Failure:
  - Missing `body`, or missing both `url` and (`repo` + `prId`) → tool call returns a validation error
  - Missing region, invalid AWS credentials, PR metadata fetch failure, or `PostComment` failure → tool call returns the same guidance as the CLI path
- Note:
  - This is the first write-side MCP tool. Each successful call produces a real comment in CodeCommit; the tool description and README must make this explicit so the MCP host's per-call approval gate is meaningful.
  - `--body-file` and stdin (`-`) input forms remain CLI-only — MCP callers pass the comment body as a string argument.
