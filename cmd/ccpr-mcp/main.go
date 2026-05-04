package main

import (
	"context"
	"log"

	"github.com/hidetzu/ccpr/internal/app"
	"github.com/hidetzu/ccpr/internal/codecommit"
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

func newCodeCommitClient(ctx context.Context, region, profile string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region, profile)
}
