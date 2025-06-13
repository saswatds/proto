package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saswatds/proto/pkg/proto"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "proto",
		Usage: "A CLI tool for managing Protocol Buffer files",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize configuration for proto repository",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "url",
						Usage:    "GitHub repository URL",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "branch",
						Usage: "Git branch name",
						Value: "main",
					},
					&cli.StringFlag{
						Name:  "remote-path",
						Usage: "Path within the repository containing proto files",
					},
					&cli.StringFlag{
						Name:  "proto-dir",
						Usage: "Directory for synced proto files",
						Value: "./proto",
					},
					&cli.StringFlag{
						Name:  "build-dir",
						Usage: "Directory for generated SDKs",
						Value: "./gen",
					},
				},
				Action: func(c *cli.Context) error {
					initCmd(c)
					return nil
				},
			},
			{
				Name:  "sync",
				Usage: "Sync proto files from the repository",
				Action: func(c *cli.Context) error {
					syncCmd()
					return nil
				},
			},
			{
				Name:  "gen",
				Usage: "Generate SDKs from proto files",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						fmt.Println("Error: Please specify the SDK type (go or python)")
						os.Exit(1)
					}
					genCmd(c.Args().Get(0))
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func initCmd(c *cli.Context) {
	config := &proto.Config{
		GitHubURL:  c.String("url"),
		Branch:     c.String("branch"),
		RemotePath: c.String("remote-path"),
		ProtoDir:   c.String("proto-dir"),
		BuildDir:   c.String("build-dir"),
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

	// Get project type and module path
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	var modulePath string
	var projectType string

	// Check for Go project
	goModPath := filepath.Join(workDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		goModData, err := os.ReadFile(goModPath)
		if err != nil {
			fmt.Printf("Error reading go.mod: %v\n", err)
			os.Exit(1)
		}
		moduleLine := strings.Split(string(goModData), "\n")[0]
		modulePath = strings.TrimPrefix(moduleLine, "module ")
		modulePath = strings.TrimSpace(modulePath)
		projectType = "go"
	} else {
		// Check for Python project
		setupPyPath := filepath.Join(workDir, "setup.py")
		pyProjectPath := filepath.Join(workDir, "pyproject.toml")

		if _, err := os.Stat(setupPyPath); err == nil {
			// Read setup.py to get package name
			setupPyData, err := os.ReadFile(setupPyPath)
			if err != nil {
				fmt.Printf("Error reading setup.py: %v\n", err)
				os.Exit(1)
			}
			content := string(setupPyData)
			// Simple regex to find package name
			if strings.Contains(content, "name=") {
				start := strings.Index(content, "name=") + 5
				end := strings.Index(content[start:], ",")
				if end == -1 {
					end = strings.Index(content[start:], ")")
				}
				if end != -1 {
					modulePath = strings.Trim(content[start:start+end], `"' `)
					projectType = "python"
				}
			}
		} else if _, err := os.Stat(pyProjectPath); err == nil {
			// Read pyproject.toml to get package name
			pyProjectData, err := os.ReadFile(pyProjectPath)
			if err != nil {
				fmt.Printf("Error reading pyproject.toml: %v\n", err)
				os.Exit(1)
			}
			content := string(pyProjectData)
			if strings.Contains(content, "name =") {
				start := strings.Index(content, "name =") + 6
				end := strings.Index(content[start:], "\n")
				if end != -1 {
					modulePath = strings.Trim(content[start:start+end], `"' `)
					projectType = "python"
				}
			}
		}
	}

	if modulePath == "" {
		fmt.Println("Error: Could not determine project type. Please ensure you're in a Go or Python project directory")
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
	if string(commitID) == config.GitHead {
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
		// Remove any quotes from the remote path
		cleanPath := strings.Trim(config.RemotePath, `"'`)
		sourceDir = filepath.Join(tempDir, cleanPath)
	}

	// Copy proto files
	protoFiles, err := filepath.Glob(filepath.Join(sourceDir, "*.proto"))
	if err != nil {
		fmt.Printf("Error finding proto files: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Printf("No proto files found in %s\n", sourceDir)

		// List all files and directories recursively from the root
		fmt.Println("\nRepository structure:")
		fmt.Println("----------------------------------------")
		err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Skip the temp directory itself
			if path == tempDir {
				return nil
			}
			// Skip hidden files and directories
			if strings.HasPrefix(info.Name(), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// Get relative path from temp directory
			relPath, err := filepath.Rel(tempDir, path)
			if err != nil {
				return err
			}
			// Calculate indentation based on depth
			depth := strings.Count(relPath, string(os.PathSeparator))
			indent := strings.Repeat("  ", depth)

			if info.IsDir() {
				fmt.Printf("%s%s/\n", indent, info.Name())
			} else {
				fmt.Printf("%s- %s (%d bytes)\n", indent, info.Name(), info.Size())
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
		}
		fmt.Println("----------------------------------------")
		fmt.Println("\nPlease check if:")
		fmt.Println("1. The remote_path is correct")
		fmt.Println("2. The repository contains .proto files")
		fmt.Println("3. The files are in the expected location")
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

		// Update package and go_package options
		content := string(data)
		lines := strings.Split(content, "\n")
		var updatedLines []string
		packageFound := false
		goPackageFound := false

		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Update package name
			if strings.HasPrefix(trimmedLine, "package ") {
				packageFound = true
				updatedLines = append(updatedLines, "package proto;")
				continue
			}

			// Update go_package option based on project type
			if strings.HasPrefix(trimmedLine, "option go_package") {
				goPackageFound = true
				if projectType == "go" {
					updatedLines = append(updatedLines, fmt.Sprintf(`option go_package = "%s";`, modulePath))
				} else {
					// For Python projects, use the package name as the go_package
					updatedLines = append(updatedLines, fmt.Sprintf(`option go_package = "%s";`, modulePath))
				}
				continue
			}

			updatedLines = append(updatedLines, line)
		}

		// Add package and go_package if not found
		if !packageFound {
			updatedLines = append([]string{"package proto;"}, updatedLines...)
		}
		if !goPackageFound {
			// Find the first line after syntax declaration
			for i, line := range updatedLines {
				if strings.HasPrefix(strings.TrimSpace(line), "syntax =") {
					if projectType == "go" {
						updatedLines = append(updatedLines[:i+1], append([]string{fmt.Sprintf(`option go_package = "%s";`, modulePath)}, updatedLines[i+1:]...)...)
					} else {
						updatedLines = append(updatedLines[:i+1], append([]string{fmt.Sprintf(`option go_package = "%s";`, modulePath)}, updatedLines[i+1:]...)...)
					}
					break
				}
			}
		}

		// Write the updated content
		if err := os.WriteFile(destPath, []byte(strings.Join(updatedLines, "\n")), 0644); err != nil {
			fmt.Printf("Error writing proto file: %v\n", err)
			os.Exit(1)
		}
	}

	// Update git head
	config.GitHead = string(commitID)
	if err := proto.SaveConfig(config); err != nil {
		fmt.Printf("Error updating config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Proto files synced successfully")
}

func genCmd(sdkType string) {
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
		os.Exit(1)
	}

	// Build proto files
	switch sdkType {
	case "go":
		// Generate Go SDK
		for _, protoFile := range protoFiles {
			cmd := exec.Command("protoc",
				"--go_out="+config.BuildDir,
				"--go_opt=paths=source_relative",
				protoFile)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error generating Go SDK: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Println("Go SDK generated successfully in", config.BuildDir)

	case "python":
		// Generate Python SDK
		for _, protoFile := range protoFiles {
			cmd := exec.Command("protoc",
				"--python_out="+config.BuildDir,
				protoFile)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error generating Python SDK: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Println("Python SDK generated successfully in", config.BuildDir)

	default:
		fmt.Println("Error: Unsupported SDK type. Use 'go' or 'python'")
		os.Exit(1)
	}
}
