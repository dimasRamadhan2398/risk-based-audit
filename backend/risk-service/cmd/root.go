package main

import (
	"fmt"
	"os"

	"risk-service/pkg/config"
	"risk-service/pkg/logger"

	"github.com/spf13/cobra"
)

// Variables set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var (
	configPath string
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "risk-service",
	Short: "Risk Service CLI",
	Long:  `Risk Service - A microservice for managing risks, controls, and mitigations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Global pre-run logic can go here
	},
	SilenceUsage: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of risk-service",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(GetVersion())
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add version command
	rootCmd.AddCommand(versionCmd)

	// Persistent flags for all commands
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "pkg/config/config.yaml", "Path to config file")
}

// GetVersion returns version info
func GetVersion() string {
	return fmt.Sprintf("Version: %s\nBuild Time: %s\nGit Commit: %s", Version, BuildTime, GitCommit)
}

// GetRootCmd returns the root command for testing
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func initConfig() error {
	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func initLogger() error {
	if err := logger.Init(nil); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	return nil
}
