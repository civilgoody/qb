package main

import (
	"qb/internal/routes"
	"qb/pkg/database"
	"qb/pkg/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	// Load .env file once at application startup
	utils.LoadDotEnv()

	// Connect to the database
	database.ConnectDB()

	router := gin.Default()
	routes.SetupRoutes(router)
	router.RedirectTrailingSlash = false
	router.Run(":8080")

	// Start the server
	port := utils.GetEnvFatal("PORT")

	log.Printf("Server starting on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
