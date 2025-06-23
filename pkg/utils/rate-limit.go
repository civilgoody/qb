package utils

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple IP-based rate limiter
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter with specified limit and time window
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// Start cleanup goroutine
	go rl.startCleanupRoutine()
	
	return rl
}

// IsAllowed checks if a request from the given IP is allowed
func (rl *RateLimiter) IsAllowed(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-rl.window)
	
	// Get or create request history for this IP
	if _, exists := rl.requests[ip]; !exists {
		rl.requests[ip] = make([]time.Time, 0)
	}
	
	// Remove old requests outside the time window
	var validRequests []time.Time
	for _, reqTime := range rl.requests[ip] {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	rl.requests[ip] = validRequests
	
	// Check if limit is exceeded
	if len(rl.requests[ip]) >= rl.limit {
		return false
	}
	
	// Add current request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}

// startCleanupRoutine periodically cleans up old request histories
func (rl *RateLimiter) startCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute) // Clean up every 10 minutes
	defer ticker.Stop()
	
	for range ticker.C {
		rl.cleanupOldRequests()
	}
}

// cleanupOldRequests removes request histories that are completely outside the time window
func (rl *RateLimiter) cleanupOldRequests() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	cutoff := time.Now().Add(-rl.window)
	
	for ip, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		
		if len(validRequests) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = validRequests
		}
	}
}

// Rate limiters for different endpoints
var (
	UploadRateLimiter   *RateLimiter // 50 uploads per hour per IP
	GeneralRateLimiter  *RateLimiter // 200 requests per hour per IP
)

// InitRateLimiters initializes the rate limiters
func InitRateLimiters() {
	UploadRateLimiter = NewRateLimiter(50, time.Hour)   // 50 uploads per hour
	GeneralRateLimiter = NewRateLimiter(200, time.Hour) // 200 requests per hour
}

// RateLimitMiddleware creates a gin middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
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

// ValidateImageFile validates uploaded image files
func ValidateImageFile(fileHeader *multipart.FileHeader) error {
	// Check file size (10MB limit)
	if fileHeader.Size > 10*1024*1024 {
		return fmt.Errorf("file size exceeds 10MB limit")
	}
	
	// Check MIME type
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}
	
	contentType := http.DetectContentType(buffer)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	
	if !allowedTypes[contentType] {
		return fmt.Errorf("unsupported file type: %s. Allowed types: JPEG, PNG, WebP", contentType)
	}
	
	return nil
} 
