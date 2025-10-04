package pkg

import (
	"testing"
	"time"

	"terminal-terrace/auth-service/config"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessToken(t *testing.T) {
	// 初始化配置
	config.Conf = &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			ExpireTime: 24,
		},
	}

	tests := []struct {
		name     string
		userID   int
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "生成有效的访问令牌",
			userID:   1,
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
		{
			name:     "用户ID为0",
			userID:   0,
			username: "testuser",
			email:    "test@example.com",
			wantErr:  false,
		},
		{
			name:     "用户名为空",
			userID:   1,
			username: "",
			email:    "test@example.com",
			wantErr:  false,
		},
		{
			name:     "邮箱为空",
			userID:   1,
			username: "testuser",
			email:    "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateAccessToken(tt.userID, tt.username, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestParseAccessToken(t *testing.T) {
	// 初始化配置
	config.Conf = &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			ExpireTime: 24,
		},
	}

	// 生成一个有效的令牌用于测试
	userID := 1
	username := "testuser"
	email := "test@example.com"
	validToken, err := GenerateAccessToken(userID, username, email)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		wantErr   bool
		expectErr error
	}{
		{
			name:      "解析有效的令牌",
			token:     validToken,
			wantErr:   false,
			expectErr: nil,
		},
		{
			name:      "解析空令牌",
			token:     "",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
		{
			name:      "解析无效的令牌",
			token:     "invalid.token.string",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
		{
			name:      "解析格式错误的令牌",
			token:     "not-a-jwt-token",
			wantErr:   true,
			expectErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ParseAccessToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.expectErr != nil {
					assert.ErrorIs(t, err, tt.expectErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, username, claims.Username)
				assert.Equal(t, email, claims.Email)
			}
		})
	}
}

func TestParseAccessToken_WithDifferentSecret(t *testing.T) {
	// 使用一个密钥生成令牌
	config.Conf = &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:     "secret-key-1",
			ExpireTime: 24,
		},
	}

	token, err := GenerateAccessToken(1, "testuser", "test@example.com")
	assert.NoError(t, err)

	// 使用不同的密钥尝试解析
	config.Conf.JWT.Secret = "secret-key-2"

	claims, err := ParseAccessToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestParseAccessToken_ExpiredToken(t *testing.T) {
	// 生成一个立即过期的令牌
	config.Conf = &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			ExpireTime: -1, // 负数会导致令牌立即过期
		},
	}

	token, err := GenerateAccessToken(1, "testuser", "test@example.com")
	assert.NoError(t, err)

	// 等待一小段时间确保令牌过期
	time.Sleep(10 * time.Millisecond)

	claims, err := ParseAccessToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestClaims_Fields(t *testing.T) {
	config.Conf = &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			ExpireTime: 24,
		},
	}

	userID := 12345
	username := "testuser123"
	email := "test123@example.com"

	token, err := GenerateAccessToken(userID, username, email)
	assert.NoError(t, err)

	claims, err := ParseAccessToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// 验证所有字段
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, email, claims.Email)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)

	// 验证时间字段的合理性
	now := time.Now()
	assert.True(t, claims.IssuedAt.Time.Before(now) || claims.IssuedAt.Time.Equal(now))
	assert.True(t, claims.ExpiresAt.Time.After(now))
}
