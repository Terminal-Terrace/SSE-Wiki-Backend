package pkg

import (
	"context"
	"time"

	"terminal-terrace/auth-service/internal/database"
)

const (
	// State 有效期：10分钟
	StateExpiration = 10 * time.Minute
	// State Redis key 前缀
	StatePrefix = "auth_state:"
)

// SaveStateWithRedirect 保存 state 和重定向地址到 Redis
func SaveStateWithRedirect(state, redirectUrl string) error {
	ctx := context.Background()
	key := StatePrefix + state

	return database.RedisDB.Set(ctx, key, redirectUrl, StateExpiration).Err()
}

// GetRedirectByState 根据 state 获取重定向地址
func GetRedirectByState(state string) (string, error) {
	ctx := context.Background()
	key := StatePrefix + state

	redirectUrl, err := database.RedisDB.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return redirectUrl, nil
}

// DeleteState 删除 state（使用后删除，防止重复使用）
func DeleteState(state string) error {
	ctx := context.Background()
	key := StatePrefix + state

	return database.RedisDB.Del(ctx, key).Err()
}