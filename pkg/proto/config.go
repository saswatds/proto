package proto

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the tool's configuration
type Config struct {
	GitHubURL  string `yaml:"github_url"`
	Branch     string `yaml:"branch"`
	RemotePath string `yaml:"remote_path"`
	ProtoDir   string `yaml:"proto_dir"`
	BuildDir   string `yaml:"build_dir"`
	GitHead    string `yaml:"gitHead"`
}

// LoadConfig loads the configuration from .protorc
func LoadConfig() (*Config, error) {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}

	configPath := filepath.Join(workDir, ".protorc")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to .protorc
func SaveConfig(config *Config) error {
	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	configPath := filepath.Join(workDir, ".protorc")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(configPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}
