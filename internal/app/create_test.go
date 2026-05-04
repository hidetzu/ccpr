package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hidetzu/ccpr/internal/codecommit"
)

type fakeCreateClient struct {
	createResult       codecommit.CreatePRResult
	createErr          error
	gotRepo            string
	gotTitle           string
	gotDescription     string
	gotSource          string
	gotDest            string
}

func (f *fakeCreateClient) GetPRMetadata(context.Context, string, string) (codecommit.PRMetadata, error) {
	return codecommit.PRMetadata{}, nil
}

func (f *fakeCreateClient) GetPRComments(context.Context, string, string, string, string) ([]codecommit.Comment, error) {
	return nil, nil
}

func (f *fakeCreateClient) ListPRs(context.Context, string, string) ([]codecommit.PRSummary, error) {
	return nil, nil
}

func (f *fakeCreateClient) PostComment(context.Context, string, string, string, string, string) (codecommit.PostCommentResult, error) {
	return codecommit.PostCommentResult{}, nil
}

func (f *fakeCreateClient) CreatePR(_ context.Context, repo, title, description, source, dest string) (codecommit.CreatePRResult, error) {
	f.gotRepo = repo
	f.gotTitle = title
	f.gotDescription = description
	f.gotSource = source
	f.gotDest = dest
	return f.createResult, f.createErr
}

func TestCreatePullRequestRequiresFields(t *testing.T) {
	cases := []struct {
		name string
		opts CreatePullRequestOptions
		want string
	}{
		{"missing repo", CreatePullRequestOptions{Title: "t", SourceBranch: "s", DestinationBranch: "d"}, "repo is required"},
		{"missing title", CreatePullRequestOptions{Repo: "r", SourceBranch: "s", DestinationBranch: "d"}, "title is required"},
		{"missing source", CreatePullRequestOptions{Repo: "r", Title: "t", DestinationBranch: "d"}, "sourceBranch is required"},
		{"missing dest", CreatePullRequestOptions{Repo: "r", Title: "t", SourceBranch: "s"}, "destinationBranch is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := CreatePullRequest(context.Background(), tc.opts, nil)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestCreatePullRequestRejectsSameBranch(t *testing.T) {
	_, err := CreatePullRequest(context.Background(), CreatePullRequestOptions{
		Repo:              "r",
		Title:             "t",
		SourceBranch:      "main",
		DestinationBranch: "main",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "same as destination") {
		t.Fatalf("error = %v, want same-branch guard", err)
	}
}

func TestCreatePullRequestBuildsResult(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\n")
	cc := &fakeCreateClient{
		createResult: codecommit.CreatePRResult{
			PRId:              "42",
			Title:             "Add feature X",
			SourceBranch:      "feature/x",
			DestinationBranch: "main",
		},
	}

	gotRegion := ""
	got, err := CreatePullRequest(context.Background(), CreatePullRequestOptions{
		Repo:              "my-repo",
		Title:             "Add feature X",
		SourceBranch:      "feature/x",
		DestinationBranch: "main",
		Description:       "details",
		Config:            configPath,
	}, func(_ context.Context, region, _ string) (codecommit.Client, error) {
		gotRegion = region
		return cc, nil
	})
	if err != nil {
		t.Fatalf("CreatePullRequest returned error: %v", err)
	}
	if gotRegion != "ap-northeast-1" {
		t.Fatalf("region = %q", gotRegion)
	}
	if cc.gotRepo != "my-repo" || cc.gotTitle != "Add feature X" || cc.gotDescription != "details" || cc.gotSource != "feature/x" || cc.gotDest != "main" {
		t.Fatalf("CreatePR called with unexpected args: %+v", cc)
	}
	if got.PRId != "42" || got.Repository != "my-repo" {
		t.Fatalf("result = %+v", got)
	}
	wantURL := "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/my-repo/pull-requests/42"
	if got.URL != wantURL {
		t.Fatalf("URL = %q, want %q", got.URL, wantURL)
	}
}

func TestCreatePullRequestRequiresRegion(t *testing.T) {
	configPath := writeListConfig(t, "")
	t.Setenv("AWS_REGION", "")
	t.Setenv("AWS_DEFAULT_REGION", "")
	_, err := CreatePullRequest(context.Background(), CreatePullRequestOptions{
		Repo:              "r",
		Title:             "t",
		SourceBranch:      "s",
		DestinationBranch: "d",
		Config:            configPath,
	}, func(context.Context, string, string) (codecommit.Client, error) {
		return &fakeCreateClient{}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "region is required") {
		t.Fatalf("error = %v, want region guidance", err)
	}
}

func TestCreatePullRequestWrapsAWSFailuresAsSystemError(t *testing.T) {
	configPath := writeListConfig(t, "region: ap-northeast-1\n")
	cc := &fakeCreateClient{createErr: errors.New("boom")}

	_, err := CreatePullRequest(context.Background(), CreatePullRequestOptions{
		Repo:              "r",
		Title:             "t",
		SourceBranch:      "s",
		DestinationBranch: "d",
		Config:            configPath,
	}, func(context.Context, string, string) (codecommit.Client, error) {
		return cc, nil
	})
	if err == nil {
		t.Fatal("CreatePullRequest returned nil error")
	}
	var sysErr *SystemError
	if !errors.As(err, &sysErr) {
		t.Fatalf("error = %T %v, want *SystemError", err, err)
	}
	if !strings.Contains(err.Error(), "creating pull request") {
		t.Fatalf("error message = %q", err.Error())
	}
}
