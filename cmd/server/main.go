package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err = db.AutoMigrate(
		&models.User{},
		&models.Home{},
		&models.HomeMembership{},
		&models.Task{},
		&models.TaskAssignment{},
	); err != nil {
		panic(err)
	}

	// OAuth
	goth.UseProviders(
		google.New(cfg.ClientId, cfg.ClientSecret, cfg.CallbackURL),
	)
	// repos
	userRepo := repository.NewUserRepository(db)
	homeRepo := repository.NewHomeRepository(db)

	// services
	authSvc := services.NewAuthService(userRepo, []byte(cfg.JWTSecret), 24*time.Hour, cfg.ClientURL)
	homeSvc := services.NewHomeService(homeRepo)

	// handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	homeHandler := handlers.NewHomeHandler(homeSvc)

	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Get("/:provider", authHandler.SignInWithProvider)
			r.Get("/:provider/callback", authHandler.CallbackHandler)
		})

		r.Route("/home", func(r chi.Router) {
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).Post("/create", homeHandler.Create)
			r.With(middleware.JWTAuth([]byte(cfg.JWTSecret))).Post("/join", homeHandler.Join)
			r.With(middleware.RequireMember(homeRepo)).Get("/:id", homeHandler.GetByID)
			r.With(middleware.RequireAdmin(homeRepo)).Delete("/:id", homeHandler.Delete)
			r.With(middleware.RequireMember(homeRepo)).Post("/leave", homeHandler.Leave)
			r.With(middleware.RequireAdmin(homeRepo)).Delete("/remove_member", homeHandler.RemoveMember)
		})
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode("Hi")
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth([]byte(cfg.JWTSecret)))
		r.Get("/protected", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Protected content"))
		})
	})

	http.ListenAndServe(":"+cfg.Port, r)
}
