package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// 注意：以下测试需要 Redis 实例运行
// 这些是集成测试，测试真实的 Redis 交互
// 如果需要纯单元测试，可以使用 Redis mock 库如 go-redis/redismock

func TestRefreshTokenConstants(t *testing.T) {
	// 测试常量定义
	assert.Equal(t, "refresh_token:", RefreshTokenPrefix)
	assert.Equal(t, "user_refresh_tokens:", UserRefreshTokensPrefix)
	assert.Greater(t, int(RefreshTokenExpiration.Hours()), 0)
}

// 以下测试需要 Redis 实例
// 如果要运行这些测试，需要在测试前初始化 Redis 连接

/*
func TestGenerateAndValidateRefreshToken_Integration(t *testing.T) {
	// 注意：此测试需要 Redis 实例
	// 初始化 Redis 连接
	// database.InitRedis(...)

	userID := 1
	username := "testuser"
	email := "test@example.com"

	// 生成 token
	token, err := GenerateRefreshToken(userID, username, email)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证 token
	validatedUserID, validatedUsername, validatedEmail, err := ValidateRefreshToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedUserID)
	assert.Equal(t, username, validatedUsername)
	assert.Equal(t, email, validatedEmail)

	// 清理：撤销 token
	err = RevokeRefreshToken(token)
	assert.NoError(t, err)

	// 再次验证应该失败
	_, _, _, err = ValidateRefreshToken(token)
	assert.Error(t, err)
}

func TestRevokeAllUserRefreshTokens_Integration(t *testing.T) {
	// 注意：此测试需要 Redis 实例
	userID := 1

	// 生成多个 token
	token1, err := GenerateRefreshToken(userID, "user1", "user1@test.com")
	assert.NoError(t, err)

	token2, err := GenerateRefreshToken(userID, "user1", "user1@test.com")
	assert.NoError(t, err)

	// 验证两个 token 都有效
	_, _, _, err = ValidateRefreshToken(token1)
	assert.NoError(t, err)

	_, _, _, err = ValidateRefreshToken(token2)
	assert.NoError(t, err)

	// 撤销该用户的所有 token
	err = RevokeAllUserRefreshTokens(userID)
	assert.NoError(t, err)

	// 验证两个 token 都已失效
	_, _, _, err = ValidateRefreshToken(token1)
	assert.Error(t, err)

	_, _, _, err = ValidateRefreshToken(token2)
	assert.Error(t, err)
}

func TestGetUserActiveSessions_Integration(t *testing.T) {
	// 注意：此测试需要 Redis 实例
	userID := 2

	// 清理旧数据
	RevokeAllUserRefreshTokens(userID)

	// 初始应该没有 session
	count, err := GetUserActiveSessions(userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	// 生成两个 token
	token1, err := GenerateRefreshToken(userID, "user2", "user2@test.com")
	assert.NoError(t, err)

	token2, err := GenerateRefreshToken(userID, "user2", "user2@test.com")
	assert.NoError(t, err)

	// 应该有 2 个活跃 session
	count, err = GetUserActiveSessions(userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// 撤销一个 token
	err = RevokeRefreshToken(token1)
	assert.NoError(t, err)

	// 应该还剩 1 个
	count, err = GetUserActiveSessions(userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// 清理
	RevokeRefreshToken(token2)
}
*/
