package cache

import (
	"context"

	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedisClient() *redis.Client {
	cfg := config.Load()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisADDR,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	return rdb
}
