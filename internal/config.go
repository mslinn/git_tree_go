package internal

import (
  "fmt"
  "os"
  "path/filepath"
  "strconv"
  "strings"

  "gopkg.in/yaml.v2"
)

// Config represents the git-tree configuration.
type Config struct {
  GitTimeout   int      `yaml:"git_timeout"`
  Verbosity    int      `yaml:"verbosity"`
  DefaultRoots []string `yaml:"default_roots"`
}

// NewConfig creates a new Config with default values.
// It loads configuration in this order of precedence:
// 1. Environment variables (GIT_TREE_*)
// 2. User config file (~/.treeconfig.yml)
// 3. Default values
func NewConfig() *Config {
  config := &Config{
    GitTimeout:   300,
    Verbosity:    LogNormal,
    DefaultRoots: []string{"sites", "sitesUbuntu", "work"},
  }

  // Try to load from config file
  config.loadFromFile()

  // Override with environment variables
  config.loadFromEnv()

  return config
}

// loadFromFile loads configuration from ~/.treeconfig.yml
func (c *Config) loadFromFile() {
  home, err := os.UserHomeDir()
  if err != nil {
    return // Silently fail if we can't get home directory
  }

  configPath := filepath.Join(home, ".treeconfig.yml")
  data, err := os.ReadFile(configPath)
  if err != nil {
    return // Silently fail if config file doesn't exist
  }

  // Parse YAML
  if err := yaml.Unmarshal(data, c); err != nil {
    fmt.Fprintf(os.Stderr, "Warning: failed to parse config file: %v\n", err)
  }
}

// loadFromEnv loads configuration from environment variables.
// Environment variables must be prefixed with GIT_TREE_ and be in uppercase.
func (c *Config) loadFromEnv() {
  if val := os.Getenv("GIT_TREE_GIT_TIMEOUT"); val != "" {
    if timeout, err := strconv.Atoi(val); err == nil {
      c.GitTimeout = timeout
    }
  }

  if val := os.Getenv("GIT_TREE_VERBOSITY"); val != "" {
    if verbosity, err := strconv.Atoi(val); err == nil {
      c.Verbosity = verbosity
    }
  }

  if val := os.Getenv("GIT_TREE_DEFAULT_ROOTS"); val != "" {
    c.DefaultRoots = strings.Fields(val)
  }
}

// SaveToFile saves the configuration to ~/.treeconfig.yml
func (c *Config) SaveToFile() error {
  home, err := os.UserHomeDir()
  if err != nil {
    return fmt.Errorf("failed to get home directory: %w", err)
  }

  configPath := filepath.Join(home, ".treeconfig.yml")

  data, err := yaml.Marshal(c)
  if err != nil {
    return fmt.Errorf("failed to marshal config: %w", err)
  }

  if err := os.WriteFile(configPath, data, 0644); err != nil {
    return fmt.Errorf("failed to write config file: %w", err)
  }

  return nil
}
