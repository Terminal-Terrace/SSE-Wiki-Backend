package login

import (
	"terminal-terrace/response"
)

type SSEWikiLoginService struct{}

func init() {
	registerLoginService("sse-wiki", &SSEWikiLoginService{})
}

// 我们自己的登录服务, 使用账号密码登录
func (s *SSEWikiLoginService) Login(req LoginRequest) (LoginResponse, *response.BusinessError) {
	// TODO: 业务逻辑

	// 检查参数

	// 检查state

	// 校验账号密码

	// 生成token

	// 返回结果
	return LoginResponse{}, nil
}