package codecommit

import (
	"context"
	"time"
)

// PRMetadata contains the metadata of a CodeCommit pull request.
type PRMetadata struct {
	Title             string
	Description       string
	AuthorARN         string
	SourceBranch      string
	DestinationBranch string
	SourceCommit      string
	DestinationCommit string
	Status            string
	CreationDate      time.Time
}

// Comment represents a single comment on a pull request.
// FilePath is empty for PR-level comments.
type Comment struct {
	CommentId string
	InReplyTo string
	Author    string
	Content   string
	Timestamp time.Time
	FilePath  string
}

// PRSummary contains minimal info for listing pull requests.
type PRSummary struct {
	PRId              string
	Title             string
	AuthorARN         string
	SourceBranch      string
	DestinationBranch string
	Status            string
	CreationDate      time.Time
}

// PostCommentResult contains the result of posting a comment.
type PostCommentResult struct {
	CommentID    string
	AuthorARN    string
	CreationDate time.Time
}

// CreatePRResult contains the result of creating a pull request.
type CreatePRResult struct {
	PRId              string
	Title             string
	SourceBranch      string
	DestinationBranch string
}

// Client defines the interface for interacting with AWS CodeCommit.
type Client interface {
	GetPRMetadata(ctx context.Context, repo, prID string) (PRMetadata, error)
	GetPRComments(ctx context.Context, repo, prID, beforeCommit, afterCommit string) ([]Comment, error)
	ListPRs(ctx context.Context, repo, status string) ([]PRSummary, error)
	PostComment(ctx context.Context, repo, prID, beforeCommit, afterCommit, content string) (PostCommentResult, error)
	CreatePR(ctx context.Context, repo, title, description, sourceBranch, destBranch string) (CreatePRResult, error)
}
