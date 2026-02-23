package middleware

import (
	"net/http"
)

const (
	// MaxRequestBodySize limits request body to 10MB
	MaxRequestBodySize = 10 << 20 // 10MB
)

// BodySizeLimit middleware limits the size of request bodies
func BodySizeLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body size using MaxBytesReader
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

		next.ServeHTTP(w, r)
	})
}
