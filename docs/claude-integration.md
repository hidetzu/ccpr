# Using ccpr with Claude Code

This guide explains how to set up Claude Code to review AWS CodeCommit pull requests using `ccpr`.

## Prerequisites

- [ccpr](../README.md#install) installed and configured
- AWS credentials configured for your CodeCommit repositories
- Repository cloned locally and mapped in `~/.config/ccpr/config.yaml`

## Setup

Add the following to the `CLAUDE.md` in your CodeCommit-based project:

```markdown
## Code Review

This project uses AWS CodeCommit. To review a pull request:

1. Run `ccpr review <codecommit-pr-url> --json` to fetch PR data
2. Use the JSON output to review the changes

Example:
\```bash
ccpr review <codecommit-pr-url> --json
\```

Do not use `gh` for this repository — it is hosted on AWS CodeCommit, not GitHub.
```

## How it works

1. Developer shares a CodeCommit PR URL with Claude Code
2. Claude Code runs `ccpr review <url> --json`
3. ccpr returns structured JSON containing:
   - PR metadata (title, author, status, branches)
   - Comments
   - Unified diff
4. Claude Code uses the JSON to generate review feedback

## Example session

```
User: Review this PR
      https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/123

Claude: (runs ccpr review <url> --json, reads the output, provides review)
```

## Using the skill (recommended)

ccpr provides a sample Claude Code skill that wraps the review workflow into a single slash command.

### Prerequisites

- `ccpr` must be installed and available in your `PATH` (see [Install](../README.md#install))

### Install

Copy `examples/claude/ccpr-review/SKILL.md` from this repository to your project or personal skills directory:

```bash
# Project-scoped (for a specific project)
mkdir -p .claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md .claude/skills/ccpr-review/

# Global (available in all projects)
mkdir -p ~/.claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md ~/.claude/skills/ccpr-review/
```

### Usage

In Claude Code:

```
/ccpr-review <codecommit-pr-url>
```

Claude will run `ccpr review <url> --format json` and review the PR, focusing on correctness, security, performance, and readability.

### Customizing the review

To adjust the review focus or language, edit the copied `SKILL.md` file. The skill is plain Markdown — no code changes needed.

## Notes

- ccpr is a local helper CLI, not an official AWS or Anthropic integration
- GitHub repositories should continue using `gh` as usual
- Only CodeCommit repositories need `ccpr`
