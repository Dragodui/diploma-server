package cache

import (
	"context"
	"crypto/tls"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(addr, password string) *redis.Client {

	if addr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}
	var tlsConfig *tls.Config

	if strings.Contains(addr, "amazonaws.com") {
		tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		log.Println("Redis TLS enabled (AWS ElastiCache detected)")
	}
	
	client := redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  password,
		DB:        0,
		TLSConfig: tlsConfig,

		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	return client
}
