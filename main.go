package main

import (
	"log"
	"qb/internal/routes"
	"qb/pkg/database"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {

	// Load .env file once at application startup
	utils.LoadDotEnv()

	// Initialize Cloudinary
	if err := utils.InitCloudinary(); err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}

	// Connect to the database
	database.ConnectDB()

	// Initialize request tracker with database connection
	utils.InitRequestTracker(database.DB)

	// Initialize rate limiters
	utils.InitRateLimiters()

	// Use gin.New() instead of gin.Default() to have full control over middleware
	r := gin.New()
	
	// Add custom middleware
	r.Use(gin.Logger())
	r.Use(utils.CustomRecovery()) // Use our custom recovery instead of default
	
	r.RedirectTrailingSlash = false

	routes.SetupRoutes(r)

	// Start the server
	port := utils.GetEnvFatal("PORT")

	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
