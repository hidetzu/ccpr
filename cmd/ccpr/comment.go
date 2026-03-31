package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hidetzu/ccpr/internal/config"
	"github.com/hidetzu/ccpr/internal/output"
	"github.com/hidetzu/ccpr/internal/parser"
)

func runComment(args []string) error {
	fs := flag.NewFlagSet("comment", flag.ContinueOnError)

	var (
		flagBody     string
		flagBodyFile string
		flagFormat   string
		flagRepo     string
		flagRegion   string
		flagPRId     string
		flagConfig   string
		flagProfile  string
	)

	fs.StringVar(&flagBody, "body", "", "Comment body (use - to read from stdin)")
	fs.StringVar(&flagBodyFile, "body-file", "", "Path to file containing comment body")
	fs.StringVar(&flagFormat, "format", "summary", "Output format: summary, json")
	fs.StringVar(&flagRepo, "repo", "", "Repository name")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.StringVar(&flagPRId, "pr-id", "", "Pull request ID")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")

	reordered := reorderArgs(args)
	if err := fs.Parse(reordered); err != nil {
		return err
	}

	// Validate format
	switch flagFormat {
	case "summary", "json":
	default:
		return fmt.Errorf("invalid format %q: must be summary or json", flagFormat)
	}

	// Resolve body
	body, err := resolveBody(flagBody, flagBodyFile)
	if err != nil {
		return err
	}

	// Resolve PR parameters
	var region, repo, prID string

	if url := fs.Arg(0); url != "" {
		result, err := parser.Parse(url)
		if err != nil {
			return fmt.Errorf("invalid PR URL: %w", err)
		}
		region = result.Region
		repo = result.Repository
		prID = result.PRId
	} else if flagRepo != "" && flagPRId != "" {
		repo = flagRepo
		prID = flagPRId
	} else {
		return fmt.Errorf("provide a PR URL or --repo and --pr-id flags")
	}

	// Load config
	cfg, _, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if region == "" {
		region = cfg.ResolveRegion(flagRegion)
	}
	profile := cfg.ResolveProfile(flagProfile)

	// Get PR metadata for commit IDs
	ctx := context.Background()
	cc, err := newCodeCommitClient(ctx, region, profile)
	if err != nil {
		return newSystemError("creating CodeCommit client: %w", err)
	}

	metadata, err := cc.GetPRMetadata(ctx, repo, prID)
	if err != nil {
		return newSystemError("fetching PR metadata: %w", err)
	}

	// Post comment
	result, err := cc.PostComment(ctx, repo, prID, metadata.DestinationCommit, metadata.SourceCommit, body)
	if err != nil {
		return newSystemError("posting comment: %w", err)
	}

	// Output
	creationDate := result.CreationDate.Format("2006-01-02T15:04:05Z07:00")
	if flagFormat == "json" {
		return printCommentJSON(os.Stdout, prID, result.CommentID, result.AuthorARN, creationDate)
	}
	return printCommentSummary(os.Stdout, prID, result.CommentID, output.ShortAuthor(result.AuthorARN), creationDate)
}

func resolveBody(flagBody, flagBodyFile string) (string, error) {
	if flagBody != "" && flagBodyFile != "" {
		return "", fmt.Errorf("--body and --body-file are mutually exclusive")
	}

	if flagBody == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	if flagBody != "" {
		return flagBody, nil
	}

	if flagBodyFile != "" {
		data, err := os.ReadFile(flagBodyFile)
		if err != nil {
			return "", fmt.Errorf("reading body file %s: %w", flagBodyFile, err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("provide comment body via --body or --body-file")
}

type commentJSONOutput struct {
	CommentID     string `json:"commentId"`
	PullRequestID string `json:"pullRequestId"`
	AuthorARN     string `json:"authorArn"`
	CreationDate  string `json:"creationDate"`
}

func printCommentJSON(w io.Writer, prID string, commentID, authorARN string, creationDate string) error {
	out := commentJSONOutput{
		CommentID:     commentID,
		PullRequestID: prID,
		AuthorARN:     authorARN,
		CreationDate:  creationDate,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printCommentSummary(w io.Writer, prID, commentID, author, creationDate string) error {
	_, err := fmt.Fprintf(w, "Comment posted successfully.\nComment ID: %s\nPR: #%s\nAuthor: %s\nCreated: %s\n", commentID, prID, author, creationDate)
	return err
}
