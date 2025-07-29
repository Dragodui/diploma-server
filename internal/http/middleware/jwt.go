package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/Dragodui/diploma-server/pkg/security"
)

type contextKey string

const userIDKey contextKey = "userID"

func JWTAuth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				utils.JSONError(w, "missing token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			claims, err := security.ParseToken(tokenStr, secret)
			if err != nil {
				utils.JSONError(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) int {
	val := r.Context().Value(userIDKey)
	if id, ok := val.(int); ok {
		return id
	}
	return 0
}
