package pkg

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"terminal-terrace/auth-service/internal/database"
)

const (
	// RefreshToken 有效期：7天
	RefreshTokenExpiration = 7 * 24 * time.Hour
	// RefreshToken Redis key 前缀
	RefreshTokenPrefix = "refresh_token:"
	// 用户的 RefreshToken 集合 key 前缀（用于查看用户的所有活跃 session）
	UserRefreshTokensPrefix = "user_refresh_tokens:"
)

// GenerateRefreshToken 生成刷新令牌并存储到 Redis
func GenerateRefreshToken(userID int, username, email string) (string, error) {
	// 生成随机字符串
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("生成随机令牌失败: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)

	// 存储到 Redis
	ctx := context.Background()
	key := RefreshTokenPrefix + token

	// 存储令牌信息：userID, username, email
	tokenData := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"email":    email,
	}

	if err := database.RedisDB.HSet(ctx, key, tokenData).Err(); err != nil {
		return "", fmt.Errorf("存储令牌失败: %w", err)
	}

	// 设置过期时间
	if err := database.RedisDB.Expire(ctx, key, RefreshTokenExpiration).Err(); err != nil {
		return "", fmt.Errorf("设置令牌过期时间失败: %w", err)
	}

	// 将 token 添加到用户的 token 集合中（用于管理用户的所有 session）
	userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)
	if err := database.RedisDB.SAdd(ctx, userTokensKey, token).Err(); err != nil {
		return "", fmt.Errorf("添加到用户令牌集合失败: %w", err)
	}

	// 设置用户 token 集合的过期时间
	if err := database.RedisDB.Expire(ctx, userTokensKey, RefreshTokenExpiration).Err(); err != nil {
		return "", fmt.Errorf("设置用户令牌集合过期时间失败: %w", err)
	}

	return token, nil
}

// ValidateRefreshToken 验证刷新令牌并返回用户信息
func ValidateRefreshToken(token string) (userID int, username, email string, err error) {
	ctx := context.Background()
	key := RefreshTokenPrefix + token

	// 从 Redis 获取令牌信息
	tokenData, err := database.RedisDB.HGetAll(ctx, key).Result()
	if err != nil {
		return 0, "", "", fmt.Errorf("获取令牌信息失败: %w", err)
	}

	if len(tokenData) == 0 {
		return 0, "", "", fmt.Errorf("令牌不存在或已过期")
	}

	// 解析用户信息
	userIDStr, ok := tokenData["user_id"]
	if !ok {
		return 0, "", "", fmt.Errorf("令牌数据不完整")
	}

	userID, err = strconv.Atoi(userIDStr)
	if err != nil {
		return 0, "", "", fmt.Errorf("用户 ID 格式错误: %w", err)
	}

	username = tokenData["username"]
	email = tokenData["email"]

	return userID, username, email, nil
}

// RevokeRefreshToken 撤销刷新令牌（用户登出）
func RevokeRefreshToken(token string) error {
	ctx := context.Background()
	key := RefreshTokenPrefix + token

	// 先获取用户 ID，以便从用户的 token 集合中删除
	tokenData, err := database.RedisDB.HGetAll(ctx, key).Result()
	if err == nil && len(tokenData) > 0 {
		if userIDStr, ok := tokenData["user_id"]; ok {
			userID, _ := strconv.Atoi(userIDStr)
			userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)
			database.RedisDB.SRem(ctx, userTokensKey, token).Err()
		}
	}

	// 删除令牌
	if err := database.RedisDB.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("撤销令牌失败: %w", err)
	}

	return nil
}

// RevokeAllUserRefreshTokens 撤销用户的所有刷新令牌（修改密码、强制登出等场景）
func RevokeAllUserRefreshTokens(userID int) error {
	ctx := context.Background()
	userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)

	// 获取用户的所有 token
	tokens, err := database.RedisDB.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("获取用户令牌列表失败: %w", err)
	}

	// 删除所有 token
	for _, token := range tokens {
		key := RefreshTokenPrefix + token
		database.RedisDB.Del(ctx, key).Err()
	}

	// 删除用户的 token 集合
	if err := database.RedisDB.Del(ctx, userTokensKey).Err(); err != nil {
		return fmt.Errorf("删除用户令牌集合失败: %w", err)
	}

	return nil
}

// GetUserActiveSessions 获取用户的所有活跃 session 数量
func GetUserActiveSessions(userID int) (int, error) {
	ctx := context.Background()
	userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)

	count, err := database.RedisDB.SCard(ctx, userTokensKey).Result()
	if err != nil {
		return 0, fmt.Errorf("获取活跃会话数失败: %w", err)
	}

	return int(count), nil
}
