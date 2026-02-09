package event

import (
	"context"
	"encoding/json"

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

func SendEvent(ctx context.Context, cache *redis.Client, channel string, event *RealTimeEvent) {
	payload, _ := json.Marshal(event)
	cache.Publish(ctx, channel, payload)
}
