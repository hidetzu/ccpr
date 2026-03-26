package codecommit

import "testing"

func TestStripRefsHeads(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"refs/heads/main", "main"},
		{"refs/heads/feature/login", "feature/login"},
		{"main", "main"},
		{"", ""},
	}
	for _, tt := range tests {
		got := stripRefsHeads(tt.input)
		if got != tt.want {
			t.Errorf("stripRefsHeads(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestDeref(t *testing.T) {
	s := "hello"
	if got := deref(&s); got != "hello" {
		t.Errorf("deref(&%q) = %q", s, got)
	}
	if got := deref(nil); got != "" {
		t.Errorf("deref(nil) = %q, want empty", got)
	}
}
