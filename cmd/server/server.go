package main

import (
	"log"
	"net/http"
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
	router http.Handler
	port   string
}

func NewServer() *Server {
	logger.Init("app.log")
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
		panic(err)
	}

	// Seed database with test data
	if err = database.SeedDatabase(db); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	cache := cache.NewRedisClient(cfg.RedisADDR, cfg.RedisPassword)

	// Mailer
	mailer := &utils.SMTPMailer{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPass,
		From:     cfg.SMTPFrom,
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
	authSvc := services.NewAuthService(userRepo, []byte(cfg.JWTSecret), cache, 24*time.Hour, cfg.ClientURL, mailer)
	homeSvc := services.NewHomeService(homeRepo, cache)
	roomSvc := services.NewRoomService(roomRepo, cache)
	taskSvc := services.NewTaskService(taskRepo, cache)
	billSvc := services.NewBillService(billRepo, cache)
	billCategorySvc := services.NewBillCategoryService(billCategoryRepo, cache)
	shoppingSvc := services.NewShoppingService(shoppingRepo, cache)
	pollSvc := services.NewPollService(pollRepo, cache)
	notificationSvc := services.NewNotificationService(notificationRepo, cache)
	userService := services.NewUserService(userRepo, cache)

	imageService, err := services.NewImageService(cfg.AWSS3Bucket, cfg.AWSRegion)
	if err != nil {
		log.Fatalf("error running S3: %s", err.Error())
	}

	ocrSvc := services.NewOCRService()
	smartHomeSvc := services.NewSmartHomeService(smartHomeRepo, cache)

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
	router := router.SetupRoutes(cfg, authHandler, homeHandler, taskHandler, billHandler, billCategoryHandler, roomHandler, shoppingHandler, imageHandler, pollHandler, notificationHandler, userHandler, ocrHandler, smartHomeHandler, cache, homeRepo)

	return &Server{router: router, port: cfg.Port}
}

func (a *Server) Run() {
	logger.Info.Print("Starting server on port:", a.port)
	if err := http.ListenAndServe(":"+a.port, a.router); err != nil {
		log.Fatal(err)
	}
}
