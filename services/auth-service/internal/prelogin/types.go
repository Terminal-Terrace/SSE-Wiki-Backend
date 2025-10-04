package prelogin

// PreLoginRequest 预登录请求
type PreLoginRequest struct {
	RedirectUrl string `json:"redirect_url" binding:"required" example:"https://example.com/callback"` // 登录成功后的重定向地址
}

// PreLoginResponse 预登录响应
type PreLoginResponse struct {
	State string `json:"state" example:"abc123def456"` // 生成的 state
}