# middleware 目录

中间件层，用于处理 HTTP 请求的横切关注点（Cross-cutting Concerns）。

## 职责

- 认证和授权
- 日志记录
- 错误恢复
- 请求限流
- CORS 处理
- 请求追踪

## 中间件执行顺序

```
Request → Middleware 1 → Middleware 2 → ... → Handler
         ↓                ↓                      ↓
Response ← Middleware 1 ← Middleware 2 ← ... ← Handler
```

## 目录结构

```
middleware/
├── auth.go          # 认证中间件
├── logger.go        # 日志中间件
├── recovery.go      # 错误恢复中间件
├── cors.go          # CORS 中间件
└── rate_limit.go    # 限流中间件
```

## 示例代码

### 认证中间件 (auth.go)

```go
package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
    "terminal-terrace/response"
    "terminal-terrace/template/internal/dto"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 获取 token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            dto.ErrorResponse(c, response.NewBusinessError(
                response.WithErrorCode(response.Unauthorized),
                response.WithErrorMessage("未提供认证令牌"),
            ))
            c.Abort()
            return
        }

        // 2. 解析 token（Bearer xxxxx）
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            dto.ErrorResponse(c, response.NewBusinessError(
                response.WithErrorCode(response.Unauthorized),
                response.WithErrorMessage("认证令牌格式错误"),
            ))
            c.Abort()
            return
        }

        token := parts[1]

        // 3. 验证 token（调用 auth-sdk 或 JWT 库）
        userID, err := ValidateToken(token)
        if err != nil {
            dto.ErrorResponse(c, response.NewBusinessError(
                response.WithErrorCode(response.Unauthorized),
                response.WithErrorMessage("认证令牌无效"),
            ))
            c.Abort()
            return
        }

        // 4. 将用户信息存入上下文
        c.Set("userID", userID)
        c.Next()
    }
}

// OptionalAuth 可选认证中间件（不强制要求登录）
func OptionalAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) == 2 && parts[0] == "Bearer" {
                userID, _ := ValidateToken(parts[1])
                if userID != 0 {
                    c.Set("userID", userID)
                }
            }
        }
        c.Next()
    }
}

// ValidateToken 验证 JWT token（示例）
func ValidateToken(token string) (int, error) {
    // TODO: 实现 JWT 验证逻辑
    // 或调用 auth-sdk 的方法
    return 0, nil
}
```

### 日志中间件 (logger.go)

```go
package middleware

import (
    "time"
    "log"
    "github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery

        // 处理请求
        c.Next()

        // 记录日志
        latency := time.Since(start)
        clientIP := c.ClientIP()
        method := c.Request.Method
        statusCode := c.Writer.Status()
        errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

        if raw != "" {
            path = path + "?" + raw
        }

        log.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s %s",
            start.Format("2006/01/02 - 15:04:05"),
            statusCode,
            latency,
            clientIP,
            method,
            path,
            errorMessage,
        )
    }
}
```

### 错误恢复中间件 (recovery.go)

```go
package middleware

import (
    "log"
    "github.com/gin-gonic/gin"
    "terminal-terrace/response"
    "terminal-terrace/template/internal/dto"
)

// Recovery 错误恢复中间件
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                // 记录错误
                log.Printf("Panic recovered: %v", err)

                // 返回统一错误响应
                dto.ErrorResponse(c, response.NewBusinessError(
                    response.WithErrorCode(response.InternalError),
                    response.WithErrorMessage("服务器内部错误"),
                ))

                c.Abort()
            }
        }()

        c.Next()
    }
}
```

### CORS 中间件 (cors.go)

```go
package middleware

import (
    "os"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
    origin := os.Getenv("FRONTEND_URL")
    if origin == "" {
        origin = "http://localhost:5173"
    }

    return cors.New(cors.Config{
        AllowOrigins:     []string{origin},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    })
}
```

### 权限检查中间件 (permission.go)

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "terminal-terrace/response"
    "terminal-terrace/template/internal/dto"
)

// RequireRole 角色检查中间件
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("userRole")
        if !exists {
            dto.ErrorResponse(c, response.NewBusinessError(
                response.WithErrorCode(response.Forbidden),
                response.WithErrorMessage("权限不足"),
            ))
            c.Abort()
            return
        }

        roleStr := userRole.(string)
        hasPermission := false
        for _, role := range roles {
            if role == roleStr {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            dto.ErrorResponse(c, response.NewBusinessError(
                response.WithErrorCode(response.Forbidden),
                response.WithErrorMessage("权限不足"),
            ))
            c.Abort()
            return
        }

        c.Next()
    }
}
```

## 中间件使用方式

### 全局中间件

```go
router := gin.New()

// 应用全局中间件
router.Use(middleware.Recovery())
router.Use(middleware.Logger())
router.Use(middleware.CORS())
```

### 路由组中间件

```go
api := router.Group("/api/v1")
api.Use(middleware.AuthMiddleware())
{
    api.GET("/users", handler.ListUsers)
    api.POST("/users", handler.CreateUser)
}
```

### 单个路由中间件

```go
router.GET("/admin/users",
    middleware.AuthMiddleware(),
    middleware.RequireRole("admin"),
    handler.ListUsers,
)
```

## 设计原则

1. **单一职责**：每个中间件只做一件事
2. **顺序重要**：中间件的执行顺序很重要（Recovery 应该在最外层）
3. **上下文传递**：使用 `c.Set()` 将数据传递给后续处理器
4. **及时中断**：认证失败时使用 `c.Abort()` 中止请求
5. **错误处理**：中间件应该有完善的错误处理

## 推荐的中间件顺序

```go
router.Use(
    middleware.Recovery(),      // 1. 错误恢复（最外层）
    middleware.Logger(),         // 2. 日志记录
    middleware.CORS(),           // 3. CORS 处理
    middleware.RateLimit(),      // 4. 限流（可选）
)

// 需要认证的路由组
authGroup := router.Group("/api/v1")
authGroup.Use(middleware.AuthMiddleware())
```

## 注意事项

- 使用 `c.Next()` 继续执行后续中间件
- 使用 `c.Abort()` 中止请求处理
- Recovery 中间件应该放在最外层
- 认证中间件应该尽早执行（但在 Recovery 和 Logger 之后）
- 使用 `c.Set()` 和 `c.Get()` 在中间件间传递数据
