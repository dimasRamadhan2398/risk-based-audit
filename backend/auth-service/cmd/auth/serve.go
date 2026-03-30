package main

import (
	"fmt"
	"net/http"

	"auth-service/docs"
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

	// Initialize event publisher
	eventPublisher := kafkapkg.NewEventPublisher(kafkapkg.EventPublisherConfig{
		Producer: kafkaProducer,
		Enabled:  cfg.Kafka.Enabled,
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
	)

	// Initialize validator
	validator := validations.New()

	// Initialize service registry
	serviceRegistry := services.NewServiceRegistry(repoRegistry)

	// Initialize controller registry
	controllerRegistry := routes.NewControllerRegistry(serviceRegistry, validator)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)

	// Create Gin engine
	engine := gin.New()

	// Apply global middleware
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.LoggerMiddleware())

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// Swagger documentation
	docs.SwaggerInfo.Title = "Auth Service API"
	docs.SwaggerInfo.Description = "Authentication and Authorization microservice for Risk-Based Internal Audit System"
	docs.SwaggerInfo.Version = "1.0.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%d", cfg.Server.Port)
	docs.SwaggerInfo.BasePath = "/"
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
	logger.Info("Swagger documentation available at", logger.LogField("url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)))

	if err := engine.Run(addr); err != nil {
		logger.Fatal("Failed to start server", logger.LogField("error", err))
		return err
	}

	return nil
}
