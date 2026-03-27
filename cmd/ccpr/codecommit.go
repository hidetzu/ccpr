package main

import (
	"context"

	"github.com/hidetzu/ccpr/internal/codecommit"
	"github.com/hidetzu/ccpr/internal/output"
)

// newCodeCommitClient creates a CodeCommit client for the given region and profile.
func newCodeCommitClient(ctx context.Context, region, profile string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region, profile)
}

func convertComments(src []codecommit.Comment) []output.Comment {
	out := make([]output.Comment, len(src))
	for i, c := range src {
		out[i] = output.Comment{
			Author:    output.ShortAuthor(c.Author),
			AuthorARN: c.Author,
			Content:   c.Content,
			Timestamp: c.Timestamp,
			FilePath:  c.FilePath,
		}
	}
	return out
}
