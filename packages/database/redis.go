package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	rdb *redis.Client
	ctx context.Context
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func InitRedis(config *RedisConfig) (*RedisClient, error) {
	if config == nil {
		return nil, fmt.Errorf("RedisConfig cannot be nil")
	}
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB, 
	})

	// Ping
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{rdb: rdb, ctx: context.Background()}, nil
}

func (c *RedisClient) Set(key string, value string, expire time.Duration) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.Set(c.ctx, key, value, expire).Err()
}

func (c *RedisClient) Get(key string) (string, error) {
	if c.rdb == nil {
		return "", fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.Get(c.ctx, key).Result()
}

func (c *RedisClient) Delete(key string) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.Del(c.ctx, key).Err()
}

func (c *RedisClient) Close() error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.Close()
}
