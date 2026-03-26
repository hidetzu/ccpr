package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hidetzu/ccpr/internal/config"
	"github.com/hidetzu/ccpr/internal/diff"
	"github.com/hidetzu/ccpr/internal/output"
	"github.com/hidetzu/ccpr/internal/parser"
)

func runReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)

	var (
		flagJSON   bool
		flagPatch  bool
		flagRepo   string
		flagRegion string
		flagPRId   string
		flagConfig  string
		flagProfile string
	)

	fs.BoolVar(&flagJSON, "json", false, "Output as JSON (default)")
	fs.BoolVar(&flagPatch, "patch", false, "Output diff only in unified patch format")
	fs.StringVar(&flagRepo, "repo", "", "Repository name")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.StringVar(&flagPRId, "pr-id", "", "Pull request ID")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")

	// Reorder args so flags come first, allowing "ccpr review <url> --json".
	// Go's flag package stops parsing at the first non-flag argument.
	reordered := reorderArgs(args)
	if err := fs.Parse(reordered); err != nil {
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

	// Resolve AWS profile
	profile := cfg.ResolveProfile(flagProfile)

	// Fetch PR metadata and comments
	ctx := context.Background()
	cc, err := newCodeCommitClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("creating CodeCommit client: %w", err)
	}

	metadata, err := cc.GetPRMetadata(ctx, repo, prID)
	if err != nil {
		return fmt.Errorf("fetching PR metadata: %w", err)
	}

	comments, err := cc.GetPRComments(ctx, repo, prID, metadata.DestinationCommit, metadata.SourceCommit)
	if err != nil {
		return fmt.Errorf("fetching PR comments: %w", err)
	}

	// Generate diff
	gen := &diff.GitGenerator{}
	diffText, err := gen.GenerateDiff(repoPath, metadata.SourceBranch, metadata.DestinationBranch)
	if err != nil {
		return fmt.Errorf("generating diff: %w", err)
	}

	// Build review output
	review := output.ReviewOutput{
		Metadata: output.PRMetadata{
			PRId:              prID,
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

	// Output
	switch {
	case flagPatch:
		return output.FormatPatch(os.Stdout, diffText)
	case flagJSON:
		return output.FormatJSON(os.Stdout, review)
	default:
		return output.FormatSummary(os.Stdout, review)
	}
}

// reorderArgs moves flags before positional args so Go's flag package
// can parse all flags regardless of where the URL is placed.
// Handles both "-flag value" and "-flag=value" forms.
func reorderArgs(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 0 && a[0] == '-' {
			flags = append(flags, a)
			// If this flag has a separate value (not bool, not =), consume next arg
			if !isBoolFlag(a) && !containsEquals(a) && i+1 < len(args) {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, a)
		}
	}
	return append(flags, positional...)
}

var boolFlags = map[string]bool{
	"-json": true, "--json": true,
	"-patch": true, "--patch": true,
}

func isBoolFlag(s string) bool {
	return boolFlags[s]
}

func containsEquals(s string) bool {
	for _, c := range s {
		if c == '=' {
			return true
		}
	}
	return false
}
