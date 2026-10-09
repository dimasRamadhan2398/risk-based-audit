package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"master-service/controllers"
	"master-service/pkg/database"
	"master-service/pkg/logger"
	"master-service/pkg/middleware"
	"master-service/pkg/validations"
	"master-service/routes"
	"master-service/services"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the master-service HTTP server",
	Long:  `Start the master-service HTTP server`,
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// master-service has shipped pkg/middleware/auth.go since the start but
	// never applied it, so the organisational master data - companies,
	// departments and the full employee roster - was readable and writable by
	// anyone who could reach the gateway. Kong verifies no tokens for this
	// service, so this is the only check there is.
	//
	// Refuse to start without a secret rather than come up accepting any token.
	if strings.TrimSpace(cfg.JWT.Secret) == "" {
		return errors.New("jwt secret is empty: set JWT_SECRET (or jwt.secret in config.yaml)")
	}
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)

	validator := validations.New()
	serviceRegistry := services.NewServiceRegistry(db)
	controllerRegistry := controllers.NewControllerRegistry(serviceRegistry, validator)

	mode := strings.ToLower(cfg.Server.Mode)
	switch mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		gin.SetMode(mode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		gin.Logger(),
		middleware.RecoveryMiddleware(),
		middleware.CORSMiddleware(),
		middleware.ResponseCache(),
	)
	routes.RegisterRoutes(router, controllerRegistry, db, authMiddleware)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting master-service server", zap.String("addr", addr))

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.ReadTimeout) * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server stopped with error: %w", err)
	}

	return nil
}
