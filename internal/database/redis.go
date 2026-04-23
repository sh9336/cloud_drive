// internal/database/redis.go
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	*redis.Client
}

func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	// Parse URL if provided
	var options *redis.Options
	var err error

	// If addr starts with redis:// or rediss://, treat it as a URL
	if len(addr) > 8 && (addr[:8] == "redis://" || addr[:9] == "rediss://") {
		options, err = redis.ParseURL(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis URL: %w", err)
		}
	} else {
		options = &redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}
	}

	client := redis.NewClient(options)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Println("✅ Redis connection established")

	return &RedisClient{client}, nil
}

func (r *RedisClient) Close() error {
	log.Println("Closing redis connection...")
	return r.Client.Close()
}

func (r *RedisClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}
