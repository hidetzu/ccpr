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
	Status            string
	CreationDate      time.Time
}

// Comment represents a single comment on a pull request.
// FilePath is empty for PR-level comments.
type Comment struct {
	Author    string
	Content   string
	Timestamp time.Time
	FilePath  string
}

// Client defines the interface for interacting with AWS CodeCommit.
type Client interface {
	GetPRMetadata(ctx context.Context, repo, prID string) (PRMetadata, error)
	GetPRComments(ctx context.Context, repo, prID string) ([]Comment, error)
}
