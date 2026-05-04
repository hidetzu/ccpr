package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hidetzu/ccpr/internal/codecommit"
)

type fakeCodeCommitClient struct {
	repo   string
	status string
	prs    []codecommit.PRSummary
}

func (f *fakeCodeCommitClient) GetPRMetadata(context.Context, string, string) (codecommit.PRMetadata, error) {
	return codecommit.PRMetadata{}, nil
}

func (f *fakeCodeCommitClient) GetPRComments(context.Context, string, string, string, string) ([]codecommit.Comment, error) {
	return nil, nil
}

func (f *fakeCodeCommitClient) ListPRs(_ context.Context, repo, status string) ([]codecommit.PRSummary, error) {
	f.repo = repo
	f.status = status
	return f.prs, nil
}

func (f *fakeCodeCommitClient) PostComment(context.Context, string, string, string, string, string) (codecommit.PostCommentResult, error) {
	return codecommit.PostCommentResult{}, nil
}

func (f *fakeCodeCommitClient) CreatePR(context.Context, string, string, string, string, string) (codecommit.CreatePRResult, error) {
	return codecommit.CreatePRResult{}, nil
}

func TestListPullRequestsDefaultsStatusAndFormatsOutput(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\n")
	created := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	fake := &fakeCodeCommitClient{
		prs: []codecommit.PRSummary{{
			PRId:              "42",
			Title:             "Add MCP",
			AuthorARN:         "arn:aws:iam::123456789012:user/example",
			SourceBranch:      "feature/mcp",
			DestinationBranch: "main",
			Status:            "OPEN",
			CreationDate:      created,
		}},
	}

	got, err := ListPullRequests(context.Background(), ListPullRequestsOptions{
		Repo:   "my-repo",
		Config: configPath,
	}, func(ctx context.Context, region, profile string) (codecommit.Client, error) {
		if region != "ap-northeast-1" {
			t.Fatalf("region = %q, want ap-northeast-1", region)
		}
		if profile != "" {
			t.Fatalf("profile = %q, want empty", profile)
		}
		return fake, nil
	})
	if err != nil {
		t.Fatalf("ListPullRequests returned error: %v", err)
	}

	if fake.repo != "my-repo" {
		t.Fatalf("repo = %q, want my-repo", fake.repo)
	}
	if fake.status != "open" {
		t.Fatalf("status = %q, want open", fake.status)
	}
	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0].CreationDate != "2026-04-01T10:00:00Z" {
		t.Fatalf("CreationDate = %q", got[0].CreationDate)
	}
}

func TestListPullRequestsRejectsInvalidStatus(t *testing.T) {
	_, err := ListPullRequests(context.Background(), ListPullRequestsOptions{
		Repo:   "my-repo",
		Status: "merged",
	}, nil)
	if err == nil {
		t.Fatal("ListPullRequests returned nil error")
	}
}

func writeListConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
