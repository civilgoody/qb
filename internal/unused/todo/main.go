package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"todo-app/database"
	_ "todo-app/docs" // This import is crucial for gin-swagger to find the generated docs
	"todo-app/middleware"
	"todo-app/routes"
	"todo-app/utils"
)

// @title Go To-Do API
// @version 1.0
// @description This is a simple To-Do list API built with Go, Gin, and GORM.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Load .env file once at application startup
	utils.LoadDotEnv()

	// Connect to the database
	database.ConnectDB()

	// Set Gin to Release Mode for production
	env := os.Getenv("GIN_MODE")
	if env == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.Default()

	// Disable Gin's default trailing slash redirect
	router.RedirectTrailingSlash = false

	// Apply custom middleware to normalize trailing slashes
	router.Use(middleware.NormalizeTrailingSlash())

	// Setup routes
	routes.SetupRoutes(router)

	// ADD SWAGGER UI ROUTE
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server
	port := utils.GetEnvFatal("PORT")

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
 