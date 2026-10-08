package main

import (
	"log"

	"analytics-service/middleware"
	"analytics-service/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// "analytics-service/cmd/docs"

func main() {
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Adjust this in production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Cache-Control", "ETag", "Last-Modified", "Expires"},
		AllowCredentials: true,
	}))

	// Add response caching middleware for dashboard endpoints
	r.Use(middleware.ResponseCache())

	// Initialize routes
	routes.SetupRoutes(r)

	// Start server
	log.Println("Analytics Service starting on port 8084...")
	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
