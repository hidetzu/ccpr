# Technical Specification

## Module Structure

```
ccpr/
├── cmd/            # CLI entrypoint and command definitions
├── internal/
│   ├── parser/     # PR URL parsing (FR-01)
│   ├── codecommit/ # AWS CodeCommit client (FR-02, FR-03)
│   ├── diff/       # Local Git diff generation (FR-04)
│   ├── config/     # Configuration loading, repo mapping (FR-07)
│   └── output/     # JSON/patch formatting (FR-05, FR-06)
└── main.go
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
- `os/exec` — local Git invocation for diff
- `gopkg.in/yaml.v3` — configuration file parsing
- Standard library for JSON, I/O, flag parsing
