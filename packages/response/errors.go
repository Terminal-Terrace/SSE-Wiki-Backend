package response

// 业务错误码
const (
	// 失败
	Fail ResponseCode = 0
	// 参数解析错误
	ParseError ResponseCode = 1
	// 参数错误
	InvalidParameter ResponseCode = 2
	// 未授权
	Unauthorized ResponseCode = 401
	// 禁止访问
	Forbidden ResponseCode = 403
	// 未找到
	NotFound ResponseCode = 404
)

type BusinessError struct {
	Code  ResponseCode
	Msg   string
	Err   error
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	if e.Err != nil {
		return e.Msg + ": " + e.Err.Error()
	}
	return e.Msg
}

type ErrorOption func(*BusinessError)

func WithErrorCode(code ResponseCode) ErrorOption {
	return func(be *BusinessError) {
		be.Code = code
	}
}

func WithErrorMessage(msg string) ErrorOption {
	return func(be *BusinessError) {
		be.Msg = msg
	}
}

func WithError(err error) ErrorOption {
	return func(be *BusinessError) {
		be.Err = err
	}
}

func NewBusinessError(opts ...ErrorOption) *BusinessError {
	err := &BusinessError{
		Code: Fail,
		Msg:  "business error",
		Err:  nil,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}
