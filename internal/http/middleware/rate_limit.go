package middleware

import (
	"net/http"
	"strings"

	ratelimiter "github.com/Dragodui/diploma-server/internal/http/rate_limiter"
	"github.com/Dragodui/diploma-server/internal/utils"
)

// RateLimitMiddleware creates a middleware that limits requests per IP
func RateLimitMiddleware(limiter *ratelimiter.IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			rateLimiter := limiter.GetLimiter(ip)
			if !rateLimiter.Allow() {
				utils.JSONError(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StrictRateLimitMiddleware creates a stricter rate limiter for sensitive endpoints
func StrictRateLimitMiddleware(limiter *ratelimiter.IPRateLimiter, maxRequests float64, refillRate float64) func(http.Handler) http.Handler {
	strictLimiters := make(map[string]*ratelimiter.RateLimiter)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)

			// Get or create strict limiter for this IP
			strictLimiter, exists := strictLimiters[ip]
			if !exists {
				strictLimiter = ratelimiter.NewRateLimiter(maxRequests, refillRate)
				strictLimiters[ip] = strictLimiter
			}

			if !strictLimiter.Allow() {
				utils.JSONError(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the real IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the chain
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
