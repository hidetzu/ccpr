package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/hidetzu/ccpr/internal/app"
	"github.com/hidetzu/ccpr/internal/output"
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

	switch flagFormat {
	case "summary", "json":
	default:
		return fmt.Errorf("invalid format %q: must be summary or json", flagFormat)
	}

	body, err := resolveBody(flagBody, flagBodyFile)
	if err != nil {
		return err
	}

	result, err := app.PostComment(context.Background(), app.PostCommentOptions{
		URL:     fs.Arg(0),
		Repo:    flagRepo,
		PRId:    flagPRId,
		Body:    body,
		Region:  flagRegion,
		Profile: flagProfile,
		Config:  flagConfig,
	}, newCodeCommitClient)
	if err != nil {
		return err
	}

	if flagFormat == "json" {
		return printCommentJSON(os.Stdout, result)
	}
	return printCommentSummary(os.Stdout, result.PullRequestID, result.CommentID, output.ShortAuthor(result.AuthorARN), result.CreationDate)
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

func printCommentJSON(w io.Writer, result app.PostedComment) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func printCommentSummary(w io.Writer, prID, commentID, author, creationDate string) error {
	_, err := fmt.Fprintf(w, "Comment posted successfully.\nComment ID: %s\nPR: #%s\nAuthor: %s\nCreated: %s\n", commentID, prID, author, creationDate)
	return err
}
