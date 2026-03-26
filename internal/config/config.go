package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration loaded from file.
type Config struct {
	Profile      string            `yaml:"profile"`
	RepoMappings map[string]string `yaml:"repoMappings"`
}

// ResolveProfile returns the AWS profile to use.
// Priority: flagProfile > config file > AWS_PROFILE env > "default".
func (c *Config) ResolveProfile(flagProfile string) string {
	if flagProfile != "" {
		return flagProfile
	}
	if c.Profile != "" {
		return c.Profile
	}
	if env := os.Getenv("AWS_PROFILE"); env != "" {
		return env
	}
	return ""
}

// Load searches for a configuration file and returns the parsed Config.
// Search order:
//  1. explicit path (if non-empty)
//  2. .ccpr.yaml in current directory
//  3. ~/.config/ccpr/config.yaml
func Load(path string) (*Config, error) {
	if path != "" {
		return loadFrom(path)
	}

	candidates := []string{".ccpr.yaml"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "ccpr", "config.yaml"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return loadFrom(c)
		}
	}

	return nil, fmt.Errorf("config file not found (searched .ccpr.yaml and ~/.config/ccpr/config.yaml)")
}

// ResolveRepoPath returns the local filesystem path for a CodeCommit repository name.
// Returns an error if no mapping is configured for the given name.
func (c *Config) ResolveRepoPath(repoName string) (string, error) {
	path, ok := c.RepoMappings[repoName]
	if !ok {
		return "", fmt.Errorf("no local path mapping for repository %q", repoName)
	}
	return path, nil
}

func loadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}
