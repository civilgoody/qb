package handlers

import (
	"log"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHelper handles auth-related HTTP concerns
type AuthHelper struct{}

var Auth *AuthHelper

// InitAuthHelper initializes the shared auth helper
func InitAuthHelper() {
	Auth = &AuthHelper{}
}

// JWTAuthMiddleware validates JWT tokens and sets user context
func (h *AuthHelper) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Res.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			Res.Unauthorized(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate the token
		claims, err := utils.ValidateToken(token)
		if err != nil {
			Res.Unauthorized(c, "Invalid or expired token")
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
func (h *AuthHelper) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			Res.Forbidden(c, "User role not found in context")
			c.Abort()
			return
		}

		if userRole != "ADMIN" {
			Res.Forbidden(c, "Admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth middleware that sets user context if token is provided but doesn't require it
func (h *AuthHelper) OptionalAuth() gin.HandlerFunc {
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
		claims, err := utils.ValidateToken(token)
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

// GetCurrentUserID extracts user ID from authenticated context
func (h *AuthHelper) GetCurrentUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", models.ErrInternal
	}
	id := userID.(string)
	return id, nil
}

// GetCurrentUserRole extracts user role from authenticated context
func (h *AuthHelper) GetCurrentUserRole(c *gin.Context) (string, error) {
	userRole, exists := c.Get("userRole")
	if !exists {
		return "", models.ErrInternal
	}
	role := userRole.(string)
	return role, nil
}

func (h *AuthHelper) CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic for debugging
				log.Printf("Panic occurred: %v", err)
				
				Res.sendError(c, models.ErrInternal)
				c.Abort()
			}
		}()
		c.Next()
	}
} 
