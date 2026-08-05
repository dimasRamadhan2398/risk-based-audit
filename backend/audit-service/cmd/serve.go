package main

import (
	"fmt"
	"net/http"

	ctrlActivity "audit-service/controllers/audit_activity"
	ctrlAssignment "audit-service/controllers/audit_assignment"
	ctrlCharter "audit-service/controllers/audit_charter"
	ctrlMandate "audit-service/controllers/audit_mandate"
	ctrlMedia "audit-service/controllers/media"
	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"
	"audit-service/pkg/media"
	"audit-service/pkg/middleware"
	"audit-service/pkg/redis"
	"audit-service/repositories"
	"audit-service/routes"
	svcActivity "audit-service/services/audit_activity"
	svcAssignment "audit-service/services/audit_assignment"
	svcCharter "audit-service/services/audit_charter"
	svcMandate "audit-service/services/audit_mandate"
	svcMedia "audit-service/services/media"

	docs "audit-service/docs"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Audit Service HTTP server",
	Long: `Start the Audit Service HTTP server with all routes and middleware configured.
This will start the Gin HTTP server on the configured port (default: 8002).

Example:
  audit serve
  audit serve --config /path/to/config.yaml`,
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

	logger.Info("Starting Audit Service", logger.LogField("port", cfg.Server.Port))
	fmt.Printf("GDrive Enabled Config Value: %v\n", cfg.GDrive.Enabled)

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

	// Auto-migrate uploaded document tables on startup
	_ = db.AutoMigrate(
		&models.UploadedPlanDocument{},
		&models.UploadedAnnualPlan{},
		&models.UploadedAssignmentLetter{},
		&models.UploadedAuditResultReport{},
		&models.UploadedExecutiveSummary{},
		&models.UploadedExecutiveSummaryReport{},
		&models.UploadedConsultingDocument{},
		&models.UploadedPerformanceReport{},
	)

	// Initialize Redis
	redisClient, err := redis.NewRedisConnection(&cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis", logger.LogField("error", err))
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Initialize base repository
	baseRepo := repositories.NewBaseRepository(db)

	// Initialize repositories
	auditCharterRepo := repositories.NewAuditCharterRepository(baseRepo)
	auditMandateRepo := repositories.NewAuditMandateRepository(baseRepo)
	auditAssignmentRepo := repositories.NewAuditAssignmentRepository(baseRepo)
	auditActivityRepo := repositories.NewAuditActivityRepository(baseRepo)

	// Initialize Media Provider
	var mediaProvider media.MediaProvider
	if cfg.GDrive.Enabled {
		var err error
		mediaProvider, err = media.NewGDriveProvider(&cfg.GDrive)
		if err != nil {
			logger.Fatal("Failed to initialize Google Drive provider", logger.LogField("error", err))
			return err
		}
		logger.Info("Using Google Drive as media provider")
	} else {
		mediaProvider = media.NewLocalProvider()
		logger.Info("Using Local Dummy Provider as media provider")
	}

	// Initialize services
	auditCharterSvc := svcCharter.NewAuditCharterService(auditCharterRepo)
	auditMandateSvc := svcMandate.NewAuditMandateService(auditMandateRepo)
	auditAssignmentSvc := svcAssignment.NewAuditAssignmentService(auditAssignmentRepo)
	auditActivitySvc := svcActivity.NewAuditActivityService(auditActivityRepo)
	mediaSvc := svcMedia.NewMediaService(mediaProvider)

	// Initialize controllers
	auditCharterCtrl := ctrlCharter.NewAuditCharterController(auditCharterSvc, mediaSvc)
	auditMandateCtrl := ctrlMandate.NewAuditMandateController(auditMandateSvc)
	auditAssignmentCtrl := ctrlAssignment.NewAuditAssignmentController(auditAssignmentSvc)
	auditActivityCtrl := ctrlActivity.NewAuditActivityController(auditActivitySvc)
	mediaCtrl := ctrlMedia.NewMediaController(mediaSvc)

	// Initialize route registry
	routeRegistry := routes.NewRouteRegistry()
	routeRegistry.SetAuditCharterController(auditCharterCtrl)
	routeRegistry.SetAuditMandateController(auditMandateCtrl)
	routeRegistry.SetAuditAssignmentController(auditAssignmentCtrl)
	routeRegistry.SetAuditActivityController(auditActivityCtrl)
	routeRegistry.SetMediaController(mediaCtrl)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)

	// Create Gin engine
	engine := gin.New()

	// Apply global middleware
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORSMiddleware())
	engine.Use(middleware.LoggerMiddleware())

	// Serve static uploads
	engine.Static("/uploads", "./uploads")

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "audit-service",
		})
	})

	docs.SwaggerInfo.Title = "Audit Service API"
	docs.SwaggerInfo.Description = "API documentation for Audit Service"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8002"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Swagger UI endpoint
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize route handler
	routeHandler := routes.NewRouteHandler(engine, authMiddleware, db)
	routeHandler.SetRegistry(routeRegistry)
	routeHandler.RegisterRoutes()

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting Audit Service HTTP server", logger.LogField("address", addr))

	if err := engine.Run(addr); err != nil {
		logger.Fatal("Failed to start server", logger.LogField("error", err))
		return err
	}

	return nil
}
