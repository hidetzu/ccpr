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
  another-repo: /home/user/projects/another-repo
```

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

Output (default: summary):
  --json       Output as JSON (for AI tools / scripts)
  --patch      Output diff only in unified patch format

Flags:
  --profile    AWS profile name
  --repo       Repository name (alternative to URL)
  --region     AWS region (alternative to URL)
  --pr-id      Pull request ID (alternative to URL)
  --config     Path to configuration file
```

### Flag Exclusivity

`--json` and `--patch` are mutually exclusive. If both are specified, the CLI
exits with code 1 and prints an error to stderr:

```
error: --json and --patch are mutually exclusive
```

If neither is specified, `--json` is the default.

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
- In JSON mode, errors are also written as JSON to stderr (FR: NFR-03):

```json
{
  "error": {
    "code": "INVALID_URL",
    "message": "could not parse CodeCommit PR URL: missing region"
  }
}
```

Error codes:
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

## Dependencies

- `aws-sdk-go-v2` — CodeCommit API calls
- `os/exec` — local Git invocation for diff
- `gopkg.in/yaml.v3` — configuration file parsing
- Standard library for JSON, I/O, flag parsing
