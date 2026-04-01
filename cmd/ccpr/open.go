package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/hidetzu/ccpr/internal/config"
	"github.com/hidetzu/ccpr/internal/parser"
)

func runOpen(args []string) error {
	fs := flag.NewFlagSet("open", flag.ContinueOnError)

	var (
		flagRepo    string
		flagPRId    string
		flagRegion  string
		flagConfig  string
		flagProfile string
	)

	fs.StringVar(&flagRepo, "repo", "", "Repository name")
	fs.StringVar(&flagPRId, "pr-id", "", "Pull request ID")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")

	reordered := reorderArgs(args)
	if err := fs.Parse(reordered); err != nil {
		return err
	}

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

	cfg, _, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	if region == "" {
		region = cfg.ResolveRegion(flagRegion)
	}
	if region == "" {
		return fmt.Errorf("region is required: use --region flag or set region in config file")
	}

	openURL := buildConsoleURL(region, repo, prID)

	return openBrowser(openURL)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		fmt.Println(url)
		return nil
	}
	if err := cmd.Start(); err != nil {
		// Fallback: print URL if browser cannot be opened
		fmt.Println(url)
	}
	return nil
}
