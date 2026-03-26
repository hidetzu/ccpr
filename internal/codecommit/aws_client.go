package codecommit

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	cc "github.com/aws/aws-sdk-go-v2/service/codecommit"
)

// AWSClient implements Client using the AWS SDK.
type AWSClient struct {
	client *cc.Client
}

// NewAWSClient creates a new CodeCommit client for the given region.
func NewAWSClient(ctx context.Context, region string) (*AWSClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return &AWSClient{client: cc.NewFromConfig(cfg)}, nil
}

func (c *AWSClient) GetPRMetadata(ctx context.Context, repo, prID string) (PRMetadata, error) {
	out, err := c.client.GetPullRequest(ctx, &cc.GetPullRequestInput{
		PullRequestId: aws.String(prID),
	})
	if err != nil {
		return PRMetadata{}, fmt.Errorf("GetPullRequest: %w", err)
	}

	pr := out.PullRequest

	meta := PRMetadata{
		Title:  deref(pr.Title),
		Status: string(pr.PullRequestStatus),
	}

	if pr.Description != nil {
		meta.Description = *pr.Description
	}
	if pr.AuthorArn != nil {
		meta.AuthorARN = *pr.AuthorArn
	}
	if pr.CreationDate != nil {
		meta.CreationDate = *pr.CreationDate
	}

	// Extract branch info from the first target matching the repo
	for _, t := range pr.PullRequestTargets {
		if deref(t.RepositoryName) == repo {
			meta.SourceBranch = stripRefsHeads(deref(t.SourceReference))
			meta.DestinationBranch = stripRefsHeads(deref(t.DestinationReference))
			break
		}
	}

	return meta, nil
}

func (c *AWSClient) GetPRComments(ctx context.Context, repo, prID string) ([]Comment, error) {
	paginator := cc.NewGetCommentsForPullRequestPaginator(c.client, &cc.GetCommentsForPullRequestInput{
		PullRequestId: aws.String(prID),
		RepositoryName: aws.String(repo),
	})

	var comments []Comment
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("GetCommentsForPullRequest: %w", err)
		}

		for _, group := range page.CommentsForPullRequestData {
			filePath := ""
			if group.Location != nil {
				filePath = deref(group.Location.FilePath)
			}

			for _, c := range group.Comments {
				if c.Deleted {
					continue
				}
				comment := Comment{
					Author:   deref(c.AuthorArn),
					Content:  deref(c.Content),
					FilePath: filePath,
				}
				if c.CreationDate != nil {
					comment.Timestamp = *c.CreationDate
				}
				comments = append(comments, comment)
			}
		}
	}

	return comments, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func stripRefsHeads(ref string) string {
	const prefix = "refs/heads/"
	if len(ref) > len(prefix) && ref[:len(prefix)] == prefix {
		return ref[len(prefix):]
	}
	return ref
}
