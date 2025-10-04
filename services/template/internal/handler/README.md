# handler 目录

Handler 层（也称 Controller 层）负责处理 HTTP 请求，是 API 的入口点。

## 职责

- 接收 HTTP 请求
- 验证请求参数
- 调用 Service 层执行业务逻辑
- 处理错误并返回响应
- 不包含业务逻辑（业务逻辑应在 Service 层）

## 目录结构

```
handler/
├── user_handler.go      # 用户相关的 handler
├── article_handler.go   # 文章相关的 handler
└── auth_handler.go      # 认证相关的 handler
```

## 示例代码

### Handler 定义

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "terminal-terrace/response"
    "terminal-terrace/template/internal/dto"
    "terminal-terrace/template/internal/service"
)

type UserHandler struct {
    userService *service.UserService
}

// NewUserHandler 构造函数（依赖注入）
func NewUserHandler(userService *service.UserService) *UserHandler {
    return &UserHandler{
        userService: userService,
    }
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
    // 1. 绑定并验证请求参数
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        dto.ErrorResponse(c, response.NewBusinessError(
            response.WithErrorCode(response.ParseError),
            response.WithErrorMessage("参数验证失败: " + err.Error()),
        ))
        return
    }

    // 2. 调用 Service 层
    user, err := h.userService.CreateUser(&req)
    if err != nil {
        dto.ErrorResponse(c, err)
        return
    }

    // 3. 转换为 Response DTO 并返回
    dto.SuccessResponse(c, dto.ToUserResponse(user))
}

// GetUser 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
    // 获取路径参数
    userID := c.Param("id")

    user, err := h.userService.GetUserByID(userID)
    if err != nil {
        dto.ErrorResponse(c, err)
        return
    }

    dto.SuccessResponse(c, dto.ToUserResponse(user))
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(c *gin.Context) {
    userID := c.Param("id")

    var req dto.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        dto.ErrorResponse(c, response.NewBusinessError(
            response.WithErrorCode(response.ParseError),
            response.WithErrorMessage(err.Error()),
        ))
        return
    }

    user, err := h.userService.UpdateUser(userID, &req)
    if err != nil {
        dto.ErrorResponse(c, err)
        return
    }

    dto.SuccessResponse(c, dto.ToUserResponse(user))
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
    userID := c.Param("id")

    err := h.userService.DeleteUser(userID)
    if err != nil {
        dto.ErrorResponse(c, err)
        return
    }

    dto.SuccessResponse(c, gin.H{"message": "删除成功"})
}

// ListUsers 获取用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
    // 获取查询参数
    page := c.DefaultQuery("page", "1")
    pageSize := c.DefaultQuery("page_size", "20")

    users, total, err := h.userService.ListUsers(page, pageSize)
    if err != nil {
        dto.ErrorResponse(c, err)
        return
    }

    dto.SuccessResponse(c, gin.H{
        "users": dto.ToUserResponseList(users),
        "total": total,
        "page":  page,
    })
}
```

## Handler 的典型结构

```go
func (h *XxxHandler) SomeAction(c *gin.Context) {
    // 1. 参数获取和验证
    var req dto.SomeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        dto.ErrorResponse(c, ...)
        return
    }

    // 2. 调用 Service 层
    result, err := h.service.DoSomething(&req)
    if err != nil {
        dto.ErrorResponse(c, err)
        return
    }

    // 3. 返回响应
    dto.SuccessResponse(c, result)
}
```

## 参数绑定方式

Gin 提供了多种参数绑定方法：

```go
// JSON 请求体绑定
c.ShouldBindJSON(&req)

// 查询参数绑定
c.ShouldBindQuery(&req)

// 表单绑定
c.ShouldBind(&req)

// URI 参数
c.Param("id")

// 查询参数（带默认值）
c.DefaultQuery("page", "1")

// 请求头
c.GetHeader("Authorization")
```

## 错误处理

Handler 层应该：
1. 捕获 Service 层返回的 `BusinessError`
2. 使用 `dto.ErrorResponse` 统一返回错误
3. 不要直接处理业务错误，应该由 Service 层返回

```go
result, err := h.service.DoSomething()
if err != nil {
    // 直接返回 Service 层的错误
    dto.ErrorResponse(c, err)
    return
}
```

## 设计原则

1. **薄层原则**：Handler 应该很薄，只做参数验证和调用转发
2. **无业务逻辑**：所有业务逻辑都应该在 Service 层
3. **依赖注入**：通过构造函数注入 Service 依赖
4. **统一响应**：使用 dto 中定义的统一响应函数
5. **错误传递**：不要在 Handler 中创建业务错误，应该由 Service 返回

## 注意事项

- 不要在 Handler 中直接操作数据库
- 不要在 Handler 中写业务逻辑
- 使用 `ShouldBindJSON` 而非 `BindJSON`（前者不会自动返回 400）
- 参数验证失败应返回 `ParseError` 错误码
- Handler 的方法签名固定为 `func(c *gin.Context)`
