package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/hidetzu/ccpr/internal/app"
	"github.com/hidetzu/ccpr/internal/config"
)

func runCreate(args []string) error {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)

	var (
		flagRepo            string
		flagTitle           string
		flagDest            string
		flagSource          string
		flagDescription     string
		flagDescriptionFile string
		flagFormat          string
		flagRegion          string
		flagProfile         string
		flagConfig          string
	)

	fs.StringVar(&flagRepo, "repo", "", "Repository name (required)")
	fs.StringVar(&flagTitle, "title", "", "Pull request title (required)")
	fs.StringVar(&flagDest, "dest", "", "Destination branch (required)")
	fs.StringVar(&flagSource, "source", "", "Source branch (defaults to current branch)")
	fs.StringVar(&flagDescription, "description", "", "PR description (use - to read from stdin)")
	fs.StringVar(&flagDescriptionFile, "description-file", "", "Path to file containing PR description")
	fs.StringVar(&flagFormat, "format", "summary", "Output format: summary, json")
	fs.StringVar(&flagRegion, "region", "", "AWS region")
	fs.StringVar(&flagProfile, "profile", "", "AWS profile name")
	fs.StringVar(&flagConfig, "config", "", "Path to configuration file")

	reordered := reorderArgs(args)
	if err := fs.Parse(reordered); err != nil {
		return err
	}

	if flagRepo == "" {
		return fmt.Errorf("--repo is required")
	}
	if flagTitle == "" {
		return fmt.Errorf("--title is required")
	}
	if flagDest == "" {
		return fmt.Errorf("--dest is required")
	}

	switch flagFormat {
	case "summary", "json":
	default:
		return fmt.Errorf("invalid format %q: must be summary or json", flagFormat)
	}

	description, err := resolveDescription(flagDescription, flagDescriptionFile)
	if err != nil {
		return err
	}

	source, err := resolveSourceBranch(flagSource, flagRepo, flagConfig)
	if err != nil {
		return err
	}

	result, err := app.CreatePullRequest(context.Background(), app.CreatePullRequestOptions{
		Repo:              flagRepo,
		Title:             flagTitle,
		SourceBranch:      source,
		DestinationBranch: flagDest,
		Description:       description,
		Region:            flagRegion,
		Profile:           flagProfile,
		Config:            flagConfig,
	}, newCodeCommitClient)
	if err != nil {
		return translateAppError(err)
	}

	if flagFormat == "json" {
		return printCreateJSON(os.Stdout, result.PRId, result.Title, result.Repository, result.SourceBranch, result.DestinationBranch, result.URL)
	}
	return printCreateSummary(os.Stdout, result.PRId, result.Title, result.Repository, result.SourceBranch, result.DestinationBranch, result.URL)
}

// resolveSourceBranch returns the explicit --source value if set, otherwise
// detects the current branch via local Git in the configured repo path.
// This default-from-Git behavior is CLI-only; the MCP path requires source
// to be passed explicitly.
func resolveSourceBranch(flagSource, repo, configPath string) (string, error) {
	if flagSource != "" {
		return flagSource, nil
	}
	cfg, _, err := config.Load(configPath)
	if err != nil {
		return "", fmt.Errorf("config: %w", err)
	}
	repoPath, err := cfg.ResolveRepoPath(repo)
	if err != nil {
		return "", fmt.Errorf("resolving repo path: %w", err)
	}
	source, err := currentBranch(repoPath)
	if err != nil {
		return "", newSystemError("detecting current branch: %w", err)
	}
	return source, nil
}

func resolveDescription(flagDesc, flagDescFile string) (string, error) {
	if flagDesc != "" && flagDescFile != "" {
		return "", fmt.Errorf("--description and --description-file are mutually exclusive")
	}

	if flagDesc == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}

	if flagDesc != "" {
		return flagDesc, nil
	}

	if flagDescFile != "" {
		data, err := os.ReadFile(flagDescFile)
		if err != nil {
			return "", fmt.Errorf("reading description file %s: %w", flagDescFile, err)
		}
		return string(data), nil
	}

	return "", nil
}

func currentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

type createJSONOutput struct {
	PRId              string `json:"prId"`
	Title             string `json:"title"`
	Repository        string `json:"repository"`
	SourceBranch      string `json:"sourceBranch"`
	DestinationBranch string `json:"destinationBranch"`
	URL               string `json:"url"`
}

func printCreateJSON(w io.Writer, prID, title, repo, source, dest, url string) error {
	out := createJSONOutput{
		PRId:              prID,
		Title:             title,
		Repository:        repo,
		SourceBranch:      source,
		DestinationBranch: dest,
		URL:               url,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printCreateSummary(w io.Writer, prID, title, repo, source, dest, url string) error {
	_, err := fmt.Fprintf(w, "✔ Pull request created\n  PR #%s: %s\n  Repository: %s\n  Source: %s → Destination: %s\n  URL: %s\n", prID, title, repo, source, dest, url)
	return err
}
