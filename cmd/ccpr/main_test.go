package main

import (
	"testing"
)

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

func TestRunReview_MutuallyExclusiveFlags(t *testing.T) {
	err := runReview([]string{"-json", "-patch", "https://ap-northeast-1.console.aws.amazon.com/codesuite/codecommit/repositories/repo/pull-requests/1"})
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	want := "--json and --patch are mutually exclusive"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
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
