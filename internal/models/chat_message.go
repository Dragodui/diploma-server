package models

import "time"

// ChatMessage is a single message in a home's shared chat. Mentions follow the
// same many2many pattern as notes, so the client can render and link them the
// same way it already does for note content.
type ChatMessage struct {
	ID        int    `gorm:"autoIncrement; primaryKey" json:"id"`
	HomeID    int    `gorm:"not null;index" json:"home_id"`
	CreatedBy int    `gorm:"not null" json:"created_by"`
	Content   string `gorm:"not null" json:"content"`
	// RecipientID is nil for the shared home chat, and set to the other
	// member's id for a one-to-one conversation inside the home.
	RecipientID *int       `gorm:"index" json:"recipient_id"`
	ImageURL    *string    `json:"image_url"`
	EditedAt    *time.Time `json:"edited_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime;index" json:"created_at"`

	// relations
	Home      *Home `gorm:"foreignKey:HomeID;constraint:OnDelete:CASCADE" json:"home,omitempty"`
	Creator   *User `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE" json:"creator,omitempty"`
	Recipient *User `gorm:"foreignKey:RecipientID;constraint:OnDelete:CASCADE" json:"recipient,omitempty"`

	// MentionsAll is set when the message used @all to ping every member.
	MentionsAll bool `gorm:"default:false" json:"mentions_all"`

	// Mentions relations
	MentionedUsers              []User             `gorm:"many2many:chat_message_user_mentions;constraint:OnDelete:CASCADE" json:"mentioned_users,omitempty"`
	MentionedTasks              []Task             `gorm:"many2many:chat_message_task_mentions;constraint:OnDelete:CASCADE" json:"mentioned_tasks,omitempty"`
	MentionedBills              []Bill             `gorm:"many2many:chat_message_bill_mentions;constraint:OnDelete:CASCADE" json:"mentioned_bills,omitempty"`
	MentionedShoppingItems      []ShoppingItem     `gorm:"many2many:chat_message_shopping_item_mentions;constraint:OnDelete:CASCADE" json:"mentioned_shopping_items,omitempty"`
	MentionedNoteCategories     []NoteCategory     `gorm:"many2many:chat_message_note_category_mentions;constraint:OnDelete:CASCADE" json:"mentioned_note_categories,omitempty"`
	MentionedBillCategories     []BillCategory     `gorm:"many2many:chat_message_bill_category_mentions;constraint:OnDelete:CASCADE" json:"mentioned_bill_categories,omitempty"`
	MentionedShoppingCategories []ShoppingCategory `gorm:"many2many:chat_message_shopping_category_mentions;constraint:OnDelete:CASCADE" json:"mentioned_shopping_categories,omitempty"`

	// Reads carries the per-user read receipts for this message.
	Reads []ChatMessageRead `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE" json:"reads,omitempty"`
}

// ChatMessageRead records that a member read a specific message, and when.
// One row per (message, user) - a household has few members, so per-message
// receipts stay cheap while giving exact "read at" times.
type ChatMessageRead struct {
	ID        int       `gorm:"autoIncrement; primaryKey" json:"id"`
	MessageID int       `gorm:"not null;uniqueIndex:idx_chat_read_message_user" json:"message_id"`
	UserID    int       `gorm:"not null;uniqueIndex:idx_chat_read_message_user" json:"user_id"`
	ReadAt    time.Time `gorm:"autoCreateTime" json:"read_at"`

	// relations
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

type CreateChatMessageRequest struct {
	Content string `json:"content"`
	// RecipientID targets a direct conversation; omit it to post to the
	// shared home chat.
	RecipientID                  *int    `json:"recipient_id"`
	ImageURL                     *string `json:"image_url"`
	MentionsAll                  bool    `json:"mentions_all"`
	MentionedUserIDs             []int   `json:"mentioned_user_ids"`
	MentionedTaskIDs             []int   `json:"mentioned_task_ids"`
	MentionedBillIDs             []int   `json:"mentioned_bill_ids"`
	MentionedShoppingItemIDs     []int   `json:"mentioned_shopping_item_ids"`
	MentionedNoteCategoryIDs     []int   `json:"mentioned_note_category_ids"`
	MentionedBillCategoryIDs     []int   `json:"mentioned_bill_category_ids"`
	MentionedShoppingCategoryIDs []int   `json:"mentioned_shopping_category_ids"`
}

type UpdateChatMessageRequest struct {
	Content                      *string `json:"content" validate:"omitempty,min=1"`
	ImageURL                     *string `json:"image_url"`
	MentionsAll                  *bool   `json:"mentions_all"`
	MentionedUserIDs             *[]int  `json:"mentioned_user_ids"`
	MentionedTaskIDs             *[]int  `json:"mentioned_task_ids"`
	MentionedBillIDs             *[]int  `json:"mentioned_bill_ids"`
	MentionedShoppingItemIDs     *[]int  `json:"mentioned_shopping_item_ids"`
	MentionedNoteCategoryIDs     *[]int  `json:"mentioned_note_category_ids"`
	MentionedBillCategoryIDs     *[]int  `json:"mentioned_bill_category_ids"`
	MentionedShoppingCategoryIDs *[]int  `json:"mentioned_shopping_category_ids"`
}

// MarkChatReadRequest marks every message in the home up to and including
// LastMessageID as read by the caller.
type MarkChatReadRequest struct {
	LastMessageID int `json:"last_message_id" validate:"required"`
	// RecipientID scopes the read receipt to one direct conversation; omit
	// it to mark the shared home chat as read.
	RecipientID *int `json:"recipient_id"`
}

// ChatConversationSummary is one row of the conversation list: the shared home
// chat (PeerID nil) or a direct chat with one member.
type ChatConversationSummary struct {
	PeerID      *int         `json:"peer_id"`
	Peer        *User        `json:"peer,omitempty"`
	LastMessage *ChatMessage `json:"last_message,omitempty"`
	UnreadCount int64        `json:"unread_count"`
}
