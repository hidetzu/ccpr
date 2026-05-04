package app

import (
	"context"
	"fmt"

	"github.com/hidetzu/ccpr/internal/config"
	"github.com/hidetzu/ccpr/internal/parser"
)

// PostCommentOptions contains inputs shared by the CLI and MCP comment paths.
type PostCommentOptions struct {
	URL     string
	Repo    string
	PRId    string
	Body    string
	Region  string
	Profile string
	Config  string
}

// PostedComment is the structured result of posting a comment, returned to
// both the CLI (for JSON formatting) and the MCP tool. Field names match the
// existing `ccpr comment --format json` schema.
type PostedComment struct {
	CommentID     string `json:"commentId"`
	PullRequestID string `json:"pullRequestId"`
	AuthorARN     string `json:"authorArn"`
	CreationDate  string `json:"creationDate"`
}

// PostComment posts a comment to a CodeCommit pull request and returns the
// resulting metadata. AWS-side failures are wrapped in *SystemError so callers
// can distinguish them from user input errors.
func PostComment(
	ctx context.Context,
	opts PostCommentOptions,
	newClient CodeCommitClientFactory,
) (PostedComment, error) {
	if opts.Body == "" {
		return PostedComment{}, fmt.Errorf("comment body is required")
	}

	var region, repo, prID string
	if opts.URL != "" {
		result, err := parser.Parse(opts.URL)
		if err != nil {
			return PostedComment{}, fmt.Errorf("invalid PR URL: %w", err)
		}
		region = result.Region
		repo = result.Repository
		prID = result.PRId
	} else if opts.Repo != "" && opts.PRId != "" {
		repo = opts.Repo
		prID = opts.PRId
	} else {
		return PostedComment{}, fmt.Errorf("provide a PR URL or repo and prId")
	}

	cfg, _, err := config.Load(opts.Config)
	if err != nil {
		return PostedComment{}, fmt.Errorf("config: %w", err)
	}

	if region == "" {
		region = cfg.ResolveRegion(opts.Region)
	}
	if region == "" {
		return PostedComment{}, fmt.Errorf("region is required: use region option, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env")
	}

	profile := cfg.ResolveProfile(opts.Profile)

	cc, err := newClient(ctx, region, profile)
	if err != nil {
		return PostedComment{}, newSystemError("creating CodeCommit client: %w", err)
	}

	metadata, err := cc.GetPRMetadata(ctx, repo, prID)
	if err != nil {
		return PostedComment{}, newSystemError("fetching PR metadata: %w", err)
	}

	result, err := cc.PostComment(ctx, repo, prID, metadata.DestinationCommit, metadata.SourceCommit, opts.Body)
	if err != nil {
		return PostedComment{}, newSystemError("posting comment: %w", err)
	}

	return PostedComment{
		CommentID:     result.CommentID,
		PullRequestID: prID,
		AuthorARN:     result.AuthorARN,
		CreationDate:  result.CreationDate.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
