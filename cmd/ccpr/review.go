package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hidetzu/ccpr/internal/app"
	"github.com/hidetzu/ccpr/internal/diff"
	"github.com/hidetzu/ccpr/internal/output"
)

func runReview(args []string) error {
	fs := flag.NewFlagSet("review", flag.ContinueOnError)

	var (
		flagFormat  string
		flagRepo    string
		flagRegion  string
		flagPRId    string
		flagConfig  string
		flagProfile string
	)

	fs.StringVar(&flagFormat, "format", "summary", "Output format: summary, json, patch")
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
	case "summary", "json", "patch":
	default:
		return fmt.Errorf("invalid format %q: must be summary, json, or patch", flagFormat)
	}

	review, err := app.GetReview(context.Background(), app.GetReviewOptions{
		URL:     fs.Arg(0),
		Repo:    flagRepo,
		PRId:    flagPRId,
		Region:  flagRegion,
		Profile: flagProfile,
		Config:  flagConfig,
	}, newCodeCommitClient, defaultDiffGenerator)
	if err != nil {
		return err
	}

	switch flagFormat {
	case "patch":
		return output.FormatPatch(os.Stdout, review.Diff)
	case "json":
		return output.FormatJSON(os.Stdout, review)
	default:
		return output.FormatSummary(os.Stdout, review)
	}
}

func defaultDiffGenerator() diff.Generator {
	return &diff.GitGenerator{}
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
			if !containsEquals(a) && i+1 < len(args) && (len(args[i+1]) == 0 || args[i+1][0] != '-') {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, a)
		}
	}
	return append(flags, positional...)
}

func containsEquals(s string) bool {
	for _, c := range s {
		if c == '=' {
			return true
		}
	}
	return false
}
