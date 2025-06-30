package routes

import (
	"qb/internal/handlers"
	"qb/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		v1.GET("/", handlers.Status)
		
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/profile", handlers.Auth.JWTAuthMiddleware(), handlers.GetProfile)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Faculty routes
		faculty := v1.Group("/faculty")
		{
			faculty.GET("", handlers.GetFaculties) // Public read
			faculty.POST("", handlers.Auth.JWTAuthMiddleware(), handlers.CreateFaculty) // Protected
			faculty.DELETE("/:id", handlers.Auth.JWTAuthMiddleware(), handlers.DeleteFaculty) // Protected
		}

		// Level routes
		level := v1.Group("/level")
		{
			level.GET("", handlers.GetLevels) // Public read
			level.POST("", handlers.Auth.JWTAuthMiddleware(), handlers.CreateLevel) // Protected
			level.DELETE("/:id", handlers.Auth.JWTAuthMiddleware(), handlers.DeleteLevel) // Protected
		}

		// Session routes
		session := v1.Group("/session")
		{
			session.GET("", handlers.GetSessions) // Public read
			session.POST("", handlers.Auth.JWTAuthMiddleware(), handlers.CreateSession) // Protected
			session.DELETE("/:id", handlers.Auth.JWTAuthMiddleware(), handlers.DeleteSession) // Protected
		}

		// Course routes
		course := v1.Group("/course")
		{
			course.GET("", handlers.GetAllCourses) // Public read
			course.GET("/:dept/:level/:semester", handlers.FilterCourses) // Public read
			course.POST("", handlers.Auth.JWTAuthMiddleware(), handlers.CreateCourse) // Protected
			course.DELETE("/:id", handlers.Auth.JWTAuthMiddleware(), handlers.DeleteCourse) // Protected
		}

		// Department routes
		department := v1.Group("/department")
		{
			department.GET("", handlers.GetDepartments) // Public read
			department.POST("", handlers.Auth.JWTAuthMiddleware(), handlers.CreateDepartment) // Protected
			department.DELETE("/:id", handlers.Auth.JWTAuthMiddleware(), handlers.DeleteDepartment) // Protected
		}

		// Question routes
		question := v1.Group("/question")
		{
			question.GET("", handlers.GetQuestions) // Public read
			question.GET("/:id", handlers.GetQuestionByID) // Public read
			question.POST("", handlers.Auth.JWTAuthMiddleware(), handlers.CreateQuestion) // Protected
		}

		// Request routes
		request := v1.Group("/request")
		{
			request.GET("", handlers.Auth.JWTAuthMiddleware(), handlers.GetRequests) // Protected
		}

		// Image upload endpoint with rate limiting and auth
		v1.POST("/upload-images", 
			middleware.UploadRateLimit(), 
			handlers.Auth.JWTAuthMiddleware(), 
			handlers.UploadImages,
		) // Protected
	}
}
