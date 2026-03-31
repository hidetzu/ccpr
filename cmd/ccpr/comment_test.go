package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveBody_Inline(t *testing.T) {
	body, err := resolveBody("hello", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "hello" {
		t.Errorf("body = %q, want hello", body)
	}
}

func TestResolveBody_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.md")
	if err := os.WriteFile(path, []byte("from file"), 0o644); err != nil {
		t.Fatal(err)
	}

	body, err := resolveBody("", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "from file" {
		t.Errorf("body = %q, want 'from file'", body)
	}
}

func TestResolveBody_MutuallyExclusive(t *testing.T) {
	_, err := resolveBody("inline", "file.md")
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveBody_Missing(t *testing.T) {
	_, err := resolveBody("", "")
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestResolveBody_FileNotFound(t *testing.T) {
	_, err := resolveBody("", "/nonexistent/file.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRunComment_NoArgs(t *testing.T) {
	err := runComment([]string{})
	if err == nil {
		t.Fatal("expected error for no arguments")
	}
}

func TestRunComment_NoBody(t *testing.T) {
	err := runComment([]string{"https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/repo/pull-requests/1"})
	if err == nil {
		t.Fatal("expected error for missing body")
	}
}

func TestRunComment_InvalidFormat(t *testing.T) {
	err := runComment([]string{"--format", "patch", "--body", "test", "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/repo/pull-requests/1"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestRunComment_PartialFlags(t *testing.T) {
	err := runComment([]string{"--repo", "my-repo", "--body", "test"})
	if err == nil {
		t.Fatal("expected error when --pr-id is missing")
	}
}

func TestPrintCommentSummary(t *testing.T) {
	var buf bytes.Buffer
	err := printCommentSummary(&buf, "123", "abc-def", "user1", "2026-03-31T17:00:00+09:00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Comment posted successfully.") {
		t.Errorf("missing success message in output: %s", out)
	}
	if !strings.Contains(out, "abc-def") {
		t.Errorf("missing comment ID in output: %s", out)
	}
	if !strings.Contains(out, "#123") {
		t.Errorf("missing PR ID in output: %s", out)
	}
	// Verify no leading spaces on lines after the first
	for _, line := range strings.Split(strings.TrimSpace(out), "\n")[1:] {
		if strings.HasPrefix(line, " ") {
			t.Errorf("unexpected leading space: %q", line)
		}
	}
}

func TestPrintCommentJSON(t *testing.T) {
	var buf bytes.Buffer
	err := printCommentJSON(&buf, "123", "abc-def", "arn:aws:iam::123456789012:user/test", "2026-03-31T17:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"commentId": "abc-def"`) {
		t.Errorf("missing commentId in JSON output: %s", out)
	}
	if !strings.Contains(out, `"pullRequestId": "123"`) {
		t.Errorf("missing pullRequestId in JSON output: %s", out)
	}
}
