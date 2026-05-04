package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hidetzu/ccpr/internal/codecommit"
)

type fakeCommentClient struct {
	metadata          codecommit.PRMetadata
	postResult        codecommit.PostCommentResult
	postErr           error
	gotMetadataRepo   string
	gotMetadataPRID   string
	gotPostRepo       string
	gotPostPRID       string
	gotPostBefore     string
	gotPostAfter      string
	gotPostBody       string
}

func (f *fakeCommentClient) GetPRMetadata(_ context.Context, repo, prID string) (codecommit.PRMetadata, error) {
	f.gotMetadataRepo = repo
	f.gotMetadataPRID = prID
	return f.metadata, nil
}

func (f *fakeCommentClient) GetPRComments(context.Context, string, string, string, string) ([]codecommit.Comment, error) {
	return nil, nil
}

func (f *fakeCommentClient) ListPRs(context.Context, string, string) ([]codecommit.PRSummary, error) {
	return nil, nil
}

func (f *fakeCommentClient) PostComment(_ context.Context, repo, prID, before, after, body string) (codecommit.PostCommentResult, error) {
	f.gotPostRepo = repo
	f.gotPostPRID = prID
	f.gotPostBefore = before
	f.gotPostAfter = after
	f.gotPostBody = body
	return f.postResult, f.postErr
}

func (f *fakeCommentClient) CreatePR(context.Context, string, string, string, string, string) (codecommit.CreatePRResult, error) {
	return codecommit.CreatePRResult{}, nil
}

func TestPostCommentRequiresBody(t *testing.T) {
	_, err := PostComment(context.Background(), PostCommentOptions{Repo: "r", PRId: "1"}, nil)
	if err == nil || !strings.Contains(err.Error(), "body is required") {
		t.Fatalf("error = %v, want body-required", err)
	}
}

func TestPostCommentRequiresURLOrRepoAndPRId(t *testing.T) {
	_, err := PostComment(context.Background(), PostCommentOptions{Body: "hi"}, nil)
	if !errors.Is(err, ErrMissingPRRef) {
		t.Fatalf("error = %v, want ErrMissingPRRef", err)
	}
}

func TestPostCommentRejectsInvalidURL(t *testing.T) {
	_, err := PostComment(context.Background(), PostCommentOptions{Body: "hi", URL: "not-a-url"}, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid PR URL") {
		t.Fatalf("error = %v, want invalid URL", err)
	}
}

func TestPostCommentURLDrivesRegionAndBuildsResult(t *testing.T) {
	configPath := writeListConfig(t, "region: us-east-1\n")
	created := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	cc := &fakeCommentClient{
		metadata: codecommit.PRMetadata{
			SourceCommit:      "src-commit",
			DestinationCommit: "dst-commit",
		},
		postResult: codecommit.PostCommentResult{
			CommentID:    "c-1",
			AuthorARN:    "arn:aws:iam::123456789012:user/example",
			CreationDate: created,
		},
	}

	gotRegion := ""
	got, err := PostComment(context.Background(), PostCommentOptions{
		URL:    "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42",
		Body:   "looks good",
		Config: configPath,
	}, func(_ context.Context, region, _ string) (codecommit.Client, error) {
		gotRegion = region
		return cc, nil
	})
	if err != nil {
		t.Fatalf("PostComment returned error: %v", err)
	}
	if gotRegion != "ap-northeast-1" {
		t.Fatalf("region = %q, want ap-northeast-1 (URL wins over config)", gotRegion)
	}
	if cc.gotMetadataRepo != "my-repo" || cc.gotMetadataPRID != "42" {
		t.Fatalf("GetPRMetadata called with repo=%q prID=%q", cc.gotMetadataRepo, cc.gotMetadataPRID)
	}
	if cc.gotPostBefore != "dst-commit" || cc.gotPostAfter != "src-commit" {
		t.Fatalf("PostComment called with before=%q after=%q", cc.gotPostBefore, cc.gotPostAfter)
	}
	if cc.gotPostBody != "looks good" {
		t.Fatalf("PostComment body = %q", cc.gotPostBody)
	}
	if got.CommentID != "c-1" || got.PullRequestID != "42" {
		t.Fatalf("result = %+v", got)
	}
	if got.CreationDate != "2026-05-04T10:00:00Z" {
		t.Fatalf("CreationDate = %q", got.CreationDate)
	}
}

func TestPostCommentRepoPRIdUsesConfigRegion(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\n")
	cc := &fakeCommentClient{
		metadata:   codecommit.PRMetadata{SourceCommit: "s", DestinationCommit: "d"},
		postResult: codecommit.PostCommentResult{CommentID: "c-1", CreationDate: time.Now()},
	}

	gotRegion := ""
	_, err := PostComment(context.Background(), PostCommentOptions{
		Repo:   "my-repo",
		PRId:   "42",
		Body:   "hi",
		Config: configPath,
	}, func(_ context.Context, region, _ string) (codecommit.Client, error) {
		gotRegion = region
		return cc, nil
	})
	if err != nil {
		t.Fatalf("PostComment returned error: %v", err)
	}
	if gotRegion != "ap-northeast-1" {
		t.Fatalf("region = %q, want ap-northeast-1 from config", gotRegion)
	}
}

func TestPostCommentWrapsAWSFailuresAsSystemError(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\n")
	cc := &fakeCommentClient{
		metadata: codecommit.PRMetadata{SourceCommit: "s", DestinationCommit: "d"},
		postErr:  errors.New("boom"),
	}

	_, err := PostComment(context.Background(), PostCommentOptions{
		Repo:   "my-repo",
		PRId:   "42",
		Body:   "hi",
		Config: configPath,
	}, func(context.Context, string, string) (codecommit.Client, error) {
		return cc, nil
	})
	if err == nil {
		t.Fatal("PostComment returned nil error")
	}
	var sysErr *SystemError
	if !errors.As(err, &sysErr) {
		t.Fatalf("error = %T %v, want *SystemError", err, err)
	}
	if !strings.Contains(err.Error(), "posting comment") {
		t.Fatalf("error message = %q, want posting-comment context", err.Error())
	}
}
