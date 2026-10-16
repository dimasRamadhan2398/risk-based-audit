package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"master-service/controllers"
	"master-service/pkg/database"
	"master-service/models"
	"master-service/pkg/kafka"
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
	// Department events feed audit-service's annual plan sync. Best-effort: the
	// service runs without them if the broker is unreachable at start-up.
	var domainEvents *kafka.DomainPublisher
	if cfg.Kafka.Enabled {
		// Events go to the transactional outbox and are published by the relay, so
		// a broker that is down at start-up delays department events, it does not drop them.
		if err := db.AutoMigrate(&models.KafkaOutboxEvent{}); err != nil {
			return fmt.Errorf("migrate kafka outbox: %w", err)
		}
		domainEvents = kafka.NewDomainPublisher(db, cfg.Kafka.ServiceName)
		go kafka.NewRelay(db, cfg.Kafka.Brokers, kafka.TopicDepartments).Run(context.Background())
	}
	serviceRegistry := services.NewServiceRegistry(db, domainEvents)
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
	routes.RegisterRoutes(router, controllerRegistry, db, authMiddleware, domainEvents)

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
