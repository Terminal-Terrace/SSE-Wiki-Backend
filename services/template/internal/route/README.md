# route 目录

Route 层负责定义 HTTP 路由，将 URL 路径映射到对应的 Handler 方法。

## 职责

- 定义 API 路由
- 配置路由中间件
- 路由分组和版本控制
- 初始化依赖并注入到 Handler

## 目录结构

```
route/
├── route.go          # 主路由设置
├── user_routes.go    # 用户相关路由
├── article_routes.go # 文章相关路由
└── auth_routes.go    # 认证相关路由
```

## 示例代码

### 主路由文件 (route.go)

```go
package route

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "terminal-terrace/template/internal/handler"
    "terminal-terrace/template/internal/middleware"
    "terminal-terrace/template/internal/repository"
    "terminal-terrace/template/internal/service"
)

// SetupRouter 初始化路由
func SetupRouter(db *gorm.DB) *gin.Engine {
    router := gin.New()

    // 1. 全局中间件
    router.Use(
        middleware.Recovery(),    // 错误恢复
        middleware.Logger(),       // 日志
        middleware.CORS(),         // CORS
    )

    // 2. 初始化依赖（Repository -> Service -> Handler）
    // User 相关
    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)

    // Article 相关
    articleRepo := repository.NewArticleRepository(db)
    articleService := service.NewArticleService(articleRepo, userRepo)
    articleHandler := handler.NewArticleHandler(articleService)

    // Auth 相关
    authService := service.NewAuthService(userRepo)
    authHandler := handler.NewAuthHandler(authService)

    // 3. 注册路由
    registerAuthRoutes(router, authHandler)
    registerUserRoutes(router, userHandler)
    registerArticleRoutes(router, articleHandler)

    return router
}
```

### 认证路由 (auth_routes.go)

```go
package route

import (
    "github.com/gin-gonic/gin"
    "terminal-terrace/template/internal/handler"
)

// registerAuthRoutes 注册认证相关路由
func registerAuthRoutes(router *gin.Engine, h *handler.AuthHandler) {
    auth := router.Group("/api/v1/auth")
    {
        auth.POST("/register", h.Register)
        auth.POST("/login", h.Login)
        auth.POST("/logout", h.Logout)
        auth.POST("/refresh", h.RefreshToken)
    }
}
```

### 用户路由 (user_routes.go)

```go
package route

import (
    "github.com/gin-gonic/gin"
    "terminal-terrace/template/internal/handler"
    "terminal-terrace/template/internal/middleware"
)

// registerUserRoutes 注册用户相关路由
func registerUserRoutes(router *gin.Engine, h *handler.UserHandler) {
    users := router.Group("/api/v1/users")
    {
        // 公开路由
        users.GET("", h.ListUsers)                    // 获取用户列表
        users.GET("/:id", h.GetUser)                  // 获取用户详情

        // 需要认证的路由
        auth := users.Group("")
        auth.Use(middleware.AuthMiddleware())
        {
            auth.PUT("/:id", h.UpdateUser)            // 更新用户
            auth.DELETE("/:id", h.DeleteUser)         // 删除用户
            auth.GET("/:id/profile", h.GetProfile)    // 获取用户资料
        }

        // 需要管理员权限的路由
        admin := users.Group("/admin")
        admin.Use(
            middleware.AuthMiddleware(),
            middleware.RequireRole("admin"),
        )
        {
            admin.POST("", h.CreateUser)              // 创建用户
            admin.DELETE("/:id", h.DeleteUser)        // 删除用户
        }
    }
}
```

### 文章路由 (article_routes.go)

```go
package route

import (
    "github.com/gin-gonic/gin"
    "terminal-terrace/template/internal/handler"
    "terminal-terrace/template/internal/middleware"
)

// registerArticleRoutes 注册文章相关路由
func registerArticleRoutes(router *gin.Engine, h *handler.ArticleHandler) {
    articles := router.Group("/api/v1/articles")
    {
        // 公开路由
        articles.GET("", h.ListArticles)                           // 文章列表
        articles.GET("/:id", h.GetArticle)                         // 文章详情

        // 需要认证的路由
        auth := articles.Group("")
        auth.Use(middleware.AuthMiddleware())
        {
            auth.POST("", h.CreateArticle)                         // 创建文章
            auth.PUT("/:id", h.UpdateArticle)                      // 更新文章
            auth.DELETE("/:id", h.DeleteArticle)                   // 删除文章
        }
    }

    // 模块相关的文章路由
    modules := router.Group("/api/v1/modules")
    {
        modules.GET("/:moduleId/articles", h.GetArticlesByModule)  // 获取模块下的文章
    }
}
```

## 路由组织方式

### 1. 按版本分组

```go
// API v1
v1 := router.Group("/api/v1")
{
    v1.GET("/users", handler.ListUsers)
    v1.GET("/articles", handler.ListArticles)
}

// API v2
v2 := router.Group("/api/v2")
{
    v2.GET("/users", handlerV2.ListUsers)
}
```

### 2. 按功能模块分组

```go
// 用户模块
users := router.Group("/api/v1/users")
{
    users.GET("", handler.ListUsers)
    users.POST("", handler.CreateUser)
    users.GET("/:id", handler.GetUser)
}

// 文章模块
articles := router.Group("/api/v1/articles")
{
    articles.GET("", handler.ListArticles)
    articles.POST("", handler.CreateArticle)
}
```

### 3. 按权限分组

```go
api := router.Group("/api/v1")

// 公开路由
public := api.Group("")
{
    public.POST("/auth/login", authHandler.Login)
    public.GET("/articles", articleHandler.List)
}

// 需要认证的路由
auth := api.Group("")
auth.Use(middleware.AuthMiddleware())
{
    auth.POST("/articles", articleHandler.Create)
    auth.PUT("/articles/:id", articleHandler.Update)
}

// 管理员路由
admin := api.Group("/admin")
admin.Use(
    middleware.AuthMiddleware(),
    middleware.RequireRole("admin"),
)
{
    admin.DELETE("/users/:id", userHandler.Delete)
    admin.GET("/stats", statsHandler.Get)
}
```

## 中间件应用

### 全局中间件

```go
router.Use(middleware.Recovery())
router.Use(middleware.Logger())
router.Use(middleware.CORS())
```

### 路由组中间件

```go
api := router.Group("/api/v1")
api.Use(middleware.AuthMiddleware())
{
    api.GET("/profile", handler.GetProfile)
}
```

### 单个路由中间件

```go
router.GET("/admin/stats",
    middleware.AuthMiddleware(),
    middleware.RequireRole("admin"),
    middleware.RateLimit(),
    handler.GetStats,
)
```

## 路径参数和查询参数

```go
// 路径参数
router.GET("/users/:id", handler.GetUser)           // /users/123
router.GET("/users/:id/posts/:postId", handler)     // /users/123/posts/456

// 查询参数
router.GET("/users", handler.ListUsers)             // /users?page=1&size=20

// 组合使用
router.GET("/users/:id/articles", handler)          // /users/123/articles?page=1
```

## RESTful API 路由规范

```go
// 资源的 CRUD 操作
router.GET("/users", handler.ListUsers)           // 列表
router.POST("/users", handler.CreateUser)         // 创建
router.GET("/users/:id", handler.GetUser)         // 详情
router.PUT("/users/:id", handler.UpdateUser)      // 更新（全量）
router.PATCH("/users/:id", handler.PatchUser)     // 更新（部分）
router.DELETE("/users/:id", handler.DeleteUser)   // 删除

// 子资源
router.GET("/users/:id/articles", handler.GetUserArticles)
router.POST("/users/:id/articles", handler.CreateUserArticle)

// 非 CRUD 操作（动词）
router.POST("/users/:id/follow", handler.FollowUser)
router.POST("/users/:id/unfollow", handler.UnfollowUser)
router.POST("/articles/:id/publish", handler.PublishArticle)
```

## 依赖注入模式

```go
func SetupRouter(db *gorm.DB) *gin.Engine {
    router := gin.New()

    // 初始化依赖链：Repository -> Service -> Handler
    userRepo := repository.NewUserRepository(db)
    userService := service.NewUserService(userRepo)
    userHandler := handler.NewUserHandler(userService)

    // 注册路由
    router.GET("/users", userHandler.ListUsers)

    return router
}
```

## 路由文件组织建议

### 方案一：按模块分文件

```
route/
├── route.go              # 主入口
├── user_routes.go        # 用户路由
├── article_routes.go     # 文章路由
└── auth_routes.go        # 认证路由
```

### 方案二：按版本分文件

```
route/
├── route.go              # 主入口
├── v1_routes.go          # v1 API
└── v2_routes.go          # v2 API
```

### 方案三：综合方式

```
route/
├── route.go              # 主入口，组装所有路由
├── v1/
│   ├── user_routes.go
│   └── article_routes.go
└── v2/
    └── user_routes.go
```

## 设计原则

1. **RESTful 风格**：遵循 RESTful API 设计规范
2. **版本控制**：通过路径前缀区分 API 版本（`/api/v1`, `/api/v2`）
3. **模块化**：按业务模块拆分路由文件
4. **中间件分层**：全局 -> 路由组 -> 单个路由
5. **依赖注入**：在路由初始化时完成依赖注入

## 注意事项

- 路由定义应该简洁明了，一目了然
- 相同前缀的路由使用路由组
- 中间件的顺序很重要（Recovery 应该最外层）
- 路径参数使用 `:id` 形式，查询参数在 Handler 中获取
- 避免在路由文件中写业务逻辑
- 使用明确的 HTTP 方法（GET, POST, PUT, DELETE, PATCH）
