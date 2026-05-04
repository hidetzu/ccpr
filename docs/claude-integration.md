# Using ccpr with Claude Code

This guide explains how to install and set up ccpr for AI-assisted CodeCommit PR reviews with Claude Code.

## Install ccpr

```bash
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
```

Requires [Go](https://go.dev/dl/) 1.25 or later. Verify the installation:

```bash
ccpr --version
```

## Initial setup

### 1. Create config file

```bash
ccpr init --profile <your-aws-profile> --region <your-region>
```

This creates `~/.config/ccpr/config.yaml`. If you don't specify flags, ccpr auto-detects from `AWS_PROFILE` and AWS CLI config.

### 2. Add repository mappings

Edit `~/.config/ccpr/config.yaml` and map your CodeCommit repository names to local clone paths:

```yaml
profile: your-aws-profile
region: ap-northeast-1
repoMappings:
  my-repo: /path/to/local/clone
```

ccpr uses local Git to generate diffs, so each repository must be cloned locally.

### 3. Validate setup

```bash
ccpr doctor
```

All checks should pass. Fix any issues before proceeding.

## Review a PR

```bash
ccpr review <codecommit-pr-url> --format json
```

This outputs structured JSON containing PR metadata, comments, and diff. Use this output to review the changes.

## Post a comment

```bash
ccpr comment <codecommit-pr-url> --body "Review comment text"
ccpr comment <codecommit-pr-url> --body-file review.md
```

## Create a PR

```bash
ccpr create --repo <repo> --title "PR title" --dest main
```

## Add to your project's CLAUDE.md

Add the following to the `CLAUDE.md` in your CodeCommit-based project so Claude Code knows how to review PRs:

```markdown
## Code Review

This project uses AWS CodeCommit. To review a pull request:

1. Run `ccpr review <codecommit-pr-url> --format json` to fetch PR data
2. Use the JSON output to review the changes

Do not use `gh` for this repository — it is hosted on AWS CodeCommit, not GitHub.
```

## Using the skill (recommended)

ccpr provides a sample Claude Code skill that wraps the review workflow into a single slash command.

### Install the skill

```bash
# Project-scoped (for a specific project)
mkdir -p .claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md .claude/skills/ccpr-review/

# Global (available in all projects)
mkdir -p ~/.claude/skills/ccpr-review
cp /path/to/ccpr/examples/claude/ccpr-review/SKILL.md ~/.claude/skills/ccpr-review/
```

### Use the skill

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
