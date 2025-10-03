package login

import (
	"terminal-terrace/response"
)

type LoginService interface {
	Login(req LoginRequest) (LoginResponse, *response.BusinessError)
}

// provider: loginService
var loginServices = make(map[string]LoginService)

// 在init调用, 之后不再修改
func registerLoginService(name string, service LoginService) {
	loginServices[name] = service
}
