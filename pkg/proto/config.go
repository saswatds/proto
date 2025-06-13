package proto

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the proto configuration
type Config struct {
	GitHubURL  string `yaml:"github_url"`
	Branch     string `yaml:"branch"`
	RemotePath string `yaml:"remote_path"`
	ProtoDir   string `yaml:"proto_dir"`
	BuildDir   string `yaml:"build_dir"`
}

// getCachePath returns the path to the cache file
func getCachePath(protoDir string) string {
	return filepath.Join(protoDir, ".proto_cache")
}

// LoadConfig loads the proto configuration from .protorc
func LoadConfig() (*Config, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}

	configPath := filepath.Join(workDir, ".protorc")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

// SaveConfig saves the proto configuration to .protorc
func SaveConfig(config *Config) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// Ensure proto directory exists
	if err := os.MkdirAll(config.ProtoDir, 0755); err != nil {
		return fmt.Errorf("error creating proto directory: %v", err)
	}

	// Save config file
	configPath := filepath.Join(workDir, ".protorc")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// SaveCache saves the git head to the cache file
func SaveCache(config *Config, gitHead string) error {
	// Ensure proto directory exists
	if err := os.MkdirAll(config.ProtoDir, 0755); err != nil {
		return fmt.Errorf("error creating proto directory: %v", err)
	}

	// Save cache file
	cacheData := map[string]string{
		"git_head": gitHead,
	}
	cacheYAML, err := yaml.Marshal(cacheData)
	if err != nil {
		return fmt.Errorf("error marshaling cache data: %v", err)
	}

	cachePath := getCachePath(config.ProtoDir)
	if err := os.WriteFile(cachePath, cacheYAML, 0644); err != nil {
		return fmt.Errorf("error writing cache file: %v", err)
	}

	return nil
}

// LoadCache loads the cache data from the cache file
func LoadCache(config *Config) (string, error) {
	cachePath := getCachePath(config.ProtoDir)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("error reading cache file: %v", err)
	}

	var cacheData map[string]string
	if err := yaml.Unmarshal(data, &cacheData); err != nil {
		return "", fmt.Errorf("error parsing cache file: %v", err)
	}

	return cacheData["git_head"], nil
}
