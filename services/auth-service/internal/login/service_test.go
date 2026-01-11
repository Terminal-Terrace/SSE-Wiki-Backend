package login

import (
	"os"
	"testing"

	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/pkg"
	"terminal-terrace/auth-service/internal/testutils"
	"terminal-terrace/response"

	"github.com/stretchr/testify/assert"
)

const loginTypeSSEWiki = "sse-wiki"

// TestSSEWikiLoginServiceLogin 测试 SSE-Wiki 登录服务
func TestSSEWikiLoginServiceLogin(t *testing.T) {
	// Setup JWT secret for testing
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	}
	config.MustLoad("config.yaml")

	db := testutils.SetupTestDB(t)
	service := &SSEWikiLoginService{}

	// Create test user
	testUser := testutils.CreateTestUser(db, testutils.WithPassword("password123"))
	username := ""
	if testUser.Username != nil {
		username = *testUser.Username
	}

	// Generate valid state
	state, err := pkg.GenerateState()
	assert.NoError(t, err)
	redirectURL := "https://example.com/home"
	err = pkg.SaveStateWithRedirect(state, redirectURL)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		req         LoginRequest
		expectError bool
		errorMsg    string
		checkResult func(t *testing.T, resp LoginResponse, err *response.BusinessError)
	}{
		{
			name: "successful login with username",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    state,
				Username: username,
				Password: "password123",
			},
			expectError: false,
			checkResult: func(t *testing.T, resp LoginResponse, err *response.BusinessError) {
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
				assert.Equal(t, redirectURL, resp.RedirectUrl)
			},
		},
		{
			name: "successful login with email",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    state,
				Username: testUser.Email,
				Password: "password123",
			},
			expectError: false,
			checkResult: func(t *testing.T, resp LoginResponse, err *response.BusinessError) {
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			},
		},
		{
			name: "invalid password",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    state,
				Username: username,
				Password: "wrongpassword",
			},
			expectError: true,
			errorMsg:    "用户名或密码错误",
		},
		{
			name: "user not found",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    state,
				Username: "nonexistent",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "用户名或密码错误",
		},
		{
			name: "empty state",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    "",
				Username: username,
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "state 不能为空",
		},
		{
			name: "empty username",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    state,
				Username: "",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "用户名不能为空",
		},
		{
			name: "empty password",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    state,
				Username: username,
				Password: "",
			},
			expectError: true,
			errorMsg:    "密码不能为空",
		},
		{
			name: "invalid state",
			req: LoginRequest{
				Type:     loginTypeSSEWiki,
				State:    "invalid_state",
				Username: username,
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "无效的 state 或 state 已过期",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For successful login tests, generate a new state for each test
			// because state is deleted after successful login
			if !tt.expectError && tt.req.State == state {
				newState, err := pkg.GenerateState()
				assert.NoError(t, err)
				err = pkg.SaveStateWithRedirect(newState, redirectURL)
				assert.NoError(t, err)
				tt.req.State = newState
			}

			resp, bizErr := service.Login(tt.req)

			if tt.expectError {
				assert.NotNil(t, bizErr)
				if tt.errorMsg != "" {
					assert.Contains(t, bizErr.Msg, tt.errorMsg)
				}
			} else {
				assert.Nil(t, bizErr)
				if tt.checkResult != nil {
					tt.checkResult(t, resp, bizErr)
				}
			}
		})
	}
}

// TestDoLoginUnsupportedType 测试不支持的登录类型
func TestDoLoginUnsupportedType(t *testing.T) {
	req := LoginRequest{
		Type:     "unsupported",
		State:    "test_state",
		Username: "test",
		Password: "password",
	}

	resp, bizErr := DoLogin(req)

	assert.Empty(t, resp)
	assert.NotNil(t, bizErr)
	assert.Contains(t, bizErr.Msg, "不支持的登录类型")
}

