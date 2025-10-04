package register

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username        string `json:"username" binding:"required" example:"newuser"`         // 用户名
	Password        string `json:"password" binding:"required" example:"password123"`     // 密码
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"password123"` // 确认密码
	Email           string `json:"email" binding:"required,email" example:"user@example.com"` // 邮箱
	Code            string `json:"code" binding:"required" example:"123456"`              // 邮箱验证码
	State           string `json:"state" binding:"required" example:"abc123def456"`       // CSRF 防护 state
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token_xxx"`       // 刷新令牌
	RedirectUrl  string `json:"redirect_url" example:"https://example.com/home"` // 重定向 URL
}