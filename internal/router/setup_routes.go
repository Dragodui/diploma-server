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

// SetupRoutes configures all application routes
func SetupRoutes(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	homeHandler *handlers.HomeHandler,
	taskHandler *handlers.TaskHandler,
	billHandler *handlers.BillHandler,
	roomHandler *handlers.RoomHandler,
	shoppingHandler *handlers.ShoppingHandler,
	imageHandler *handlers.ImageHandler,
	// home repo for middleware
	homeRepo repository.HomeRepository,

) http.Handler {
	r := chi.NewRouter()
	// HTTP request logger
	r.Use(middleware.RequestResponseLogger)
	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.ClientURL},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Public healthcheck
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode("OK")
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes (no JWT required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Get("/{provider}", authHandler.SignInWithProvider)
			r.Get("/{provider}/callback", authHandler.CallbackHandler)
			r.Get("/verify", authHandler.VerifyEmail)
			r.Post("/forgot", authHandler.ForgotPassword)
			r.Post("/reset", authHandler.ResetPassword)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			// JWT authentication for all following routes
			r.Use(middleware.JWTAuth([]byte(cfg.JWTSecret)))

			// upload images
			r.Post("/upload", imageHandler.UploadImage)

			// Homes and nested resources
			r.Route("/homes", func(r chi.Router) {
				r.Post("/create", homeHandler.Create) // Create home
				r.Post("/join", homeHandler.Join)     // Join home
				r.Get("/my", homeHandler.GetUserHome) // Get user home

				// Home-specific actions
				r.Route("/{home_id}", func(r chi.Router) {
					r.With(middleware.RequireMember(homeRepo)).Get("/", homeHandler.GetByID)
					r.With(middleware.RequireAdmin(homeRepo)).Delete("/", homeHandler.Delete)
					r.With(middleware.RequireMember(homeRepo)).Post("/leave", homeHandler.Leave)
					r.With(middleware.RequireAdmin(homeRepo)).Delete("/members/{user_id}", homeHandler.RemoveMember)

					// Rooms under a home
					r.Route("/rooms", func(r chi.Router) {
						r.With(middleware.RequireAdmin(homeRepo)).Post("/", roomHandler.Create)
						r.With(middleware.RequireMember(homeRepo)).Get("/", roomHandler.GetByHomeID)
						r.With(middleware.RequireMember(homeRepo)).Get("/{room_id}", roomHandler.GetByID)
						r.With(middleware.RequireAdmin(homeRepo)).Delete("/{room_id}", roomHandler.Delete)
					})

					// Tasks under a home
					r.Route("/tasks", func(r chi.Router) {
						r.With(middleware.RequireMember(homeRepo)).Post("/", taskHandler.Create)
						r.With(middleware.RequireMember(homeRepo)).Get("/", taskHandler.GetTasksByHomeID)
						r.With(middleware.RequireMember(homeRepo)).Get("/{task_id}", taskHandler.GetByID)
						r.With(middleware.RequireAdmin(homeRepo)).Delete("/{task_id}", taskHandler.DeleteTask)
						// Assignments
						r.With(middleware.RequireMember(homeRepo)).Post("/{task_id}/assign", taskHandler.AssignUser)
						r.With(middleware.RequireMember(homeRepo)).Patch("/{task_id}/reassign-room", taskHandler.ReassignRoom)
						r.With(middleware.RequireMember(homeRepo)).Patch("/{task_id}/mark-completed", taskHandler.MarkAssignmentCompleted)
						r.With(middleware.RequireAdmin(homeRepo)).Delete("/{task_id}/assignments/{assignment_id}", taskHandler.DeleteAssignment)
					})

					// User assignments (not scoped to a specific home)
					r.Route("/users/{user_id}/assignments", func(r chi.Router) {
						r.With(middleware.RequireMember(homeRepo)).Get("/", taskHandler.GetAssignmentsForUser)
						r.With(middleware.RequireMember(homeRepo)).Get("/closest", taskHandler.GetClosestAssignmentForUser)
					})

					// Bills under a home
					r.Route("/bills", func(r chi.Router) {
						r.With(middleware.RequireMember(homeRepo)).Post("/", billHandler.Create)
						r.With(middleware.RequireMember(homeRepo)).Get("/{bill_id}", billHandler.GetById)
						r.With(middleware.RequireAdmin(homeRepo)).Delete("/{bill_id}", billHandler.Delete)
						r.With(middleware.RequireMember(homeRepo)).Patch("/{bill_id}", billHandler.MarkPayed)
					})

					// Shopping
					r.Route("/shopping", func(r chi.Router) {
						r.Route("/categories", func(r chi.Router) {
							r.With(middleware.RequireMember(homeRepo)).Post("/", shoppingHandler.CreateCategory)
							r.With(middleware.RequireMember(homeRepo)).Get("/all", shoppingHandler.GetAllCategories)
							r.With(middleware.RequireMember(homeRepo)).Get("/{category_id}", shoppingHandler.GetCategoryByID)
							r.With(middleware.RequireMember(homeRepo)).Delete("/{category_id}", shoppingHandler.DeleteCategory)
							r.With(middleware.RequireMember(homeRepo)).Put("/{category_id}", shoppingHandler.EditCategory)
						})
						r.Route("/items", func(r chi.Router) {
							r.With(middleware.RequireMember(homeRepo)).Post("/", shoppingHandler.CreateItem)
							r.With(middleware.RequireMember(homeRepo)).Get("/category/{category_id}", shoppingHandler.GetItemsByCategoryID)
							r.With(middleware.RequireMember(homeRepo)).Get("/{item_id}", shoppingHandler.GetItemByID)
							r.With(middleware.RequireMember(homeRepo)).Delete("/{item_id}", shoppingHandler.DeleteItem)
							r.With(middleware.RequireMember(homeRepo)).Put("/{item_id}", shoppingHandler.EditItem)
						})
					})
				})
			})
		})
	})

	return r
}
