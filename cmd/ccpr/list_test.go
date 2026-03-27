package main

import (
	"testing"
)

func TestRunList_NoRepo(t *testing.T) {
	err := runList([]string{})
	if err == nil {
		t.Fatal("expected error for missing --repo")
	}
}

func TestRunList_RepoRequired(t *testing.T) {
	err := runList([]string{"--status", "open"})
	if err == nil {
		t.Fatal("expected error for missing --repo")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a long title that needs truncation", 20, "this is a long ti..."},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
