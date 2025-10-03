package login

import "terminal-terrace/response"

type SSEMarketLoginService struct{}

func init() {
	registerLoginService("sse-market", &SSEMarketLoginService{})
}

func (s *SSEMarketLoginService) Login(req LoginRequest) (LoginResponse, *response.BusinessError) {
	// TODO: 业务逻辑
	// 突然发现软工集市也做了oauth, 应该更方便了
	return LoginResponse{}, nil
}