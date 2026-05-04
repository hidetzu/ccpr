package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hidetzu/ccpr/internal/codecommit"
	"github.com/hidetzu/ccpr/internal/diff"
)

type fakeReviewClient struct {
	metadata          codecommit.PRMetadata
	comments          []codecommit.Comment
	gotMetadataRepo   string
	gotMetadataPRID   string
	gotCommentsRepo   string
	gotCommentsPRID   string
	gotCommentsBefore string
	gotCommentsAfter  string
}

func (f *fakeReviewClient) GetPRMetadata(_ context.Context, repo, prID string) (codecommit.PRMetadata, error) {
	f.gotMetadataRepo = repo
	f.gotMetadataPRID = prID
	return f.metadata, nil
}

func (f *fakeReviewClient) GetPRComments(_ context.Context, repo, prID, before, after string) ([]codecommit.Comment, error) {
	f.gotCommentsRepo = repo
	f.gotCommentsPRID = prID
	f.gotCommentsBefore = before
	f.gotCommentsAfter = after
	return f.comments, nil
}

func (f *fakeReviewClient) ListPRs(context.Context, string, string) ([]codecommit.PRSummary, error) {
	return nil, nil
}

func (f *fakeReviewClient) PostComment(context.Context, string, string, string, string, string) (codecommit.PostCommentResult, error) {
	return codecommit.PostCommentResult{}, nil
}

func (f *fakeReviewClient) CreatePR(context.Context, string, string, string, string, string) (codecommit.CreatePRResult, error) {
	return codecommit.CreatePRResult{}, nil
}

type fakeDiffGen struct {
	diff       string
	err        error
	gotRepo    string
	gotSource  string
	gotDestBr  string
	called     bool
}

func (f *fakeDiffGen) GenerateDiff(repoPath, sourceBranch, destBranch string) (string, error) {
	f.called = true
	f.gotRepo = repoPath
	f.gotSource = sourceBranch
	f.gotDestBr = destBranch
	return f.diff, f.err
}

func TestGetReviewRequiresURLOrRepoAndPRId(t *testing.T) {
	_, err := GetReview(context.Background(), GetReviewOptions{}, nil, nil)
	if !errors.Is(err, ErrMissingPRRef) {
		t.Fatalf("error = %v, want ErrMissingPRRef", err)
	}
}

func TestGetReviewRejectsInvalidURL(t *testing.T) {
	_, err := GetReview(context.Background(), GetReviewOptions{URL: "not-a-url"}, nil, nil)
	if err == nil {
		t.Fatal("GetReview returned nil error")
	}
	if !strings.Contains(err.Error(), "invalid PR URL") {
		t.Fatalf("error = %q, want invalid URL", err.Error())
	}
}

func TestGetReviewURLDrivesRegionAndBuildsPayload(t *testing.T) {
	configPath := writeListConfig(t, "region: us-east-1\nrepoMappings:\n  my-repo: /tmp/my-repo\n")
	created := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	commented := time.Date(2026, 4, 2, 9, 30, 0, 0, time.UTC)

	cc := &fakeReviewClient{
		metadata: codecommit.PRMetadata{
			Title:             "Add feature X",
			Description:       "desc",
			AuthorARN:         "arn:aws:iam::123456789012:user/example",
			SourceBranch:      "feature/x",
			DestinationBranch: "main",
			SourceCommit:      "src-commit",
			DestinationCommit: "dst-commit",
			Status:            "OPEN",
			CreationDate:      created,
		},
		comments: []codecommit.Comment{{
			CommentId: "c1",
			Author:    "arn:aws:iam::123456789012:user/example",
			Content:   "looks good",
			Timestamp: commented,
		}},
	}
	dg := &fakeDiffGen{diff: "diff text"}

	gotRegion := ""
	got, err := GetReview(context.Background(), GetReviewOptions{
		URL:    "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42",
		Config: configPath,
	}, func(_ context.Context, region, _ string) (codecommit.Client, error) {
		gotRegion = region
		return cc, nil
	}, func() diff.Generator {
		return dg
	})
	if err != nil {
		t.Fatalf("GetReview returned error: %v", err)
	}

	if gotRegion != "ap-northeast-1" {
		t.Fatalf("region passed to client factory = %q, want ap-northeast-1 (URL wins over config)", gotRegion)
	}
	if cc.gotMetadataRepo != "my-repo" || cc.gotMetadataPRID != "42" {
		t.Fatalf("GetPRMetadata called with repo=%q prID=%q", cc.gotMetadataRepo, cc.gotMetadataPRID)
	}
	if cc.gotCommentsBefore != "dst-commit" || cc.gotCommentsAfter != "src-commit" {
		t.Fatalf("GetPRComments called with before=%q after=%q", cc.gotCommentsBefore, cc.gotCommentsAfter)
	}
	if !dg.called {
		t.Fatal("diff generator was not invoked")
	}
	if dg.gotRepo != "/tmp/my-repo" || dg.gotSource != "feature/x" || dg.gotDestBr != "main" {
		t.Fatalf("GenerateDiff called with repo=%q source=%q dest=%q", dg.gotRepo, dg.gotSource, dg.gotDestBr)
	}

	if got.Metadata.PRId != "42" || got.Metadata.Title != "Add feature X" {
		t.Fatalf("metadata = %+v", got.Metadata)
	}
	if got.Metadata.Author != "example" {
		t.Fatalf("Author = %q, want short form 'example'", got.Metadata.Author)
	}
	if got.Metadata.CreationDate != "2026-04-01T10:00:00Z" {
		t.Fatalf("CreationDate = %q", got.Metadata.CreationDate)
	}
	if got.Diff != "diff text" {
		t.Fatalf("Diff = %q", got.Diff)
	}
	if len(got.Comments) != 1 || got.Comments[0].Timestamp != "2026-04-02T09:30:00Z" {
		t.Fatalf("comments = %+v", got.Comments)
	}
}

func TestGetReviewRepoPRIdUsesConfigRegion(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\nrepoMappings:\n  my-repo: /tmp/my-repo\n")
	cc := &fakeReviewClient{
		metadata: codecommit.PRMetadata{
			SourceBranch:      "feature/x",
			DestinationBranch: "main",
			CreationDate:      time.Now(),
		},
	}
	dg := &fakeDiffGen{diff: "ok"}

	gotRegion := ""
	_, err := GetReview(context.Background(), GetReviewOptions{
		Repo:   "my-repo",
		PRId:   "42",
		Config: configPath,
	}, func(_ context.Context, region, _ string) (codecommit.Client, error) {
		gotRegion = region
		return cc, nil
	}, func() diff.Generator { return dg })
	if err != nil {
		t.Fatalf("GetReview returned error: %v", err)
	}
	if gotRegion != "ap-northeast-1" {
		t.Fatalf("region = %q, want ap-northeast-1 from config", gotRegion)
	}
}

func TestGetReviewSurfacesRepoMappingError(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\n")
	_, err := GetReview(context.Background(), GetReviewOptions{
		Repo:   "unmapped-repo",
		PRId:   "42",
		Config: configPath,
	}, func(context.Context, string, string) (codecommit.Client, error) {
		return &fakeReviewClient{}, nil
	}, func() diff.Generator { return &fakeDiffGen{} })
	if err == nil || !strings.Contains(err.Error(), "no local path mapping") {
		t.Fatalf("error = %v, want repo mapping failure", err)
	}
}

func TestGetReviewSurfacesDiffError(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\nrepoMappings:\n  my-repo: /tmp/my-repo\n")
	cc := &fakeReviewClient{
		metadata: codecommit.PRMetadata{
			SourceBranch:      "feature/x",
			DestinationBranch: "main",
			CreationDate:      time.Now(),
		},
	}
	dg := &fakeDiffGen{err: errors.New("boom")}

	_, err := GetReview(context.Background(), GetReviewOptions{
		Repo:   "my-repo",
		PRId:   "42",
		Config: configPath,
	}, func(context.Context, string, string) (codecommit.Client, error) {
		return cc, nil
	}, func() diff.Generator { return dg })
	if err == nil || !strings.Contains(err.Error(), "generating diff") {
		t.Fatalf("error = %v, want diff generation failure", err)
	}
}
