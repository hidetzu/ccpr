package output

import (
	"io"
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
	Title             string `json:"title"`
	Description       string `json:"description"`
	AuthorARN         string `json:"authorArn"`
	SourceBranch      string `json:"sourceBranch"`
	DestinationBranch string `json:"destinationBranch"`
	Status            string `json:"status"`
	CreationDate      string `json:"creationDate"`
}

// Comment is the JSON-serializable representation of a PR comment.
type Comment struct {
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"filePath,omitempty"`
}

// Formatter defines the interface for writing review output.
type Formatter interface {
	// FormatJSON serializes ReviewOutput as JSON and writes to w.
	FormatJSON(w io.Writer, output ReviewOutput) error

	// FormatPatch writes the raw unified diff to w.
	FormatPatch(w io.Writer, diff string) error
}
