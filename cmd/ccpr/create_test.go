package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDescription_Inline(t *testing.T) {
	desc, err := resolveDescription("inline text", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "inline text" {
		t.Errorf("description = %q, want %q", desc, "inline text")
	}
}

func TestResolveDescription_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "desc.md")
	if err := os.WriteFile(path, []byte("from file"), 0o644); err != nil {
		t.Fatal(err)
	}

	desc, err := resolveDescription("", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "from file" {
		t.Errorf("description = %q, want %q", desc, "from file")
	}
}

func TestResolveDescription_MutuallyExclusive(t *testing.T) {
	_, err := resolveDescription("inline", "file.md")
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolveDescription_Empty(t *testing.T) {
	desc, err := resolveDescription("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if desc != "" {
		t.Errorf("description = %q, want empty", desc)
	}
}

func TestResolveDescription_FileNotFound(t *testing.T) {
	_, err := resolveDescription("", "/nonexistent/file.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestRunCreate_MissingRepo(t *testing.T) {
	err := runCreate([]string{"--title", "test", "--dest", "main"})
	if err == nil {
		t.Fatal("expected error for missing --repo")
	}
	if !strings.Contains(err.Error(), "--repo") {
		t.Errorf("error should mention --repo: %v", err)
	}
}

func TestRunCreate_MissingTitle(t *testing.T) {
	err := runCreate([]string{"--repo", "my-repo", "--dest", "main"})
	if err == nil {
		t.Fatal("expected error for missing --title")
	}
	if !strings.Contains(err.Error(), "--title") {
		t.Errorf("error should mention --title: %v", err)
	}
}

func TestRunCreate_MissingDest(t *testing.T) {
	err := runCreate([]string{"--repo", "my-repo", "--title", "test"})
	if err == nil {
		t.Fatal("expected error for missing --dest")
	}
	if !strings.Contains(err.Error(), "--dest") {
		t.Errorf("error should mention --dest: %v", err)
	}
}

func TestRunCreate_InvalidFormat(t *testing.T) {
	err := runCreate([]string{"--repo", "my-repo", "--title", "test", "--dest", "main", "--format", "patch"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunCreate_MissingRegion(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("repoMappings:\n  my-repo: .\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runCreate([]string{"--config", cfgPath, "--repo", "my-repo", "--title", "test", "--dest", "main", "--source", "feature"})
	if err == nil {
		t.Fatal("expected error for missing region")
	}
	if !strings.Contains(err.Error(), "region is required") {
		t.Errorf("error should mention region: %v", err)
	}
}

func TestBuildConsoleURL(t *testing.T) {
	url := buildConsoleURL("ap-northeast-1", "my-repo", "42")
	want := "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42"
	if url != want {
		t.Errorf("url = %q, want %q", url, want)
	}
}

func TestPrintCreateSummary(t *testing.T) {
	var buf bytes.Buffer
	err := printCreateSummary(&buf, "42", "Add feature X", "my-repo", "feature/add-x", "main",
		"https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Pull request created") {
		t.Errorf("missing success message: %s", out)
	}
	if !strings.Contains(out, "#42") {
		t.Errorf("missing PR ID: %s", out)
	}
	if !strings.Contains(out, "Add feature X") {
		t.Errorf("missing title: %s", out)
	}
	if !strings.Contains(out, "feature/add-x") {
		t.Errorf("missing source branch: %s", out)
	}
	if !strings.Contains(out, "main") {
		t.Errorf("missing destination branch: %s", out)
	}
}

func TestPrintCreateJSON(t *testing.T) {
	var buf bytes.Buffer
	err := printCreateJSON(&buf, "42", "Add feature X", "my-repo", "feature/add-x", "main",
		"https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"prId": "42"`) {
		t.Errorf("missing prId in JSON: %s", out)
	}
	if !strings.Contains(out, `"title": "Add feature X"`) {
		t.Errorf("missing title in JSON: %s", out)
	}
	if !strings.Contains(out, `"repository": "my-repo"`) {
		t.Errorf("missing repository in JSON: %s", out)
	}
	if !strings.Contains(out, `"sourceBranch": "feature/add-x"`) {
		t.Errorf("missing sourceBranch in JSON: %s", out)
	}
	if !strings.Contains(out, `"url"`) {
		t.Errorf("missing url in JSON: %s", out)
	}
}

func TestCurrentBranch(t *testing.T) {
	// Use the current repo itself to test
	branch, err := currentBranch(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}
