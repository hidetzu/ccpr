package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

var testOutput = ReviewOutput{
	Metadata: PRMetadata{
		PRId:              "42",
		Title:             "Fix login bug",
		Description:       "Fixes timeout on login",
		Author:            "dev",
		AuthorARN:         "arn:aws:iam::123:user/dev",
		SourceBranch:      "fix/login",
		DestinationBranch: "main",
		Status:            "OPEN",
		CreationDate:      "2026-01-15T10:30:00Z",
	},
	Comments: []Comment{
		{
			CommentId: "c1",
			Author:    "reviewer",
			AuthorARN: "arn:aws:iam::123:user/reviewer",
			Content:   "Looks good",
			Timestamp: "2026-01-15T10:30:00Z",
		},
		{
			CommentId: "c2",
			InReplyTo: "c1",
			Author:    "reviewer",
			AuthorARN: "arn:aws:iam::123:user/reviewer",
			Content:   "Check this file",
			Timestamp: "2026-01-15T10:30:00Z",
			FilePath:  "src/login.go",
		},
	},
	Diff: "diff --git a/file1 b/file1\n--- a/file1\n+++ b/file1\n@@ -1 +1 @@\n-old\n+new\ndiff --git a/file2 b/file2\n--- a/file2\n+++ b/file2\n@@ -1 +1 @@\n-foo\n+bar\n",
}

func TestFormatJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatJSON(&buf, testOutput); err != nil {
		t.Fatalf("FormatJSON() error: %v", err)
	}

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

	if parsed.Comments[0].CommentId != "c1" {
		t.Errorf("comment[0].commentId = %q, want c1", parsed.Comments[0].CommentId)
	}
	if parsed.Comments[1].InReplyTo != "c1" {
		t.Errorf("comment[1].inReplyTo = %q, want c1", parsed.Comments[1].InReplyTo)
	}

	if bytes.Contains(buf.Bytes(), []byte(`"filePath":""`)) {
		t.Error("empty filePath should be omitted from JSON")
	}
	if bytes.Contains(buf.Bytes(), []byte(`"inReplyTo":""`)) {
		t.Error("empty inReplyTo should be omitted from JSON")
	}
}

func TestFormatPatch(t *testing.T) {
	diff := "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new\n"

	var buf bytes.Buffer
	if err := FormatPatch(&buf, diff); err != nil {
		t.Fatalf("FormatPatch() error: %v", err)
	}

	if buf.String() != diff {
		t.Errorf("output = %q, want %q", buf.String(), diff)
	}
}

func TestFormatSummary(t *testing.T) {
	var buf bytes.Buffer
	if err := FormatSummary(&buf, testOutput); err != nil {
		t.Fatalf("FormatSummary() error: %v", err)
	}

	got := buf.String()

	checks := []string{
		"PR #42: Fix login bug",
		"Author:   dev",
		"Status:   OPEN",
		"Branch:   fix/login → main",
		"Created:  2026-01-15 10:30",
		"Comments: 2 (1 thread)",
		"Files:    2 changed",
		"## Description",
		"Fixes timeout on login",
	}

	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Errorf("summary missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestFormatSummary_NoDescription(t *testing.T) {
	out := testOutput
	out.Metadata.Description = ""

	var buf bytes.Buffer
	if err := FormatSummary(&buf, out); err != nil {
		t.Fatalf("FormatSummary() error: %v", err)
	}

	if strings.Contains(buf.String(), "Description") {
		t.Error("summary should not show description section when empty")
	}
}

func TestShortAuthor(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"arn:aws:iam::123456789012:user/example-user", "example-user"},
		{"arn:aws:iam::123456789012:assumed-role/role/session", "session"},
		{"plain-name", "plain-name"},
		{"", ""},
	}
	for _, tt := range tests {
		got := ShortAuthor(tt.input)
		if got != tt.want {
			t.Errorf("ShortAuthor(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
