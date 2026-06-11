package main

import (
	"fmt"
	"os"

	"audit-service/pkg/config"
	"audit-service/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "audit-service",
	Short: "Audit Service - Risk-Based Internal Audit Microservice",
	Long: `Audit Service is a microservice for handling audit-related operations.
It provides features including:
- Audit charter management
- Audit mandate management
- Audit assignment management
- Audit fieldwork documentation`,
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./pkg/config/config.yaml", "config file path")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug mode")
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}

func initLogger() {
	if err := logger.Init(&cfg.Log); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}
