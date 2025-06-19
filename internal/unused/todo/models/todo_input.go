package models

// CreateTodoInput defines the structure for creating a new To-Do item via API request.
// It only includes fields that the client is allowed to provide.
type CreateTodoInput struct {
	Title       string  `json:"title" binding:"required" validate:"required,min=3,max=100" example:"Plan weekend trip"`
	Description *string `json:"description" validate:"omitempty,max=500" example:"Research destinations and activities"`
	Completed   bool    `json:"completed" example:"false"`
}

// UpdateTodoInput defines the structure for updating an existing To-Do item via API request.
// Fields are pointers to allow for partial updates (omitempty).
type UpdateTodoInput struct {
	Title       *string `json:"title" validate:"omitempty,min=3,max=100" example:"Book flight tickets"`
	Description *string `json:"description" validate:"omitempty,max=500" example:"Confirm dates and prices"`
	Completed   *bool   `json:"completed" example:"true"`
} 
