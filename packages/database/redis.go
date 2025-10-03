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

// HSet 设置 hash 字段
func (c *RedisClient) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.HSet(ctx, key, values).Err()
}

// HGetAll 获取 hash 所有字段
func (c *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if c.rdb == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.HGetAll(ctx, key).Result()
}

// Expire 设置 key 过期时间
func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.Expire(ctx, key, expiration).Err()
}

// SAdd 添加元素到集合
func (c *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.SAdd(ctx, key, members...).Err()
}

// SRem 从集合中删除元素
func (c *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.SRem(ctx, key, members...).Err()
}

// SMembers 获取集合所有成员
func (c *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	if c.rdb == nil {
		return nil, fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.SMembers(ctx, key).Result()
}

// SCard 获取集合成员数量
func (c *RedisClient) SCard(ctx context.Context, key string) (int64, error) {
	if c.rdb == nil {
		return 0, fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.SCard(ctx, key).Result()
}

// Del 删除一个或多个 key
func (c *RedisClient) Del(ctx context.Context, keys ...string) error {
	if c.rdb == nil {
		return fmt.Errorf("Redis client is not initialized")
	}
	return c.rdb.Del(ctx, keys...).Err()
}
