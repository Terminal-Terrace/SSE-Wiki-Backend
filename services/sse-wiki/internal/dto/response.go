package dto

import (
	"fmt"
	"strings"

	res "terminal-terrace/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func SuccessResponse(c *gin.Context, data any) {
	c.JSON(200, res.SuccessResponse(data))
}

func ErrorResponse(c *gin.Context, err *res.BusinessError) {
	c.JSON(200, res.ErrorResponse(err.Code, err.Msg))
}

// ValidationErrorResponse 处理验证错误，返回友好的JSON字段名
func ValidationErrorResponse(c *gin.Context, err error) {
	// 尝试转换为 validator.ValidationErrors
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		// 获取第一个错误
		if len(validationErrs) > 0 {
			firstErr := validationErrs[0]

			// 获取字段的JSON标签名
			jsonField := getJSONFieldName(firstErr)

			// 构造友好的错误消息
			var message string
			switch firstErr.Tag() {
			case "required":
				message = fmt.Sprintf("字段 '%s' 是必填项", jsonField)
			case "max":
				message = fmt.Sprintf("字段 '%s' 长度不能超过 %s", jsonField, firstErr.Param())
			case "min":
				message = fmt.Sprintf("字段 '%s' 长度不能少于 %s", jsonField, firstErr.Param())
			case "oneof":
				message = fmt.Sprintf("字段 '%s' 必须是以下值之一: %s", jsonField, firstErr.Param())
			default:
				message = fmt.Sprintf("字段 '%s' 验证失败: %s", jsonField, firstErr.Tag())
			}

			ErrorResponse(c, res.NewBusinessError(
				res.WithErrorCode(res.ParseError),
				res.WithErrorMessage(message),
			))
			return
		}
	}

	// 如果不是 validation 错误，返回原始错误消息
	ErrorResponse(c, res.NewBusinessError(
		res.WithErrorCode(res.ParseError),
		res.WithErrorMessage("参数错误: "+err.Error()),
	))
}

// getJSONFieldName 获取字段的JSON标签名称
func getJSONFieldName(fe validator.FieldError) string {
	// 获取字段所在的结构体命名空间
	field := fe.StructNamespace()

	// 尝试通过反射获取JSON标签
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		if len(parts) > 1 {
			// 获取最后一个字段名（去掉结构体名称前缀）
			fieldName := parts[len(parts)-1]

			// 尝试从错误中获取结构体类型并查找JSON标签
			// 由于validator不直接提供结构体实例，我们只能返回字段名的snake_case版本
			return toSnakeCase(fieldName)
		}
	}

	return toSnakeCase(fe.Field())
}

// toSnakeCase 将PascalCase转换为snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}