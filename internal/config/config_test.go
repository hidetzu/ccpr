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

	cfg, err := Load(cfgPath)
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
	t.Cleanup(func() { os.Chdir(orig) })
	os.Chdir(dir)

	cfg, err := Load("")
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
	t.Cleanup(func() { os.Chdir(orig) })
	os.Chdir(dir)

	// Override home to empty dir so ~/.config/ccpr/config.yaml also doesn't exist
	t.Setenv("HOME", dir)

	_, err := Load("")
	if err == nil {
		t.Fatal("Load(\"\") expected error for missing config, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.yaml")
	writeFile(t, cfgPath, `[[[invalid`)

	_, err := Load(cfgPath)
	if err == nil {
		t.Fatal("Load() expected error for invalid YAML, got nil")
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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
