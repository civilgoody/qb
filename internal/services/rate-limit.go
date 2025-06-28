package services

import (
	"sync"
	"time"
)

// RateLimitService handles rate limiting business logic
type RateLimitService struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimitService creates a new rate limiter with specified limit and time window
func NewRateLimitService(limit int, window time.Duration) *RateLimitService {
	service := &RateLimitService{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	
	// Start cleanup goroutine
	go service.startCleanupRoutine()
	
	return service
}

// IsAllowed checks if a request from the given IP is allowed
func (s *RateLimitService) IsAllowed(ip string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-s.window)
	
	// Get or create request history for this IP
	if _, exists := s.requests[ip]; !exists {
		s.requests[ip] = make([]time.Time, 0)
	}
	
	// Remove old requests outside the time window
	var validRequests []time.Time
	for _, reqTime := range s.requests[ip] {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	s.requests[ip] = validRequests
	
	// Check if limit is exceeded
	if len(s.requests[ip]) >= s.limit {
		return false
	}
	
	// Add current request
	s.requests[ip] = append(s.requests[ip], now)
	return true
}

// GetRemainingRequests returns how many requests the IP has left in the current window
func (s *RateLimitService) GetRemainingRequests(ip string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if requests, exists := s.requests[ip]; exists {
		return s.limit - len(requests)
	}
	return s.limit
}

// GetWindowResetTime returns when the rate limit window resets for the given IP
func (s *RateLimitService) GetWindowResetTime(ip string) time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	if requests, exists := s.requests[ip]; exists && len(requests) > 0 {
		return requests[0].Add(s.window)
	}
	return time.Now().Add(s.window)
}

// startCleanupRoutine periodically cleans up old request histories
func (s *RateLimitService) startCleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute) // Clean up every 10 minutes
	defer ticker.Stop()
	
	for range ticker.C {
		s.cleanupOldRequests()
	}
}

// cleanupOldRequests removes request histories that are completely outside the time window
func (s *RateLimitService) cleanupOldRequests() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	cutoff := time.Now().Add(-s.window)
	
	for ip, requests := range s.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		
		if len(validRequests) == 0 {
			delete(s.requests, ip)
		} else {
			s.requests[ip] = validRequests
		}
	}
} 
