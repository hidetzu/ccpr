package main

import "testing"

func TestRunOpen_NoRepo(t *testing.T) {
	err := runOpen([]string{"--pr-id", "123"})
	if err == nil {
		t.Fatal("expected error for missing --repo")
	}
}

func TestRunOpen_NoPRId(t *testing.T) {
	err := runOpen([]string{"--repo", "my-repo"})
	if err == nil {
		t.Fatal("expected error for missing --pr-id")
	}
}

func TestRunOpen_NoArgs(t *testing.T) {
	err := runOpen([]string{})
	if err == nil {
		t.Fatal("expected error for missing args")
	}
}
