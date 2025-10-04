package login

// LoginRequest 登录请求
type LoginRequest struct {
	Type     string `json:"type" binding:"required" example:"sse-wiki" enums:"sse-wiki,github,sse-market"`  // 登录服务提供者
	State    string `json:"state" binding:"required" example:"abc123def456"`                               // CSRF 防护用的 state
	Username string `json:"username" example:"admin"`                                                      // SSE-Wiki 用户名
	Password string `json:"password" example:"password123"`                                                // SSE-Wiki 密码
	Code     string `json:"code" example:"github_oauth_code"`                                              // 第三方 OAuth 授权码
}

// LoginResponse 登录响应
type LoginResponse struct {
	RefreshToken string `json:"refresh_token,omitempty" example:"refresh_token_xxx"` // 刷新令牌
	RedirectUrl  string `json:"redirect_url,omitempty" example:"https://example.com/home"`  // 重定向 URL
}