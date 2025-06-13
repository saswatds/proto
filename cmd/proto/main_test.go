package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// Run tests
	os.Exit(m.Run())
}

func TestInitCommand(t *testing.T) {
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
		args    []string
		wantErr bool
	}{
		{
			name:    "missing url",
			args:    []string{"init"},
			wantErr: true,
		},
		{
			name:    "valid init",
			args:    []string{"init", "--url", "https://github.com/example/proto", "--branch", "main", "--remote-path", "./protos"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = append([]string{"proto"}, tt.args...)

			// Run the command
			main()

			// Check if .protorc exists
			configPath := filepath.Join(tempDir, ".protorc")
			if _, err := os.Stat(configPath); (err != nil) != tt.wantErr {
				t.Errorf("Config file existence = %v, want %v", err == nil, !tt.wantErr)
			}
		})
	}
}

func TestSyncCommand(t *testing.T) {
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

	// Initialize with a test repository
	os.Args = []string{"proto", "init", "--url", "https://github.com/example/proto", "--branch", "main", "--remote-path", "./protos"}
	main()

	// Test sync command
	os.Args = []string{"proto", "sync"}
	main()

	// Check if output directory exists
	outputDir := filepath.Join(tempDir, "protos")
	if _, err := os.Stat(outputDir); err != nil {
		t.Errorf("Output directory not created: %v", err)
	}
}

func TestBuildCommand(t *testing.T) {
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

	// Initialize with a test repository
	os.Args = []string{"proto", "init", "--url", "https://github.com/example/proto", "--branch", "main", "--remote-path", "./protos"}
	main()

	// Test cases
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "missing language",
			args:    []string{"build"},
			wantErr: true,
		},
		{
			name:    "invalid language",
			args:    []string{"build", "invalid"},
			wantErr: true,
		},
		{
			name:    "build go",
			args:    []string{"build", "go"},
			wantErr: false,
		},
		{
			name:    "build python",
			args:    []string{"build", "python"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			// Set test args
			os.Args = append([]string{"proto"}, tt.args...)

			// Run the command
			main()
		})
	}
}

func TestCommandNotFound(t *testing.T) {
	// Test unknown command
	os.Args = []string{"proto", "unknown"}
	main()
}
