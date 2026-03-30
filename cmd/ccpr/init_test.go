package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit_CreatesConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("AWS_PROFILE", "")

	err := runInit([]string{"--profile", "test-profile", "--region", "us-west-2"})
	if err != nil {
		t.Fatalf("runInit() unexpected error: %v", err)
	}

	configPath := filepath.Join(dir, ".config", "ccpr", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "profile: test-profile") {
		t.Errorf("config missing profile, got:\n%s", content)
	}
	if !strings.Contains(content, "region: us-west-2") {
		t.Errorf("config missing region, got:\n%s", content)
	}
}

func TestRunInit_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "ccpr")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit([]string{})
	if err == nil {
		t.Fatal("expected error when config exists")
	}
	if !strings.Contains(err.Error(), "already exists") || !strings.Contains(err.Error(), "ccpr init --force") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunInit_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("AWS_PROFILE", "")

	configDir := filepath.Join(dir, ".config", "ccpr")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := runInit([]string{"--force", "--profile", "new-profile", "--region", "eu-west-1"})
	if err != nil {
		t.Fatalf("runInit(--force) unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "new-profile") {
		t.Errorf("config not overwritten, got:\n%s", string(data))
	}
}

func TestRunInit_DetectsAWSProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("AWS_PROFILE", "env-profile")

	err := runInit([]string{"--region", "us-east-1"})
	if err != nil {
		t.Fatalf("runInit() unexpected error: %v", err)
	}

	configPath := filepath.Join(dir, ".config", "ccpr", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "profile: env-profile") {
		t.Errorf("expected AWS_PROFILE detection, got:\n%s", string(data))
	}
}
