package proto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "proto-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test cases
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				GitHubURL:    "https://github.com/example/repo",
				Branch:       "main",
				RemotePath:   "api/proto",
				ProtoDir:     "./proto",
				BuildDir:     "./gen",
				LastCommitID: "abc123",
			},
			wantErr: false,
		},
		{
			name: "empty config",
			config: &Config{
				GitHubURL:    "",
				Branch:       "",
				RemotePath:   "",
				ProtoDir:     "",
				BuildDir:     "",
				LastCommitID: "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test SaveConfig
			err := SaveConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Test LoadConfig
			got, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.GitHubURL != tt.config.GitHubURL {
					t.Errorf("LoadConfig() GitHubURL = %v, want %v", got.GitHubURL, tt.config.GitHubURL)
				}
				if got.Branch != tt.config.Branch {
					t.Errorf("LoadConfig() Branch = %v, want %v", got.Branch, tt.config.Branch)
				}
				if got.RemotePath != tt.config.RemotePath {
					t.Errorf("LoadConfig() RemotePath = %v, want %v", got.RemotePath, tt.config.RemotePath)
				}
				if got.ProtoDir != tt.config.ProtoDir {
					t.Errorf("LoadConfig() ProtoDir = %v, want %v", got.ProtoDir, tt.config.ProtoDir)
				}
				if got.BuildDir != tt.config.BuildDir {
					t.Errorf("LoadConfig() BuildDir = %v, want %v", got.BuildDir, tt.config.BuildDir)
				}
				if got.LastCommitID != tt.config.LastCommitID {
					t.Errorf("LoadConfig() LastCommitID = %v, want %v", got.LastCommitID, tt.config.LastCommitID)
				}
			}
		})
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "proto-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test loading non-existent config
	config, err := LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig() error = %v, want nil", err)
	}
	if config == nil {
		t.Error("LoadConfig() returned nil config, want empty config")
	}
	if config.GitHubURL != "" || config.Branch != "" || config.RemotePath != "" || config.ProtoDir != "" || config.BuildDir != "" || config.LastCommitID != "" {
		t.Error("LoadConfig() returned non-empty config for non-existent file")
	}
}

func TestSaveConfigInvalidPath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "proto-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override home directory for testing with an invalid path
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", filepath.Join(tempDir, "nonexistent"))
	defer os.Setenv("HOME", originalHome)

	config := &Config{
		GitHubURL:    "https://github.com/example/repo",
		Branch:       "main",
		RemotePath:   "api/proto",
		ProtoDir:     "./proto",
		BuildDir:     "./gen",
		LastCommitID: "abc123",
	}

	// Test saving to invalid path
	err = SaveConfig(config)
	if err == nil {
		t.Error("SaveConfig() error = nil, want error")
	}
}
