package app

import (
	"context"
	"fmt"

	"github.com/hidetzu/ccpr/internal/codecommit"
	"github.com/hidetzu/ccpr/internal/config"
	"github.com/hidetzu/ccpr/internal/diff"
	"github.com/hidetzu/ccpr/internal/output"
	"github.com/hidetzu/ccpr/internal/parser"
)

// DiffGeneratorFactory creates a diff generator for the resolved repo path.
type DiffGeneratorFactory func() diff.Generator

// GetReviewOptions contains inputs shared by the CLI and MCP review paths.
type GetReviewOptions struct {
	URL     string
	Repo    string
	PRId    string
	Region  string
	Profile string
	Config  string
}

// ReviewPayload is the structured review payload returned to both the CLI
// (for JSON formatting) and the MCP tool. It aliases output.ReviewOutput so
// the JSON shape remains consistent across surfaces.
type ReviewPayload = output.ReviewOutput

// GetReview fetches PR metadata, comments, and the local-Git-generated diff
// for a single pull request and returns it as a ReviewPayload.
func GetReview(
	ctx context.Context,
	opts GetReviewOptions,
	newClient CodeCommitClientFactory,
	newDiff DiffGeneratorFactory,
) (ReviewPayload, error) {
	var region, repo, prID string

	if opts.URL != "" {
		result, err := parser.Parse(opts.URL)
		if err != nil {
			return ReviewPayload{}, fmt.Errorf("invalid PR URL: %w", err)
		}
		region = result.Region
		repo = result.Repository
		prID = result.PRId
	} else if opts.Repo != "" && opts.PRId != "" {
		repo = opts.Repo
		prID = opts.PRId
	} else {
		return ReviewPayload{}, fmt.Errorf("provide a PR URL or repo and prId")
	}

	cfg, _, err := config.Load(opts.Config)
	if err != nil {
		return ReviewPayload{}, fmt.Errorf("config: %w", err)
	}

	if region == "" {
		region = cfg.ResolveRegion(opts.Region)
	}
	if region == "" {
		return ReviewPayload{}, fmt.Errorf("region is required: use region option, set region in config file, or set AWS_REGION/AWS_DEFAULT_REGION env")
	}

	repoPath, err := cfg.ResolveRepoPath(repo)
	if err != nil {
		return ReviewPayload{}, err
	}

	profile := cfg.ResolveProfile(opts.Profile)

	cc, err := newClient(ctx, region, profile)
	if err != nil {
		return ReviewPayload{}, fmt.Errorf("creating CodeCommit client: %w", err)
	}

	metadata, err := cc.GetPRMetadata(ctx, repo, prID)
	if err != nil {
		return ReviewPayload{}, fmt.Errorf("fetching PR metadata: %w", err)
	}

	comments, err := cc.GetPRComments(ctx, repo, prID, metadata.DestinationCommit, metadata.SourceCommit)
	if err != nil {
		return ReviewPayload{}, fmt.Errorf("fetching PR comments: %w", err)
	}

	gen := newDiff()
	diffText, err := gen.GenerateDiff(repoPath, metadata.SourceBranch, metadata.DestinationBranch)
	if err != nil {
		return ReviewPayload{}, fmt.Errorf("generating diff: %w", err)
	}

	return ReviewPayload{
		Metadata: output.PRMetadata{
			PRId:              prID,
			Title:             metadata.Title,
			Description:       metadata.Description,
			Author:            output.ShortAuthor(metadata.AuthorARN),
			AuthorARN:         metadata.AuthorARN,
			SourceBranch:      metadata.SourceBranch,
			DestinationBranch: metadata.DestinationBranch,
			Status:            metadata.Status,
			CreationDate:      metadata.CreationDate.Format("2006-01-02T15:04:05Z07:00"),
		},
		Comments: convertComments(comments),
		Diff:     diffText,
	}, nil
}

func convertComments(src []codecommit.Comment) []output.Comment {
	out := make([]output.Comment, len(src))
	for i, c := range src {
		out[i] = output.Comment{
			CommentId: c.CommentId,
			InReplyTo: c.InReplyTo,
			Author:    output.ShortAuthor(c.Author),
			AuthorARN: c.Author,
			Content:   c.Content,
			Timestamp: c.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			FilePath:  c.FilePath,
		}
	}
	return out
}
