package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/hidetzu/ccpr/internal/config"
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

	fs.StringVar(&flagRepo, "repo", "", "Repository name (required)")
	fs.StringVar(&flagPRId, "pr-id", "", "Pull request ID (required)")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if flagRepo == "" {
		return fmt.Errorf("--repo is required for open command")
	}
	if flagPRId == "" {
		return fmt.Errorf("--pr-id is required for open command")
	}

	cfg, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	region := cfg.ResolveRegion(flagRegion)
	if region == "" {
		return fmt.Errorf("region is required: use --region flag or set region in config file")
	}

	url := fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codecommit/repositories/%s/pull-requests/%s",
		region, flagRepo, flagPRId)

	return openBrowser(url)
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
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
