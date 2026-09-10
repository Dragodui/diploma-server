package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ChatRepository interface {
	Create(ctx context.Context, message *models.ChatMessage) error
	FindByID(ctx context.Context, id int) (*models.ChatMessage, error)
	// FindConversation returns one conversation's messages newest-first, paging
	// backwards from beforeID when it is non-nil. peerID nil means the shared
	// home chat; otherwise it's the direct chat between userID and peerID.
	FindConversation(ctx context.Context, homeID, userID int, peerID *int, limit int, beforeID *int) ([]models.ChatMessage, error)
	// FindLastMessage returns the newest message of one conversation, or nil.
	FindLastMessage(ctx context.Context, homeID, userID int, peerID *int) (*models.ChatMessage, error)
	Update(ctx context.Context, message *models.ChatMessage) error
	Delete(ctx context.Context, id int) error

	// read receipts
	MarkReadUpTo(ctx context.Context, homeID, userID, lastMessageID int, peerID *int, readAt time.Time) error
	CountUnread(ctx context.Context, homeID, userID int, peerID *int) (int64, error)
}

type chatRepo struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepo{db: db}
}

// preloadAll attaches the creator, every mention list and the read receipts,
// which is what the client needs to render a message in full.
func preloadAll(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Creator").
		Preload("MentionedUsers").
		Preload("MentionedTasks").
		Preload("MentionedBills").
		Preload("MentionedShoppingItems").
		Preload("MentionedNoteCategories").
		Preload("MentionedBillCategories").
		Preload("MentionedShoppingCategories").
		Preload("Reads").
		Preload("Reads.User")
}

func (r *chatRepo) Create(ctx context.Context, message *models.ChatMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *chatRepo) FindByID(ctx context.Context, id int) (*models.ChatMessage, error) {
	var message models.ChatMessage
	err := preloadAll(r.db.WithContext(ctx)).First(&message, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

// scopeConversation narrows a query to a single conversation: the shared home
// chat (peerID nil) or the two-way direct thread between userID and peerID.
func scopeConversation(query *gorm.DB, homeID, userID int, peerID *int) *gorm.DB {
	query = query.Where("home_id = ?", homeID)
	if peerID == nil {
		return query.Where("recipient_id IS NULL")
	}
	return query.Where(
		"(created_by = ? AND recipient_id = ?) OR (created_by = ? AND recipient_id = ?)",
		userID, *peerID, *peerID, userID,
	)
}

func (r *chatRepo) FindConversation(ctx context.Context, homeID, userID int, peerID *int, limit int, beforeID *int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	query := scopeConversation(r.db.WithContext(ctx), homeID, userID, peerID)
	if beforeID != nil {
		query = query.Where("id < ?", *beforeID)
	}

	err := preloadAll(query).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
}

func (r *chatRepo) FindLastMessage(ctx context.Context, homeID, userID int, peerID *int) (*models.ChatMessage, error) {
	var message models.ChatMessage
	err := scopeConversation(r.db.WithContext(ctx), homeID, userID, peerID).
		Preload("Creator").
		Order("id DESC").
		First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &message, nil
}

func (r *chatRepo) Update(ctx context.Context, message *models.ChatMessage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Creator", "Home", "Reads").Save(message).Error; err != nil {
			return err
		}
		// Mentions are replaced wholesale, same as notes do on edit.
		if err := tx.Model(message).Association("MentionedUsers").Replace(message.MentionedUsers); err != nil {
			return err
		}
		if err := tx.Model(message).Association("MentionedTasks").Replace(message.MentionedTasks); err != nil {
			return err
		}
		if err := tx.Model(message).Association("MentionedBills").Replace(message.MentionedBills); err != nil {
			return err
		}
		if err := tx.Model(message).Association("MentionedShoppingItems").Replace(message.MentionedShoppingItems); err != nil {
			return err
		}
		if err := tx.Model(message).Association("MentionedNoteCategories").Replace(message.MentionedNoteCategories); err != nil {
			return err
		}
		if err := tx.Model(message).Association("MentionedBillCategories").Replace(message.MentionedBillCategories); err != nil {
			return err
		}
		if err := tx.Model(message).Association("MentionedShoppingCategories").Replace(message.MentionedShoppingCategories); err != nil {
			return err
		}
		return nil
	})
}

func (r *chatRepo) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.ChatMessage{}, id).Error
}

// MarkReadUpTo inserts a read receipt for every message in the home at or
// below lastMessageID that the user hasn't already read. Own messages are
// skipped - a receipt on your own message carries no information.
func (r *chatRepo) MarkReadUpTo(ctx context.Context, homeID, userID, lastMessageID int, peerID *int, readAt time.Time) error {
	var messageIDs []int
	query := scopeConversation(r.db.WithContext(ctx).Model(&models.ChatMessage{}), homeID, userID, peerID)
	if err := query.
		Where("id <= ? AND created_by <> ?", lastMessageID, userID).
		Pluck("id", &messageIDs).Error; err != nil {
		return err
	}
	if len(messageIDs) == 0 {
		return nil
	}

	reads := make([]models.ChatMessageRead, 0, len(messageIDs))
	for _, id := range messageIDs {
		reads = append(reads, models.ChatMessageRead{MessageID: id, UserID: userID, ReadAt: readAt})
	}

	// Already-read messages keep their original read_at instead of being bumped.
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).
		Create(&reads).Error
}

func (r *chatRepo) CountUnread(ctx context.Context, homeID, userID int, peerID *int) (int64, error) {
	var count int64
	err := scopeConversation(r.db.WithContext(ctx).Model(&models.ChatMessage{}), homeID, userID, peerID).
		Where("created_by <> ?", userID).
		Where("NOT EXISTS (SELECT 1 FROM chat_message_reads WHERE chat_message_reads.message_id = chat_messages.id AND chat_message_reads.user_id = ?)", userID).
		Count(&count).Error
	return count, err
}
