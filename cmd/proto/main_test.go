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

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "proto_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test proto file
	testProto := `
syntax = "proto3";

package test;

service TestService {
    rpc TestMethod (TestRequest) returns (TestResponse);
}

message TestRequest {
    string message = 1;
}

message TestResponse {
    string reply = 1;
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "test.proto"), []byte(testProto), 0644); err != nil {
		t.Fatalf("Failed to write test proto file: %v", err)
	}

	// Create test go.mod
	goMod := `module github.com/test/proto`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to write go.mod file: %v", err)
	}

	// Create .protorc
	protorc := `{
		"build_dir": "build",
		"proto_dir": "."
	}`
	if err := os.WriteFile(filepath.Join(tempDir, ".protorc"), []byte(protorc), 0644); err != nil {
		t.Fatalf("Failed to write .protorc file: %v", err)
	}

	// Create build directory
	if err := os.MkdirAll(filepath.Join(tempDir, "build"), 0755); err != nil {
		t.Fatalf("Failed to create build directory: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func TestGenerateSDK(t *testing.T) {
	// Setup test environment
	tempDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test cases
	tests := []struct {
		name           string
		args           []string
		expectedFiles  []string
		expectedOutput string
	}{
		{
			name: "Generate Go SDK",
			args: []string{"gen", "go"},
			expectedFiles: []string{
				"build/test.pb.go",
				"build/test_grpc.pb.go",
			},
			expectedOutput: "Go SDK (with gRPC) generated successfully in build\n",
		},
		{
			name: "Generate Python SDK",
			args: []string{"gen", "python"},
			expectedFiles: []string{
				"build/test_pb2.py",
				"build/test_pb2_grpc.py",
			},
			expectedOutput: "Python SDK (with gRPC) generated successfully in build\n",
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

			// Verify generated files
			for _, file := range tt.expectedFiles {
				filePath := filepath.Join(tempDir, file)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Expected file %s was not generated", file)
				}
			}
		})
	}
}
