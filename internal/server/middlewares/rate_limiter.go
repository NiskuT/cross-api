package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// EndpointLimit defines rate limiting configuration for an endpoint
type EndpointLimit struct {
	MaxAttempts int
	Window      time.Duration
}

// RateLimiter handles rate limiting for different endpoints
type RateLimiter struct {
	limits   map[string]*EndpointLimit // endpoint -> limit config
	attempts map[string][]time.Time    // IP -> attempt timestamps
	mutex    sync.RWMutex
	stopChan chan struct{}
}

// NewRateLimiter creates a new rate limiter with default configurations
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		limits: map[string]*EndpointLimit{
			"login": {
				MaxAttempts: 5,
				Window:      5 * time.Minute,
			},
			"forgot-password": {
				MaxAttempts: 3,
				Window:      1 * time.Hour,
			},
		},
		attempts: make(map[string][]time.Time),
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// SetLimit sets or updates the rate limit for a specific endpoint
func (rl *RateLimiter) SetLimit(endpoint string, maxAttempts int, window time.Duration) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.limits[endpoint] = &EndpointLimit{
		MaxAttempts: maxAttempts,
		Window:      window,
	}
}

// getClientIP extracts the real client IP, considering trusted proxies
func (rl *RateLimiter) getClientIP(c *gin.Context) string {
	// Use Gin's ClientIP() which handles X-Forwarded-For when trusted proxies are set
	clientIP := c.ClientIP()

	// Validate IP format and fallback to RemoteAddr if invalid
	if net.ParseIP(clientIP) == nil {
		clientIP = c.Request.RemoteAddr
		// Strip port if present
		if host, _, err := net.SplitHostPort(clientIP); err == nil {
			clientIP = host
		}
	}

	return clientIP
}

// isAllowed checks if a request from the given IP is allowed for the endpoint
func (rl *RateLimiter) isAllowed(endpoint, clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limit, exists := rl.limits[endpoint]
	if !exists {
		// No limit configured for this endpoint, allow by default
		return true
	}

	now := time.Now()
	key := endpoint + ":" + clientIP

	// Get existing attempts for this IP+endpoint
	attempts, exists := rl.attempts[key]
	if !exists {
		attempts = []time.Time{}
	}

	// Remove attempts outside the time window
	var validAttempts []time.Time
	for _, attempt := range attempts {
		if now.Sub(attempt) < limit.Window {
			validAttempts = append(validAttempts, attempt)
		}
	}

	// Check if we're under the limit
	if len(validAttempts) >= limit.MaxAttempts {
		// Update the attempts list (without adding new attempt)
		rl.attempts[key] = validAttempts
		return false
	}

	// Add current attempt and allow
	validAttempts = append(validAttempts, now)
	rl.attempts[key] = validAttempts

	return true
}

// getRetryAfter calculates when the client can retry (in seconds)
func (rl *RateLimiter) getRetryAfter(endpoint, clientIP string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	limit, exists := rl.limits[endpoint]
	if !exists {
		return 0
	}

	key := endpoint + ":" + clientIP
	attempts, exists := rl.attempts[key]
	if !exists || len(attempts) == 0 {
		return 0
	}

	// Find the oldest attempt within the window
	now := time.Now()
	var oldestAttempt time.Time

	for _, attempt := range attempts {
		if now.Sub(attempt) < limit.Window {
			if oldestAttempt.IsZero() || attempt.Before(oldestAttempt) {
				oldestAttempt = attempt
			}
		}
	}

	if oldestAttempt.IsZero() {
		return 0
	}

	// Calculate when the oldest attempt will expire
	expiry := oldestAttempt.Add(limit.Window)
	retryAfter := int(expiry.Sub(now).Seconds())

	if retryAfter < 0 {
		return 0
	}

	return retryAfter
}

// Limit returns a middleware function that enforces rate limiting for the specified endpoint
func (rl *RateLimiter) Limit(endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := rl.getClientIP(c)

		if !rl.isAllowed(endpoint, clientIP) {
			retryAfter := rl.getRetryAfter(endpoint, clientIP)

			c.Header("X-RateLimit-Limit", "5")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "Too many requests",
				"message":             "Rate limit exceeded. Please try again later.",
				"retry_after_seconds": retryAfter,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanup periodically removes old attempts to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute) // Cleanup every 10 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()

			now := time.Now()
			for key, attempts := range rl.attempts {
				// Extract endpoint from key (format: "endpoint:ip")
				var maxWindow time.Duration
				for endpoint, limit := range rl.limits {
					if len(key) > len(endpoint) && key[:len(endpoint)] == endpoint {
						maxWindow = limit.Window
						break
					}
				}

				// Remove attempts older than the maximum window + buffer
				var validAttempts []time.Time
				cutoff := maxWindow + 5*time.Minute // 5 minute buffer

				for _, attempt := range attempts {
					if now.Sub(attempt) < cutoff {
						validAttempts = append(validAttempts, attempt)
					}
				}

				if len(validAttempts) == 0 {
					delete(rl.attempts, key)
				} else {
					rl.attempts[key] = validAttempts
				}
			}

			rl.mutex.Unlock()

		case <-rl.stopChan:
			return
		}
	}
}

// Stop gracefully stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// ResetAttempts clears the attempt history for a specific endpoint and IP
// This should be called when authentication is successful to reset the rate limit
func (rl *RateLimiter) ResetAttempts(endpoint, clientIP string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	key := endpoint + ":" + clientIP
	delete(rl.attempts, key)
}

// GetClientIP extracts the real client IP from the gin context (public method)
func (rl *RateLimiter) GetClientIP(c *gin.Context) string {
	return rl.getClientIP(c)
}
