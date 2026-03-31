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
	Region       string            `yaml:"region"`
	RepoMappings map[string]string `yaml:"repoMappings"`
}

// ResolveRegion returns the AWS region to use.
// Priority: flagRegion > config file > "".
func (c *Config) ResolveRegion(flagRegion string) string {
	if flagRegion != "" {
		return flagRegion
	}
	return c.Region
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

// Load searches for a configuration file and returns the parsed Config
// along with the resolved file path.
// Search order:
//  1. explicit path (if non-empty)
//  2. .ccpr.yaml in current directory
//  3. ~/.config/ccpr/config.yaml
func Load(path string) (*Config, string, error) {
	if path != "" {
		cfg, err := loadFrom(path)
		return cfg, path, err
	}

	candidates := []string{".ccpr.yaml"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "ccpr", "config.yaml"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			cfg, err := loadFrom(c)
			return cfg, c, err
		}
	}

	return nil, "", fmt.Errorf("config file not found (searched .ccpr.yaml and ~/.config/ccpr/config.yaml)")
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

// DefaultPath returns the default config file path: ~/.config/ccpr/config.yaml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "ccpr", "config.yaml"), nil
}

// Write generates a config file with comments and writes it to the given path.
// It creates parent directories as needed.
func Write(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	content := generateConfigYAML(cfg)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing config %s: %w", path, err)
	}
	return nil
}

func generateConfigYAML(cfg *Config) string {
	var s string
	s += fmt.Sprintf("profile: %s\n", cfg.Profile)
	s += fmt.Sprintf("region: %s\n", cfg.Region)
	s += "\n"
	s += "repoMappings:\n"
	s += "  # (optional) Map CodeCommit repository to local path\n"
	s += "  # format:\n"
	s += "  #   <repo-name>: <local-path>\n"
	s += "  # example:\n"
	s += "  #   my-repo: ~/src/my-repo\n"
	return s
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
