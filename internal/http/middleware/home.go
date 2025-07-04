package middleware

import (
	"net/http"
	"strconv"

	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/go-chi/chi/v5"
)

func RequireAdmin(homeRepo repository.HomeRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			homeIDStr := chi.URLParam(r, "homeID")
			homeID, err := strconv.Atoi(homeIDStr)
			if err != nil {
				http.Error(w, "invalid home ID", http.StatusBadRequest)
				return
			}

			isAdmin, err := homeRepo.IsAdmin(homeID, userID)
			if err != nil || !isAdmin {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireMember(homeRepo repository.HomeRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			homeIDStr := chi.URLParam(r, "homeID")
			homeID, err := strconv.Atoi(homeIDStr)
			if err != nil {
				http.Error(w, "invalid home ID", http.StatusBadRequest)
				return
			}

			isMember, err := homeRepo.IsMember(homeID, userID)
			if err != nil || !isMember {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
