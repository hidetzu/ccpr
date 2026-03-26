package main

import (
	"github.com/kawashima/ccpr/internal/codecommit"
	"github.com/kawashima/ccpr/internal/output"
)

// newCodeCommitClient creates a CodeCommit client for the given region.
// TODO: replace stub with real AWS SDK implementation.
func newCodeCommitClient(region string) codecommit.Client {
	_ = region
	panic("codecommit client not yet implemented")
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
