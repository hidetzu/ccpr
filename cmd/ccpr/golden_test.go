package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hidetzu/ccpr/internal/output"
)

// loadGolden reads a golden file from testdata/ and returns its content.
func loadGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", name, err)
	}
	return string(data)
}

// normalizeJSON re-encodes JSON to normalize whitespace for comparison.
func normalizeJSON(t *testing.T, data string) string {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return string(out) + "\n"
}

func TestGolden_ReviewJSON(t *testing.T) {
	review := output.ReviewOutput{
		Metadata: output.PRMetadata{
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
		Comments: []output.Comment{
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
		Diff: "diff --git a/file1 b/file1\n--- a/file1\n+++ b/file1\n@@ -1 +1 @@\n-old\n+new\n",
	}

	var buf bytes.Buffer
	if err := output.FormatJSON(&buf, review); err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	got := normalizeJSON(t, buf.String())
	want := normalizeJSON(t, loadGolden(t, "review.json"))

	if got != want {
		t.Errorf("review JSON mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGolden_ReviewJSON_EmptyComments(t *testing.T) {
	review := output.ReviewOutput{
		Metadata: output.PRMetadata{
			PRId:              "42",
			Title:             "Fix login bug",
			Description:       "",
			Author:            "dev",
			AuthorARN:         "arn:aws:iam::123:user/dev",
			SourceBranch:      "fix/login",
			DestinationBranch: "main",
			Status:            "OPEN",
			CreationDate:      "2026-01-15T10:30:00Z",
		},
		Comments: []output.Comment{},
		Diff:     "",
	}

	var buf bytes.Buffer
	if err := output.FormatJSON(&buf, review); err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	got := normalizeJSON(t, buf.String())
	want := normalizeJSON(t, loadGolden(t, "review_empty_comments.json"))

	if got != want {
		t.Errorf("review empty comments JSON mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGolden_ListJSON(t *testing.T) {
	prs := []listJSONItem{
		{
			PRId:              "42",
			Title:             "Add feature X",
			AuthorARN:         "arn:aws:iam::123:user/dev",
			SourceBranch:      "feature/x",
			DestinationBranch: "main",
			Status:            "OPEN",
			CreationDate:      "2026-04-01T10:00:00Z",
		},
		{
			PRId:              "41",
			Title:             "Fix bug Y",
			AuthorARN:         "arn:aws:iam::123:user/reviewer",
			SourceBranch:      "fix/y",
			DestinationBranch: "main",
			Status:            "OPEN",
			CreationDate:      "2026-03-30T09:00:00Z",
		},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(prs); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	got := normalizeJSON(t, buf.String())
	want := normalizeJSON(t, loadGolden(t, "list.json"))

	if got != want {
		t.Errorf("list JSON mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGolden_ListJSON_Empty(t *testing.T) {
	prs := []listJSONItem{}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(prs); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	got := normalizeJSON(t, buf.String())
	want := normalizeJSON(t, loadGolden(t, "list_empty.json"))

	if got != want {
		t.Errorf("list empty JSON mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGolden_CommentJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := printCommentJSON(&buf, "42", "eb596ff8-5133-438f-88d6-7c94f693302b", "arn:aws:iam::123:user/dev", "2026-04-01T10:00:00+09:00"); err != nil {
		t.Fatalf("printCommentJSON error: %v", err)
	}

	got := normalizeJSON(t, buf.String())
	want := normalizeJSON(t, loadGolden(t, "comment.json"))

	if got != want {
		t.Errorf("comment JSON mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestGolden_CreateJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := printCreateJSON(&buf, "42", "Add feature X", "my-repo", "feature/add-x", "main",
		"https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42"); err != nil {
		t.Fatalf("printCreateJSON error: %v", err)
	}

	got := normalizeJSON(t, buf.String())
	want := normalizeJSON(t, loadGolden(t, "create.json"))

	if got != want {
		t.Errorf("create JSON mismatch.\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestGolden_SchemaValidation verifies that all JSON outputs contain
// the required fields defined in docs/json-schema.md.
func TestGolden_SchemaValidation(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		isArray  bool
		required []string
	}{
		{
			name:     "review",
			file:     "review.json",
			required: []string{"metadata", "comments", "diff"},
		},
		{
			name:     "review_empty_comments",
			file:     "review_empty_comments.json",
			required: []string{"metadata", "comments", "diff"},
		},
		{
			name:     "list",
			file:     "list.json",
			isArray:  true,
			required: []string{"prId", "title", "authorArn", "sourceBranch", "destinationBranch", "status", "creationDate"},
		},
		{
			name:     "comment",
			file:     "comment.json",
			required: []string{"commentId", "pullRequestId", "authorArn", "creationDate"},
		},
		{
			name:     "create",
			file:     "create.json",
			required: []string{"prId", "title", "repository", "sourceBranch", "destinationBranch", "url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadGolden(t, tt.file)

			if tt.isArray {
				var arr []map[string]any
				if err := json.Unmarshal([]byte(data), &arr); err != nil {
					t.Fatalf("invalid JSON array: %v", err)
				}
				if len(arr) == 0 {
					t.Skip("empty array, skipping field check")
				}
				for _, obj := range arr {
					checkRequiredFields(t, obj, tt.required)
				}
			} else {
				var obj map[string]any
				if err := json.Unmarshal([]byte(data), &obj); err != nil {
					t.Fatalf("invalid JSON object: %v", err)
				}
				checkRequiredFields(t, obj, tt.required)
			}
		})
	}
}

// TestGolden_ReviewMetadataFields validates the metadata object has all required fields.
func TestGolden_ReviewMetadataFields(t *testing.T) {
	data := loadGolden(t, "review.json")
	var obj map[string]any
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	metadata, ok := obj["metadata"].(map[string]any)
	if !ok {
		t.Fatal("metadata is not an object")
	}

	required := []string{"prId", "title", "description", "author", "authorArn", "sourceBranch", "destinationBranch", "status", "creationDate"}
	checkRequiredFields(t, metadata, required)
}

// TestGolden_Indentation verifies that actual JSON output uses 2-space indentation
// by checking raw output directly (not via normalizeJSON which re-encodes).
func TestGolden_Indentation(t *testing.T) {
	tests := []struct {
		name   string
		output func() string
	}{
		{
			name: "review",
			output: func() string {
				review := output.ReviewOutput{
					Metadata: output.PRMetadata{PRId: "1", Title: "t", Author: "a", AuthorARN: "arn", SourceBranch: "s", DestinationBranch: "d", Status: "OPEN", CreationDate: "2026-01-01T00:00:00Z"},
					Comments: []output.Comment{},
					Diff:     "",
				}
				var buf bytes.Buffer
				_ = output.FormatJSON(&buf, review)
				return buf.String()
			},
		},
		{
			name: "list",
			output: func() string {
				items := []listJSONItem{{PRId: "1", Title: "t", AuthorARN: "arn", SourceBranch: "s", DestinationBranch: "d", Status: "OPEN", CreationDate: "2026-01-01T00:00:00Z"}}
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetIndent("", "  ")
				_ = enc.Encode(items)
				return buf.String()
			},
		},
		{
			name: "comment",
			output: func() string {
				var buf bytes.Buffer
				_ = printCommentJSON(&buf, "1", "c1", "arn", "2026-01-01T00:00:00Z")
				return buf.String()
			},
		},
		{
			name: "create",
			output: func() string {
				var buf bytes.Buffer
				_ = printCreateJSON(&buf, "1", "t", "r", "s", "d", "https://example.com")
				return buf.String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := tt.output()
			lines := strings.Split(raw, "\n")
			for i, line := range lines {
				if len(line) == 0 {
					continue
				}
				// Lines with indentation must use exactly 2-space multiples
				trimmed := strings.TrimLeft(line, " ")
				indent := len(line) - len(trimmed)
				if indent > 0 && indent%2 != 0 {
					t.Errorf("line %d: odd indentation (%d spaces): %q", i+1, indent, line)
				}
				if strings.HasPrefix(line, "\t") {
					t.Errorf("line %d: tab indentation found: %q", i+1, line)
				}
			}
		})
	}
}

// TestGolden_FieldTypes validates that fields have the correct JSON types
// as defined in docs/json-schema.md.
func TestGolden_FieldTypes(t *testing.T) {
	t.Run("review", func(t *testing.T) {
		data := loadGolden(t, "review.json")
		var obj map[string]any
		if err := json.Unmarshal([]byte(data), &obj); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// top-level types
		assertType(t, obj, "metadata", "object")
		assertType(t, obj, "comments", "array")
		assertType(t, obj, "diff", "string")

		// metadata field types
		metadata := obj["metadata"].(map[string]any)
		for _, field := range []string{"prId", "title", "description", "author", "authorArn", "sourceBranch", "destinationBranch", "status", "creationDate"} {
			assertType(t, metadata, field, "string")
		}

		// comment item types
		comments := obj["comments"].([]any)
		for i, item := range comments {
			comment := item.(map[string]any)
			for _, field := range []string{"commentId", "author", "authorArn", "content", "timestamp"} {
				assertTypeF(t, comment, field, "string", "comment[%d]", i)
			}
		}
	})

	t.Run("list", func(t *testing.T) {
		data := loadGolden(t, "list.json")
		var arr []map[string]any
		if err := json.Unmarshal([]byte(data), &arr); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		for i, obj := range arr {
			for _, field := range []string{"prId", "title", "authorArn", "sourceBranch", "destinationBranch", "status", "creationDate"} {
				assertTypeF(t, obj, field, "string", "item[%d]", i)
			}
		}
	})

	t.Run("comment", func(t *testing.T) {
		data := loadGolden(t, "comment.json")
		var obj map[string]any
		if err := json.Unmarshal([]byte(data), &obj); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		for _, field := range []string{"commentId", "pullRequestId", "authorArn", "creationDate"} {
			assertType(t, obj, field, "string")
		}
	})

	t.Run("create", func(t *testing.T) {
		data := loadGolden(t, "create.json")
		var obj map[string]any
		if err := json.Unmarshal([]byte(data), &obj); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		for _, field := range []string{"prId", "title", "repository", "sourceBranch", "destinationBranch", "url"} {
			assertType(t, obj, field, "string")
		}
	})
}

func assertType(t *testing.T, obj map[string]any, field, expectedType string) {
	t.Helper()
	assertTypeF(t, obj, field, expectedType, "%s", field)
}

func assertTypeF(t *testing.T, obj map[string]any, field, expectedType, ctxFmt string, ctxArgs ...any) {
	t.Helper()
	val, ok := obj[field]
	if !ok {
		return // missing field is caught by required field tests
	}
	ctx := fmt.Sprintf(ctxFmt, ctxArgs...)
	switch expectedType {
	case "string":
		if _, ok := val.(string); !ok {
			t.Errorf("%s.%s: expected string, got %T", ctx, field, val)
		}
	case "object":
		if _, ok := val.(map[string]any); !ok {
			t.Errorf("%s.%s: expected object, got %T", ctx, field, val)
		}
	case "array":
		if _, ok := val.([]any); !ok {
			t.Errorf("%s.%s: expected array, got %T", ctx, field, val)
		}
	}
}

// TestGolden_TimestampFormat extracts all timestamp fields from golden files
// and validates they parse as ISO 8601.
func TestGolden_TimestampFormat(t *testing.T) {
	tests := []struct {
		name           string
		file           string
		timestampPaths [][]string // paths to timestamp fields in the JSON
		isArray        bool
	}{
		{
			name:           "review metadata",
			file:           "review.json",
			timestampPaths: [][]string{{"metadata", "creationDate"}},
		},
		{
			name:           "review comments",
			file:           "review.json",
			timestampPaths: [][]string{{"timestamp"}},
			isArray:        true, // items inside "comments" array
		},
		{
			name:           "list",
			file:           "list.json",
			timestampPaths: [][]string{{"creationDate"}},
			isArray:        true,
		},
		{
			name:           "comment",
			file:           "comment.json",
			timestampPaths: [][]string{{"creationDate"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadGolden(t, tt.file)

			if tt.isArray {
				// For "review comments", extract the comments array first
				var root any
				if err := json.Unmarshal([]byte(data), &root); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}

				var arr []any
				switch v := root.(type) {
				case []any:
					arr = v
				case map[string]any:
					// extract "comments" array from review
					if comments, ok := v["comments"].([]any); ok {
						arr = comments
					} else {
						t.Fatal("could not find array in golden file")
					}
				}

				for i, item := range arr {
					obj, ok := item.(map[string]any)
					if !ok {
						continue
					}
					for _, path := range tt.timestampPaths {
						val := extractField(obj, path)
						if s, ok := val.(string); ok {
							if _, err := time.Parse("2006-01-02T15:04:05Z07:00", s); err != nil {
								t.Errorf("item[%d].%s: %q is not valid ISO 8601: %v", i, path[len(path)-1], s, err)
							}
						}
					}
				}
			} else {
				var obj map[string]any
				if err := json.Unmarshal([]byte(data), &obj); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				for _, path := range tt.timestampPaths {
					val := extractField(obj, path)
					if s, ok := val.(string); ok {
						if _, err := time.Parse("2006-01-02T15:04:05Z07:00", s); err != nil {
							t.Errorf("%s: %q is not valid ISO 8601: %v", path[len(path)-1], s, err)
						}
					} else {
						t.Errorf("timestamp field %v is not a string", path)
					}
				}
			}
		})
	}
}

// TestGolden_EmptyArrayNotNull verifies that empty results serialize as []
// rather than null.
func TestGolden_EmptyArrayNotNull(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"list_empty", "list_empty.json"},
		{"review_empty_comments", "review_empty_comments.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := loadGolden(t, tt.file)

			if tt.name == "review_empty_comments" {
				var obj map[string]any
				if err := json.Unmarshal([]byte(data), &obj); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				comments, ok := obj["comments"]
				if !ok {
					t.Fatal("missing comments field")
				}
				arr, ok := comments.([]any)
				if !ok {
					t.Fatalf("comments is not an array (got %T)", comments)
				}
				if arr == nil {
					t.Error("comments is null, expected []")
				}
			} else {
				var arr []any
				if err := json.Unmarshal([]byte(data), &arr); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if arr == nil {
					t.Error("array is null, expected []")
				}
			}
		})
	}
}

// TestGolden_ReviewCommentFields validates required fields in comment array items.
func TestGolden_ReviewCommentFields(t *testing.T) {
	data := loadGolden(t, "review.json")
	var obj map[string]any
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	comments, ok := obj["comments"].([]any)
	if !ok {
		t.Fatal("comments is not an array")
	}

	required := []string{"commentId", "author", "authorArn", "content", "timestamp"}
	for i, item := range comments {
		comment, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("comment[%d] is not an object", i)
		}
		for _, field := range required {
			if _, ok := comment[field]; !ok {
				t.Errorf("comment[%d] missing required field %q", i, field)
			}
		}
	}
}

// extractField navigates a nested map by the given path keys.
func extractField(obj map[string]any, path []string) any {
	var current any = obj
	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = m[key]
	}
	return current
}

func checkRequiredFields(t *testing.T, obj map[string]any, required []string) {
	t.Helper()
	for _, field := range required {
		if _, ok := obj[field]; !ok {
			t.Errorf("missing required field %q", field)
		}
	}
}
