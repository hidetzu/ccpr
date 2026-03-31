package codecommit

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	cc "github.com/aws/aws-sdk-go-v2/service/codecommit"
	"github.com/aws/aws-sdk-go-v2/service/codecommit/types"
)

// AWSClient implements Client using the AWS SDK.
type AWSClient struct {
	client *cc.Client
}

// NewAWSClient creates a new CodeCommit client for the given region and profile.
// If profile is empty, the default credential chain is used.
func NewAWSClient(ctx context.Context, region, profile string) (*AWSClient, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
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

	// Extract branch and commit info from the first target matching the repo
	for _, t := range pr.PullRequestTargets {
		if deref(t.RepositoryName) == repo {
			meta.SourceBranch = stripRefsHeads(deref(t.SourceReference))
			meta.DestinationBranch = stripRefsHeads(deref(t.DestinationReference))
			meta.SourceCommit = deref(t.SourceCommit)
			meta.DestinationCommit = deref(t.DestinationCommit)
			break
		}
	}

	return meta, nil
}

func (c *AWSClient) GetPRComments(ctx context.Context, repo, prID, beforeCommit, afterCommit string) ([]Comment, error) {
	paginator := cc.NewGetCommentsForPullRequestPaginator(c.client, &cc.GetCommentsForPullRequestInput{
		PullRequestId:  aws.String(prID),
		RepositoryName: aws.String(repo),
		BeforeCommitId: aws.String(beforeCommit),
		AfterCommitId:  aws.String(afterCommit),
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
					CommentId: deref(c.CommentId),
					InReplyTo: deref(c.InReplyTo),
					Author:    deref(c.AuthorArn),
					Content:   deref(c.Content),
					FilePath:  filePath,
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

func (c *AWSClient) ListPRs(ctx context.Context, repo, status string) ([]PRSummary, error) {
	input := &cc.ListPullRequestsInput{
		RepositoryName: aws.String(repo),
	}
	switch status {
	case "open":
		input.PullRequestStatus = types.PullRequestStatusEnumOpen
	case "closed":
		input.PullRequestStatus = types.PullRequestStatusEnumClosed
	case "all":
		// no filter
	default:
		input.PullRequestStatus = types.PullRequestStatusEnumOpen
	}

	paginator := cc.NewListPullRequestsPaginator(c.client, input)

	var ids []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("ListPullRequests: %w", err)
		}
		ids = append(ids, page.PullRequestIds...)
	}

	var summaries []PRSummary
	for _, id := range ids {
		meta, err := c.GetPRMetadata(ctx, repo, id)
		if err != nil {
			return nil, fmt.Errorf("fetching PR %s: %w", id, err)
		}
		summaries = append(summaries, PRSummary{
			PRId:              id,
			Title:             meta.Title,
			AuthorARN:         meta.AuthorARN,
			SourceBranch:      meta.SourceBranch,
			DestinationBranch: meta.DestinationBranch,
			Status:            meta.Status,
			CreationDate:      meta.CreationDate,
		})
	}

	return summaries, nil
}

func (c *AWSClient) PostComment(ctx context.Context, repo, prID, beforeCommit, afterCommit, content string) (PostCommentResult, error) {
	out, err := c.client.PostCommentForPullRequest(ctx, &cc.PostCommentForPullRequestInput{
		PullRequestId:  aws.String(prID),
		RepositoryName: aws.String(repo),
		BeforeCommitId: aws.String(beforeCommit),
		AfterCommitId:  aws.String(afterCommit),
		Content:        aws.String(content),
	})
	if err != nil {
		return PostCommentResult{}, fmt.Errorf("PostCommentForPullRequest: %w", err)
	}

	result := PostCommentResult{}
	if out.Comment != nil {
		result.CommentID = deref(out.Comment.CommentId)
		result.AuthorARN = deref(out.Comment.AuthorArn)
		if out.Comment.CreationDate != nil {
			result.CreationDate = *out.Comment.CreationDate
		}
	}
	return result, nil
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
