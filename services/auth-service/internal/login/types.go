package login

type LoginRequest struct {
	Type     string `json:"type" binding:"required"`  // 登录服务提供者，如 "sse-wiki", "github" 等
	State    string `json:"state" binding:"required"` // csrf 防护用的 state
	Username string `json:"username"`                 // "sse-wiki" 类型的用户名
	Password string `json:"password"`                 // "sse-wiki" 类型的密码
	Code     string `json:"code"`                     // 第三方登录的授权码，如 GitHub OAuth 的 code
}

type LoginResponse struct {
	RefreshToken string `json:"refresh_token,omitempty"` // 刷新令牌
	RedirectUrl  string `json:"redirect_url,omitempty"`  // 第三方登录的重定向 URL
}