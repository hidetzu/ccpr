# Technical Specification

## Module Structure

```
ccpr/
├── cmd/            # CLI entrypoint and command definitions
├── internal/
│   ├── parser/     # PR URL parsing (FR-01)
│   ├── app/        # Shared use cases for CLI and MCP
│   ├── codecommit/ # AWS CodeCommit client (FR-02, FR-03)
│   ├── diff/       # Local Git diff generation (FR-04)
│   ├── config/     # Configuration loading, repo mapping (FR-07)
│   └── output/     # JSON/patch formatting (FR-05, FR-06)
├── cmd/ccpr/       # Human-facing CLI
└── cmd/ccpr-mcp/   # MCP server over stdio
```

## Key Interfaces

### URLParser

```go
// ParseResult holds the extracted components from a CodeCommit PR URL.
type ParseResult struct {
    Region     string
    Repository string
    PRId       string
}

// Parse extracts region, repository, and PR ID from a CodeCommit PR URL.
func Parse(rawURL string) (ParseResult, error)
```

### CodeCommitClient

```go
type PRMetadata struct {
    Title             string
    Description       string
    AuthorARN         string
    SourceBranch      string
    DestinationBranch string
    Status            string
    CreationDate      time.Time
}

type Comment struct {
    Author    string
    Content   string
    Timestamp time.Time
    FilePath  string // empty if PR-level comment
}

type Client interface {
    GetPRMetadata(ctx context.Context, repo, prID string) (PRMetadata, error)
    GetPRComments(ctx context.Context, repo, prID string) ([]Comment, error)
}
```

### DiffGenerator

```go
// GenerateDiff produces a unified diff using merge-base strategy.
//
// Steps:
//   1. Resolve merge-base: git merge-base <dest> <source>
//   2. Generate diff:      git diff <merge-base> origin/<source>
//
// This ensures the diff only contains changes introduced by the source branch,
// excluding unrelated changes merged into the destination after the branch point.
func GenerateDiff(repoPath, sourceBranch, destBranch string) (string, error)
```

### Config

```go
// Config holds application configuration loaded from file.
type Config struct {
    RepoMappings map[string]string `yaml:"repoMappings"`
}

// ResolveRepoPath returns the local filesystem path for a CodeCommit repository name.
// Returns an error if no mapping is configured.
func (c *Config) ResolveRepoPath(repoName string) (string, error)
```

Configuration file location (searched in order):
1. `--config` flag (explicit path)
2. `.ccpr.yaml` in current directory
3. `~/.config/ccpr/config.yaml`

Example `.ccpr.yaml`:

```yaml
repoMappings:
  my-repo: /work/src/my-repo
  another-repo: ~/projects/another-repo
```

### Path Expansion

`repoMappings` values support a leading `~` or `~/` which is expanded to the
current user's home directory at config load time. Other forms such as
`~otheruser/foo` are returned unchanged.

### Output

```go
type ReviewOutput struct {
    Metadata PRMetadata `json:"metadata"`
    Comments []Comment  `json:"comments"`
    Diff     string     `json:"diff"`
}

// FormatJSON serializes ReviewOutput to JSON and writes to w.
func FormatJSON(w io.Writer, output ReviewOutput) error

// FormatPatch writes the raw diff to w.
func FormatPatch(w io.Writer, diff string) error
```

## CLI

Built with standard `flag` package or a lightweight CLI library (e.g., cobra).

```
ccpr review <url> [flags]

Flags:
  --format     Output format: summary (default), json, patch
  --profile    AWS profile name
  --repo       Repository name (alternative to URL)
  --region     AWS region (alternative to URL)
  --pr-id      Pull request ID (alternative to URL)
  --config     Path to configuration file
```

### Format Flag

`--format` accepts one of: `summary` (default), `json`, `patch`.

Invalid values cause the CLI to exit with code 1 and print an error to stderr.

## Shared Use Cases

The CLI and the MCP server share command logic via `internal/app`. Each shared use case is responsible for input validation, config/profile/region resolution, AWS calls, and producing the same data structure both surfaces serialize.

### List

`ccpr list` and the `ccpr_list` MCP tool both call the same internal use case.

```go
type ListPullRequestsOptions struct {
    Repo    string
    Status  string
    Config  string
    Profile string
    Region  string
}

type ListPullRequest struct {
    PRId              string `json:"prId"`
    Title             string `json:"title"`
    AuthorARN         string `json:"authorArn"`
    SourceBranch      string `json:"sourceBranch"`
    DestinationBranch string `json:"destinationBranch"`
    Status            string `json:"status"`
    CreationDate      string `json:"creationDate"`
}
```

Behavior:
1. Validate `repo` and `status`
2. Load config
3. Resolve profile and region using the standard priority rules
4. Create a CodeCommit client
5. Call `ListPRs`
6. Return `[]ListPullRequest`

### Review

`ccpr review` and the `ccpr_review` MCP tool both call the same internal use case.

```go
type GetReviewOptions struct {
    URL     string
    Repo    string
    PRId    string
    Region  string
    Profile string
    Config  string
}

// ReviewPayload mirrors output.ReviewOutput so the JSON shape stays
// stable across the CLI and MCP paths.
type ReviewPayload = output.ReviewOutput
```

Behavior:
1. If `URL` is set, parse it to extract region, repo, and PR ID. Otherwise require `Repo` and `PRId`.
2. Load config
3. Resolve profile and region using the standard priority rules (URL-derived region wins when present)
4. Resolve the local repo path via the config repo mapping
5. Create a CodeCommit client
6. Fetch PR metadata and PR comments
7. Generate the unified diff via local Git (merge-base strategy, see "Diff Strategy")
8. Return a `ReviewPayload` with `metadata`, `comments`, and `diff`

This keeps the CLI summary/JSON/patch rendering and the MCP tool response backed by the same data retrieval path.

## MCP Server (FR-18)

### Binary

```
ccpr-mcp
```

Source location:

```
cmd/ccpr-mcp/
```

The MCP server is a separate binary from `ccpr` and communicates over stdio.

### Tool: `ccpr_list`

Lists CodeCommit pull requests for a repository.

Input schema:

```json
{
  "repo": "my-repo",
  "status": "open",
  "config": "/path/to/config.yaml",
  "profile": "my-profile",
  "region": "ap-northeast-1"
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `repo` | string | yes | | CodeCommit repository name |
| `status` | string | no | `open` | PR status filter: `open`, `closed`, or `all` |
| `config` | string | no | | Path to ccpr config file |
| `profile` | string | no | | AWS profile override |
| `region` | string | no | | AWS region override |

Output schema:

```json
{
  "pullRequests": [
    {
      "prId": "42",
      "title": "Add feature X",
      "authorArn": "arn:aws:iam::123456789012:user/example",
      "sourceBranch": "feature/x",
      "destinationBranch": "main",
      "status": "OPEN",
      "creationDate": "2026-04-01T10:00:00Z"
    }
  ]
}
```

The MCP tool wraps the PR summary array in an object under `pullRequests` because MCP tool output schemas must be objects. Each element of `pullRequests` uses the same field names and types as `ccpr list --format json`. Consumers must ignore unknown fields for forward compatibility.

### Tool: `ccpr_review`

Fetches PR metadata, comments, and the local-Git-generated diff for a single pull request.

Input schema:

```json
{
  "url": "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42",
  "repo": "my-repo",
  "prId": "42",
  "region": "ap-northeast-1",
  "profile": "my-profile",
  "config": "/path/to/config.yaml"
}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | conditional | | Full CodeCommit PR URL. When provided, takes priority for region/repo/PR ID resolution |
| `repo` | string | conditional | | CodeCommit repository name. Required when `url` is not provided |
| `prId` | string | conditional | | Pull request ID. Required when `url` is not provided |
| `region` | string | no | | AWS region override |
| `profile` | string | no | | AWS profile override |
| `config` | string | no | | Path to ccpr config file |

At least one of (`url`) or (`repo` and `prId`) must be provided.

Output schema:

```json
{
  "metadata": {
    "prId": "42",
    "title": "Add feature X",
    "description": "...",
    "author": "example",
    "authorArn": "arn:aws:iam::123456789012:user/example",
    "sourceBranch": "feature/x",
    "destinationBranch": "main",
    "status": "OPEN",
    "creationDate": "2026-04-01T10:00:00Z"
  },
  "comments": [],
  "diff": "diff --git a/... b/...\n..."
}
```

The output is the same schema as `ccpr review --format json` (already an object, so no wrapper is needed). See `docs/json-schema.md` for the full field-level contract. Consumers must ignore unknown fields for forward compatibility.

The MCP tool only exposes this structured shape. The CLI's `summary` and `patch` formats are not surfaced via MCP.

### Error Handling

MCP tool errors are returned as tool call errors. Error messages follow the same validation and AWS/config guidance as the shared use cases.

`ccpr_list`:

- missing `repo`
- invalid `status`
- config load failure
- missing region
- CodeCommit client creation failure
- CodeCommit list failure

`ccpr_review`:

- missing both `url` and (`repo` + `prId`)
- invalid `url`
- config load failure
- missing region (when `url` does not supply one)
- repo not mapped to a local Git path
- CodeCommit client creation failure
- PR metadata or comment fetch failure
- local Git diff generation failure

Structured MCP-specific error codes are out of scope for this release.

## Diff Strategy

The diff is generated locally using a merge-base approach:

```
      A---B---C  (destination)
       \
        D---E    (source)

  merge-base = A
  diff = git diff A origin/source
```

1. **Fetch latest refs**: `git fetch origin <source> <dest>`
2. **Find merge-base**: `git merge-base origin/<dest> origin/<source>`
3. **Generate diff**: `git diff <merge-base> origin/<source>`

This approach:
- Shows only changes introduced by the source branch
- Excludes commits merged into destination after the branch point
- Matches what a reviewer expects to see in a PR diff

If merge-base resolution fails (e.g., unrelated histories), return an error
with exit code 2.

## Error Handling

- All errors returned as `error` values, not panics
- CLI layer formats errors to stderr with context
- Exit codes: 0 = success, 1 = user error (bad input), 2 = system error (AWS/Git)
- In JSON mode, errors are currently printed as plain text to stderr (same as non-JSON mode). Structured JSON error output (NFR-03) is planned for a future release:

```json
{
  "error": {
    "code": "INVALID_URL",
    "message": "could not parse CodeCommit PR URL: missing region"
  }
}
```

Reserved error codes (for future use):
- `INVALID_URL` — malformed PR URL
- `INVALID_FLAGS` — conflicting or missing CLI flags
- `CONFIG_NOT_FOUND` — configuration file not found
- `REPO_NOT_MAPPED` — repository has no local path mapping
- `AWS_ERROR` — CodeCommit API failure
- `GIT_ERROR` — local Git operation failure

## Claude Code Skill (FR-11)

The repository provides a Claude Code skill for direct PR review integration.

### File Location

```
examples/claude/ccpr-review/SKILL.md
```

### Skill Design

- **Thin wrapper**: The skill only invokes `ccpr review` and defines review behavior
- **No ccpr-specific logic**: All data fetching and formatting is handled by the CLI
- **Fixed review focus**: correctness, security, performance, readability
- **User-customizable**: Users copy the skill and edit as needed

### Installation

Copy `examples/claude/ccpr-review/SKILL.md` to:
- `.claude/skills/ccpr-review/SKILL.md` (project-scoped)
- `~/.claude/skills/ccpr-review/SKILL.md` (global)

### Frontmatter

```yaml
---
name: ccpr-review
description: Review a CodeCommit pull request using ccpr structured output
allowed-tools: Bash(ccpr *)
---
```

- `allowed-tools` restricts the skill to only run ccpr commands

## Init Command (FR-12)

### Behavior

```
ccpr init [--profile <name>] [--region <region>] [--force]
```

1. Determine config path: `~/.config/ccpr/config.yaml`
2. If file exists and `--force` not set → exit with error
3. Detect defaults:
   - Profile: `--profile` flag > `AWS_PROFILE` env > `""`
   - Region: `--region` flag > parsed from `~/.aws/config` > `""`
4. Create parent directory if needed (`~/.config/ccpr/`)
5. Write YAML config
6. Print summary to stdout

### Output

```
Config written to ~/.config/ccpr/config.yaml
  profile: default
  region:  ap-northeast-1
```

### Error Cases

- `CONFIG_EXISTS` — config file already exists (exit code 1)
- `INIT_WRITE_ERROR` — cannot create directory or write file (exit code 2)

## Doctor Command (FR-13)

### Behavior

```
ccpr doctor [--config <path>] [--profile <name>] [--region <region>]
```

Runs the following checks in order:

1. **Config file** — exists and is valid YAML
2. **AWS credentials** — STS GetCallerIdentity succeeds
3. **Repo mappings** — each path exists and is a git repository

### Output

```
✔ Config file: ~/.config/ccpr/config.yaml
✔ AWS credentials: arn:aws:iam::123456789012:user/example-user
✔ Repo mapping: my-repo → /home/user/src/my-repo (git OK)
✖ Repo mapping: other-repo → /tmp/missing (path not found)
  → Check repoMappings in ~/.config/ccpr/config.yaml

3/4 checks passed
```

### Exit Codes

- `0` — all checks passed
- `1` — one or more checks failed

## Open Command (FR-15)

### Behavior

```
ccpr open <PR_URL>
ccpr open --repo <repo> --pr-id <id> --region <region>
```

1. Resolve repository name, PR ID, and region from URL or flags
2. Load config, resolve region (required)
3. Build CodeCommit console URL
4. Open URL in default browser (`xdg-open` on Linux, `open` on macOS, `cmd /c start` on Windows)
5. If browser cannot be opened, print URL to stdout as fallback

### Error Cases

- `MISSING_REGION` — region could not be resolved from URL, `--region`, config, `AWS_REGION`, or `AWS_DEFAULT_REGION` (exit code 1)
- `INVALID_URL` — malformed PR URL (exit code 1)

## Comment Command (FR-14)

### Behavior

```
ccpr comment <PR_URL> --body "comment text"
ccpr comment <PR_URL> --body-file review.md
ccpr comment <PR_URL> --body -
ccpr comment --repo <repo> --pr-id <id> --body "comment text"
```

1. Resolve repository name, PR ID, and region from URL or flags
2. Load config, resolve profile/region
3. Call `GetPullRequest` to retrieve `sourceCommit` and `destinationCommit`
4. Call `PostCommentForPullRequest` with resolved parameters and body
5. Print result

### Body Input Priority

1. `--body` flag (if not `-`)
2. `--body -` (read stdin)
3. `--body-file` path

`--body` and `--body-file` are mutually exclusive.

### Output

```
Comment posted successfully.
Comment ID: eb596ff8-5133-438f-88d6-7c94f693302b
PR: #957
Author: example-user
Created: 2026-03-31T17:22:24+09:00
```

With `--format json`:

```json
{
  "commentId": "eb596ff8-...",
  "pullRequestId": "957",
  "authorArn": "arn:aws:iam::123456789012:user/example-user",
  "creationDate": "2026-03-31T17:22:24Z"
}
```

### Error Cases

- `MISSING_BODY` — no body provided (exit code 1)
- `BODY_CONFLICT` — both `--body` and `--body-file` specified (exit code 1)
- `AWS_ERROR` — PostCommentForPullRequest failure (exit code 2)

## Create Command (FR-15)

### Behavior

```
ccpr create --repo <repo> --title "PR title" --dest main
ccpr create --repo <repo> --title "PR title" --dest main --source feature/x
ccpr create --repo <repo> --title "PR title" --dest main --description "Description text"
ccpr create --repo <repo> --title "PR title" --dest main --description -
ccpr create --repo <repo> --title "PR title" --dest main --description-file desc.md
ccpr create --repo <repo> --title "PR title" --dest main --format json
```

1. Validate required flags: `--repo`, `--title`, `--dest`
2. Resolve source branch: `--source` flag > current Git branch (from repo mapping path)
3. Resolve description: `--description` / `--description -` / `--description-file` (same pattern as comment command's body resolution)
4. Load config, resolve profile/region
5. Call `CreatePullRequest` API with title, description, source branch, destination branch, repository name
6. Build console URL from region and repository name
7. Print result

### Description Input Priority

1. `--description` flag (if not `-`)
2. `--description -` (read stdin)
3. `--description-file` path

`--description` and `--description-file` are mutually exclusive. Description is optional — if none provided, the PR is created without a description.

### Output

```
✔ Pull request created
  PR #42: Add feature X
  Repository: my-repo
  Source: feature/add-x → Destination: main
  URL: https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42
```

With `--format json`:

```json
{
  "prId": "42",
  "title": "Add feature X",
  "repository": "my-repo",
  "sourceBranch": "feature/add-x",
  "destinationBranch": "main",
  "url": "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42"
}
```

### Source Branch Resolution

When `--source` is not specified:
1. Resolve local repo path from config repo mappings using `--repo`
2. Run `git -C <repo-path> rev-parse --abbrev-ref HEAD` to get the current branch
3. Error if the resolved branch is the same as `--dest`

### Console URL Format

```
https://<region>.console.aws.amazon.com/codesuite/codecommit/repositories/<repo>/pull-requests/<prId>
```

### Error Cases

- `MISSING_FLAGS` — required flags not provided (exit code 1)
- `DESCRIPTION_CONFLICT` — both `--description` and `--description-file` specified (exit code 1)
- `SAME_BRANCH` — source and destination branches are the same (exit code 1)
- `AWS_ERROR` — CreatePullRequest API failure (exit code 2)
- `GIT_ERROR` — cannot determine current branch (exit code 2)

## Dependencies

- `aws-sdk-go-v2` — CodeCommit API calls
- `modelcontextprotocol/go-sdk` — MCP server and stdio transport
- `os/exec` — local Git invocation for diff
- `gopkg.in/yaml.v3` — configuration file parsing
- Standard library for JSON, I/O, flag parsing
