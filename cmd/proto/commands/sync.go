package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/saswatds/proto/pkg/proto"
)

// SyncCmd handles syncing proto files from the repository
func SyncCmd() {
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
		fmt.Println("\nCommon issues:")
		fmt.Println("1. Incorrect repository URL")
		fmt.Println("2. Private repository (requires authentication)")
		fmt.Println("3. Incorrect branch name")
		fmt.Println("4. Network connectivity issues")
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

	// Load cached git head
	cachedGitHead, err := proto.LoadCache(config)
	if err != nil {
		fmt.Printf("Warning: Could not load cache file: %v\n", err)
	}

	// If commit ID hasn't changed, exit
	if string(commitID) == cachedGitHead {
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

		// Verify the remote path exists
		if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
			fmt.Printf("Error: Remote path '%s' does not exist in the repository\n", config.RemotePath)
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
			fmt.Println("2. The path exists in the repository")
			fmt.Println("3. The path is properly formatted")
			os.Exit(1)
		}
	}

	// Find all proto files in the source directory and its subdirectories
	var protoFiles []string
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error searching for proto files: %v\n", err)
		os.Exit(1)
	}

	if len(protoFiles) == 0 {
		fmt.Printf("No proto files found in %s\n", sourceDir)
		fmt.Println("\nDirectory structure:")
		fmt.Println("----------------------------------------")
		err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Skip the source directory itself
			if path == sourceDir {
				return nil
			}
			// Skip hidden files and directories
			if strings.HasPrefix(info.Name(), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// Get relative path from source directory
			relPath, err := filepath.Rel(sourceDir, path)
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

	// Copy proto files to the proto directory
	for _, protoFile := range protoFiles {
		// Get relative path from source directory
		relPath, err := filepath.Rel(sourceDir, protoFile)
		if err != nil {
			fmt.Printf("Error getting relative path: %v\n", err)
			continue
		}

		// Create destination path
		destPath := filepath.Join(config.ProtoDir, relPath)

		// Create parent directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			fmt.Printf("Error creating directory for %s: %v\n", relPath, err)
			continue
		}

		// Read and copy the file
		data, err := os.ReadFile(protoFile)
		if err != nil {
			fmt.Printf("Error reading proto file %s: %v\n", relPath, err)
			continue
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			fmt.Printf("Error writing proto file %s: %v\n", relPath, err)
			continue
		}
	}

	// Update git head in cache
	if err := proto.SaveCache(config, string(commitID)); err != nil {
		fmt.Printf("Error updating cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Proto files synced successfully")
}
