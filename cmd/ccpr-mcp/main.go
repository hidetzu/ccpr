package main

import (
	"context"
	"log"

	"github.com/hidetzu/ccpr/internal/app"
	"github.com/hidetzu/ccpr/internal/codecommit"
	"github.com/hidetzu/ccpr/internal/diff"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const version = "dev"

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

func main() {
	server := newServer()
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func newServer() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ccpr-mcp",
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_list",
		Description: "List AWS CodeCommit pull requests for a repository.",
	}, listPullRequests)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_review",
		Description: "Fetch a CodeCommit pull request's metadata, comments, and unified diff for AI-assisted review.",
	}, reviewPullRequest)

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

func newCodeCommitClient(ctx context.Context, region, profile string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region, profile)
}

func defaultDiffGenerator() diff.Generator {
	return &diff.GitGenerator{}
}
