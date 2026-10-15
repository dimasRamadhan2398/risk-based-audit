package main

import (
	"log"
	"net/http"
	"os"

	"analytics-service/middleware"
	"analytics-service/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// "analytics-service/cmd/docs"

func main() {
	// analytics-service has no config file, so the shared stack secret comes
	// from the environment. Fatal rather than optional: starting without it is
	// how every dashboard aggregate ended up anonymously readable.
	authMiddleware, err := middleware.NewAuthMiddleware(os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatalf("Refusing to start analytics-service: %v", err)
	}

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

	// Unauthenticated on purpose, and registered before the API group so Kong's
	// active healthcheck (which sends no token) can reach it.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "analytics-service"})
	})

	// Initialize routes
	routes.SetupRoutes(r, authMiddleware)

	// Start server
	log.Println("Analytics Service starting on port 8084...")
	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
