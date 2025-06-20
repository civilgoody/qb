package routes

import (
	"qb/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/", handlers.Status)
		faculty := v1.Group("/faculty")
		{
			faculty.POST("", handlers.CreateFaculty)
			faculty.GET("", handlers.GetFaculties)
		}
	}
}
