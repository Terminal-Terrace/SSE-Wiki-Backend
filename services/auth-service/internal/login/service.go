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

// DoLogin is the entry point for login, exposed for gRPC usage
// It selects the appropriate login service based on the request type
func DoLogin(req LoginRequest) (LoginResponse, *response.BusinessError) {
	service, exists := loginServices[req.Type]
	if !exists {
		return LoginResponse{}, response.NewBusinessError(
			response.WithErrorCode(response.Fail),
			response.WithErrorMessage("不支持的登录类型"),
		)
	}
	return service.Login(req)
}
