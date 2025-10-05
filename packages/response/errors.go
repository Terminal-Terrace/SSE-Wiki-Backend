package response

// 业务错误码
const (
	// 失败
	Fail ResponseCode = 0
	// 参数解析错误
	ParseError ResponseCode = 1
	// 参数错误
	InvalidParameter ResponseCode = 2
)

type BusinessError struct {
	Code  ResponseCode
	Msg   string
	Err   error
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
