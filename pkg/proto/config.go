package proto

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration for proto repository
type Config struct {
	GitHubURL    string `yaml:"github_url"`
	Branch       string `yaml:"branch"`
	RemotePath   string `yaml:"remote_path"` // Path within the repository containing proto files
	ProtoDir     string `yaml:"proto_dir"`   // Directory for synced proto files
	BuildDir     string `yaml:"build_dir"`   // Directory for generated SDKs
	LastCommitID string `yaml:"last_commit_id"`
}

// LoadConfig loads the configuration from .protorc file
func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".protorc")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves the configuration to .protorc file
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".protorc")
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
