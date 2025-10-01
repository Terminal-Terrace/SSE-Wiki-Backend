package service

import (
	"terminal-terrace/response"
)

type ExampleService struct {
}

func NewExampleService() *ExampleService {
	return &ExampleService{}
}

func (s *ExampleService) DoSomeGood() (string, *response.BusinessError) {
	return "Good", nil
}

func (s *ExampleService) DoSomeBad() (string, *response.BusinessError) {
	return "", response.NewBusinessError(
		response.WithErrorCode(response.Fail),
		response.WithErrorMessage("something went wrong"),
	)
}
