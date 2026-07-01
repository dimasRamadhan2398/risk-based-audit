package main

import (
	"log"
	"net/http"
	"os"

	"risk-service/controllers"
	"risk-service/models"
	"risk-service/pkg/database"
	"risk-service/repositories"
	"risk-service/routes"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start risk-service HTTP server",
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

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}

	// Seed initial risks if the DB is empty
	if err := seedInitialRisks(db); err != nil {
		log.Printf("Warning: Seeding failed: %v", err)
	}

	// Seed initial appetite statements if empty
	if err := seedInitialAppetite(db); err != nil {
		log.Printf("Warning: Appetite seeding failed: %v", err)
	}

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Initialize Layers
	riskRepo := repositories.NewRiskRepository(db)
	mitigationRepo := repositories.NewMitigationRepository(db)

	riskServ := services.NewRiskService(riskRepo)
	mitigationServ := services.NewMitigationService(mitigationRepo)

	riskCtrl := controllers.NewRiskController(riskServ)
	mitigationCtrl := controllers.NewMitigationController(mitigationServ)

	// Register Routes
	riskRoute := routes.NewRiskRoute(riskCtrl, mitigationCtrl, &r.RouterGroup)
	riskRoute.Run()

	// Risk Appetite Handlers (Migrated to Gin)
	r.GET("/api/v1/risk-appetite", func(c *gin.Context) {
		var appetites []models.RiskAppetite
		if err := db.Order("created_at desc").Find(&appetites).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query risk appetites: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, appetites)
	})

	r.POST("/api/v1/risk-appetite", func(c *gin.Context) {
		var req models.RiskAppetite
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
		req.ID = uuid.New()
		if req.Status == "" {
			req.Status = models.RiskAppetiteStatusDraft
		}
		if err := db.Create(&req).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create risk appetite: " + err.Error()})
			return
		}
		c.JSON(http.StatusCreated, req)
	})

	r.PUT("/api/v1/risk-appetite/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		appID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid risk appetite ID format"})
			return
		}

		var req models.RiskAppetite
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		var appetite models.RiskAppetite
		if err := db.First(&appetite, "id = ?", appID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Risk appetite not found"})
			return
		}

		// Update fields
		appetite.Statement = req.Statement
		appetite.ThresholdLimit = req.ThresholdLimit
		appetite.Status = req.Status

		if err := db.Save(&appetite).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update risk appetite: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, appetite)
	})

	r.DELETE("/api/v1/risk-appetite/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		appID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid risk appetite ID format"})
			return
		}

		var appetite models.RiskAppetite
		if err := db.First(&appetite, "id = ?", appID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Risk appetite not found"})
			return
		}

		if err := db.Delete(&appetite).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete risk appetite: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Risk appetite deleted successfully",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	addr := ":" + port
	log.Printf("Starting risk-service HTTP server on %s", addr)

	return r.Run(addr)
}

func seedInitialAppetite(db *gorm.DB) error {
	var count int64
	db.Model(&models.RiskAppetite{}).Count(&count)
	if count > 0 {
		return nil
	}

	log.Println("Seeding initial risk appetite statements...")
	seeds := []models.RiskAppetite{
		{
			ID:             uuid.New(),
			Statement:      "Risiko tingkat Low dan Low to Moderate dapat diterima dan tidak perlu dilakukan mitigasi risiko.",
			ThresholdLimit: 10.00,
			Status:         models.RiskAppetiteStatusApproved,
		},
		{
			ID:             uuid.New(),
			Statement:      "Risiko tingkat Moderate, Moderate to High, dan High harus ditangani dengan melakukan mitigasi risiko yang tepat untuk menurunkan tingkat risiko.",
			ThresholdLimit: 15.00,
			Status:         models.RiskAppetiteStatusApproved,
		},
	}

	for _, s := range seeds {
		if err := db.Create(&s).Error; err != nil {
			return err
		}
	}
	return nil
}

