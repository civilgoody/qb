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
			faculty.DELETE("/:id", handlers.DeleteFaculty)
		}
		level := v1.Group("/level")
		{
			level.POST("", handlers.CreateLevel)
			level.GET("", handlers.GetLevels)
			level.DELETE("/:id", handlers.DeleteLevel)
		}
		session := v1.Group("/session")
		{
			session.POST("", handlers.CreateSession)
			session.GET("", handlers.GetSessions)
			session.DELETE("/:id", handlers.DeleteSession)
		}
		course := v1.Group("/course")
		{
			course.POST("", handlers.CreateCourse)
			course.GET("", handlers.GetCourses)
			course.GET("/:dept/:level/:semester", handlers.FilterCourses)
			course.DELETE("/:id", handlers.DeleteCourse)
		}
		department := v1.Group("/department")
		{
			department.POST("", handlers.CreateDepartment)
			department.GET("", handlers.GetDepartments)
			department.DELETE("/:id", handlers.DeleteDepartment)
		}
	}
}
