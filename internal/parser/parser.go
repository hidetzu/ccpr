package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParseResult holds the extracted components from a CodeCommit PR URL.
type ParseResult struct {
	Region     string
	Repository string
	PRId       string
}

var hostPattern = regexp.MustCompile(`^(.+)\.console\.aws\.amazon\.com$`)

// Parse extracts region, repository, and PR ID from a CodeCommit PR URL.
// Expected format:
//
//	https://<region>.console.aws.amazon.com/codesuite/codecommit/repositories/<repo>/pull-requests/<pr-id>
func Parse(rawURL string) (ParseResult, error) {
	if rawURL == "" {
		return ParseResult{}, fmt.Errorf("empty URL")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return ParseResult{}, fmt.Errorf("invalid URL: %w", err)
	}

	matches := hostPattern.FindStringSubmatch(u.Hostname())
	if matches == nil {
		return ParseResult{}, fmt.Errorf("not a CodeCommit console URL: %s", u.Hostname())
	}
	region := matches[1]

	// Expected path: /codesuite/codecommit/repositories/<repo>/pull-requests/<pr-id>
	segments := splitPath(u.Path)
	repoIdx, prIdx := -1, -1
	for i, s := range segments {
		switch s {
		case "repositories":
			repoIdx = i
		case "pull-requests":
			prIdx = i
		}
	}

	if repoIdx < 0 || repoIdx+1 >= len(segments) {
		return ParseResult{}, fmt.Errorf("missing repository name in URL path")
	}
	repo := segments[repoIdx+1]
	if repo == "" || repo == "pull-requests" {
		return ParseResult{}, fmt.Errorf("missing repository name in URL path")
	}

	if prIdx < 0 || prIdx+1 >= len(segments) {
		return ParseResult{}, fmt.Errorf("missing pull request ID in URL path")
	}
	prID := segments[prIdx+1]
	if prID == "" {
		return ParseResult{}, fmt.Errorf("missing pull request ID in URL path")
	}

	if !isNumeric(prID) {
		return ParseResult{}, fmt.Errorf("pull request ID must be numeric, got %q", prID)
	}

	return ParseResult{
		Region:     region,
		Repository: repo,
		PRId:       prID,
	}, nil
}

func splitPath(path string) []string {
	var segments []string
	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segments = append(segments, s)
		}
	}
	return segments
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
