package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kawashima/ccpr/internal/config"
	"github.com/kawashima/ccpr/internal/diff"
	"github.com/kawashima/ccpr/internal/output"
	"github.com/kawashima/ccpr/internal/parser"
)

func runReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)

	var (
		flagJSON   bool
		flagPatch  bool
		flagRepo   string
		flagRegion string
		flagPRId   string
		flagConfig string
	)

	fs.BoolVar(&flagJSON, "json", false, "Output as JSON (default)")
	fs.BoolVar(&flagPatch, "patch", false, "Output diff only in unified patch format")
	fs.StringVar(&flagRepo, "repo", "", "Repository name")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.StringVar(&flagPRId, "pr-id", "", "Pull request ID")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Flag exclusivity check
	if flagJSON && flagPatch {
		return fmt.Errorf("--json and --patch are mutually exclusive")
	}

	// Resolve PR parameters: URL or explicit flags
	var region, repo, prID string

	if url := fs.Arg(0); url != "" {
		result, err := parser.Parse(url)
		if err != nil {
			return fmt.Errorf("invalid PR URL: %w", err)
		}
		region = result.Region
		repo = result.Repository
		prID = result.PRId
	} else if flagRepo != "" && flagRegion != "" && flagPRId != "" {
		region = flagRegion
		repo = flagRepo
		prID = flagPRId
	} else {
		return fmt.Errorf("provide a PR URL or --repo, --region, and --pr-id flags")
	}

	// Load config for repo mapping
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	repoPath, err := cfg.ResolveRepoPath(repo)
	if err != nil {
		return err
	}

	// Fetch PR metadata and comments
	cc := newCodeCommitClient(region)
	ctx := context.Background()

	metadata, err := cc.GetPRMetadata(ctx, repo, prID)
	if err != nil {
		return fmt.Errorf("fetching PR metadata: %w", err)
	}

	comments, err := cc.GetPRComments(ctx, repo, prID)
	if err != nil {
		return fmt.Errorf("fetching PR comments: %w", err)
	}

	// Generate diff
	gen := &diff.GitGenerator{}
	diffText, err := gen.GenerateDiff(repoPath, metadata.SourceBranch, metadata.DestinationBranch)
	if err != nil {
		return fmt.Errorf("generating diff: %w", err)
	}

	// Output
	if flagPatch {
		f := &output.PatchFormatter{}
		return f.FormatPatch(os.Stdout, diffText)
	}

	review := output.ReviewOutput{
		Metadata: output.PRMetadata{
			Title:             metadata.Title,
			Description:       metadata.Description,
			AuthorARN:         metadata.AuthorARN,
			SourceBranch:      metadata.SourceBranch,
			DestinationBranch: metadata.DestinationBranch,
			Status:            metadata.Status,
			CreationDate:      metadata.CreationDate.Format("2006-01-02T15:04:05Z"),
		},
		Comments: convertComments(comments),
		Diff:     diffText,
	}

	f := &output.JSONFormatter{}
	return f.FormatJSON(os.Stdout, review)
}
