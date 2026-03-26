package diff

import (
	"fmt"
	"os/exec"
	"strings"
)

// Generator defines the interface for producing diffs between branches.
type Generator interface {
	// GenerateDiff produces a unified diff using merge-base strategy.
	//
	// Steps:
	//   1. git fetch origin <sourceBranch> <destBranch>
	//   2. git merge-base origin/<destBranch> origin/<sourceBranch>
	//   3. git diff <merge-base>...origin/<sourceBranch>
	GenerateDiff(repoPath, sourceBranch, destBranch string) (string, error)
}

// GitGenerator implements Generator using local Git commands.
type GitGenerator struct{}

func (g *GitGenerator) GenerateDiff(repoPath, sourceBranch, destBranch string) (string, error) {
	// Step 1: fetch latest refs
	if err := g.gitRun(repoPath, "fetch", "origin", sourceBranch, destBranch); err != nil {
		return "", fmt.Errorf("git fetch: %w", err)
	}

	// Step 2: find merge-base
	mergeBase, err := g.gitOutput(repoPath, "merge-base", "origin/"+destBranch, "origin/"+sourceBranch)
	if err != nil {
		return "", fmt.Errorf("git merge-base: %w", err)
	}

	// Step 3: generate diff
	diff, err := g.gitOutput(repoPath, "diff", mergeBase+"...origin/"+sourceBranch)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}

	return diff, nil
}

func (g *GitGenerator) gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *GitGenerator) gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
