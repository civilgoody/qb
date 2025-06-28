package handlers

import (
	"qb/internal/services"
	"time"
)

// Rate limiter instances
var (
	uploadRateLimiter   *services.RateLimitService
	generalRateLimiter  *services.RateLimitService
)

// InitRateLimitServices initializes the rate limiting services
func InitRateLimitServices() {
	uploadRateLimiter = services.NewRateLimitService(50, time.Hour)   // 50 uploads per hour
	generalRateLimiter = services.NewRateLimitService(200, time.Hour) // 200 requests per hour
}

// GetUploadRateLimiter returns the upload rate limiter
func GetUploadRateLimiter() *services.RateLimitService {
	return uploadRateLimiter
}

// GetGeneralRateLimiter returns the general rate limiter
func GetGeneralRateLimiter() *services.RateLimitService {
	return generalRateLimiter
} 
