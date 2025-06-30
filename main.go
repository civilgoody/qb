package main

import (
	"log"
	"qb/internal/handlers"
	"qb/internal/routes"
	"qb/internal/services"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	utils.LoadDotEnv()
	utils.InitJwt()
	
	// Initialize all services with shared dependencies
	services.InitServices()
	
	// Initialize handlers
	handlers.InitResponseHelper()
	handlers.InitAuthHelper()
}

func main() {
	r := gin.New()
	
	r.Use(gin.Logger())
	r.Use(handlers.Auth.CustomRecovery())
	
	r.RedirectTrailingSlash = false

	routes.SetupRoutes(r)

	port := utils.GetEnvFatal("PORT")

	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
