package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Module string

const (
	ModuleBillCategory     Module = "BILL_CATEGORY"
	ModuleBill             Module = "BILL"
	ModuleHome             Module = "HOME"
	ModuleNotification     Module = "NOTIFICATION"
	ModuleHomeNotification Module = "HOME_NOTIFICATION"
	ModulePoll             Module = "POLL"
	ModuleRoom             Module = "ROOM"
	ModuleShoppingCategory Module = "SHOPPING_CATEGORY"
	ModuleShoppingItem     Module = "SHOPPING_ITEM"
	ModuleTask             Module = "TASK"
	ModuleUser             Module = "USER"
	ModuleNote             Module = "NOTE"
	ModuleNoteCategory     Module = "NOTE_CATEGORY"
	ModuleChat             Module = "CHAT"
)

type Action string

const (
	ActionCreated       Action = "CREATED"
	ActionUpdated       Action = "UPDATED"
	ActionDeleted       Action = "DELETED"
	ActionMarkedPayed   Action = "MARKED_PAYED"
	ActionClosed        Action = "CLOSED"
	ActionVoted         Action = "VOTED"
	ActionUnvoted       Action = "UNVOTED"
	ActionMemberJoined  Action = "MEMBER_JOINED"
	ActionMemberLeft    Action = "MEMBER_LEFT"
	ActionMemberRemoved Action = "MEMBER_REMOVED"
	ActionAssigned      Action = "ASSIGNED"
	ActionCompleted     Action = "COMPLETED"
	ActionUncompleted   Action = "UNCOMPLETED"
	ActionMarkRead      Action = "MARK_READ"
)

type RealTimeEvent struct {
	Module Module `json:"module"`
	Action Action `json:"action"`
	Data   any    `json:"data"`
}

// HomeChannel returns the Redis pub/sub channel name for a specific home.
func HomeChannel(homeID int) string {
	return fmt.Sprintf("home:%d:updates", homeID)
}

// SendEvent publishes an event to the given Redis channel.
func SendEvent(ctx context.Context, cache *redis.Client, channel string, event *RealTimeEvent) {
	payload, _ := json.Marshal(event)
	cache.Publish(ctx, channel, payload)
}

// SendHomeEvent publishes an event scoped to a specific home.
func SendHomeEvent(ctx context.Context, cache *redis.Client, homeID int, event *RealTimeEvent) {
	SendEvent(ctx, cache, HomeChannel(homeID), event)
}

// UserChannel returns the Redis pub/sub channel name for a single user.
func UserChannel(userID int) string {
	return fmt.Sprintf("user:%d:updates", userID)
}

// SendUserEvent publishes an event only that user's sockets receive - used for
// private payloads such as direct messages, which must not reach the whole home.
func SendUserEvent(ctx context.Context, cache *redis.Client, userID int, event *RealTimeEvent) {
	SendEvent(ctx, cache, UserChannel(userID), event)
}
