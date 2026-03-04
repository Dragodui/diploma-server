package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Dragodui/diploma-server/internal/metrics"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/Dragodui/diploma-server/pkg/security"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const tokenBlacklistPrefix = "blacklist:"

func JWTAuth(secret []byte, cache *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				metrics.AuthTokensValidated.WithLabelValues("invalid").Inc()
				utils.JSONError(w, "missing token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")

			// Check if token has been revoked (logout)
			if val, err := cache.Exists(r.Context(), tokenBlacklistPrefix+tokenStr).Result(); err == nil && val > 0 {
				metrics.AuthTokensValidated.WithLabelValues("invalid").Inc()
				utils.JSONError(w, "token revoked", http.StatusUnauthorized)
				return
			}

			claims, err := security.ParseToken(tokenStr, secret)
			if err != nil {
				metrics.AuthTokensValidated.WithLabelValues("invalid").Inc()
				utils.JSONError(w, "invalid token", http.StatusUnauthorized)
				return
			}

			metrics.AuthTokensValidated.WithLabelValues("valid").Inc()
			ctx := context.WithValue(r.Context(), utils.UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) int {
	val := r.Context().Value(utils.UserIDKey)
	if id, ok := val.(int); ok {
		return id
	}
	return 0
}

func GetUser(r *http.Request, db *gorm.DB) (*models.User, error) {
	id := GetUserID(r)
	if id == 0 {
		return nil, errors.New("invalid user ID")
	}

	userRepo := repository.NewUserRepository(db)
	user, err := userRepo.FindByID(r.Context(), id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
