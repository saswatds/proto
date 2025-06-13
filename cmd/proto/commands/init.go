package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/saswatds/proto/pkg/proto"
)

// InitCmd handles initializing the proto configuration
func InitCmd(githubURL, branch, remotePath, protoDir, buildDir string) {
	config := &proto.Config{
		GitHubURL:  githubURL,
		Branch:     branch,
		RemotePath: remotePath,
		ProtoDir:   protoDir,
		BuildDir:   buildDir,
	}

	// Create proto and gen directories if they don't exist
	if err := os.MkdirAll(config.ProtoDir, 0755); err != nil {
		fmt.Printf("Error creating proto directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(config.BuildDir, 0755); err != nil {
		fmt.Printf("Error creating build directory: %v\n", err)
		os.Exit(1)
	}

	if err := proto.SaveConfig(config); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	// Read and print the config file
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	configPath := filepath.Join(workDir, ".protorc")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Configuration initialized successfully")
	fmt.Println("\nConfiguration file (.protorc):")
	fmt.Println("----------------------------------------")
	fmt.Println(string(data))
	fmt.Println("----------------------------------------")
	fmt.Printf("\nCreated directories:\n")
	fmt.Printf("- %s (for proto files)\n", config.ProtoDir)
	fmt.Printf("- %s (for generated SDKs)\n", config.BuildDir)
}
