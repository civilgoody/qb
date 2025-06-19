package routes

import (
	"github.com/gin-gonic/gin"

	"todo-app/handlers"
)

// SetupRoutes configures all the API routes for the To-Do application.
func SetupRoutes(router *gin.Engine) {
	// Group routes under /api/v1
	v1 := router.Group("/api/v1")
	{
		todos := v1.Group("/todos")
		{
			todos.POST("", handlers.CreateTodo)
			todos.GET("", handlers.GetTodos)

			todos.GET("/:id", handlers.GetTodoByID)
			todos.PUT("/:id", handlers.UpdateTodo)
			todos.DELETE("/:id", handlers.DeleteTodo)
		}
	}
}
 