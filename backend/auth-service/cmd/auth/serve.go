package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"auth-service/docs"
	"auth-service/pkg/activitylog"
	"auth-service/pkg/database"
	kafkapkg "auth-service/pkg/kafka"
	"auth-service/pkg/logger"
	"auth-service/pkg/middleware"
	"auth-service/pkg/redis"
	"auth-service/pkg/validations"
	"auth-service/repositories"
	"auth-service/routes"
	"auth-service/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Auth Service HTTP server",
	Long: `Start the Auth Service HTTP server with all routes and middleware configured.
This will start the Gin HTTP server on the configured port (default: 8001).

Example:
  auth serve
  auth serve --config /path/to/config.yaml`,
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

	logger.Info("Starting Auth Service", logger.LogField("port", cfg.Server.Port))
	if err := cfg.ValidateJWT(); err != nil {
		logger.Fatal("Refusing to start: invalid JWT_SECRET", logger.LogField("error", err.Error()))
	}

	// Set Gin mode
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	// Initialize Redis
	redisClient, err := redis.NewRedisConnection(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", logger.LogField("error", err))
		return err
	}
	defer redisClient.Close()

	redisRepo := redis.NewRepository(redisClient)

	// Initialize Kafka producer
	kafkaProducer, err := kafkapkg.NewProducer(kafkapkg.Config{
		Brokers:     cfg.Kafka.Brokers,
		Topic:       cfg.Kafka.Topic,
		ServiceName: cfg.Kafka.ServiceName,
		Enabled:     cfg.Kafka.Enabled,
	}, logger.GetLogger())
	if err != nil {
		logger.Warn("Failed to initialize Kafka producer", logger.LogField("error", err))
	}

	// Admin user activity log: publish to this stack's activity topic, fall
	// back to a direct write when Kafka is unavailable.
	sysLogRepo := repositories.NewSysAlertLogRepository(db)
	activitySource := activitylog.ResolveSource(cfg.Kafka.ServiceName)
	activityTopic := activitylog.ResolveTopic(cfg.ActivityLog.Topic, activitylog.EnvTopicPrefix(), activitySource)
	var activityPublisher activitylog.Publisher
	if kafkaProducer != nil && kafkaProducer.IsEnabled() {
		activityPublisher = kafkaProducer
	}
	activityRecorder := activitylog.NewKafkaRecorder(activitylog.RecorderConfig{
		Topic:  activityTopic,
		Source: activitySource,
	}, activityPublisher, sysLogRepo, logger.GetLogger())
	defer activityRecorder.Close(5 * time.Second)
	logger.Info("Activity log configured",
		logger.LogField("topic", activityTopic),
		logger.LogField("source", activitySource),
		logger.LogField("kafka_publish", activityPublisher != nil),
	)

	// The consumer stores this stack's activity events. It starts in the
	// background and retries with backoff, so a broker that is down at boot
	// neither delays nor crashes the server.
	if cfg.Kafka.Enabled && cfg.ActivityLog.ConsumerEnabled {
		activityConsumer := activitylog.StartConsumer(activitylog.ConsumerConfig{
			Brokers: cfg.Kafka.Brokers,
			Topic:   activityTopic,
			GroupID: activitylog.ResolveGroupID(cfg.ActivityLog.ConsumerGroupID, activitySource),
			Version: cfg.KafkaConsumer.Version,
		}, activitylog.NewMessageProcessor(sysLogRepo, activitySource, logger.GetLogger()), logger.GetLogger())
		defer activityConsumer.Stop()
	} else {
		logger.Info("Activity log consumer disabled; events are written directly when Kafka publish is off")
	}

	if cfg.ActivityLog.PurgeEnabled {
		retention := activitylog.StartRetention(sysLogRepo, cfg.ActivityLog.EffectiveRetentionDays(), 24*time.Hour, logger.GetLogger())
		defer retention.Stop()
	}

	// Initialize event publisher. The tap feeds every published auth event
	// (MFA, roles, trusted devices, permission.denied, step-up) to the activity log.
	eventPublisher := kafkapkg.NewEventPublisher(kafkapkg.EventPublisherConfig{
		Producer: kafkaProducer,
		Enabled:  cfg.Kafka.Enabled,
		Tap:      activitylog.Tap(activityRecorder),
	})

	// Initialize cache repository
	cacheRepo := repositories.NewCacheRepository(redisRepo, redisClient)

	// Initialize repository registry
	repoRegistry := repositories.NewRegistryWithAll(
		db,
		redisClient,
		redisRepo,
		kafkaProducer,
		eventPublisher,
		cacheRepo,
		cfg,
		activityRecorder,
	)

	// Initialize validator
	validator := validations.New()

	// Initialize service registry
	serviceRegistry := services.NewServiceRegistry(repoRegistry)

	// Initialize controller registry
	controllerRegistry := routes.NewControllerRegistry(serviceRegistry, validator)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret).WithSessionCheck(loginSessionCheck(redisClient))

	// Create Gin engine
	engine := gin.New()

	// Apply global middleware
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.LoggerMiddleware())
	engine.Use(activitylog.ContextMiddleware())

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// 404 handler — unknown route
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "ENDPOINT_NOT_FOUND",
				"message": "The requested endpoint does not exist. Please check the URL and try again.",
			},
		})
	})

	// 405 handler — wrong HTTP method
	engine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "METHOD_NOT_ALLOWED",
				"message": "HTTP method not allowed for this endpoint.",
			},
		})
	})

	// Swagger documentation
	docs.SwaggerInfo.Title = "Auth Service API"
	docs.SwaggerInfo.Description = "Authentication and Authorization microservice for Risk-Based Internal Audit System"
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.Server.Port)
	docs.SwaggerInfo.BasePath = "/"
	// Swagger UI is off unless ENABLE_SWAGGER=true.
	if swaggerEnabled() {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		logger.Info("Swagger documentation available at", logger.LogField("url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)))
	}

	// API v1 routes
	apiV1 := engine.Group("/api/v1")

	// Initialize route registry
	routeRegistry := routes.NewRouteRegistry(
		apiV1,
		authMiddleware,
	)
	routeRegistry.SetController(controllerRegistry)
	routeRegistry.Serve()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting Auth Service HTTP server", logger.LogField("address", addr))

	// Graceful shutdown on SIGINT/SIGTERM so queued activity events are
	// flushed (the deferred Close/Stop above) before the process exits.
	srv := &http.Server{Addr: addr, Handler: engine}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.ListenAndServe() }()

	select {
	case err := <-serveErr:
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.LogField("error", err))
			return err
		}
	case <-ctx.Done():
		logger.Info("Shutting down Auth Service")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Warn("HTTP server shutdown", logger.LogField("error", err))
		}
	}

	return nil
}

// loginSessionCheck reports whether a token still has the session that login
// stored under "token:<jwt>" (services/auth). Logout deletes that key, so a
// logged-out token stops working here before its JWT expiry.
func loginSessionCheck(rdb *redis.Client) middleware.SessionChecker {
	return func(ctx context.Context, token string) (bool, error) {
		n, err := rdb.Exists(ctx, "token:"+token).Result()
		return n > 0, err
	}
}

// swaggerEnabled reports whether the Swagger UI is mounted (ENABLE_SWAGGER=true).
func swaggerEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("ENABLE_SWAGGER")), "true")
}
