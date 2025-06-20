package utils

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// CustomRecovery provides enhanced panic recovery with structured error responses
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic for debugging
				log.Printf("Panic occurred: %v", err)
				
				// Return structured error response
				ErrorResponse(c, 500, ErrInternal, fmt.Sprintf("Panic: %v", err))
				c.Abort()
			}
		}()
		c.Next()
	}
} 
