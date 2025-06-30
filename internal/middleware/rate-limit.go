package middleware

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

// UploadRateLimit creates middleware for upload rate limiting
func UploadRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(services.GetUploadRateLimiter())
}

// GeneralRateLimit creates middleware for general rate limiting
func GeneralRateLimit() gin.HandlerFunc {
	return RateLimitMiddleware(services.GetGeneralRateLimiter())
} 
