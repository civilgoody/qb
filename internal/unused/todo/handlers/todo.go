package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"todo-app/database"
	"todo-app/models"
	"todo-app/utils"
)

// Initialize a validator instance
var validate = validator.New()

// formatValidationErrors converts validator.ValidationErrors into a map for JSON response.
func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	for _, fieldErr := range err.(validator.ValidationErrors) {
		errors[fieldErr.Field()] = fieldErr.Tag()
	}
	return errors
}

// CreateTodo handles the creation of a new To-Do item.
// @Summary Create a new To-Do
// @Description Create a new To-Do item in the database.
// @Tags todos
// @Accept json
// @Produce json
// @Param todo body models.CreateTodoInput true "To-Do object to be created"
// @Success 201 {object} models.Todo
// @Failure 400 {object} models.ValidationErrorResponse "Bad Request: Validation Errors"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /todos [post]
func CreateTodo(c *gin.Context) {
	var input models.CreateTodoInput // Use the input DTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"validation_errors": formatValidationErrors(validationErrors)})
			return
		}
		utils.HandleInternalServerError(c, err)
		return
	}

	// Map input DTO to database model
	todo := models.Todo{
		Title:       input.Title,
		Description: input.Description,
		Completed:   input.Completed,
	}

	result := database.DB.Create(&todo)
	if result.Error != nil {
		utils.HandleInternalServerError(c, result.Error)
		return
	}

	c.JSON(http.StatusCreated, todo)
}

// GetTodos handles retrieving all To-Do items.
// @Summary Get all To-Dos
// @Description Retrieve a list of all To-Do items from the database.
// @Tags todos
// @Produce json
// @Success 200 {array} models.Todo
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /todos [get]
func GetTodos(c *gin.Context) {
	var todos []models.Todo
	result := database.DB.Find(&todos)
	if result.Error != nil {
		utils.HandleInternalServerError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, todos)
}

// GetTodoByID handles retrieving a single To-Do item by its ID.
// @Summary Get To-Do by ID
// @Description Retrieve a single To-Do item from the database using its ID.
// @Tags todos
// @Produce json
// @Param id path string true "To-Do ID"
// @Success 200 {object} models.Todo
// @Failure 404 {object} models.ErrorResponse "Not Found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /todos/{id} [get]
func GetTodoByID(c *gin.Context) {
	id := c.Param("id")

	var todo models.Todo
	result := database.DB.First(&todo, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		utils.HandleInternalServerError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, todo)
}

// UpdateTodo handles updating an existing To-Do item.
// @Summary Update a To-Do
// @Description Update an existing To-Do item in the database.
// @Tags todos
// @Accept json
// @Produce json
// @Param id path string true "To-Do ID"
// @Param todo body models.UpdateTodoInput true "To-Do object with updated fields"
// @Success 200 {object} models.Todo
// @Failure 400 {object} models.ValidationErrorResponse "Bad Request: Validation Errors"
// @Failure 404 {object} models.ErrorResponse "Not Found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /todos/{id} [put]
func UpdateTodo(c *gin.Context) {
	id := c.Param("id")

	var existingTodo models.Todo
	result := database.DB.First(&existingTodo, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		utils.HandleInternalServerError(c, result.Error)
		return
	}

	var input models.UpdateTodoInput // Use the input DTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validate.Struct(input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{"validation_errors": formatValidationErrors(validationErrors)})
			return
		}
		utils.HandleInternalServerError(c, err)
		return
	}

	// Apply updates from DTO to the existing Todo model
	updates := make(map[string]interface{})
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Completed != nil {
		updates["completed"] = *input.Completed
	}

	result = database.DB.Model(&existingTodo).Updates(updates)
	if result.Error != nil {
		utils.HandleInternalServerError(c, result.Error)
		return
	}

	c.JSON(http.StatusOK, existingTodo)
}

// DeleteTodo handles deleting a To-Do item by its ID.
// @Summary Delete a To-Do
// @Description Delete a To-Do item from the database using its ID.
// @Tags todos
// @Produce json
// @Param id path string true "To-Do ID"
// @Success 204 "No Content"
// @Failure 404 {object} models.ErrorResponse "Not Found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /todos/{id} [delete]
func DeleteTodo(c *gin.Context) {
	id := c.Param("id")

	result := database.DB.Delete(&models.Todo{}, "id = ?", id)
	if result.Error != nil {
		utils.HandleInternalServerError(c, result.Error)
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
 