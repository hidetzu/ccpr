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

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ccpr-mcp",
		Version: version,
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "ccpr_list",
		Description: "List AWS CodeCommit pull requests for a repository.",
	}, listPullRequests)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

func listPullRequests(ctx context.Context, _ *mcp.CallToolRequest, input listInput) (*mcp.CallToolResult, []app.ListPullRequest, error) {
	prs, err := app.ListPullRequests(ctx, app.ListPullRequestsOptions{
		Repo:    input.Repo,
		Status:  input.Status,
		Config:  input.Config,
		Profile: input.Profile,
		Region:  input.Region,
	}, newCodeCommitClient)
	if err != nil {
		return nil, nil, err
	}
	return nil, prs, nil
}

func newCodeCommitClient(ctx context.Context, region, profile string) (codecommit.Client, error) {
	return codecommit.NewAWSClient(ctx, region, profile)
}
