package main

import (
	"log"
	"qb/internal/handlers"
	"qb/internal/routes"
	"qb/pkg/database"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

func init() {
	utils.LoadDotEnv()
	utils.InitJwt()
	utils.InitCloudinary()
	database.ConnectDB()
	utils.InitRequestTracker(database.DB)
	utils.InitRateLimiters()
	handlers.InitGenericService()
	handlers.InitCourseService()
}

func main() {
	r := gin.New()
	
	r.Use(gin.Logger())
	r.Use(utils.CustomRecovery())
	
	r.RedirectTrailingSlash = false

	routes.SetupRoutes(r)

	port := utils.GetEnvFatal("PORT")

	log.Printf("Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
