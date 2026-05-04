# Requirements (Revised)

## Functional Requirements

### FR-01 PR URL Parsing

- Parse CodeCommit PR URL to extract: AWS region, repository name, PR ID
- Expected URL format:
  https://<region>.console.aws.amazon.com/codesuite/codecommit/repositories/<repo>/pull-requests/<pr-id>
- Return structured result with extracted fields
- Return clear error for malformed URLs

---

### FR-02 PR Metadata Fetching

- Fetch PR metadata via AWS SDK
- Required fields:
  - title
  - description
  - author
  - source branch
  - destination branch
  - status
  - creation date

---

### FR-03 PR Comments Fetching

- Fetch all comment threads
- Flatten threads into a list
- Include:
  - author
  - content
  - timestamp
  - file path (optional)

Constraints (MVP):
- Line numbers are NOT required
- Thread structure is NOT preserved

---

### FR-04 Diff Generation

- Use local Git (not AWS API)
- Determine merge-base between source and destination
- Generate diff using:
  git diff <merge-base>...<source>
- Output must be unified diff format

---

### FR-05 Summary Output (default)

- Default output mode when no format flag is specified
- Display human-readable summary:
  - PR title, author, status, branches, creation date
  - Comment count
  - Changed files count
- Designed for CLI users

---

### FR-06 JSON Output

- Triggered by `--format json` flag
- Combine metadata, comments, and diff
- Output as a single JSON document
- JSON schema must be stable (see docs/json-schema.md)
- Designed for AI tools and script integration

Constraints:
- Field names must not change without versioning
- Optional fields must be documented

---

### FR-07 Patch Output

- Triggered by `--format patch` flag
- Output raw diff only

Constraints:
- `--format` accepts one value only; `json` and `patch` cannot be combined

---

### FR-08 Local Repository Resolution

- Map repository name to local filesystem path
- Configuration file defines mapping
- Path values support `~` and `~/` as a prefix for the user's home directory
  - `~/src/repo` → `/home/<user>/src/repo` (expanded at config load time)
  - `~user/foo` form is **not** supported (returned as-is)

Example:

repoMappings:
  my-repo: /work/src/my-repo
  another-repo: ~/src/another-repo

- Return error if mapping is missing

---

### FR-09 AWS Profile Resolution

- Resolve AWS profile in the following order:
  1. `--profile` flag (explicit)
  2. `profile` field in config file
  3. `AWS_PROFILE` environment variable
  4. default

---

### FR-10 PR List

- List pull requests for a given repository
- Default filter: OPEN
- Support `--status` flag: `open` (default), `closed`, `all`
- Summary output: PR ID, title, branches, status, creation date
- Support `--format json` for machine-readable output
- Repository resolved via `--repo` flag (required)

---

### FR-11 Claude Code Skill

- Provide a sample Claude Code skill at `examples/claude/ccpr-review/SKILL.md`
- The skill invokes `ccpr review $ARGUMENTS --format json` and reviews the output
- Fixed review focus: correctness, security, performance, readability
- Users copy the skill to their `.claude/skills/` and customize as needed

---

### FR-12 Config Initialization

- Generate `~/.config/ccpr/config.yaml` with sensible defaults
- Auto-detect:
  - AWS profile from `AWS_PROFILE` environment variable
  - Region from AWS CLI shared config (`~/.aws/config`)
- Accept explicit overrides via `--profile` and `--region` flags
- Refuse to overwrite existing config unless `--force` is specified
- Print generated config path and values to stdout

---

### FR-13 Environment Validation (doctor)

- Check config file existence and YAML validity
- Validate AWS credentials via STS GetCallerIdentity
- Check each repoMappings entry:
  - Path exists on filesystem
  - Path is a git repository
- Print checklist with pass (✔) / fail (✖) markers
- Suggest fix for each failure
- Exit code 0 if all checks pass, 1 if any check fails

---

### FR-14 Post Comment to PR

- Post a comment to a CodeCommit pull request
- Accept comment body via:
  - `--body` flag (inline text)
  - `--body-file` flag (read from file)
  - `--body -` (read from stdin)
- Resolve PR parameters from URL or `--repo` + `--pr-id` flags
- Retrieve source/destination commit IDs via GetPullRequest API
- Call PostCommentForPullRequest API
- Print comment ID and timestamp on success
- Support `--format json` for machine-readable output

---

### FR-15 Open PR in Browser

- Open a CodeCommit PR in the default browser
- Resolve PR parameters from URL or `--repo` + `--pr-id` flags
- Build CodeCommit console URL from region, repository, and PR ID
- Region is required (from URL, or resolved per FR-17)
- Fall back to printing URL to stdout if browser cannot be opened

---

### FR-16 Create Pull Request

- Create a CodeCommit pull request via the CreatePullRequest API
- Required flags: `--repo`, `--title`, `--dest`
- Optional flags: `--source`, `--description`, `--description-file`, `--format`, `--region`, `--profile`, `--config`
- Source branch defaults to the current Git branch if `--source` is not specified
- Accept description via:
  - `--description` flag (inline text)
  - `--description-file` flag (read from file)
  - `--description -` (read from stdin)
- `--description` and `--description-file` are mutually exclusive
- Summary output (default): PR ID, title, source/destination branches, console URL
- JSON output (`--format json`): machine-readable result for piping to other commands
- Follow AWS profile resolution priority (FR-09) and region resolution priority (FR-17)

---

### FR-17 AWS Region Resolution

- Resolve AWS region in the following order:
  1. `--region` flag (explicit)
  2. `region` field in config file
  3. `AWS_REGION` environment variable
  4. `AWS_DEFAULT_REGION` environment variable
  5. empty (commands that require a region will error with guidance)
- The env-variable fallbacks match the AWS SDK convention so users with `AWS_REGION` already set in their shell do not need to duplicate it in `~/.config/ccpr/config.yaml`

---

### FR-18 MCP Server

- Provide a separate MCP server binary at `cmd/ccpr-mcp`
- The MCP server must not change the existing `ccpr` CLI behavior or JSON output
- Transport: stdio
- Exposed tools:
  - `ccpr_list` — list pull requests (read-only)
  - `ccpr_review` — fetch a PR's metadata, comments, and diff (read-only)
  - `ccpr_comment` — post a comment to a pull request (write-side)
  - `ccpr_create` — create a pull request (write-side)
- `ccpr_list` input mirrors `ccpr list` flags where applicable:
  - `repo` (required)
  - `status` (optional, default `open`; accepted values: `open`, `closed`, `all`)
  - `config` (optional)
  - `profile` (optional)
  - `region` (optional)
- `ccpr_list` output must reuse the existing `ccpr list --format json` PR summary schema (wrapped under `pullRequests` because MCP tool output schemas must be objects)
- `ccpr_review` input mirrors `ccpr review` parameters:
  - `url` (optional) — full CodeCommit PR URL; takes priority when present
  - `repo` (optional) — required when `url` is not provided
  - `prId` (optional) — required when `url` is not provided
  - `region` (optional)
  - `profile` (optional)
  - `config` (optional)
  - At least one of (`url`) or (`repo` and `prId`) must be provided
- `ccpr_review` output must reuse the existing `ccpr review --format json` schema (`{metadata, comments, diff}`); no wrapper is needed because the payload is already an object
- `ccpr_comment` input mirrors `ccpr comment` parameters:
  - `url` (optional) — full CodeCommit PR URL; takes priority when present
  - `repo` (optional) — required when `url` is not provided
  - `prId` (optional) — required when `url` is not provided
  - `body` (required) — comment body
  - `region` (optional)
  - `profile` (optional)
  - `config` (optional)
  - At least one of (`url`) or (`repo` and `prId`) must be provided
- `ccpr_comment` output must reuse the existing `ccpr comment --format json` schema (`{commentId, pullRequestId, authorArn, creationDate}`); no wrapper is needed because the payload is already an object
- `ccpr_comment` is write-side: each successful call posts a real comment to CodeCommit. The tool description and README must call this out so MCP hosts can prompt users before invocation
- `ccpr_create` input mirrors `ccpr create` parameters with one intentional difference:
  - `repo` (required)
  - `title` (required)
  - `sourceBranch` (required) — MCP does not auto-detect a "current" Git branch. CLI behavior of falling back to the local repo's current branch when `--source` is empty is **not** carried over
  - `destinationBranch` (required)
  - `description` (optional)
  - `region` (optional)
  - `profile` (optional)
  - `config` (optional)
- `ccpr_create` output must reuse the existing `ccpr create --format json` schema (`{prId, title, repository, sourceBranch, destinationBranch, url}`); no wrapper is needed because the payload is already an object
- `ccpr_create` is write-side: each successful call creates a real PR in CodeCommit. The tool description and README must call this out so MCP hosts can prompt users before invocation
- `ccpr_create` rejects requests where `sourceBranch == destinationBranch` (matches the CLI guard)
- The MCP server exposes only the structured (JSON-equivalent) output. `summary`, `patch`, `--body-file`, `--description-file`, and stdin (`-`) input forms remain CLI-only
- AWS profile resolution follows FR-09
- AWS region resolution follows FR-17, with the following tool-specific exception: when `ccpr_review` or `ccpr_comment` is invoked with a `url`, the region embedded in the URL takes priority over `region`/config/env (matching the corresponding CLI behavior). Otherwise FR-17 applies as-is
- The implementation must share the internal use cases with the CLI so the CLI and MCP paths do not diverge

Constraints:
- No new authentication or permission model — write-side MCP tools rely on the MCP host's per-call approval gate plus the same AWS profile resolution as the CLI
- MCP versions of `open` and `doctor` are out of scope. `ccpr_open` is intentionally not provided — MCP hosts do not own a browser, and PR URLs are already returned by `ccpr_list` / `ccpr_review` / `ccpr_create`. `ccpr_doctor` may be revisited later if there is demand

---

## Non-Functional Requirements

### NFR-01 CLI Behavior

- Single binary
- Exit codes:
  - 0 = success
  - 1 = user error
  - 2 = system error

- Errors printed to stderr

---

### NFR-02 Composability

- Output must be machine-readable
- Support piping

---

### NFR-03 Error Format (JSON mode)

Planned for a future release. Currently, errors are printed as plain text to stderr regardless of output format.

Target format:

{
  "error": {
    "code": "ERROR_CODE",
    "message": "description"
  }
}
