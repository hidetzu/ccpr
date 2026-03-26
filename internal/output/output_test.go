package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestFormatJSON(t *testing.T) {
	ts := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	out := ReviewOutput{
		Metadata: PRMetadata{
			Title:             "Fix login bug",
			Description:       "Fixes timeout on login",
			AuthorARN:         "arn:aws:iam::123:user/dev",
			SourceBranch:      "fix/login",
			DestinationBranch: "main",
			Status:            "OPEN",
			CreationDate:      "2026-01-15T10:30:00Z",
		},
		Comments: []Comment{
			{
				Author:    "reviewer",
				Content:   "Looks good",
				Timestamp: ts,
			},
			{
				Author:    "reviewer",
				Content:   "Check this file",
				Timestamp: ts,
				FilePath:  "src/login.go",
			},
		},
		Diff: "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new\n",
	}

	var buf bytes.Buffer
	f := &JSONFormatter{}
	if err := f.FormatJSON(&buf, out); err != nil {
		t.Fatalf("FormatJSON() error: %v", err)
	}

	// Verify it's valid JSON
	var parsed ReviewOutput
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.Metadata.Title != "Fix login bug" {
		t.Errorf("title = %q, want %q", parsed.Metadata.Title, "Fix login bug")
	}
	if len(parsed.Comments) != 2 {
		t.Errorf("comments count = %d, want 2", len(parsed.Comments))
	}
	if parsed.Comments[1].FilePath != "src/login.go" {
		t.Errorf("comment[1].filePath = %q, want %q", parsed.Comments[1].FilePath, "src/login.go")
	}

	// filePath omitted when empty
	if bytes.Contains(buf.Bytes(), []byte(`"filePath":""`) ) {
		t.Error("empty filePath should be omitted from JSON")
	}
}

func TestFormatPatch(t *testing.T) {
	diff := "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new\n"

	var buf bytes.Buffer
	f := &PatchFormatter{}
	if err := f.FormatPatch(&buf, diff); err != nil {
		t.Fatalf("FormatPatch() error: %v", err)
	}

	if buf.String() != diff {
		t.Errorf("output = %q, want %q", buf.String(), diff)
	}
}
