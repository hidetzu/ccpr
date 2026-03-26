## UC-01 Review a PR with Claude Code

- Actor: Developer
- Trigger: A CodeCommit PR URL is available
- Main flow:
  1. Developer gives PR URL to Claude Code
  2. Claude runs `ccpr review <url> --json`
  3. Tool returns PR data and diff
  4. Claude generates review feedback
- Success:
  - Developer gets review suggestions
