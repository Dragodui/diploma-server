package router

import (
	"encoding/json"
	"net/http"

	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRoutes(cfg *config.Config, authHandler *handlers.AuthHandler, homeHandler *handlers.HomeHandler, taskHandler *handlers.TaskHandler, billHandler *handlers.BillHandler, roomHandler *handlers.RoomHandler, homeRepo repository.HomeRepository) http.Handler {
	r := chi.NewRouter()
	// add cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.ClientURL},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Get("/{provider}", authHandler.SignInWithProvider)
			r.Get("/{provider}/callback", authHandler.CallbackHandler)
		})

		r.Route("/home", func(r chi.Router) {
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).Post("/create", homeHandler.Create)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).Post("/join", homeHandler.Join)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/{home_id}", homeHandler.GetByID)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Delete("/{home_id}", homeHandler.Delete)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Post("/leave", homeHandler.Leave)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Delete("/remove_member", homeHandler.RemoveMember)
		})

		r.Route("/room", func(r chi.Router) {
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Post("/create", roomHandler.Create)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/{room_id}", roomHandler.GetByID)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/home/{home_id}", roomHandler.GetByHomeID)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Delete("/{room_id}", roomHandler.Delete)
		})

		r.Route("/task", func(r chi.Router) {
			r.Use(middleware.JWTAuth([]byte(cfg.JWTSecret)))
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Post("/create", taskHandler.Create)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/{task_id}", taskHandler.GetByID)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/home/{home_id}", taskHandler.GetTasksByHomeID)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Delete("/{task_id}", taskHandler.DeleteTask)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Post("/assign_user", taskHandler.AssignUser)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Post("/{user_id}", taskHandler.GetAssignmentsForUser)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/user/{user_id}", taskHandler.GetClosestAssignmentForUser)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Patch("/mark_completed", taskHandler.MarkAssignmentCompleted)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Delete("/{assignment_id}", taskHandler.DeleteAssignment)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Patch("/reassign_room", taskHandler.ReassignRoom)
		})

		r.Route("/bill", func(r chi.Router) {
			r.Use(middleware.JWTAuth([]byte(cfg.JWTSecret)))
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Post("/create", billHandler.Create)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Get("/{bill_id}", billHandler.GetById)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireAdmin(homeRepo)).Delete("/{bill_id}", billHandler.Delete)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).With(middleware.RequireMember(homeRepo)).Patch("/{bill_id}", billHandler.MarkPayed)

		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode("Hi")
	})

	return r
}
