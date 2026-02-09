package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
)

func RequireAdmin(homeRepo repository.HomeRepository) func(http.Handler) http.Handler {
	type bodyWithHomeID struct {
		HomeID int `json:"home_id"`
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)

			// Try to get from URL first
			homeIDStr := chi.URLParam(r, "home_id")
			if homeIDStr != "" {
				if homeID, err := strconv.Atoi(homeIDStr); err == nil {
					if ok, _ := homeRepo.IsMember(r.Context(), homeID, userID); ok {
						next.ServeHTTP(w, r)
						return
					}
					utils.JSONError(w, "you are not a member", http.StatusUnauthorized)
					return
				}
			}

			// Try to get from body
			var bodyCopy bytes.Buffer
			tee := io.TeeReader(r.Body, &bodyCopy)

			var req bodyWithHomeID
			if err := json.NewDecoder(tee).Decode(&req); err != nil || req.HomeID == 0 {
				utils.JSONError(w, "invalid or missing home_id", http.StatusBadRequest)
				return
			}

			// Restore the original body for the next handler
			r.Body = io.NopCloser(&bodyCopy)

			ok, err := homeRepo.IsAdmin(r.Context(), req.HomeID, userID)
			if err != nil {
				utils.JSONError(w, "forbidden", http.StatusForbidden)
				return
			}

			if !ok {
				utils.JSONError(w, "you are not an admin", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireMember(homeRepo repository.HomeRepository) func(http.Handler) http.Handler {
	type bodyWithHomeID struct {
		HomeID int `json:"home_id"`
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)

			// Try to get from URL first
			homeIDStr := chi.URLParam(r, "home_id")
			if homeIDStr != "" {
				if homeID, err := strconv.Atoi(homeIDStr); err == nil {
					if ok, _ := homeRepo.IsMember(r.Context(), homeID, userID); ok {
						next.ServeHTTP(w, r)
						return
					}
					utils.JSONError(w, "you are not a member", http.StatusUnauthorized)
					return
				}
			}

			// Try to get from body
			var bodyCopy bytes.Buffer
			tee := io.TeeReader(r.Body, &bodyCopy)

			var req bodyWithHomeID
			if err := json.NewDecoder(tee).Decode(&req); err != nil || req.HomeID == 0 {
				utils.JSONError(w, "invalid or missing home_id", http.StatusBadRequest)
				return
			}

			// Restore the original body for the next handler
			r.Body = io.NopCloser(&bodyCopy)

			ok, err := homeRepo.IsMember(r.Context(), req.HomeID, userID)
			if err != nil {
				utils.JSONError(w, "forbidden", http.StatusForbidden)
				return
			}

			if !ok {
				utils.JSONError(w, "you are not a member", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
