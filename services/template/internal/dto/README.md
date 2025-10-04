# dto 目录

DTO (Data Transfer Object) 数据传输对象，定义 API 的请求和响应结构。

## 职责

- 定义 HTTP 请求体的结构（Request DTO）
- 定义 HTTP 响应体的结构（Response DTO）
- 提供统一的响应封装函数
- 数据验证标签定义

## 为什么需要 DTO？

1. **解耦**：将 API 层的数据结构与数据库模型分离
2. **控制**：精确控制哪些字段暴露给前端
3. **验证**：使用 tag 进行参数验证
4. **版本控制**：同一个 Model 可以有多个 DTO 版本

## 目录结构

```
dto/
├── request.go       # 请求 DTO
├── response.go      # 响应 DTO 和封装函数
└── converter.go     # Model 和 DTO 之间的转换函数（可选）
```

## 示例代码

### 请求 DTO (request.go)

```go
package dto

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
    Username *string `json:"username" binding:"omitempty,min=3,max=20"`
    Email    *string `json:"email" binding:"omitempty,email"`
}

// LoginRequest 登录请求
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}
```

### 响应 DTO (response.go)

```go
package dto

import (
    "github.com/gin-gonic/gin"
    res "terminal-terrace/response"
)

// UserResponse 用户响应
type UserResponse struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    // 注意：不包含 Password 等敏感字段
}

// LoginResponse 登录响应
type LoginResponse struct {
    Token string       `json:"token"`
    User  UserResponse `json:"user"`
}

// 统一响应函数
func SuccessResponse(c *gin.Context, data any) {
    c.JSON(200, res.SuccessResponse(data))
}

func ErrorResponse(c *gin.Context, err *res.BusinessError) {
    c.JSON(200, res.ErrorResponse(err.Code, err.Msg))
}
```

### 转换函数 (converter.go)

```go
package dto

import "terminal-terrace/template/internal/model"

// ToUserResponse 将 Model 转换为 Response DTO
func ToUserResponse(user *model.User) *UserResponse {
    return &UserResponse{
        ID:       user.ID,
        Username: user.Username,
        Email:    user.Email,
    }
}

// ToUserResponseList 批量转换
func ToUserResponseList(users []*model.User) []*UserResponse {
    result := make([]*UserResponse, len(users))
    for i, user := range users {
        result[i] = ToUserResponse(user)
    }
    return result
}
```

## Gin 验证标签

常用的 binding 标签：

- `required` - 必填字段
- `email` - 邮箱格式
- `min=n` - 最小长度/值
- `max=n` - 最大长度/值
- `len=n` - 固定长度
- `oneof=red green` - 枚举值
- `omitempty` - 可选字段（用于 Update DTO）

## 设计原则

1. **请求和响应分离**：不要混用同一个结构体
2. **只包含必要字段**：响应 DTO 不要包含敏感信息
3. **使用指针处理可选字段**：Update 请求中使用 `*string` 区分 "未提供" 和 "空值"
4. **添加验证标签**：在请求 DTO 上添加 binding 标签
5. **提供转换函数**：在 converter.go 中统一处理转换逻辑

## 注意事项

- Request DTO 中使用 `binding` 标签进行验证
- Response DTO 中不要包含密码等敏感字段
- 使用 `json` 标签控制 JSON 序列化的字段名
- Update 请求使用指针类型，方便区分"不更新"和"更新为空"
