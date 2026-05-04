package app

import (
	"context"
	"fmt"

	"github.com/hidetzu/ccpr/internal/codecommit"
	"github.com/hidetzu/ccpr/internal/config"
)

// CodeCommitClientFactory creates a CodeCommit client for resolved AWS settings.
type CodeCommitClientFactory func(ctx context.Context, region, profile string) (codecommit.Client, error)

// ListPullRequestsOptions contains inputs shared by the CLI and MCP list paths.
type ListPullRequestsOptions struct {
	Repo    string
	Status  string
	Config  string
	Profile string
	Region  string
}

// ListPullRequest is the stable machine-readable PR summary schema.
type ListPullRequest struct {
	PRId              string `json:"prId"`
	Title             string `json:"title"`
	AuthorARN         string `json:"authorArn"`
	SourceBranch      string `json:"sourceBranch"`
	DestinationBranch string `json:"destinationBranch"`
	Status            string `json:"status"`
	CreationDate      string `json:"creationDate"`
}

// ListPullRequests returns pull request summaries for a repository.
func ListPullRequests(ctx context.Context, opts ListPullRequestsOptions, newClient CodeCommitClientFactory) ([]ListPullRequest, error) {
	if opts.Repo == "" {
		return nil, fmt.Errorf("--repo is required for list command")
	}

	status := opts.Status
	if status == "" {
		status = "open"
	}
	switch status {
	case "open", "closed", "all":
	default:
		return nil, fmt.Errorf("invalid status %q: must be open, closed, or all", status)
	}

	cfg, _, err := config.Load(opts.Config)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	profile := cfg.ResolveProfile(opts.Profile)
	region := cfg.ResolveRegion(opts.Region)
	if region == "" {
		return nil, fmt.Errorf("region is required: use --region flag, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env")
	}

	cc, err := newClient(ctx, region, profile)
	if err != nil {
		return nil, fmt.Errorf("creating CodeCommit client: %w", err)
	}

	prs, err := cc.ListPRs(ctx, opts.Repo, status)
	if err != nil {
		return nil, fmt.Errorf("listing PRs: %w", err)
	}

	items := make([]ListPullRequest, len(prs))
	for i, pr := range prs {
		items[i] = ListPullRequest{
			PRId:              pr.PRId,
			Title:             pr.Title,
			AuthorARN:         pr.AuthorARN,
			SourceBranch:      pr.SourceBranch,
			DestinationBranch: pr.DestinationBranch,
			Status:            pr.Status,
			CreationDate:      pr.CreationDate.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return items, nil
}
