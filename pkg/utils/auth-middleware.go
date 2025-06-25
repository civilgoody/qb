package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware validates JWT tokens and sets user context
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			ErrorResponse(c, http.StatusUnauthorized, ErrUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ErrorResponse(c, http.StatusUnauthorized, ErrUnauthorized, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate the token
		claims, err := ValidateToken(token)
		if err != nil {
			ErrorResponse(c, http.StatusUnauthorized, ErrUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RequireAdmin middleware ensures only admin users can access the route
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			ErrorResponse(c, http.StatusForbidden, ErrForbidden, "User role not found in context")
			c.Abort()
			return
		}

		if userRole != "ADMIN" {
			ErrorResponse(c, http.StatusForbidden, ErrForbidden, "Admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that sets user context if token is provided but doesn't require it
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]
		claims, err := ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context if token is valid
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// GetCurrentUserID helper function to get user ID from context
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	fmt.Println("userID", userID)
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// GetCurrentUserRole helper function to get user role from context
func GetCurrentUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("userRole")
	if !exists {
		return "", false
	}
	role, ok := userRole.(string)
	return role, ok
} 
