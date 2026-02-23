package middleware

import (
	"crypto/subtle"
	"net/http"
)

// BasicAuth provides HTTP Basic Authentication
// Used for protecting /metrics and /swagger endpoints
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get credentials from request
			user, pass, ok := r.BasicAuth()

			// Check if credentials are provided and valid
			// Use constant-time comparison to prevent timing attacks
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {

				// Send 401 with WWW-Authenticate header
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
