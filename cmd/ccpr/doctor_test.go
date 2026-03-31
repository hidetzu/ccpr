package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCheckConfig_Valid(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("profile: test\nregion: us-east-1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, result := checkConfig(cfgPath)
	if !result.ok {
		t.Fatalf("expected ok, got fail: %s", result.message)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	} else if cfg.Profile != "test" {
		t.Errorf("profile = %q, want test", cfg.Profile)
	}
}

func TestCheckConfig_Missing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Point to a non-existent path in isolated dir
	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	_ = os.Chdir(dir)

	cfg, result := checkConfig("")
	if result.ok {
		t.Fatal("expected fail for missing config")
	}
	if cfg != nil {
		t.Fatal("expected nil config")
	}
	if result.fix == "" {
		t.Error("expected fix suggestion")
	}
}

func TestCheckRepoMapping_ValidGitRepo(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", dir)
	if err := cmd.Run(); err != nil {
		t.Skipf("git not available: %v", err)
	}

	result := checkRepoMapping("test-repo", dir)
	if !result.ok {
		t.Fatalf("expected ok, got fail: %s", result.message)
	}
}

func TestCheckRepoMapping_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	result := checkRepoMapping("test-repo", dir)
	if result.ok {
		t.Fatal("expected fail for non-git directory")
	}
}

func TestCheckRepoMapping_PathNotFound(t *testing.T) {
	result := checkRepoMapping("test-repo", "/nonexistent/path/12345")
	if result.ok {
		t.Fatal("expected fail for missing path")
	}
}

func TestRunDoctor_NoConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	orig, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(orig) })
	_ = os.Chdir(dir)

	err := runDoctor([]string{})
	if err == nil {
		t.Fatal("expected error when config is missing")
	}
}
