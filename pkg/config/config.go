package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config represents persistent and runtime settings for MemGit.
type Config struct {
	ServerURL      string `json:"server_url" yaml:"server_url"`
	MembussAPI     string `json:"membuss_api" yaml:"membuss_api"`
	MembussGateway string `json:"membuss_gateway" yaml:"membuss_gateway"`
	DataDir        string `json:"data_dir" yaml:"data_dir"`
	WebDir         string `json:"web_dir" yaml:"web_dir"`
	AuthorName     string `json:"author_name,omitempty" yaml:"author_name,omitempty"`
	AuthorEmail    string `json:"author_email,omitempty" yaml:"author_email,omitempty"`
}

// DefaultConfig returns the default configuration settings.
func DefaultConfig() *Config {
	return &Config{
		ServerURL:      "http://localhost:8500",
		MembussAPI:     "http://127.0.0.1:5004",
		MembussGateway: "https://gateway.membuss.dpdns.org",
		DataDir:        "./data",
		WebDir:         "./web/dist",
		AuthorName:     "Membuss Developer",
		AuthorEmail:    "dev@membuss.network",
	}
}

// ConfigDir returns the default user config directory (~/.memgit).
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".memgit")
}

// ConfigPath returns the path to the config file (~/.memgit/config.json).
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// Load reads the configuration from disk, applying environment variable overrides.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := ConfigPath()
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Environment variable overrides
	if env := os.Getenv("MEMGIT_SERVER_URL"); env != "" {
		cfg.ServerURL = strings.TrimRight(env, "/")
	}
	if env := os.Getenv("MEMBUSS_API_URL"); env != "" {
		cfg.MembussAPI = strings.TrimRight(env, "/")
	}
	if env := os.Getenv("MEMBUSS_GW_URL"); env != "" {
		cfg.MembussGateway = strings.TrimRight(env, "/")
	}
	if env := os.Getenv("MEMGIT_DATA_DIR"); env != "" {
		cfg.DataDir = env
	}

	return cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0o644)
}

// SetKey updates a single configuration key.
func (c *Config) SetKey(key, value string) error {
	switch strings.ToLower(key) {
	case "server", "server_url", "serverurl":
		c.ServerURL = strings.TrimRight(value, "/")
	case "api", "api_url", "apiurl", "membuss_api", "membussapi":
		c.MembussAPI = strings.TrimRight(value, "/")
	case "gw", "gateway", "gw_url", "gwurl", "membuss_gw", "membuss_gateway":
		c.MembussGateway = strings.TrimRight(value, "/")
	case "datadir", "data_dir":
		c.DataDir = value
	case "webdir", "web_dir":
		c.WebDir = value
	case "author", "author_name", "name":
		c.AuthorName = value
	case "email", "author_email":
		c.AuthorEmail = value
	default:
		return fmt.Errorf("unknown configuration key %q (valid keys: server_url, membuss_api, membuss_gateway, data_dir, web_dir, author_name, author_email)", key)
	}
	return c.Save()
}
