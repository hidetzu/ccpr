package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	writeFile(t, cfgPath, `repoMappings:
  my-repo: /work/src/my-repo
  other: /tmp/other
`)

	cfg, _, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load(%q) unexpected error: %v", cfgPath, err)
	}
	if len(cfg.RepoMappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(cfg.RepoMappings))
	}
	if cfg.RepoMappings["my-repo"] != "/work/src/my-repo" {
		t.Errorf("my-repo = %q, want /work/src/my-repo", cfg.RepoMappings["my-repo"])
	}
}

func TestLoad_DotCcprYaml(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".ccpr.yaml"), `repoMappings:
  local-repo: /home/user/local-repo
`)

	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := Load("")
	if err != nil {
		t.Fatalf("Load(\"\") unexpected error: %v", err)
	}
	if cfg.RepoMappings["local-repo"] != "/home/user/local-repo" {
		t.Errorf("local-repo = %q, want /home/user/local-repo", cfg.RepoMappings["local-repo"])
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	dir := t.TempDir()

	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// Override home to empty dir so ~/.config/ccpr/config.yaml also doesn't exist
	t.Setenv("HOME", dir)

	_, _, err := Load("")
	if err == nil {
		t.Fatal("Load(\"\") expected error for missing config, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")
	writeFile(t, cfgPath, `[[[invalid`)

	_, _, err := Load(cfgPath)
	if err == nil {
		t.Fatal("Load() expected error for invalid YAML, got nil")
	}
}

func TestLoad_ResolvedPathMatchesActualFile(t *testing.T) {
	dir := t.TempDir()

	// Create both .ccpr.yaml and ~/.config/ccpr/config.yaml
	dotCcpr := filepath.Join(dir, ".ccpr.yaml")
	writeFile(t, dotCcpr, "profile: local\n")

	defaultDir := filepath.Join(dir, ".config", "ccpr")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(defaultDir, "config.yaml"), "profile: global\n")

	t.Setenv("HOME", dir)
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	_ = os.Chdir(dir)

	// Load("") should pick .ccpr.yaml and report that path
	cfg, resolvedPath, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "local" {
		t.Errorf("profile = %q, want local", cfg.Profile)
	}
	if resolvedPath != ".ccpr.yaml" {
		t.Errorf("resolvedPath = %q, want .ccpr.yaml", resolvedPath)
	}
}

func TestResolveRepoPath(t *testing.T) {
	cfg := &Config{
		RepoMappings: map[string]string{
			"my-repo": "/work/src/my-repo",
		},
	}

	path, err := cfg.ResolveRepoPath("my-repo")
	if err != nil {
		t.Fatalf("ResolveRepoPath(my-repo) unexpected error: %v", err)
	}
	if path != "/work/src/my-repo" {
		t.Errorf("got %q, want /work/src/my-repo", path)
	}
}

func TestResolveRepoPath_NotMapped(t *testing.T) {
	cfg := &Config{
		RepoMappings: map[string]string{},
	}

	_, err := cfg.ResolveRepoPath("unknown")
	if err == nil {
		t.Fatal("ResolveRepoPath(unknown) expected error, got nil")
	}
}

func TestResolveRegion(t *testing.T) {
	cfg := &Config{Region: "ap-northeast-1"}

	// Flag takes priority
	if got := cfg.ResolveRegion("us-east-1"); got != "us-east-1" {
		t.Errorf("flag priority: got %q, want us-east-1", got)
	}
	// Config fallback
	if got := cfg.ResolveRegion(""); got != "ap-northeast-1" {
		t.Errorf("config fallback: got %q, want ap-northeast-1", got)
	}
	// Empty config, empty flag
	empty := &Config{}
	if got := empty.ResolveRegion(""); got != "" {
		t.Errorf("empty: got %q, want empty", got)
	}
}

func TestLoad_ExpandsHomeInRepoMappings(t *testing.T) {
	dir := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgPath := filepath.Join(dir, "config.yaml")
	// Note: bare `~` is YAML null, so the literal "~" string must be quoted.
	writeFile(t, cfgPath, `repoMappings:
  with-tilde-slash: ~/src/repo
  bare-tilde: '~'
  absolute: /work/src/abs
  relative: rel/path
  other-user: ~someone/else
`)

	cfg, _, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load(%q) unexpected error: %v", cfgPath, err)
	}

	cases := map[string]string{
		"with-tilde-slash": filepath.Join(home, "src/repo"),
		"bare-tilde":       home,
		"absolute":         "/work/src/abs",
		"relative":         "rel/path",
		"other-user":       "~someone/else",
	}
	for name, want := range cases {
		if got := cfg.RepoMappings[name]; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
