package main

import (
	"fmt"
	"os"

	"auth-service/pkg/config"
	"auth-service/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "auth-service",
	Short: "Auth Service - Authentication and Authorization Microservice",
	Long: `Auth Service is a microservice for handling authentication and authorization.
It provides features including:
- User authentication (login/register)
- Multi-factor authentication (MFA/TOTP)
- Trusted device management
- Role-based access control (RBAC)
- JWT token management
- Event publishing to Kafka`,
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
