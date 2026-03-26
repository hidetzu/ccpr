package diff

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
