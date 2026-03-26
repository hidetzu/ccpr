package config

// Config holds application configuration loaded from file.
type Config struct {
	RepoMappings map[string]string `yaml:"repoMappings"`
}

// Load searches for a configuration file and returns the parsed Config.
// Search order:
//  1. explicit path (if non-empty)
//  2. .ccpr.yaml in current directory
//  3. ~/.config/ccpr/config.yaml
func Load(path string) (*Config, error) {
	panic("not implemented")
}

// ResolveRepoPath returns the local filesystem path for a CodeCommit repository name.
// Returns an error if no mapping is configured for the given name.
func (c *Config) ResolveRepoPath(repoName string) (string, error) {
	panic("not implemented")
}
