package middleware

import (
	"net/http"
)

// SecurityHeaders adds security-related HTTP headers to responses
// These headers protect against common web vulnerabilities
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src * data:; font-src 'self' data:")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
