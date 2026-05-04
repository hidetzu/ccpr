package main

import (
	"context"

	"github.com/hidetzu/ccpr/internal/codecommit"
)

// newCodeCommitClient creates a CodeCommit client for the given region and profile.
func newCodeCommitClient(ctx context.Context, region, profile string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region, profile)
}
