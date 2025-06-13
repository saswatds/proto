package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/saswatds/proto/pkg/proto"
)

func printHelp() {
	fmt.Println("Proto CLI Tool - A command-line tool for managing and syncing Protocol Buffer files")
	fmt.Println("\nUsage:")
	fmt.Println("  proto <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  init    - Initialize proto configuration")
	fmt.Println("  sync    - Sync proto files from repository")
	fmt.Println("  build   - Build SDKs (go|python)")
	fmt.Println("  help    - Show this help message")
	fmt.Println("\nOptions:")
	fmt.Println("\ninit:")
	fmt.Println("  --url string         GitHub repository URL (required)")
	fmt.Println("  --branch string      Branch name (default: main)")
	fmt.Println("  --remote-path string Path within the repository containing proto files")
	fmt.Println("  --proto-dir string   Directory for synced proto files (default: ./proto)")
	fmt.Println("  --build-dir string   Directory for generated SDKs (default: ./gen)")
	fmt.Println("\nbuild:")
	fmt.Println("  [go|python]          Specify the target language for SDK generation")
	fmt.Println("\nExamples:")
	fmt.Println("  proto init --url https://github.com/example/proto-files --branch main")
	fmt.Println("  proto sync")
	fmt.Println("  proto build go")
	fmt.Println("  proto build python")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "init":
		initCmd()
	case "sync":
		syncCmd()
	case "build":
		buildCmd()
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'proto help' for usage information")
		os.Exit(1)
	}
}

func initCmd() {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	githubURL := initFlags.String("url", "", "GitHub repository URL")
	branch := initFlags.String("branch", "main", "Branch name")
	remotePath := initFlags.String("remote-path", "", "Path within the repository containing proto files")
	protoDir := initFlags.String("proto-dir", "./proto", "Directory for synced proto files")
	buildDir := initFlags.String("build-dir", "./gen", "Directory for generated SDKs")

	initFlags.Parse(os.Args[2:])

	if *githubURL == "" {
		fmt.Println("Error: GitHub URL is required")
		os.Exit(1)
	}

	config := &proto.Config{
		GitHubURL:  *githubURL,
		Branch:     *branch,
		RemotePath: *remotePath,
		ProtoDir:   *protoDir,
		BuildDir:   *buildDir,
	}

	if err := proto.SaveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration initialized successfully")
}

func syncCmd() {
	config, err := proto.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if config.GitHubURL == "" {
		fmt.Println("Error: Configuration not initialized. Run 'proto init' first")
		os.Exit(1)
	}

	// Create temporary directory for cloning
	tempDir, err := os.MkdirTemp("", "proto-sync-*")
	if err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Clone repository
	cloneCmd := exec.Command("git", "clone", "-b", config.Branch, config.GitHubURL, tempDir)
	if err := cloneCmd.Run(); err != nil {
		fmt.Printf("Error cloning repository: %v\n", err)
		os.Exit(1)
	}

	// Get latest commit ID
	commitCmd := exec.Command("git", "rev-parse", "HEAD")
	commitCmd.Dir = tempDir
	commitID, err := commitCmd.Output()
	if err != nil {
		fmt.Printf("Error getting commit ID: %v\n", err)
		os.Exit(1)
	}

	// If commit ID hasn't changed, exit
	if string(commitID) == config.LastCommitID {
		fmt.Println("Already up to date")
		return
	}

	// Create proto directory if it doesn't exist
	if err := os.MkdirAll(config.ProtoDir, 0755); err != nil {
		fmt.Printf("Error creating proto directory: %v\n", err)
		os.Exit(1)
	}

	// Determine the source directory for proto files
	sourceDir := tempDir
	if config.RemotePath != "" {
		sourceDir = filepath.Join(tempDir, config.RemotePath)
	}

	// Copy proto files
	protoFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.proto"))
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Printf("No proto files found in %s\n", sourceDir)
		os.Exit(1)
	}

	for _, protoFile := range protoFiles {
		fileName := filepath.Base(protoFile)
		destPath := filepath.Join(config.ProtoDir, fileName)
		data, err := os.ReadFile(protoFile)
		if err != nil {
			fmt.Printf("Error reading proto file: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			fmt.Printf("Error writing proto file: %v\n", err)
			os.Exit(1)
		}
	}

	// Update last commit ID
	config.LastCommitID = string(commitID)
	if err := proto.SaveConfig(config); err != nil {
		fmt.Printf("Error updating config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Proto files synced successfully")
}

func buildCmd() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: proto build [go|python]")
		os.Exit(1)
	}

	config, err := proto.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	lang := os.Args[2]
	switch lang {
	case "go":
		buildGoSDK(config)
	case "python":
		buildPythonSDK(config)
	default:
		fmt.Printf("Unsupported language: %s\n", lang)
		os.Exit(1)
	}
}

func buildGoSDK(config *proto.Config) {
	// Create build directory if it doesn't exist
	if err := os.MkdirAll(config.BuildDir, 0755); err != nil {
		fmt.Printf("Error creating build directory: %v\n", err)
		os.Exit(1)
	}

	// Find all proto files
	protoFiles, err := filepath.Glob(filepath.Join(config.ProtoDir, "*.proto"))
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	for _, protoFile := range protoFiles {
		cmd := exec.Command("protoc",
			"--go_out="+config.BuildDir,
			"--go_opt=paths=source_relative",
			protoFile)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error building Go SDK: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("Go SDK built successfully")
}

func buildPythonSDK(config *proto.Config) {
	// Create build directory if it doesn't exist
	if err := os.MkdirAll(config.BuildDir, 0755); err != nil {
		fmt.Printf("Error creating build directory: %v\n", err)
		os.Exit(1)
	}

	// Find all proto files
	protoFiles, err := filepath.Glob(filepath.Join(config.ProtoDir, "*.proto"))
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	for _, protoFile := range protoFiles {
		cmd := exec.Command("protoc",
			"--python_out="+config.BuildDir,
			protoFile)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error building Python SDK: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("Python SDK built successfully")
}
