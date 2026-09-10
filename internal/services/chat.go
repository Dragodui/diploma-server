package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Dragodui/diploma-server/internal/event"
	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/repository"
	"github.com/redis/go-redis/v9"
)

// defaultChatPageSize is how many messages one history page holds.
const defaultChatPageSize = 50

type ChatService struct {
	repo         repository.ChatRepository
	homeRepo     repository.HomeRepository
	taskRepo     repository.TaskRepository
	billRepo     repository.BillRepository
	billCatRepo  repository.IBillCategoryRepository
	shoppingRepo repository.ShoppingRepository
	noteRepo     repository.NoteRepository
	cache        *redis.Client
	notifSvc     INotificationService
}

type IChatService interface {
	SendMessage(ctx context.Context, homeID, createdBy int, req models.CreateChatMessageRequest) (*models.ChatMessage, error)
	GetMessages(ctx context.Context, homeID, limit int, beforeID *int) ([]models.ChatMessage, error)
	UpdateMessage(ctx context.Context, id, homeID, userID int, req models.UpdateChatMessageRequest) (*models.ChatMessage, error)
	DeleteMessage(ctx context.Context, id, homeID, userID int) error
	MarkRead(ctx context.Context, homeID, userID, lastMessageID int) error
	GetUnreadCount(ctx context.Context, homeID, userID int) (int64, error)
}

func NewChatService(
	repo repository.ChatRepository,
	homeRepo repository.HomeRepository,
	taskRepo repository.TaskRepository,
	billRepo repository.BillRepository,
	billCatRepo repository.IBillCategoryRepository,
	shoppingRepo repository.ShoppingRepository,
	noteRepo repository.NoteRepository,
	cache *redis.Client,
	notifSvc INotificationService,
) *ChatService {
	return &ChatService{
		repo:         repo,
		homeRepo:     homeRepo,
		taskRepo:     taskRepo,
		billRepo:     billRepo,
		billCatRepo:  billCatRepo,
		shoppingRepo: shoppingRepo,
		noteRepo:     noteRepo,
		cache:        cache,
		notifSvc:     notifSvc,
	}
}

// validateMentions checks every mentioned entity actually belongs to this home,
// so a message can't link to another household's data.
func (s *ChatService) validateMentions(
	ctx context.Context,
	homeID int,
	userIDs, taskIDs, billIDs, itemIDs, noteCatIDs, billCatIDs, shoppingCatIDs []int,
) error {
	if len(userIDs) > 0 {
		members, err := s.homeRepo.GetMembers(ctx, homeID)
		if err != nil {
			return err
		}
		memberMap := make(map[int]bool, len(members))
		for _, m := range members {
			memberMap[m.UserID] = true
		}
		for _, uid := range userIDs {
			if !memberMap[uid] {
				return fmt.Errorf("user %d is not a member of this home", uid)
			}
		}
	}

	for _, tid := range taskIDs {
		t, err := s.taskRepo.FindByID(ctx, tid)
		if err != nil {
			return err
		}
		if t == nil || t.HomeID != homeID {
			return fmt.Errorf("task %d does not exist in this home", tid)
		}
	}

	for _, bid := range billIDs {
		b, err := s.billRepo.FindByID(ctx, bid)
		if err != nil {
			return err
		}
		if b == nil || b.HomeID != homeID {
			return fmt.Errorf("bill %d does not exist in this home", bid)
		}
	}

	for _, iid := range itemIDs {
		item, err := s.shoppingRepo.FindItemByID(ctx, iid)
		if err != nil {
			return err
		}
		if item == nil {
			return fmt.Errorf("shopping item %d does not exist", iid)
		}
		cat, err := s.shoppingRepo.FindCategoryByID(ctx, item.CategoryID)
		if err != nil {
			return err
		}
		if cat == nil || cat.HomeID != homeID {
			return fmt.Errorf("shopping item %d does not exist in this home", iid)
		}
	}

	for _, cid := range noteCatIDs {
		cat, err := s.noteRepo.FindCategoryByID(ctx, cid)
		if err != nil {
			return err
		}
		if cat == nil || cat.HomeID != homeID {
			return fmt.Errorf("note category %d does not exist in this home", cid)
		}
	}

	for _, cid := range billCatIDs {
		cat, err := s.billCatRepo.GetByID(ctx, cid)
		if err != nil {
			return err
		}
		if cat == nil || cat.HomeID != homeID {
			return fmt.Errorf("bill category %d does not exist in this home", cid)
		}
	}

	for _, cid := range shoppingCatIDs {
		cat, err := s.shoppingRepo.FindCategoryByID(ctx, cid)
		if err != nil {
			return err
		}
		if cat == nil || cat.HomeID != homeID {
			return fmt.Errorf("shopping category %d does not exist in this home", cid)
		}
	}

	return nil
}

// attachMentions fills the association slices with ID-only placeholders, which
// is all GORM needs to write the many2many join rows.
func attachMentions(
	message *models.ChatMessage,
	userIDs, taskIDs, billIDs, itemIDs, noteCatIDs, billCatIDs, shoppingCatIDs []int,
) {
	message.MentionedUsers = nil
	message.MentionedTasks = nil
	message.MentionedBills = nil
	message.MentionedShoppingItems = nil
	message.MentionedNoteCategories = nil
	message.MentionedBillCategories = nil
	message.MentionedShoppingCategories = nil

	for _, id := range userIDs {
		message.MentionedUsers = append(message.MentionedUsers, models.User{ID: id})
	}
	for _, id := range taskIDs {
		message.MentionedTasks = append(message.MentionedTasks, models.Task{ID: id})
	}
	for _, id := range billIDs {
		message.MentionedBills = append(message.MentionedBills, models.Bill{ID: id})
	}
	for _, id := range itemIDs {
		message.MentionedShoppingItems = append(message.MentionedShoppingItems, models.ShoppingItem{ID: id})
	}
	for _, id := range noteCatIDs {
		message.MentionedNoteCategories = append(message.MentionedNoteCategories, models.NoteCategory{ID: id})
	}
	for _, id := range billCatIDs {
		message.MentionedBillCategories = append(message.MentionedBillCategories, models.BillCategory{ID: id})
	}
	for _, id := range shoppingCatIDs {
		message.MentionedShoppingCategories = append(message.MentionedShoppingCategories, models.ShoppingCategory{ID: id})
	}
}

// notifyMentioned pings the people a message called out: everyone in the home
// for @all, or just the named users otherwise. The author is never notified
// about their own message.
func (s *ChatService) notifyMentioned(ctx context.Context, message *models.ChatMessage, authorName string) {
	recipients := make(map[int]bool)

	if message.MentionsAll {
		members, err := s.homeRepo.GetMembers(ctx, message.HomeID)
		if err == nil {
			for _, m := range members {
				recipients[m.UserID] = true
			}
		}
	}
	for _, u := range message.MentionedUsers {
		recipients[u.ID] = true
	}
	delete(recipients, message.CreatedBy)

	if len(recipients) == 0 {
		return
	}

	preview := message.Content
	if strings.TrimSpace(preview) == "" && message.ImageURL != nil {
		preview = "[image]"
	}
	if len(preview) > 80 {
		preview = preview[:80] + "..."
	}
	description := fmt.Sprintf("%s mentioned you in the home chat: %s", authorName, preview)

	from := message.CreatedBy
	for userID := range recipients {
		_ = s.notifSvc.Create(ctx, &from, userID, &message.HomeID, description)
	}
}

func (s *ChatService) SendMessage(ctx context.Context, homeID, createdBy int, req models.CreateChatMessageRequest) (*models.ChatMessage, error) {
	// A message needs to carry something - text, an image, or both.
	if strings.TrimSpace(req.Content) == "" && (req.ImageURL == nil || *req.ImageURL == "") {
		return nil, errors.New("message must have text or an image")
	}

	if err := s.validateMentions(
		ctx, homeID,
		req.MentionedUserIDs, req.MentionedTaskIDs, req.MentionedBillIDs, req.MentionedShoppingItemIDs,
		req.MentionedNoteCategoryIDs, req.MentionedBillCategoryIDs, req.MentionedShoppingCategoryIDs,
	); err != nil {
		return nil, err
	}

	message := &models.ChatMessage{
		HomeID:      homeID,
		CreatedBy:   createdBy,
		Content:     req.Content,
		ImageURL:    req.ImageURL,
		MentionsAll: req.MentionsAll,
		CreatedAt:   time.Now(),
	}
	attachMentions(
		message,
		req.MentionedUserIDs, req.MentionedTaskIDs, req.MentionedBillIDs, req.MentionedShoppingItemIDs,
		req.MentionedNoteCategoryIDs, req.MentionedBillCategoryIDs, req.MentionedShoppingCategoryIDs,
	)

	if err := s.repo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Re-read so the response carries the creator and resolved mention entities.
	saved, err := s.repo.FindByID(ctx, message.ID)
	if err != nil {
		return nil, err
	}
	if saved == nil {
		return nil, errors.New("message not found after creation")
	}

	authorName := ""
	if saved.Creator != nil {
		authorName = saved.Creator.Name
	}
	s.notifyMentioned(ctx, saved, authorName)

	event.SendHomeEvent(ctx, s.cache, homeID, &event.RealTimeEvent{
		Module: event.ModuleChat,
		Action: event.ActionCreated,
		Data:   saved,
	})

	return saved, nil
}

func (s *ChatService) GetMessages(ctx context.Context, homeID, limit int, beforeID *int) ([]models.ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = defaultChatPageSize
	}
	return s.repo.FindByHomeID(ctx, homeID, limit, beforeID)
}

func (s *ChatService) UpdateMessage(ctx context.Context, id, homeID, userID int, req models.UpdateChatMessageRequest) (*models.ChatMessage, error) {
	message, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if message == nil || message.HomeID != homeID {
		return nil, errors.New("message not found")
	}
	// Editing is author-only; moderation is limited to deletion.
	if message.CreatedBy != userID {
		return nil, errors.New("you can only edit your own messages")
	}

	if err := s.validateMentions(
		ctx, homeID,
		derefIDs(req.MentionedUserIDs), derefIDs(req.MentionedTaskIDs), derefIDs(req.MentionedBillIDs),
		derefIDs(req.MentionedShoppingItemIDs), derefIDs(req.MentionedNoteCategoryIDs),
		derefIDs(req.MentionedBillCategoryIDs), derefIDs(req.MentionedShoppingCategoryIDs),
	); err != nil {
		return nil, err
	}

	if req.Content != nil {
		message.Content = *req.Content
	}
	if req.ImageURL != nil {
		// An empty string clears the attachment.
		if *req.ImageURL == "" {
			message.ImageURL = nil
		} else {
			message.ImageURL = req.ImageURL
		}
	}
	if req.MentionsAll != nil {
		message.MentionsAll = *req.MentionsAll
	}
	editedAt := time.Now()
	message.EditedAt = &editedAt

	attachMentions(
		message,
		mentionIDs(req.MentionedUserIDs, message.MentionedUsers, func(u models.User) int { return u.ID }),
		mentionIDs(req.MentionedTaskIDs, message.MentionedTasks, func(t models.Task) int { return t.ID }),
		mentionIDs(req.MentionedBillIDs, message.MentionedBills, func(b models.Bill) int { return b.ID }),
		mentionIDs(req.MentionedShoppingItemIDs, message.MentionedShoppingItems, func(i models.ShoppingItem) int { return i.ID }),
		mentionIDs(req.MentionedNoteCategoryIDs, message.MentionedNoteCategories, func(c models.NoteCategory) int { return c.ID }),
		mentionIDs(req.MentionedBillCategoryIDs, message.MentionedBillCategories, func(c models.BillCategory) int { return c.ID }),
		mentionIDs(req.MentionedShoppingCategoryIDs, message.MentionedShoppingCategories, func(c models.ShoppingCategory) int { return c.ID }),
	)

	if err := s.repo.Update(ctx, message); err != nil {
		return nil, err
	}

	saved, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	event.SendHomeEvent(ctx, s.cache, homeID, &event.RealTimeEvent{
		Module: event.ModuleChat,
		Action: event.ActionUpdated,
		Data:   saved,
	})

	return saved, nil
}

func (s *ChatService) DeleteMessage(ctx context.Context, id, homeID, userID int) error {
	message, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if message == nil || message.HomeID != homeID {
		return errors.New("message not found")
	}
	if message.CreatedBy != userID {
		return errors.New("you can only delete your own messages")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	event.SendHomeEvent(ctx, s.cache, homeID, &event.RealTimeEvent{
		Module: event.ModuleChat,
		Action: event.ActionDeleted,
		Data:   map[string]int{"message_id": id},
	})

	return nil
}

func (s *ChatService) MarkRead(ctx context.Context, homeID, userID, lastMessageID int) error {
	if err := s.repo.MarkReadUpTo(ctx, homeID, userID, lastMessageID, time.Now()); err != nil {
		return err
	}

	// Tell everyone else so their "read by" rows update live.
	event.SendHomeEvent(ctx, s.cache, homeID, &event.RealTimeEvent{
		Module: event.ModuleChat,
		Action: event.ActionMarkRead,
		Data:   map[string]int{"user_id": userID, "last_message_id": lastMessageID},
	})

	return nil
}

func (s *ChatService) GetUnreadCount(ctx context.Context, homeID, userID int) (int64, error) {
	return s.repo.CountUnread(ctx, homeID, userID)
}

// derefIDs returns the slice a pointer field holds, or nil when it was omitted.
func derefIDs(ids *[]int) []int {
	if ids == nil {
		return nil
	}
	return *ids
}

// mentionIDs picks the requested IDs when the field was sent, and otherwise
// keeps whatever the message already had, so a partial edit doesn't wipe
// mentions the client didn't touch.
func mentionIDs[T any](requested *[]int, existing []T, id func(T) int) []int {
	if requested != nil {
		return *requested
	}
	current := make([]int, 0, len(existing))
	for _, e := range existing {
		current = append(current, id(e))
	}
	return current
}
