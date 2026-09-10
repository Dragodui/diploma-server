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
	GetMessages(ctx context.Context, homeID, userID int, peerID *int, limit int, beforeID *int) ([]models.ChatMessage, error)
	UpdateMessage(ctx context.Context, id, homeID, userID int, req models.UpdateChatMessageRequest) (*models.ChatMessage, error)
	DeleteMessage(ctx context.Context, id, homeID, userID int) error
	MarkRead(ctx context.Context, homeID, userID, lastMessageID int, peerID *int) error
	GetUnreadCount(ctx context.Context, homeID, userID int, peerID *int) (int64, error)
	GetConversations(ctx context.Context, homeID, userID int) ([]models.ChatConversationSummary, error)
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

// notifyMentioned pings the people a message called out: in the home chat that
// is everyone for @all or just the named users, while a direct message always
// notifies its recipient. The author is never notified about their own message.
func (s *ChatService) notifyMentioned(ctx context.Context, message *models.ChatMessage, authorName string) {
	recipients := make(map[int]bool)
	isDirect := message.RecipientID != nil

	if isDirect {
		recipients[*message.RecipientID] = true
	} else if message.MentionsAll {
		members, err := s.homeRepo.GetMembers(ctx, message.HomeID)
		if err == nil {
			for _, m := range members {
				recipients[m.UserID] = true
			}
		}
	}
	for _, u := range message.MentionedUsers {
		// Mentions in a direct chat can't drag in anyone outside it.
		if isDirect && u.ID != *message.RecipientID {
			continue
		}
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
	if isDirect {
		description = fmt.Sprintf("%s: %s", authorName, preview)
	}

	from := message.CreatedBy
	for userID := range recipients {
		_ = s.notifSvc.Create(ctx, &from, userID, &message.HomeID, description)
	}
}

// publishChatEvent routes a chat event to the whole home for the shared chat,
// or only to the two participants for a direct message - otherwise a private
// message would be pushed over websockets to every member of the home.
func (s *ChatService) publishChatEvent(ctx context.Context, message *models.ChatMessage, action event.Action, data any) {
	evt := &event.RealTimeEvent{Module: event.ModuleChat, Action: action, Data: data}

	if message.RecipientID == nil {
		event.SendHomeEvent(ctx, s.cache, message.HomeID, evt)
		return
	}
	event.SendUserEvent(ctx, s.cache, message.CreatedBy, evt)
	event.SendUserEvent(ctx, s.cache, *message.RecipientID, evt)
}

func (s *ChatService) SendMessage(ctx context.Context, homeID, createdBy int, req models.CreateChatMessageRequest) (*models.ChatMessage, error) {
	// A message needs to carry something - text, an image, or both.
	if strings.TrimSpace(req.Content) == "" && (req.ImageURL == nil || *req.ImageURL == "") {
		return nil, errors.New("message must have text or an image")
	}

	// A direct message must go to another member of the same home.
	if req.RecipientID != nil {
		if *req.RecipientID == createdBy {
			return nil, errors.New("cannot send a direct message to yourself")
		}
		if err := s.assertMember(ctx, homeID, *req.RecipientID); err != nil {
			return nil, err
		}
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
		RecipientID: req.RecipientID,
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

	s.publishChatEvent(ctx, saved, event.ActionCreated, saved)

	return saved, nil
}

func (s *ChatService) GetMessages(ctx context.Context, homeID, userID int, peerID *int, limit int, beforeID *int) ([]models.ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = defaultChatPageSize
	}
	if peerID != nil {
		if err := s.assertMember(ctx, homeID, *peerID); err != nil {
			return nil, err
		}
	}
	return s.repo.FindConversation(ctx, homeID, userID, peerID, limit, beforeID)
}

// assertMember rejects peers who don't belong to the home, so a direct chat
// can't be opened with someone outside it.
func (s *ChatService) assertMember(ctx context.Context, homeID, userID int) error {
	members, err := s.homeRepo.GetMembers(ctx, homeID)
	if err != nil {
		return err
	}
	for _, m := range members {
		if m.UserID == userID {
			return nil
		}
	}
	return fmt.Errorf("user %d is not a member of this home", userID)
}

// GetConversations lists the shared home chat plus a direct chat with every
// other member, each with its last message and unread count.
func (s *ChatService) GetConversations(ctx context.Context, homeID, userID int) ([]models.ChatConversationSummary, error) {
	members, err := s.homeRepo.GetMembers(ctx, homeID)
	if err != nil {
		return nil, err
	}

	summaries := make([]models.ChatConversationSummary, 0, len(members))

	homeLast, err := s.repo.FindLastMessage(ctx, homeID, userID, nil)
	if err != nil {
		return nil, err
	}
	homeUnread, err := s.repo.CountUnread(ctx, homeID, userID, nil)
	if err != nil {
		return nil, err
	}
	summaries = append(summaries, models.ChatConversationSummary{
		PeerID:      nil,
		LastMessage: homeLast,
		UnreadCount: homeUnread,
	})

	for i := range members {
		member := members[i]
		if member.UserID == userID {
			continue
		}
		peerID := member.UserID

		last, err := s.repo.FindLastMessage(ctx, homeID, userID, &peerID)
		if err != nil {
			return nil, err
		}
		unread, err := s.repo.CountUnread(ctx, homeID, userID, &peerID)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, models.ChatConversationSummary{
			PeerID:      &peerID,
			Peer:        member.User,
			LastMessage: last,
			UnreadCount: unread,
		})
	}

	return summaries, nil
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

	if saved != nil {
		s.publishChatEvent(ctx, saved, event.ActionUpdated, saved)
	}

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

	s.publishChatEvent(ctx, message, event.ActionDeleted, map[string]int{"message_id": id})

	return nil
}

func (s *ChatService) MarkRead(ctx context.Context, homeID, userID, lastMessageID int, peerID *int) error {
	if err := s.repo.MarkReadUpTo(ctx, homeID, userID, lastMessageID, peerID, time.Now()); err != nil {
		return err
	}

	// Tell the other side so their "read by" rows update live.
	readEvent := &event.RealTimeEvent{
		Module: event.ModuleChat,
		Action: event.ActionMarkRead,
		Data:   map[string]any{"user_id": userID, "last_message_id": lastMessageID, "peer_id": peerID},
	}
	if peerID == nil {
		event.SendHomeEvent(ctx, s.cache, homeID, readEvent)
	} else {
		event.SendUserEvent(ctx, s.cache, userID, readEvent)
		event.SendUserEvent(ctx, s.cache, *peerID, readEvent)
	}

	return nil
}

func (s *ChatService) GetUnreadCount(ctx context.Context, homeID, userID int, peerID *int) (int64, error) {
	return s.repo.CountUnread(ctx, homeID, userID, peerID)
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
