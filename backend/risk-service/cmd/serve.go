package main

import (
	"log"
	"net/http"
	"os"

	"risk-service/controllers"
	"risk-service/pkg/database"
	"risk-service/repositories"
	"risk-service/routes"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	addr := ":" + port
	log.Printf("Starting risk-service HTTP server on %s", addr)

	return r.Run(addr)
}
