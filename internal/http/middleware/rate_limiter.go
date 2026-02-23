package middleware

import (
	"net"
	"net/http"

	ratelimiter "github.com/Dragodui/diploma-server/internal/http/rate_limiter"
)

func RateLimiterMiddleware(ipRateLimiter *ratelimiter.IPRateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)

		if err != nil {
			http.Error(w, "Invalid IP", http.StatusInternalServerError)
			return
		}

		limiter := ipRateLimiter.GetLimiter(ip)
		if limiter.Allow() {
			next(w, r)
		} else {
			http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		}
	}
}
