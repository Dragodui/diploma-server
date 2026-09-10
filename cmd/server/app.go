package main

import (
	"time"

	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/Dragodui/diploma-server/internal/http/handlers"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/Dragodui/diploma-server/internal/router"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type appDeps struct {
	cfg      *config.Config
	db       *gorm.DB
	cache    *redis.Client
	repos    repositories
	services serviceSet
	handlers handlerSet
}

type repositories struct {
	user         repository.UserRepository
	home         repository.HomeRepository
	room         repository.RoomRepository
	task         repository.TaskRepository
	bill         repository.BillRepository
	billCategory repository.IBillCategoryRepository
	shopping     repository.ShoppingRepository
	poll         repository.PollRepository
	notification repository.NotificationRepository
	audit        repository.AuditRepository
	smartHome    repository.SmartHomeRepository
	taskSchedule repository.TaskScheduleRepository
	pushSub      repository.PushSubscriptionRepository
	note         repository.NoteRepository
	chat         repository.ChatRepository
}

type serviceSet struct {
	pushSub      *services.PushSubscriptionService
	notification *services.NotificationService
	audit        *services.AuditService
	auth         *services.AuthService
	home         *services.HomeService
	room         *services.RoomService
	task         *services.TaskService
	bill         *services.BillService
	billCategory *services.BillCategoryService
	shopping     *services.ShoppingService
	poll         *services.PollService
	user         *services.UserService
	image        *services.ImageService
	ocr          *services.OCRService
	smartHome    services.ISmartHomeService
	taskSchedule *services.TaskScheduleService
	note         *services.NoteService
	chat         *services.ChatService
}

type handlerSet struct {
	auth         *handlers.AuthHandler
	home         *handlers.HomeHandler
	room         *handlers.RoomHandler
	task         *handlers.TaskHandler
	taskSchedule *handlers.TaskScheduleHandler
	bill         *handlers.BillHandler
	billCategory *handlers.BillCategoryHandler
	shopping     *handlers.ShoppingHandler
	image        *handlers.ImageHandler
	poll         *handlers.PollHandler
	notification *handlers.NotificationHandler
	audit        *handlers.AuditHandler
	user         *handlers.UserHandler
	ocr          *handlers.OCRHandler
	smartHome    *handlers.SmartHomeHandler
	pushSub      *handlers.PushSubscriptionHandler
	note         *handlers.NoteHandler
	chat         *handlers.ChatHandler
}

func newAppDeps(cfg *config.Config, db *gorm.DB, cache *redis.Client) (*appDeps, error) {
	app := &appDeps{
		cfg:   cfg,
		db:    db,
		cache: cache,
	}

	app.repos = newRepositories(db)

	services, err := newServices(cfg, cache, app.repos)
	if err != nil {
		return nil, err
	}
	app.services = services
	app.handlers = newHandlers(cfg, app.repos, app.services)

	return app, nil
}

func newRepositories(db *gorm.DB) repositories {
	return repositories{
		user:         repository.NewUserRepository(db),
		home:         repository.NewHomeRepository(db),
		room:         repository.NewRoomRepository(db),
		task:         repository.NewTaskRepository(db),
		bill:         repository.NewBillRepository(db),
		billCategory: repository.NewBillCategoryRepository(db),
		shopping:     repository.NewShoppingRepository(db),
		poll:         repository.NewPollRepository(db),
		notification: repository.NewNotificationRepository(db),
		audit:        repository.NewAuditRepository(db),
		smartHome:    repository.NewSmartHomeRepository(db),
		taskSchedule: repository.NewTaskScheduleRepository(db),
		pushSub:      repository.NewPushSubscriptionRepository(db),
		note:         repository.NewNoteRepository(db),
		chat:         repository.NewChatRepository(db),
	}
}

func newServices(cfg *config.Config, cache *redis.Client, repos repositories) (serviceSet, error) {
	mailer := &utils.BrevoMailer{
		APIKey: cfg.BrevoAPIKey,
		From:   cfg.SMTPFrom,
	}

	goth.UseProviders(
		google.New(cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL),
	)

	pushSubSvc := services.NewPushSubscriptionService(repos.pushSub, cfg.VapidPublicKey, cfg.VapidPrivateKey, cfg.VapidSubject)
	notificationSvc := services.NewNotificationService(repos.notification, cache, pushSubSvc, repos.home)
	auditSvc := services.NewAuditService(repos.audit)
	authSvc := services.NewAuthService(repos.user, []byte(cfg.JWTSecret), cache, 30*24*time.Hour, cfg.ClientURL, cfg.ServerURL, mailer)
	homeSvc := services.NewHomeService(repos.home, cache, notificationSvc)
	roomSvc := services.NewRoomService(repos.room, cache)
	taskSvc := services.NewTaskService(repos.task, cache, notificationSvc)
	billSvc := services.NewBillService(repos.bill, repos.billCategory, cache, notificationSvc, homeSvc)
	billCategorySvc := services.NewBillCategoryService(repos.billCategory, cache)
	shoppingSvc := services.NewShoppingService(repos.shopping, cache)
	pollSvc := services.NewPollService(repos.poll, cache, notificationSvc)
	userSvc := services.NewUserService(repos.user, cache)
	noteSvc := services.NewNoteService(
		repos.note,
		repos.user,
		repos.home,
		repos.task,
		repos.bill,
		repos.billCategory,
		repos.shopping,
		cache,
	)

	chatSvc := services.NewChatService(
		repos.chat,
		repos.home,
		repos.task,
		repos.bill,
		repos.billCategory,
		repos.shopping,
		repos.note,
		cache,
		notificationSvc,
	)

	imageSvc, err := services.NewImageService(cfg.R2S3Bucket, cfg.R2Region, cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2PublicUrl)
	if err != nil {
		return serviceSet{}, err
	}

	return serviceSet{
		pushSub:      pushSubSvc,
		notification: notificationSvc,
		audit:        auditSvc,
		auth:         authSvc,
		home:         homeSvc,
		room:         roomSvc,
		task:         taskSvc,
		bill:         billSvc,
		billCategory: billCategorySvc,
		shopping:     shoppingSvc,
		poll:         pollSvc,
		user:         userSvc,
		image:        imageSvc,
		ocr:          services.NewOCRService(cfg.GeminiAPIKey),
		smartHome:    services.NewSmartHomeService(repos.smartHome, cache, cfg.HAEncryptionKey),
		taskSchedule: services.NewTaskScheduleService(repos.taskSchedule, repos.task, cache, notificationSvc),
		note:         noteSvc,
		chat:         chatSvc,
	}, nil
}

func newHandlers(cfg *config.Config, repos repositories, services serviceSet) handlerSet {
	authHandler := handlers.NewAuthHandler(services.auth, cfg.ClientURL, cfg.Mode != "dev")
	authHandler.SetAuditService(services.audit)

	homeHandler := handlers.NewHomeHandler(services.home)
	homeHandler.SetAuditService(services.audit)

	return handlerSet{
		auth:         authHandler,
		home:         homeHandler,
		room:         handlers.NewRoomHandler(services.room, repos.home),
		task:         handlers.NewTaskHandler(services.task, repos.home),
		taskSchedule: handlers.NewTaskScheduleHandler(services.taskSchedule, repos.home),
		bill:         handlers.NewBillHandler(services.bill, repos.home),
		billCategory: handlers.NewBillCategoryHandler(services.billCategory, repos.home),
		shopping:     handlers.NewShoppingHandler(services.shopping, repos.home),
		image:        handlers.NewImageHandler(services.image),
		poll:         handlers.NewPollHandler(services.poll, repos.home),
		notification: handlers.NewNotificationHandler(services.notification),
		audit:        handlers.NewAuditHandler(services.audit),
		user:         handlers.NewUserHandler(services.user, services.image),
		ocr:          handlers.NewOCRHandler(services.ocr),
		smartHome:    handlers.NewSmartHomeHandler(services.smartHome),
		pushSub:      handlers.NewPushSubscriptionHandler(services.pushSub),
		note:         handlers.NewNoteHandler(services.note, repos.home),
		chat:         handlers.NewChatHandler(services.chat, repos.home),
	}
}

func (h handlerSet) RouterHandlers() router.HandlerSet {
	return router.HandlerSet{
		Auth:         h.auth,
		Home:         h.home,
		Task:         h.task,
		TaskSchedule: h.taskSchedule,
		Bill:         h.bill,
		BillCategory: h.billCategory,
		Room:         h.room,
		Shopping:     h.shopping,
		Image:        h.image,
		Poll:         h.poll,
		Notification: h.notification,
		Audit:        h.audit,
		User:         h.user,
		OCR:          h.ocr,
		SmartHome:    h.smartHome,
		PushSub:      h.pushSub,
		Note:         h.note,
		Chat:         h.chat,
	}
}
