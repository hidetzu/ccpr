package diff

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGitGenerator_GenerateDiff sets up a local bare repo with two branches
// and verifies the merge-base diff strategy.
func TestGitGenerator_GenerateDiff(t *testing.T) {
	// Create a bare "remote" repo
	bare := t.TempDir()
	git(t, bare, "init", "--bare", "--initial-branch=main")

	// Create a working clone
	work := filepath.Join(t.TempDir(), "work")
	gitClone(t, bare, work)

	// Initial commit on main
	writeFileIn(t, work, "base.txt", "base content\n")
	git(t, work, "add", ".")
	git(t, work, "commit", "-m", "initial")
	git(t, work, "push", "origin", "main")

	// Create feature branch with a change
	git(t, work, "checkout", "-b", "feature")
	writeFileIn(t, work, "feature.txt", "new feature\n")
	git(t, work, "add", ".")
	git(t, work, "commit", "-m", "add feature")
	git(t, work, "push", "origin", "feature")

	// Add a commit to main after the branch point (should NOT appear in diff)
	git(t, work, "checkout", "main")
	writeFileIn(t, work, "main-only.txt", "main only\n")
	git(t, work, "add", ".")
	git(t, work, "commit", "-m", "main-only change")
	git(t, work, "push", "origin", "main")

	// Run diff from a fresh clone to simulate real usage
	clone := filepath.Join(t.TempDir(), "clone")
	gitClone(t, bare, clone)

	gen := &GitGenerator{}
	result, err := gen.GenerateDiff(clone, "feature", "main")
	if err != nil {
		t.Fatalf("GenerateDiff() error: %v", err)
	}

	// Diff should contain feature.txt but not main-only.txt
	if !containsString(result, "feature.txt") {
		t.Error("diff should contain feature.txt")
	}
	if containsString(result, "main-only.txt") {
		t.Error("diff should not contain main-only.txt")
	}
}

func TestGitGenerator_InvalidRepo(t *testing.T) {
	gen := &GitGenerator{}
	_, err := gen.GenerateDiff(t.TempDir(), "feature", "main")
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func gitClone(t *testing.T, bare, dest string) {
	t.Helper()
	cmd := exec.Command("git", "clone", bare, dest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git clone failed: %v\n%s", err, out)
	}
}

func writeFileIn(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && contains(s, substr)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
