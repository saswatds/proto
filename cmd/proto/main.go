package main

import (
	"fmt"
	"os"

	"github.com/saswatds/proto/cmd/proto/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "proto",
	Short: "Proto is a tool for managing protocol buffers",
	Long:  `Proto is a tool for managing protocol buffers, including syncing from repositories and generating SDKs.`,
}

var (
	githubURL  string
	branch     string
	remotePath string
	protoDir   string
	buildDir   string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize proto configuration",
	Long:  `Initialize proto configuration with GitHub URL, branch, and proto directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if githubURL == "" {
			fmt.Println("Error: GitHub repository URL is required")
			os.Exit(1)
		}
		commands.InitCmd(githubURL, branch, remotePath, protoDir, buildDir)
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync proto files from repository",
	Long:  `Sync proto files from the configured GitHub repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		commands.SyncCmd()
	},
}

var genCmd = &cobra.Command{
	Use:   "gen [sdk_type]",
	Short: "Generate SDK from proto files",
	Long:  `Generate SDK (Go or Python) from proto files.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commands.GenCmd(args[0], "")
	},
}

func init() {
	initCmd.Flags().StringVar(&githubURL, "url", "", "GitHub repository URL")
	initCmd.Flags().StringVar(&branch, "branch", "main", "Git branch name")
	initCmd.Flags().StringVar(&remotePath, "remote-path", "proto", "Path within the repository containing proto files")
	initCmd.Flags().StringVar(&protoDir, "proto-dir", "./proto", "Directory for synced proto files")
	initCmd.Flags().StringVar(&buildDir, "build-dir", "./gen", "Directory for generated SDKs")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(genCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
