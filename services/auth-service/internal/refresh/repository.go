package refresh

import (
	"context"
	"fmt"
	"strconv"
	"terminal-terrace/database"
	"time"
)

const (
	// RefreshToken 有效期：7天
	RefreshTokenExpiration = 7 * 24 * time.Hour
	// RefreshToken Redis key 前缀
	RefreshTokenPrefix = "refresh_token:"
	// 用户的 RefreshToken 集合 key 前缀（用于查看用户的所有活跃 session）
	UserRefreshTokensPrefix = "user_refresh_tokens:"
)

// RefreshTokenRepository 刷新令牌数据访问层
type RefreshTokenRepository struct {
	redis *database.RedisClient
}

// NewRefreshTokenRepository 创建刷新令牌仓库实例
func NewRefreshTokenRepository(redisClient *database.RedisClient) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		redis: redisClient,
	}
}

// TokenData 令牌数据结构
type TokenData struct {
	UserID   int
	Username string
	Email    string
	Role     string
}

// Create 创建刷新令牌并存储到 Redis
func (r *RefreshTokenRepository) Create(token string, data TokenData) error {
	ctx := context.Background()
	key := RefreshTokenPrefix + token

	// 存储令牌信息：userID, username, email, role
	tokenData := map[string]interface{}{
		"user_id":  data.UserID,
		"username": data.Username,
		"email":    data.Email,
		"role":     data.Role,
	}

	if err := r.redis.HSet(ctx, key, tokenData).Err(); err != nil {
		return fmt.Errorf("存储令牌失败: %w", err)
	}

	// 设置过期时间
	if err := r.redis.Expire(ctx, key, RefreshTokenExpiration).Err(); err != nil {
		return fmt.Errorf("设置令牌过期时间失败: %w", err)
	}

	// 将 token 添加到用户的 token 集合中（用于管理用户的所有 session）
	userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(data.UserID)
	if err := r.redis.SAdd(ctx, userTokensKey, token).Err(); err != nil {
		return fmt.Errorf("添加到用户令牌集合失败: %w", err)
	}

	// 设置用户 token 集合的过期时间
	if err := r.redis.Expire(ctx, userTokensKey, RefreshTokenExpiration).Err(); err != nil {
		return fmt.Errorf("设置用户令牌集合过期时间失败: %w", err)
	}

	return nil
}

// Get 获取刷新令牌信息
func (r *RefreshTokenRepository) Get(token string) (*TokenData, error) {
	ctx := context.Background()
	key := RefreshTokenPrefix + token

	// 从 Redis 获取令牌信息
	tokenData, err := r.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取令牌信息失败: %w", err)
	}

	if len(tokenData) == 0 {
		return nil, fmt.Errorf("令牌不存在或已过期")
	}

	// 解析用户信息
	userIDStr, ok := tokenData["user_id"]
	if !ok {
		return nil, fmt.Errorf("令牌数据不完整")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("用户 ID 格式错误: %w", err)
	}

	return &TokenData{
		UserID:   userID,
		Username: tokenData["username"],
		Email:    tokenData["email"],
		Role:     tokenData["role"],
	}, nil
}

// Delete 删除刷新令牌（用户登出）
func (r *RefreshTokenRepository) Delete(token string) error {
	ctx := context.Background()
	key := RefreshTokenPrefix + token

	// 先获取用户 ID，以便从用户的 token 集合中删除
	tokenData, err := r.redis.HGetAll(ctx, key).Result()
	if err == nil && len(tokenData) > 0 {
		if userIDStr, ok := tokenData["user_id"]; ok {
			userID, _ := strconv.Atoi(userIDStr)
			userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)
			r.redis.SRem(ctx, userTokensKey, token).Err()
		}
	}

	// 删除令牌
	if err := r.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("撤销令牌失败: %w", err)
	}

	return nil
}

// DeleteAllByUserID 删除用户的所有刷新令牌（修改密码、强制登出等场景）
func (r *RefreshTokenRepository) DeleteAllByUserID(userID int) error {
	ctx := context.Background()
	userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)

	// 获取用户的所有 token
	tokens, err := r.redis.SMembers(ctx, userTokensKey).Result()
	if err != nil {
		return fmt.Errorf("获取用户令牌列表失败: %w", err)
	}

	// 删除所有 token
	for _, token := range tokens {
		key := RefreshTokenPrefix + token
		r.redis.Del(ctx, key).Err()
	}

	// 删除用户的 token 集合
	if err := r.redis.Del(ctx, userTokensKey).Err(); err != nil {
		return fmt.Errorf("删除用户令牌集合失败: %w", err)
	}

	return nil
}

// CountActiveSessionsByUserID 获取用户的所有活跃 session 数量
func (r *RefreshTokenRepository) CountActiveSessionsByUserID(userID int) (int, error) {
	ctx := context.Background()
	userTokensKey := UserRefreshTokensPrefix + strconv.Itoa(userID)

	count, err := r.redis.SCard(ctx, userTokensKey).Result()
	if err != nil {
		return 0, fmt.Errorf("获取活跃会话数失败: %w", err)
	}

	return int(count), nil
}
