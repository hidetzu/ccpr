package main

import (
	"context"
	"log"
	"runtime/debug"

	"github.com/hidetzu/ccpr/internal/app"
	"github.com/hidetzu/ccpr/internal/codecommit"
	"github.com/hidetzu/ccpr/internal/diff"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// version is set via ldflags at build time.
// Falls back to module version from BuildInfo (go install).
var version = ""

func getVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

type listInput struct {
	Repo    string `json:"repo" jsonschema:"CodeCommit repository name"`
	Status  string `json:"status,omitempty" jsonschema:"Pull request status filter: open, closed, or all. Defaults to open."`
	Config  string `json:"config,omitempty" jsonschema:"Path to ccpr configuration file"`
	Profile string `json:"profile,omitempty" jsonschema:"AWS profile name"`
	Region  string `json:"region,omitempty" jsonschema:"AWS region"`
}

type listOutput struct {
	PullRequests []app.ListPullRequest `json:"pullRequests" jsonschema:"PR summaries for the repository"`
}

type reviewInput struct {
	URL     string `json:"url,omitempty" jsonschema:"Full CodeCommit PR URL. When provided, takes priority for region, repo, and PR ID resolution."`
	Repo    string `json:"repo,omitempty" jsonschema:"CodeCommit repository name. Required when url is not provided."`
	PRId    string `json:"prId,omitempty" jsonschema:"Pull request ID. Required when url is not provided."`
	Region  string `json:"region,omitempty" jsonschema:"AWS region override"`
	Profile string `json:"profile,omitempty" jsonschema:"AWS profile name"`
	Config  string `json:"config,omitempty" jsonschema:"Path to ccpr configuration file"`
}

type commentInput struct {
	URL     string `json:"url,omitempty" jsonschema:"Full CodeCommit PR URL. When provided, takes priority for region, repo, and PR ID resolution."`
	Repo    string `json:"repo,omitempty" jsonschema:"CodeCommit repository name. Required when url is not provided."`
	PRId    string `json:"prId,omitempty" jsonschema:"Pull request ID. Required when url is not provided."`
	Body    string `json:"body" jsonschema:"Comment body. Each successful call posts a real comment to the PR."`
	Region  string `json:"region,omitempty" jsonschema:"AWS region override"`
	Profile string `json:"profile,omitempty" jsonschema:"AWS profile name"`
	Config  string `json:"config,omitempty" jsonschema:"Path to ccpr configuration file"`
}

type createInput struct {
	Repo              string `json:"repo" jsonschema:"CodeCommit repository name"`
	Title             string `json:"title" jsonschema:"Pull request title"`
	SourceBranch      string `json:"sourceBranch" jsonschema:"Source branch (the branch being merged from). MCP does not auto-detect a current branch."`
	DestinationBranch string `json:"destinationBranch" jsonschema:"Destination branch (the branch being merged into)"`
	Description       string `json:"description,omitempty" jsonschema:"Pull request description"`
	Region            string `json:"region,omitempty" jsonschema:"AWS region override"`
	Profile           string `json:"profile,omitempty" jsonschema:"AWS profile name"`
	Config            string `json:"config,omitempty" jsonschema:"Path to ccpr configuration file"`
}

func main() {
	server := newServer()
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ccpr-mcp",
		Version: getVersion(),
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_list",
		Description: "List AWS CodeCommit pull requests for a repository.",
	}, listPullRequests)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_review",
		Description: "Fetch a CodeCommit pull request's metadata, comments, and unified diff for AI-assisted review.",
	}, reviewPullRequest)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_comment",
		Description: "Post a comment to a CodeCommit pull request. WRITE-SIDE: each successful call creates a real comment, so the host should prompt the user before invocation.",
	}, postPullRequestComment)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_create",
		Description: "Create a CodeCommit pull request. WRITE-SIDE: each successful call creates a real PR, so the host should prompt the user before invocation. Unlike the CLI, sourceBranch is required (no implicit current-branch detection).",
	}, createPullRequest)

	return server
}

func listPullRequests(ctx context.Context, _ *mcp.CallToolRequest, input listInput) (*mcp.CallToolResult, listOutput, error) {
	prs, err := app.ListPullRequests(ctx, app.ListPullRequestsOptions{
		Repo:    input.Repo,
		Status:  input.Status,
		Config:  input.Config,
		Profile: input.Profile,
		Region:  input.Region,
	}, newCodeCommitClient)
	if err != nil {
		return nil, listOutput{}, err
	}
	return nil, listOutput{PullRequests: prs}, nil
}

func reviewPullRequest(ctx context.Context, _ *mcp.CallToolRequest, input reviewInput) (*mcp.CallToolResult, app.ReviewPayload, error) {
	payload, err := app.GetReview(ctx, app.GetReviewOptions{
		URL:     input.URL,
		Repo:    input.Repo,
		PRId:    input.PRId,
		Region:  input.Region,
		Profile: input.Profile,
		Config:  input.Config,
	}, newCodeCommitClient, defaultDiffGenerator)
	if err != nil {
		return nil, app.ReviewPayload{}, err
	}
	return nil, payload, nil
}

func postPullRequestComment(ctx context.Context, _ *mcp.CallToolRequest, input commentInput) (*mcp.CallToolResult, app.PostedComment, error) {
	posted, err := app.PostComment(ctx, app.PostCommentOptions{
		URL:     input.URL,
		Repo:    input.Repo,
		PRId:    input.PRId,
		Body:    input.Body,
		Region:  input.Region,
		Profile: input.Profile,
		Config:  input.Config,
	}, newCodeCommitClient)
	if err != nil {
		return nil, app.PostedComment{}, err
	}
	return nil, posted, nil
}

func createPullRequest(ctx context.Context, _ *mcp.CallToolRequest, input createInput) (*mcp.CallToolResult, app.CreatedPullRequest, error) {
	created, err := app.CreatePullRequest(ctx, app.CreatePullRequestOptions{
		Repo:              input.Repo,
		Title:             input.Title,
		SourceBranch:      input.SourceBranch,
		DestinationBranch: input.DestinationBranch,
		Description:       input.Description,
		Region:            input.Region,
		Profile:           input.Profile,
		Config:            input.Config,
	}, newCodeCommitClient)
	if err != nil {
		return nil, app.CreatedPullRequest{}, err
	}
	return nil, created, nil
}

func newCodeCommitClient(ctx context.Context, region, profile string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region, profile)
}

func defaultDiffGenerator() diff.Generator {
	return &diff.GitGenerator{}
}
