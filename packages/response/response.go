package response

type ResponseCode int

// 统一业务代码
const (
	Success = 100
)

type Response struct {
	Message string       `json:"message"`
	Code    ResponseCode `json:"code"`
	Data    any          `json:"data"`
}

type ResponseOptions func(*Response)

func WithMessage(message string) ResponseOptions {
	return func(r *Response) {
		r.Message = message
	}
}

func WithCode(code ResponseCode) ResponseOptions {
	return func(r *Response) {
		r.Code = code
	}
}

func WithData(data any) ResponseOptions {
	return func(r *Response) {
		r.Data = data
	}
}

func CustomResponse(opts ...ResponseOptions) Response {
	response := Response{}
	for _, opt := range opts {
		opt(&response)
	}
	return response
}

func SuccessResponse(data any) Response {
	return Response{
		Message: "success",
		Code:    Success,
		Data:    data,
	}
}

func ErrorResponse(code ResponseCode, msg string) Response {
	return Response{
		Message: msg,
		Code:    code,
		Data:    nil,
	}
}
