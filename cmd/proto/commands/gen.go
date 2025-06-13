package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saswatds/proto/pkg/proto"
)

// GenCmd handles generating SDKs from proto files
func GenCmd(sdkType string, moduleName string) {
	config, err := proto.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if config.GitHubURL == "" {
		fmt.Println("Error: Configuration not initialized. Run 'proto init' first")
		os.Exit(1)
	}

	// Create build directory if it doesn't exist
	if err := os.MkdirAll(config.BuildDir, 0755); err != nil {
		fmt.Printf("Error creating build directory: %v\n", err)
		os.Exit(1)
	}

	// Get all proto files
	protoFiles, err := filepath.Glob(filepath.Join(config.ProtoDir, "*.proto"))
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Println("Error: No proto files found in", config.ProtoDir)
		fmt.Println("\nPlease ensure:")
		fmt.Println("1. You have run 'proto sync' to download proto files")
		fmt.Println("2. The proto files are in the correct directory:", config.ProtoDir)
		os.Exit(1)
	}

	// Build proto files
	switch sdkType {
	case "go":
		// Check if protoc-gen-go is installed
		if _, err := exec.LookPath("protoc-gen-go"); err != nil {
			fmt.Println("Error: protoc-gen-go not found")
			fmt.Println("\nPlease install it using:")
			fmt.Println("go install google.golang.org/protobuf/cmd/protoc-gen-go@latest")
			os.Exit(1)
		}

		// Check if protoc-gen-go-grpc is installed
		if _, err := exec.LookPath("protoc-gen-go-grpc"); err != nil {
			fmt.Println("Error: protoc-gen-go-grpc not found")
			fmt.Println("\nPlease install it using:")
			fmt.Println("go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest")
			os.Exit(1)
		}

		// Read go.mod to get module path
		goModPath := filepath.Join(".", "go.mod")
		goModContent, err := os.ReadFile(goModPath)
		if err != nil {
			fmt.Println("Error: go.mod file not found")
			fmt.Println("Please ensure you're in a Go project directory with a go.mod file")
			os.Exit(1)
		}

		// Extract module path from go.mod
		var modulePath string
		lines := strings.Split(string(goModContent), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "module ") {
				modulePath = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}

		if modulePath == "" {
			fmt.Println("Error: Could not find module path in go.mod")
			os.Exit(1)
		}

		fmt.Printf("Using module path from go.mod: %s\n", modulePath)

		// Add go_package option to proto files
		var tmpProtoFiles []string
		for _, protoFile := range protoFiles {
			data, err := os.ReadFile(protoFile)
			if err != nil {
				fmt.Printf("Error reading proto file %s: %v\n", protoFile, err)
				continue
			}

			// Create a temporary file in the same directory as the original
			dir := filepath.Dir(protoFile)
			baseName := filepath.Base(protoFile)
			tmpFile := filepath.Join(dir, "pb_"+baseName)
			defer os.Remove(tmpFile)

			// Add or update go_package option
			content := string(data)
			lines := strings.Split(content, "\n")
			var newLines []string
			packageLineFound := false

			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "option go_package") {
					// Skip existing go_package option
					continue
				}
				newLines = append(newLines, line)
				if strings.HasPrefix(strings.TrimSpace(line), "package ") && !packageLineFound {
					newLines = append(newLines, fmt.Sprintf("option go_package = \"%s\";", modulePath))
					packageLineFound = true
				}
			}

			// Write to temp file
			if err := os.WriteFile(tmpFile, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
				fmt.Printf("Error writing temp proto file: %v\n", err)
				continue
			}

			tmpProtoFiles = append(tmpProtoFiles, tmpFile)
			fmt.Printf("Created temporary proto file: %s\n", tmpFile)
		}

		// Generate Go SDK
		args := []string{
			"--go_out=" + config.BuildDir,
			"--go_opt=paths=source_relative",
			"--go-grpc_out=" + config.BuildDir,
			"--go-grpc_opt=paths=source_relative",
			"-I", config.ProtoDir,
		}
		args = append(args, tmpProtoFiles...)
		cmd := exec.Command("protoc", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error generating Go SDK: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Go SDK (with gRPC) generated successfully in", config.BuildDir)

	case "python":
		// Check if Python protobuf is installed
		pythonCmd := exec.Command("python3", "-c", "import google.protobuf")
		if err := pythonCmd.Run(); err != nil {
			fmt.Println("Error: Python protobuf package not found")
			fmt.Println("\nPlease install it using:")
			fmt.Println("pip install protobuf grpcio grpcio-tools")
			os.Exit(1)
		}

		// Check if mypy-protobuf is installed
		pythonCmd = exec.Command("python3", "-c", "import mypy_protobuf")
		if err := pythonCmd.Run(); err != nil {
			fmt.Println("Error: mypy-protobuf not found")
			fmt.Println("\nPlease install it using:")
			fmt.Println("pip install mypy-protobuf")
			os.Exit(1)
		}

		// Generate Python SDK with gRPC
		for _, protoFile := range protoFiles {
			args := []string{
				"--python_out=" + config.BuildDir,
				"--grpc_python_out=" + config.BuildDir,
				"--mypy_out=" + config.BuildDir,
				"-I", config.ProtoDir,
				protoFile,
			}

			// Log the args
			fmt.Println("Generating Python SDK with args:", args)

			cmd := exec.Command("protoc", args...)

			// Capture both stdout and stderr
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error generating Python SDK for %s:\n", filepath.Base(protoFile))
				fmt.Println(string(output))
				fmt.Println("\nCommon issues:")
				fmt.Println("1. Missing Python protobuf or gRPC packages")
				fmt.Println("2. Syntax errors in proto file")
				fmt.Println("3. Invalid import paths")
				os.Exit(1)
			}
		}
		fmt.Println("Python SDK (with gRPC) generated successfully in", config.BuildDir)

	default:
		fmt.Println("Error: Unsupported SDK type. Use 'go' or 'python'")
		os.Exit(1)
	}
}
