package main

import (
	"context"

	"github.com/kawashima/ccpr/internal/codecommit"
	"github.com/kawashima/ccpr/internal/output"
)

// newCodeCommitClient creates a CodeCommit client for the given region.
func newCodeCommitClient(ctx context.Context, region string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region)
}

func convertComments(src []codecommit.Comment) []output.Comment {
	out := make([]output.Comment, len(src))
	for i, c := range src {
		out[i] = output.Comment{
			Author:    c.Author,
			Content:   c.Content,
			Timestamp: c.Timestamp,
			FilePath:  c.FilePath,
		}
	}
	return out
}
