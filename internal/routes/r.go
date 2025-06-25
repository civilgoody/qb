package routes

import (
	"qb/internal/handlers"
	"qb/pkg/utils"

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
			auth.GET("/profile", utils.JWTAuthMiddleware(), handlers.GetProfile)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Faculty routes
		faculty := v1.Group("/faculty")
		{
			faculty.GET("", handlers.GetFaculties) // Public read
			faculty.POST("", utils.JWTAuthMiddleware(), handlers.CreateFaculty) // Protected
			faculty.DELETE("/:id", utils.JWTAuthMiddleware(), handlers.DeleteFaculty) // Protected
		}

		// Level routes
		level := v1.Group("/level")
		{
			level.GET("", handlers.GetLevels) // Public read
			level.POST("", utils.JWTAuthMiddleware(), handlers.CreateLevel) // Protected
			level.DELETE("/:id", utils.JWTAuthMiddleware(), handlers.DeleteLevel) // Protected
		}

		// Session routes
		session := v1.Group("/session")
		{
			session.GET("", handlers.GetSessions) // Public read
			session.POST("", utils.JWTAuthMiddleware(), handlers.CreateSession) // Protected
			session.DELETE("/:id", utils.JWTAuthMiddleware(), handlers.DeleteSession) // Protected
		}

		// Course routes
		course := v1.Group("/course")
		{
			course.GET("", handlers.GetCourses) // Public read
			course.GET("/:dept/:level/:semester", handlers.FilterCourses) // Public read
			course.POST("", utils.JWTAuthMiddleware(), handlers.CreateCourse) // Protected
			course.DELETE("/:id", utils.JWTAuthMiddleware(), handlers.DeleteCourse) // Protected
		}

		// Department routes
		department := v1.Group("/department")
		{
			department.GET("", handlers.GetDepartments) // Public read
			department.POST("", utils.JWTAuthMiddleware(), handlers.CreateDepartment) // Protected
			department.DELETE("/:id", utils.JWTAuthMiddleware(), handlers.DeleteDepartment) // Protected
		}

		// Question routes
		question := v1.Group("/question")
		{
			question.GET("", handlers.GetQuestions) // Public read
			question.GET("/:id", handlers.GetQuestionByID) // Public read
			question.POST("", utils.JWTAuthMiddleware(), handlers.CreateQuestion) // Protected
		}

		// Request routes
		request := v1.Group("/request")
		{
			request.GET("", utils.JWTAuthMiddleware(), handlers.GetRequests) // Protected
		}

		// Image upload endpoint with rate limiting and auth
		v1.POST("/upload-images", 
			utils.RateLimitMiddleware(utils.UploadRateLimiter), 
			utils.JWTAuthMiddleware(), 
			handlers.UploadImages,
		) // Protected
	}
}
