package parser

// ParseResult holds the extracted components from a CodeCommit PR URL.
type ParseResult struct {
	Region     string
	Repository string
	PRId       string
}

// Parse extracts region, repository, and PR ID from a CodeCommit PR URL.
// Expected format:
//
//	https://<region>.console.aws.amazon.com/codesuite/codecommit/repositories/<repo>/pull-requests/<pr-id>
func Parse(rawURL string) (ParseResult, error) {
	panic("not implemented")
}
