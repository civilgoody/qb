package handlers

import (
	"net/http"
	"qb/internal/services"

	"github.com/gin-gonic/gin"
)

// RateLimitMiddleware creates a gin middleware for rate limiting using services
func RateLimitMiddleware(limiter *services.RateLimitService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !limiter.IsAllowed(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded. Please try again later.",
				"code":    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// UploadRateLimitMiddleware creates middleware for upload rate limiting
func UploadRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(uploadRateLimiter)
}

// GeneralRateLimitMiddleware creates middleware for general rate limiting
func GeneralRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(generalRateLimiter)
}
