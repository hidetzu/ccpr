package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ReviewOutput is the top-level structure for JSON output combining
// PR metadata, comments, and diff.
type ReviewOutput struct {
	Metadata PRMetadata `json:"metadata"`
	Comments []Comment  `json:"comments"`
	Diff     string     `json:"diff"`
}

// PRMetadata is the JSON-serializable representation of PR metadata.
type PRMetadata struct {
	PRId              string `json:"prId"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Author            string `json:"author"`
	AuthorARN         string `json:"authorArn"`
	SourceBranch      string `json:"sourceBranch"`
	DestinationBranch string `json:"destinationBranch"`
	Status            string `json:"status"`
	CreationDate      string `json:"creationDate"`
}

// Comment is the JSON-serializable representation of a PR comment.
type Comment struct {
	Author    string    `json:"author"`
	AuthorARN string    `json:"authorArn"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"filePath,omitempty"`
}

// ShortAuthor extracts the username from an ARN (last segment after /).
func ShortAuthor(arn string) string {
	if i := strings.LastIndex(arn, "/"); i >= 0 {
		return arn[i+1:]
	}
	return arn
}

// FormatJSON serializes ReviewOutput as JSON and writes to w.
func FormatJSON(w io.Writer, output ReviewOutput) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// FormatPatch writes the raw unified diff to w.
func FormatPatch(w io.Writer, diff string) error {
	_, err := io.WriteString(w, diff)
	return err
}

// FormatSummary writes a human-readable summary to w.
func FormatSummary(w io.Writer, output ReviewOutput) error {
	m := output.Metadata

	author := ShortAuthor(m.AuthorARN)

	// Count changed files from diff
	filesChanged := countChangedFiles(output.Diff)

	if _, err := fmt.Fprintf(w, "PR #%s: %s\n", m.PRId, m.Title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Author:   %s\n", author); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Status:   %s\n", m.Status); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Branch:   %s → %s\n", m.SourceBranch, m.DestinationBranch); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Created:  %s\n", formatDate(m.CreationDate)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nComments: %d\n", len(output.Comments)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Files:    %d changed\n", filesChanged); err != nil {
		return err
	}

	if m.Description != "" {
		if _, err := fmt.Fprintf(w, "\n## Description\n\n%s\n", m.Description); err != nil {
			return err
		}
	}

	return nil
}

func countChangedFiles(diff string) int {
	count := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			count++
		}
	}
	return count
}

func formatDate(raw string) string {
	t, err := time.Parse("2006-01-02T15:04:05Z", raw)
	if err != nil {
		return raw
	}
	return t.Format("2006-01-02 15:04")
}
