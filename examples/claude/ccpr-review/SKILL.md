---
name: ccpr-review
description: Review a CodeCommit pull request using ccpr structured output
allowed-tools: Bash(ccpr *)
---

If `ccpr` is not installed, install it first:

```
go install github.com/hidetzu/ccpr/cmd/ccpr@latest
```

Run `ccpr review $ARGUMENTS --format json` and review the pull request based on the structured output.

Focus on:

- Correctness: logic errors, edge cases, missing validation
- Security concerns: credential handling, injection risks, access control
- Performance impact: inefficient queries, unnecessary allocations, N+1 patterns
- Readability and maintainability: naming, structure, complexity

Return your review in this format:

1. **Summary** — What the PR does in 2-3 sentences
2. **Risks** — Issues that should be addressed before merging
3. **Suggested changes** — Specific, actionable improvements with file/line references
