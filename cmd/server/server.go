package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dragodui/diploma-server/internal/cache"
	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/Dragodui/diploma-server/internal/database"
	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/logger"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/router"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Server struct {
	router     http.Handler
	port       string
	httpServer *http.Server
	sqlCloser  interface{ Close() error }
	redis      interface{ Close() error }
}

func NewServer() (*Server, error) {
	logger.Init("app.log")
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DB_DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(
		&models.User{},
		&models.Home{},
		&models.HomeMembership{},
		&models.Task{},
		&models.TaskAssignment{},
		&models.Bill{},
		&models.BillCategory{},
		&models.ShoppingCategory{},
		&models.ShoppingItem{},
		&models.Poll{},
		&models.Option{},
		&models.Vote{},
		&models.Notification{},
		&models.HomeNotification{},
		&models.Room{},
		&models.HomeAssistantConfig{},
		&models.SmartDevice{},
	); err != nil {
		return nil, err
	}

	// Seed database with test data
	if err = database.SeedDatabase(db); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	cacheClient := cache.NewRedisClient(cfg.RedisADDR, cfg.RedisPassword, cfg.RedisTLS)

	// Mailer
	mailer := &utils.BrevoMailer{
		APIKey: cfg.BrevoAPIKey,
		From:   cfg.SMTPFrom,
	}

	// OAuth
	goth.UseProviders(
		google.New(cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL),
	)
	// repos
	userRepo := repository.NewUserRepository(db)
	homeRepo := repository.NewHomeRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	billRepo := repository.NewBillRepository(db)
	billCategoryRepo := repository.NewBillCategoryRepository(db)
	shoppingRepo := repository.NewShoppingRepository(db)
	pollRepo := repository.NewPollRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	smartHomeRepo := repository.NewSmartHomeRepository(db)

	// services
	authSvc := services.NewAuthService(userRepo, []byte(cfg.JWTSecret), cacheClient, 24*time.Hour, cfg.ClientURL, mailer)
	homeSvc := services.NewHomeService(homeRepo, cacheClient)
	roomSvc := services.NewRoomService(roomRepo, cacheClient)
	taskSvc := services.NewTaskService(taskRepo, cacheClient)
	billSvc := services.NewBillService(billRepo, cacheClient)
	billCategorySvc := services.NewBillCategoryService(billCategoryRepo, cacheClient)
	shoppingSvc := services.NewShoppingService(shoppingRepo, cacheClient)
	pollSvc := services.NewPollService(pollRepo, cacheClient)
	notificationSvc := services.NewNotificationService(notificationRepo, cacheClient)
	userService := services.NewUserService(userRepo, cacheClient)

	imageService, err := services.NewImageService(cfg.AWSS3Bucket, cfg.AWSRegion)
	if err != nil {
		log.Fatalf("error running S3: %s", err.Error())
	}

	ocrSvc := services.NewOCRService()
	smartHomeSvc := services.NewSmartHomeService(smartHomeRepo, cacheClient)

	// handlers
	authHandler := handlers.NewAuthHandler(authSvc)
	homeHandler := handlers.NewHomeHandler(homeSvc)
	roomHandler := handlers.NewRoomHandler(roomSvc)
	taskHandler := handlers.NewTaskHandler(taskSvc)
	billHandler := handlers.NewBillHandler(billSvc)
	billCategoryHandler := handlers.NewBillCategoryHandler(billCategorySvc)
	shoppingHandler := handlers.NewShoppingHandler(shoppingSvc)
	imageHandler := handlers.NewImageHandler(imageService)
	pollHandler := handlers.NewPollHandler(pollSvc)
	notificationHandler := handlers.NewNotificationHandler(notificationSvc)
	userHandler := handlers.NewUserHandler(userService, imageService)
	ocrHandler := handlers.NewOCRHandler(ocrSvc)
	smartHomeHandler := handlers.NewSmartHomeHandler(smartHomeSvc)

	// setup all routes
	router := router.SetupRoutes(cfg, authHandler, homeHandler, taskHandler, billHandler, billCategoryHandler, roomHandler, shoppingHandler, imageHandler, pollHandler, notificationHandler, userHandler, ocrHandler, smartHomeHandler, cacheClient, homeRepo)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		router:     router,
		port:       cfg.Port,
		httpServer: httpServer,
		sqlCloser:  sqlDB,
		redis:      cacheClient,
	}, nil
}

func (a *Server) Run() error {
	logger.Info.Print("Starting server on port:", a.port)
	serveErr := make(chan error, 1)

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
		logger.Info.Print("Shutdown signal received")
	case err := <-serveErr:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}

	var closeErrs []error
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}
	if a.sqlCloser != nil {
		if err := a.sqlCloser.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}

	return errors.Join(closeErrs...)
}
