package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ctrlActivity "audit-service/controllers/audit_activity"
	ctrlAssignment "audit-service/controllers/audit_assignment"
	ctrlCharter "audit-service/controllers/audit_charter"
	ctrlMandate "audit-service/controllers/audit_mandate"
	ctrlMedia "audit-service/controllers/media"
	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/kafka"
	"audit-service/pkg/logger"
	"audit-service/pkg/media"
	"audit-service/pkg/middleware"
	"audit-service/pkg/redis"
	"audit-service/repositories"
	"audit-service/routes"
	annualPlanSync "audit-service/services/annual_plan_sync"
	svcActivity "audit-service/services/audit_activity"
	svcAssignment "audit-service/services/audit_assignment"
	svcCharter "audit-service/services/audit_charter"
	svcCompletion "audit-service/services/audit_completion"
	svcMandate "audit-service/services/audit_mandate"
	svcMedia "audit-service/services/media"

	docs "audit-service/docs"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
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
	if err := cfg.ValidateJWT(); err != nil {
		logger.Fatal("Refusing to start: invalid JWT_SECRET", logger.LogField("error", err.Error()))
	}
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

	// Auto-migrate uploaded document tables and analyzer snapshot on startup
	if err := db.AutoMigrate(
		&models.UploadedStrategicPlanDocument{},
		&models.UploadedPlanDocument{},
		&models.UploadedAnnualPlan{},
		&models.UploadedAssignmentLetter{},
		&models.UploadedAuditResultReport{},
		&models.UploadedExecutiveSummary{},
		&models.UploadedExecutiveSummaryReport{},
		&models.ExecutiveSummary{},
		&models.UploadedConsultingDocument{},
		&models.UploadedPerformanceReport{},
		&models.ImportedWorkingPaper{},
		&models.AuditCompletionSnapshot{},
		&models.WorkingPaperHeader{},
		&models.WorkingPaperRisk{},
		&models.WorkingPaperSample{},
		&models.WorkingPaperCause{},
		&models.WorkingPaperPlan{},
		&models.FieldworkInterview{},
		&models.FieldworkObservation{},
		&models.FieldworkDocument{},
		// Activity ID counter (pkg/activitycode); created here too so a deploy
		// that skips migrate/seed does not break creating activities.
		&models.ActivityCodeSequence{},
		// Plan code PKAT-{year}-{seq} and its partial unique index (models/audit-annual-plan.go).
		&models.AuditAnnual{},
		&models.AuditResultReport{},
	); err != nil {
		logger.Error("Auto-migrate failed; audit tables may be missing columns or indexes", logger.LogField("error", err))
	}

	// GORM's AutoMigrate did not create the partial unique index on the plan code
	// (models/audit-annual-plan.go), so create it explicitly and idempotently.
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_audit_annuals_code_live ON audit_annuals (code) WHERE deleted_at IS NULL AND code <> ''`).Error; err != nil {
		logger.Error("Failed to create plan code index", logger.LogField("error", err))
	}

	// Executive summaries saved with the legacy three risk counts get the five
	// levels (Tinggi -> High, Sedang -> Moderate, Rendah -> Low). Idempotent.
	normalizeExecutiveSummaryRiskCounts(db)

	// Legacy annual plan statuses (e.g. "Done", "DRAFT") become the canonical
	// vocabulary before any request is served. Idempotent.
	if err := normalizeAnnualPlanStatuses(db); err != nil {
		logger.Error("Annual plan status normalization failed; status checks may reject legacy values",
			logger.LogField("error", err))
	}

	// Ensure columns in audit_result_reports exist
	db.Exec(`
		ALTER TABLE audit_result_reports ADD COLUMN IF NOT EXISTS signature_place varchar(100);
		ALTER TABLE audit_result_reports ADD COLUMN IF NOT EXISTS signature_date timestamp with time zone;
		ALTER TABLE audit_result_reports ADD COLUMN IF NOT EXISTS signatures text;
	`)

	// Ensure assignment_letters letter_number unique index is a partial index on active records only
	db.Exec(`
		DROP INDEX IF EXISTS idx_assignment_letters_letter_number;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_assignment_letters_letter_number ON assignment_letters (letter_number) WHERE deleted_at IS NULL;
	`)

	// Action taken reports: new LHA-bound layout, legacy columns dropped, ATRs
	// of approved LHAs created. Idempotent; tenant stacks start with "serve"
	// only, so this cannot be left to "migrate up" (see migrateATR).
	if err := migrateATR(db); err != nil {
		logger.Error("Action taken report migration failed; the ATR endpoints and LHA approval may fail until it succeeds",
			logger.LogField("error", err))
	}

	// Ensure population column in working_paper_samples is varchar(255)
	db.Exec("ALTER TABLE working_paper_samples ALTER COLUMN population TYPE varchar(255) USING population::varchar;")

	// Initialize Redis
	redisClient, err := redis.NewRedisConnection(&cfg.Redis)
	if err != nil {
		logger.Warn("Failed to connect to Redis", logger.LogField("error", err))
	}
	if redisClient != nil {
		defer redisClient.Close()
	}

	// Start the Audit Completion Analyzer background goroutine.
	// It runs every 5 minutes, caches results in Redis, and persists
	// snapshots for historical trending.
	analyzerCtx, cancelAnalyzer := context.WithCancel(context.Background())
	defer cancelAnalyzer()
	completionAnalyzer := svcCompletion.NewAuditCompletionAnalyzer(db, redisClient)
	go completionAnalyzer.Start(analyzerCtx)
	logger.Info("[AuditCompletionAnalyzer] Background goroutine launched")

	// Initialize base repository
	baseRepo := repositories.NewBaseRepository(db)

	// Initialize repositories
	auditCharterRepo := repositories.NewAuditCharterRepository(baseRepo)
	auditMandateRepo := repositories.NewAuditMandateRepository(baseRepo)
	auditAssignmentRepo := repositories.NewAuditAssignmentRepository(baseRepo)
	auditActivityRepo := repositories.NewAuditActivityRepository(baseRepo)

	// Initialize Media Provider
	// Priority: ImageKit (env) → GDrive (config) → Local
	var mediaProvider media.MediaProvider
	if cfg.ImageKit.PrivateKey != "" && cfg.ImageKit.UrlEndpoint != "" {
		mediaProvider = media.NewImageKitProvider(&cfg.ImageKit)
		logger.Info("Using ImageKit as media provider",
			logger.LogField("endpoint", cfg.ImageKit.UrlEndpoint))
	} else if cfg.GDrive.Enabled {
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
	// Add response caching middleware for dashboard endpoints
	engine.Use(middleware.ResponseCache())

	// Serve static uploads, except private subdirectories such as ATR evidence
	// (./uploads/action-taken-reports), which only the authenticated API serves
	routes.RegisterUploads(engine, "./uploads", authMiddleware.Authenticate())

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
	// Swagger UI is off unless ENABLE_SWAGGER=true.
	if routes.SwaggerEnabled() {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	startAnnualPlanSync(db)

	// Initialize route handler
	routeHandler := routes.NewRouteHandler(engine, authMiddleware, db, redisClient)
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

// startAnnualPlanSync consumes Audit Universe and Department events and applies
// them to annual plans (services/annual_plan_sync). It runs in the background
// and never blocks or fails start-up: if the broker is unreachable the service
// keeps serving, and the consumer group resumes from its committed offset on the
// next start.
func startAnnualPlanSync(db *gorm.DB) {
	if !cfg.KafkaConsumer.Enabled {
		logger.Info("Annual plan sync consumer disabled (kafka_consumer.enabled=false)")
		return
	}
	// The consumer is created in the background and retried until the broker
	// and the group are reachable. A broker that is down at boot must not
	// disable annual plan sync until the next restart.
	go func() {
		backoff := 2 * time.Second
		for {
			consumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
				Brokers:       cfg.KafkaConsumer.Brokers,
				GroupID:       cfg.KafkaConsumer.GroupID,
				Topics:        cfg.KafkaConsumer.Topics,
				Enabled:       cfg.KafkaConsumer.Enabled,
				Version:       cfg.KafkaConsumer.Version,
				Assignor:      cfg.KafkaConsumer.Assignor,
				InitialOffset: cfg.KafkaConsumer.Offset,
			}, logger.GetLogger())
			if err == nil {
				annualPlanSync.Register(consumer, annualPlanSync.NewHandler(annualPlanSync.New(db)))
				if err := consumer.Start(); err != nil {
					logger.Error("Annual plan sync consumer failed to start", logger.LogField("error", err))
				}
				return
			}
			logger.Warn("Annual plan sync consumer unavailable, retrying",
				logger.LogField("error", err), logger.LogField("retry_in", backoff.String()))
			time.Sleep(backoff)
			if backoff < time.Minute {
				backoff *= 2
			}
		}
	}()
}
