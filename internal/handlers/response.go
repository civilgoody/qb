package handlers

import (
	"qb/internal/services"
	"qb/pkg/models"

	"github.com/gin-gonic/gin"
)

// ResponseHelper handles all response formatting and error processing
type ResponseHelper struct {
	errorService *services.ErrorService
}

var Res *ResponseHelper

// InitResponseHelper initializes the shared response helper
func InitResponseHelper() {
	Res = &ResponseHelper{
		errorService: services.NewErrorService(),
	}
}

// Send handles all response types with optional message
func (h *ResponseHelper) Send(c *gin.Context, data interface{}, err error, message ...string) {
	if err != nil {
		h.sendError(c, err)
		return
	}
	
	msg := "success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	
	c.JSON(200, models.APIResponse{
		Code:    200,
		Message: msg,
		Data:    data,
	})
}

// Created handles 201 created responses
func (h *ResponseHelper) Created(c *gin.Context, data interface{}, err error) {
	if err != nil {
		h.sendError(c, err)
		return
	}
	
	c.JSON(201, models.APIResponse{
		Code:    201,
		Message: "created",
		Data:    data,
	})
}

// Invalid handles validation errors specifically
func (h *ResponseHelper) Invalid(c *gin.Context, err interface{}) {
	validationErr := h.errorService.Invalid(err)
	h.sendError(c, validationErr)
}

// sendError handles all error responses
func (h *ResponseHelper) sendError(c *gin.Context, err error) {
	if businessErr, ok := err.(*models.BusinessError); ok {
		c.JSON(businessErr.Code, models.APIResponse{
			Code:    businessErr.Code,
			Message: businessErr.Message,
			Error:   businessErr.Details,
		})
		return
	}
	
	// Unknown error defaults to internal server error
	c.JSON(500, models.APIResponse{
		Code:    500,
		Message: "Internal server error",
		Error:   err.Error(),
	})
} 

func (h *ResponseHelper) Unauthorized(c *gin.Context, details string) {
	Res.sendError(c, models.ErrUnauthorized)
}

func (h *ResponseHelper) Forbidden(c *gin.Context, details string) {
	Res.sendError(c, models.ErrForbidden)
}

func (h *ResponseHelper) NotFound(c *gin.Context, details string) {
	Res.sendError(c, models.ErrNotFound)
}

func (h *ResponseHelper) BadRequest(c *gin.Context, details string) {
	Res.sendError(c, models.ErrBadRequest)
}
