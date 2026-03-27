package main

import "testing"

func TestRunOpen_NoArgs(t *testing.T) {
	err := runOpen([]string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}

func TestRunOpen_RepoWithoutPRId(t *testing.T) {
	err := runOpen([]string{"--repo", "my-repo"})
	if err == nil {
		t.Fatal("expected error for missing --pr-id")
	}
}

func TestRunOpen_PRIdWithoutRepo(t *testing.T) {
	err := runOpen([]string{"--pr-id", "123"})
	if err == nil {
		t.Fatal("expected error for missing --repo")
	}
}

func TestRunOpen_InvalidURL(t *testing.T) {
	err := runOpen([]string{"https://github.com/user/repo/pull/1"})
	if err == nil {
		t.Fatal("expected error for non-CodeCommit URL")
	}
}
