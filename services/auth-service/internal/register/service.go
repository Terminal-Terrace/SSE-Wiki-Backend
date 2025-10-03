package register

import "terminal-terrace/response"

type RegisterService struct{}

// 只支持账号密码注册
func (s *RegisterService) Register(req RegisterRequest) (RegisterResponse, *response.BusinessError) {
	// TODO: 实现注册逻辑
	return RegisterResponse{}, nil
}
