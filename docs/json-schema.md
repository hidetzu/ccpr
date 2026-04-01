# JSON Output Reference

This document defines the JSON output contract for ccpr v1. All field names, types, and required/optional semantics described here are guaranteed stable within v1.x releases.

See [versioning.md](versioning.md) for the full stability policy (planned, tracked in #42).

## Conventions

- All field names use **camelCase**
- All JSON output uses 2-space indentation
- Timestamps use ISO 8601 format: `2006-01-02T15:04:05Z07:00`
- Empty arrays are serialized as `[]`, not `null`
- Fields marked **optional** use `omitempty` and are omitted when empty
- Fields marked **required** are always present (may be empty string `""`)
- Consumers should ignore unknown fields (forward compatibility)

## Commands

### `review --format json`

Top-level object combining PR metadata, comments, and diff.

```json
{
  "metadata": { ... },
  "comments": [ ... ],
  "diff": "string"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `metadata` | object | yes | PR metadata (see below) |
| `comments` | array | yes | Comment list (empty array `[]` if none) |
| `diff` | string | yes | Unified diff output |

#### `metadata` object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prId` | string | yes | Pull request ID |
| `title` | string | yes | PR title |
| `description` | string | yes | PR description (empty string if none) |
| `author` | string | yes | Short author name (extracted from ARN) |
| `authorArn` | string | yes | Full IAM ARN of the author |
| `sourceBranch` | string | yes | Source branch name |
| `destinationBranch` | string | yes | Destination branch name |
| `status` | string | yes | PR status: `OPEN` or `CLOSED` |
| `creationDate` | string | yes | ISO 8601 timestamp |

#### `comments` array item

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `commentId` | string | yes | Comment ID |
| `inReplyTo` | string | optional | Parent comment ID (omitted if root comment) |
| `author` | string | yes | Short author name |
| `authorArn` | string | yes | Full IAM ARN |
| `content` | string | yes | Comment body |
| `timestamp` | string | yes | ISO 8601 timestamp |
| `filePath` | string | optional | File path (omitted for PR-level comments) |

---

### `list --format json`

Array of PR summary objects.

```json
[
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
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prId` | string | yes | Pull request ID |
| `title` | string | yes | PR title |
| `authorArn` | string | yes | Full IAM ARN of the author |
| `sourceBranch` | string | yes | Source branch name |
| `destinationBranch` | string | yes | Destination branch name |
| `status` | string | yes | PR status: `OPEN`, `CLOSED` |
| `creationDate` | string | yes | ISO 8601 timestamp |

Empty result: `[]`

---

### `comment --format json`

Result of posting a comment.

```json
{
  "commentId": "eb596ff8-5133-438f-88d6-7c94f693302b",
  "pullRequestId": "42",
  "authorArn": "arn:aws:iam::123456789012:user/example",
  "creationDate": "2026-04-01T10:00:00+09:00"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `commentId` | string | yes | Created comment ID |
| `pullRequestId` | string | yes | Target PR ID |
| `authorArn` | string | yes | Full IAM ARN of the commenter |
| `creationDate` | string | yes | ISO 8601 timestamp |

---

### `create --format json`

Result of creating a pull request.

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

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prId` | string | yes | Created PR ID |
| `title` | string | yes | PR title |
| `repository` | string | yes | Repository name |
| `sourceBranch` | string | yes | Source branch name |
| `destinationBranch` | string | yes | Destination branch name |
| `url` | string | yes | CodeCommit console URL |

---

## Error Output

Errors are printed to stderr as plain text. In a future release, `--format json` may produce structured error output:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "description"
  }
}
```

This is **not yet implemented** in v1. Error codes are defined in [spec.md](spec.md) for future use.
