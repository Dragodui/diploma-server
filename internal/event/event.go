package event

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RealTimeEvent struct {
	// TODO: create modules type for all modules in app
	Module string
	// TODO: create action type
	Action string
	Data   any
}

func SendEvent(ctx context.Context, cache *redis.Client, channel string, event *RealTimeEvent) {
	payload, _ := json.Marshal(event)
	cache.Publish(ctx, channel, payload)
}
