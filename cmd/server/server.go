package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Dragodui/diploma-server/internal/cache"
	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/router"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	router http.Handler
	port   string
}

func NewServer() *Server {
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

	cache := cache.NewRedisClient()

	// OAuth
	goth.UseProviders(
		google.New(cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL),
	)
	// repos
	userRepo := repository.NewUserRepository(db)
	homeRepo := repository.NewHomeRepository(db)
	taskRepo := repository.NewTaskRepository(db)

	// services
	authSvc := services.NewAuthService(userRepo, []byte(cfg.JWTSecret), cache, 24*time.Hour, cfg.ClientURL)
	homeSvc := services.NewHomeService(homeRepo, cache)
	taskSvc := services.NewTaskService(taskRepo, cache)

	// handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	homeHandler := handlers.NewHomeHandler(homeSvc)
	taskHandler := handlers.NewTaskHandler(taskSvc)

	// setup all routes
	router := router.SetupRoutes(cfg, authHandler, homeHandler, taskHandler, homeRepo)

	return &Server{router: router, port: cfg.Port}
}

func (a *Server) Run() {
	log.Println("Starting server on port:", a.port)
	if err := http.ListenAndServe(":"+a.port, a.router); err != nil {
		log.Fatal(err)
	}
}
