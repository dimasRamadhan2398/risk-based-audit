package main

import (
	"fmt"
	"os"

	"kong-gateway/pkg/config"
	"kong-gateway/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "kong-gateway",
	Short: "Kong Gateway - API Gateway for Microservices",
	Long: `Kong Gateway is an API gateway that provides:
- Reverse proxy to upstream services
- Authentication & authorization
- Rate limiting
- Request/response transformation
- Circuit breaking
- Health checking`,
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config/config.yaml", "config file path")
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
