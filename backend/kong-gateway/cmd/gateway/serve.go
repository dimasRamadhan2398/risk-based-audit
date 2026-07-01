package main

import (
	"fmt"
	"net/http"

	"kong-gateway/pkg/gateway"
	"kong-gateway/pkg/handlers"
	"kong-gateway/pkg/logger"
	"kong-gateway/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Kong Gateway HTTP server",
	Long: `Start the Kong Gateway HTTP server with all routes and middleware configured.
This will start the Gin HTTP server on the configured port (default: 8080).

Example:
  kong-gateway serve
  kong-gateway serve --config /path/to/config.yaml`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	// Initialize logger
	if err := logger.Init(&cfg.Log); err != nil {
		return err
	}
	defer logger.Sync()

	logger.Info("Starting Kong Gateway", logger.LogField("port", cfg.Server.Port))

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize gateway proxy
	proxy := gateway.NewProxy(cfg)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(proxy)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(&cfg.Auth)
	rateLimiter := middleware.NewRateLimiter(&cfg.RateLimit)

	// Create Gin engine
	engine := gin.New()

	// Apply global middleware
	engine.Use(middleware.RecoveryMiddleware())
	engine.Use(middleware.RequestIDMiddleware())
	engine.Use(middleware.LoggerMiddleware())
	engine.Use(middleware.CORSMiddleware(&cfg.CORS))

	// Health check endpoints (no auth required)
	engine.GET("/health", healthHandler.Health)
	engine.GET("/healthz", healthHandler.Live)
	engine.GET("/ready", healthHandler.Ready)
	engine.GET("/live", healthHandler.Live)

	// Stats endpoints (admin only)
	adminGroup := engine.Group("/admin")
	adminGroup.Use(authMiddleware.Authenticate())
	{
		adminGroup.GET("/stats", healthHandler.Stats)
		adminGroup.GET("/services", healthHandler.Services)
		adminGroup.GET("/health", healthHandler.HealthDetailed)
	}

	// API routes with authentication and rate limiting
	apiGroup := engine.Group("/api")
	apiGroup.Use(authMiddleware.Authenticate())
	if cfg.RateLimit.Enabled {
		apiGroup.Use(middleware.RateLimitMiddleware(rateLimiter))
	}

	// Proxy all API requests to the gateway
	apiGroup.Any("/*path", func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// Catch-all route for non-API requests
	engine.NoRoute(func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting Kong Gateway HTTP server", logger.LogField("address", addr))

	server := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  0,
		WriteTimeout: 0,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server", logger.LogField("error", err))
		return err
	}

	return nil
}
