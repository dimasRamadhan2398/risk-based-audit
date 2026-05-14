package main

import (
	"log"

	"analytics-service/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Adjust this in production
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Initialize routes
	routes.SetupRoutes(r)

	// Start server
	log.Println("Analytics Service starting on port 8084...")
	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
