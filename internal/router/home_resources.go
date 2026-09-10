package router

import (
	"github.com/Dragodui/diploma-server/internal/http/middleware"
	ratelimiter "github.com/Dragodui/diploma-server/internal/http/rate_limiter"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/go-chi/chi/v5"
)

func mountHomeNotificationRoutes(r chi.Router, deps RoutesDeps) {
	r.Get("/", deps.Handlers.Notification.GetByHomeID)
	r.Delete("/{notification_id}", deps.Handlers.Notification.MarkAsReadForHome)
}

func mountRoomRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireAdmin(homeRepo)).Post("/", deps.Handlers.Room.Create)
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Room.GetByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Get("/{room_id}", deps.Handlers.Room.GetByID)
	r.With(middleware.RequireMember(homeRepo)).Get("/{room_id}/devices", deps.Handlers.SmartHome.GetDevicesByRoom)
	r.With(middleware.RequireMember(homeRepo)).Put("/{room_id}", deps.Handlers.Room.Update)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{room_id}", deps.Handlers.Room.Delete)
}

func mountTaskRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Task.Create)
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Task.GetTasksByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Get("/{task_id}", deps.Handlers.Task.GetByID)
	r.With(middleware.RequireMember(homeRepo)).Put("/{task_id}", deps.Handlers.Task.UpdateTask)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{task_id}", deps.Handlers.Task.DeleteTask)

	r.With(middleware.RequireMember(homeRepo)).Post("/{task_id}/assign", deps.Handlers.Task.AssignUser)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{task_id}/reassign-room", deps.Handlers.Task.ReassignRoom)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{task_id}/mark-completed", deps.Handlers.Task.MarkAssignmentCompleted)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{task_id}/mark-uncompleted", deps.Handlers.Task.MarkAssignmentUncompleted)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{task_id}/complete", deps.Handlers.Task.MarkTaskCompleted)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{task_id}/assignments/{assignment_id}", deps.Handlers.Task.DeleteAssignment)

	r.With(middleware.RequireAdmin(homeRepo)).Post("/schedules", deps.Handlers.TaskSchedule.CreateSchedule)
	r.With(middleware.RequireMember(homeRepo)).Get("/schedules", deps.Handlers.TaskSchedule.GetSchedulesByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Get("/{task_id}/schedule", deps.Handlers.TaskSchedule.GetScheduleByTaskID)
	r.With(middleware.RequireAdmin(homeRepo)).Delete("/schedules/{schedule_id}", deps.Handlers.TaskSchedule.DeleteSchedule)
}

func mountUserAssignmentRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Task.GetAssignmentsForUser)
	r.With(middleware.RequireMember(homeRepo)).Get("/closest", deps.Handlers.Task.GetClosestAssignmentForUser)
}

func mountBillRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Bill.GetByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Get("/private", deps.Handlers.Bill.GetByUserID)
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Bill.Create)
	r.With(middleware.RequireMember(homeRepo)).Get("/{bill_id}", deps.Handlers.Bill.GetByID)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{bill_id}", deps.Handlers.Bill.Delete)
	r.With(middleware.RequireMember(homeRepo)).Put("/{bill_id}", deps.Handlers.Bill.Update)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{bill_id}", deps.Handlers.Bill.MarkPayed)
	r.With(middleware.RequireMember(homeRepo)).Put("/{bill_id}/splits", deps.Handlers.Bill.UpdateSplits)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{bill_id}/splits/{split_id}/paid", deps.Handlers.Bill.MarkSplitPaid)
}

func mountBillCategoryRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.BillCategory.GetAll)
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.BillCategory.Create)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{category_id}", deps.Handlers.BillCategory.Delete)
	r.With(middleware.RequireMember(homeRepo)).Patch("/{category_id}", deps.Handlers.BillCategory.Update)
}

func mountShoppingRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.Route("/categories", func(r chi.Router) {
		r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Shopping.CreateCategory)
		r.With(middleware.RequireMember(homeRepo)).Get("/all", deps.Handlers.Shopping.GetAllCategories)
		r.With(middleware.RequireMember(homeRepo)).Get("/{category_id}", deps.Handlers.Shopping.GetCategoryByID)
		r.With(middleware.RequireMember(homeRepo)).Get("/{category_id}/items", deps.Handlers.Shopping.GetItemsByCategoryID)
		r.With(middleware.RequireMember(homeRepo)).Delete("/{category_id}", deps.Handlers.Shopping.DeleteCategory)
		r.With(middleware.RequireMember(homeRepo)).Put("/{category_id}", deps.Handlers.Shopping.EditCategory)
	})
	r.Route("/items", func(r chi.Router) {
		r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Shopping.CreateItem)
		r.With(middleware.RequireMember(homeRepo)).Get("/{item_id}", deps.Handlers.Shopping.GetItemByID)
		r.With(middleware.RequireMember(homeRepo)).Delete("/{item_id}", deps.Handlers.Shopping.DeleteItem)
		r.With(middleware.RequireMember(homeRepo)).Put("/{item_id}", deps.Handlers.Shopping.EditItem)
		r.With(middleware.RequireMember(homeRepo)).Patch("/{item_id}", deps.Handlers.Shopping.MarkIsBought)
	})
}

func mountPollRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Poll.Create)
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Poll.GetAllByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Get("/{poll_id}", deps.Handlers.Poll.GetByID)
	r.With(middleware.RequireAdmin(homeRepo)).Patch("/{poll_id}/close", deps.Handlers.Poll.Close)
	r.With(middleware.RequireAdmin(homeRepo)).Delete("/{poll_id}", deps.Handlers.Poll.Delete)
	r.With(middleware.RequireMember(homeRepo)).Post("/{poll_id}/vote", deps.Handlers.Poll.Vote)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{poll_id}/vote", deps.Handlers.Poll.Unvote)
}

func mountSmartHomeRoutes(r chi.Router, deps RoutesDeps, rateLimiter *ratelimiter.IPRateLimiter, homeRepo repository.HomeRepository) {
	r.With(middleware.RequireAdmin(homeRepo)).Post("/connect", deps.Handlers.SmartHome.Connect)
	r.With(middleware.RequireAdmin(homeRepo)).Delete("/disconnect", deps.Handlers.SmartHome.Disconnect)
	r.With(middleware.RequireMember(homeRepo)).Get("/status", deps.Handlers.SmartHome.Status)
	r.With(middleware.RequireMember(homeRepo)).Get("/discover", deps.Handlers.SmartHome.Discover)
	r.With(middleware.RequireMember(homeRepo)).Get("/states", deps.Handlers.SmartHome.GetAllStates)

	deviceControlLimit := middleware.StrictRateLimitMiddleware(rateLimiter, 120, 0.5)
	r.Route("/devices", func(r chi.Router) {
		r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.SmartHome.AddDevice)
		r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.SmartHome.GetDevices)
		r.With(middleware.RequireMember(homeRepo)).Get("/{device_id}", deps.Handlers.SmartHome.GetDevice)
		r.With(middleware.RequireAdmin(homeRepo)).Put("/{device_id}", deps.Handlers.SmartHome.UpdateDevice)
		r.With(middleware.RequireAdmin(homeRepo)).Delete("/{device_id}", deps.Handlers.SmartHome.DeleteDevice)
		r.With(middleware.RequireMember(homeRepo), deviceControlLimit).Post("/{device_id}/control", deps.Handlers.SmartHome.ControlDevice)
	})
}

func mountNoteRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Note.CreateNote)
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Note.GetNotesByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Get("/{note_id}", deps.Handlers.Note.GetNoteByID)
	r.With(middleware.RequireMember(homeRepo)).Put("/{note_id}", deps.Handlers.Note.UpdateNote)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{note_id}", deps.Handlers.Note.DeleteNote)
}

func mountChatRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Chat.GetMessages)
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Chat.SendMessage)
	r.With(middleware.RequireMember(homeRepo)).Get("/unread", deps.Handlers.Chat.GetUnreadCount)
	r.With(middleware.RequireMember(homeRepo)).Get("/conversations", deps.Handlers.Chat.GetConversations)
	r.With(middleware.RequireMember(homeRepo)).Post("/read", deps.Handlers.Chat.MarkRead)
	r.With(middleware.RequireMember(homeRepo)).Put("/{message_id}", deps.Handlers.Chat.UpdateMessage)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{message_id}", deps.Handlers.Chat.DeleteMessage)
}

func mountNoteCategoryRoutes(r chi.Router, deps RoutesDeps) {
	homeRepo := deps.HomeRepo
	r.With(middleware.RequireMember(homeRepo)).Get("/", deps.Handlers.Note.GetCategoriesByHomeID)
	r.With(middleware.RequireMember(homeRepo)).Post("/", deps.Handlers.Note.CreateCategory)
	r.With(middleware.RequireMember(homeRepo)).Put("/{category_id}", deps.Handlers.Note.UpdateCategory)
	r.With(middleware.RequireMember(homeRepo)).Delete("/{category_id}", deps.Handlers.Note.DeleteCategory)
}

