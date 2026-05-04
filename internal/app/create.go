package app

import (
	"context"
	"fmt"

	"github.com/hidetzu/ccpr/internal/config"
)

// CreatePullRequestOptions contains inputs shared by the CLI and MCP create paths.
// SourceBranch is required: the CLI's "default to local Git current branch when
// empty" behavior is resolved on the CLI side before calling this use case, so
// MCP callers must pass an explicit source branch.
type CreatePullRequestOptions struct {
	Repo              string
	Title             string
	SourceBranch      string
	DestinationBranch string
	Description       string
	Region            string
	Profile           string
	Config            string
}

// CreatedPullRequest is the structured result of creating a pull request,
// returned to both the CLI (for JSON formatting) and the MCP tool. Field names
// match the existing `ccpr create --format json` schema.
type CreatedPullRequest struct {
	PRId              string `json:"prId"`
	Title             string `json:"title"`
	Repository        string `json:"repository"`
	SourceBranch      string `json:"sourceBranch"`
	DestinationBranch string `json:"destinationBranch"`
	URL               string `json:"url"`
}

// CreatePullRequest creates a new CodeCommit pull request and returns the
// resulting metadata. AWS-side failures are wrapped in *SystemError so callers
// can distinguish them from user input errors.
func CreatePullRequest(
	ctx context.Context,
	opts CreatePullRequestOptions,
	newClient CodeCommitClientFactory,
) (CreatedPullRequest, error) {
	if opts.Repo == "" {
		return CreatedPullRequest{}, fmt.Errorf("repo is required")
	}
	if opts.Title == "" {
		return CreatedPullRequest{}, fmt.Errorf("title is required")
	}
	if opts.SourceBranch == "" {
		return CreatedPullRequest{}, fmt.Errorf("sourceBranch is required")
	}
	if opts.DestinationBranch == "" {
		return CreatedPullRequest{}, fmt.Errorf("destinationBranch is required")
	}
	if opts.SourceBranch == opts.DestinationBranch {
		return CreatedPullRequest{}, fmt.Errorf("source branch %q is the same as destination branch", opts.SourceBranch)
	}

	cfg, _, err := config.Load(opts.Config)
	if err != nil {
		return CreatedPullRequest{}, fmt.Errorf("config: %w", err)
	}

	region := cfg.ResolveRegion(opts.Region)
	if region == "" {
		return CreatedPullRequest{}, fmt.Errorf("region is required: use region option, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env")
	}
	profile := cfg.ResolveProfile(opts.Profile)

	cc, err := newClient(ctx, region, profile)
	if err != nil {
		return CreatedPullRequest{}, newSystemError("creating CodeCommit client: %w", err)
	}

	result, err := cc.CreatePR(ctx, opts.Repo, opts.Title, opts.Description, opts.SourceBranch, opts.DestinationBranch)
	if err != nil {
		return CreatedPullRequest{}, newSystemError("creating pull request: %w", err)
	}

	return CreatedPullRequest{
		PRId:              result.PRId,
		Title:             result.Title,
		Repository:        opts.Repo,
		SourceBranch:      result.SourceBranch,
		DestinationBranch: result.DestinationBranch,
		URL:               BuildConsoleURL(region, opts.Repo, result.PRId),
	}, nil
}

// BuildConsoleURL returns the CodeCommit web-console URL for a pull request.
func BuildConsoleURL(region, repo, prID string) string {
	return fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s/pull-requests/%s", region, repo, prID)
}
