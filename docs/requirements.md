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

- Triggered by `--json` flag
- Combine metadata, comments, and diff
- Output as a single JSON document
- JSON schema must be stable
- Designed for AI tools and script integration

Constraints:
- Field names must not change without versioning
- Optional fields must be documented

---

### FR-07 Patch Output

- Triggered by `--patch` flag
- Output raw diff only

Constraints:
- `--json`, `--patch` are mutually exclusive

---

### FR-08 Local Repository Resolution

- Map repository name to local filesystem path
- Configuration file defines mapping

Example:

repoMappings:
  my-repo: /work/src/my-repo

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
- Support `--json` for machine-readable output
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

Errors must follow:

{
  "error": {
    "code": "ERROR_CODE",
    "message": "description"
  }
}
