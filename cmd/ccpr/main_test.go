package main

import (
	"testing"
)

func TestRun_Version(t *testing.T) {
	err := run([]string{"--version"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRun_VersionSubcommand(t *testing.T) {
	err := run([]string{"version"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRun_NoArgs(t *testing.T) {
	err := run([]string{})
	if err == nil {
		t.Fatal("expected error for no arguments")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	err := run([]string{"foobar"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestRunReview_InvalidFormat(t *testing.T) {
	err := runReview([]string{"-format", "xml", "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/repo/pull-requests/1"})
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestRunReview_NoInput(t *testing.T) {
	err := runReview([]string{})
	if err == nil {
		t.Fatal("expected error for missing PR URL or flags")
	}
}

func TestRunReview_PartialFlags(t *testing.T) {
	err := runReview([]string{"-repo", "my-repo", "-region", "us-east-1"})
	if err == nil {
		t.Fatal("expected error when --pr-id is missing")
	}
}

func TestRunReview_InvalidURL(t *testing.T) {
	err := runReview([]string{"https://github.com/user/repo/pull/1"})
	if err == nil {
		t.Fatal("expected error for non-CodeCommit URL")
	}
}

func TestReorderArgs(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "format flag before URL",
			in:   []string{"-format", "json", "https://example.com"},
			want: []string{"-format", "json", "https://example.com"},
		},
		{
			name: "format flag after URL",
			in:   []string{"https://example.com", "--format", "json"},
			want: []string{"--format", "json", "https://example.com"},
		},
		{
			name: "value flag after URL",
			in:   []string{"https://example.com", "--profile", "myprof"},
			want: []string{"--profile", "myprof", "https://example.com"},
		},
		{
			name: "mixed",
			in:   []string{"https://example.com", "--format", "patch", "--profile", "myprof"},
			want: []string{"--format", "patch", "--profile", "myprof", "https://example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reorderArgs(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("reorderArgs(%v) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("reorderArgs(%v)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRunReview_FormatAfterURL(t *testing.T) {
	// Invalid format after URL should be detected
	err := runReview([]string{"https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/repo/pull-requests/1", "-format", "xml"})
	if err == nil {
		t.Fatal("expected error for invalid format after URL")
	}
}

func TestRunList_InvalidFormat(t *testing.T) {
	err := runList([]string{"--repo", "my-repo", "--format", "patch"})
	if err == nil {
		t.Fatal("expected error for invalid format in list")
	}
}
