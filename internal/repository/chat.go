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
	// FindByHomeID returns messages oldest-last (newest first), paging backwards
	// from beforeID when it is non-nil so the client can scroll into history.
	FindByHomeID(ctx context.Context, homeID int, limit int, beforeID *int) ([]models.ChatMessage, error)
	Update(ctx context.Context, message *models.ChatMessage) error
	Delete(ctx context.Context, id int) error

	// read receipts
	MarkReadUpTo(ctx context.Context, homeID, userID, lastMessageID int, readAt time.Time) error
	CountUnread(ctx context.Context, homeID, userID int) (int64, error)
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

func (r *chatRepo) FindByHomeID(ctx context.Context, homeID int, limit int, beforeID *int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	query := r.db.WithContext(ctx).Where("home_id = ?", homeID)
	if beforeID != nil {
		query = query.Where("id < ?", *beforeID)
	}

	err := preloadAll(query).
		Order("id DESC").
		Limit(limit).
		Find(&messages).Error

	return messages, err
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
func (r *chatRepo) MarkReadUpTo(ctx context.Context, homeID, userID, lastMessageID int, readAt time.Time) error {
	var messageIDs []int
	if err := r.db.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("home_id = ? AND id <= ? AND created_by <> ?", homeID, lastMessageID, userID).
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

func (r *chatRepo) CountUnread(ctx context.Context, homeID, userID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ChatMessage{}).
		Where("chat_messages.home_id = ? AND chat_messages.created_by <> ?", homeID, userID).
		Where("NOT EXISTS (SELECT 1 FROM chat_message_reads WHERE chat_message_reads.message_id = chat_messages.id AND chat_message_reads.user_id = ?)", userID).
		Count(&count).Error
	return count, err
}
