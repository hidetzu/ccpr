package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/hidetzu/ccpr/internal/config"
	"github.com/hidetzu/ccpr/internal/output"
)

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)

	var (
		flagRepo    string
		flagStatus  string
		flagFormat  string
		flagConfig  string
		flagProfile string
		flagRegion  string
	)

	fs.StringVar(&flagRepo, "repo", "", "Repository name (required)")
	fs.StringVar(&flagStatus, "status", "open", "PR status filter: open, closed, all")
	fs.StringVar(&flagFormat, "format", "summary", "Output format: summary, json")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")
	fs.StringVar(&flagRegion, "region", "", "AWS region")

	if err := fs.Parse(args); err != nil {
		return err
	}

	switch flagFormat {
	case "summary", "json":
	default:
		return fmt.Errorf("invalid format %q: must be summary or json", flagFormat)
	}

	if flagRepo == "" {
		return fmt.Errorf("--repo is required for list command")
	}

	cfg, _, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	profile := cfg.ResolveProfile(flagProfile)

	region := cfg.ResolveRegion(flagRegion)
	if region == "" {
		return fmt.Errorf("region is required: use --region flag or set region in config file")
	}

	ctx := context.Background()
	cc, err := newCodeCommitClient(ctx, region, profile)
	if err != nil {
		return fmt.Errorf("creating CodeCommit client: %w", err)
	}

	prs, err := cc.ListPRs(ctx, flagRepo, flagStatus)
	if err != nil {
		return fmt.Errorf("listing PRs: %w", err)
	}

	if flagFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(prs)
	}

	if len(prs) == 0 {
		fmt.Println("No pull requests found.")
		return nil
	}

	for _, pr := range prs {
		author := output.ShortAuthor(pr.AuthorARN)
		fmt.Printf("#%-6s %-40s %-20s %s → %s  %-6s  %s\n",
			pr.PRId,
			truncate(pr.Title, 40),
			author,
			pr.SourceBranch,
			pr.DestinationBranch,
			pr.Status,
			pr.CreationDate.Format("2006-01-02"),
		)
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
