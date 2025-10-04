package code

type CodeType int

const (
	CodeTypeRegister CodeType = iota + 1
	CodeTypeResetPassword
)

type SendCodeRequest struct {
	Email string   `json:"email" binding:"required,email"`
	Type  CodeType `json:"type" binding:"required"`
}

type SendCodeResponse struct{}
