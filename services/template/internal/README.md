# internal 目录

internal 目录存放服务的内部代码，这些代码无法被其他包导入（Go 语言特性）。

## 目录结构

```
internal/
├── dto/          # 数据传输对象（请求/响应结构）
├── handler/      # HTTP 处理器（Controller 层）
├── middleware/   # 中间件
├── model/        # 数据模型（GORM Model）
├── repository/   # 数据访问层（DAO 层）
├── route/        # 路由定义
└── service/      # 业务逻辑层
```

## 为什么使用 internal？

Go 语言规定，`internal` 目录下的包只能被其父目录及子目录下的包导入，这样可以：

1. **强制封装**：确保内部实现细节不会被外部包使用
2. **API 边界清晰**：明确哪些是公开 API，哪些是内部实现
3. **防止误用**：避免其他服务直接依赖内部实现

## 分层架构

本项目采用经典的分层架构：

```
┌─────────────────────────────────────────────┐
│              HTTP Request                    │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│  Route (路由层)                              │
│  - 定义 URL 路径                             │
│  - 绑定 Handler                              │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│  Middleware (中间件层)                       │
│  - 认证/授权                                 │
│  - 日志/监控                                 │
│  - 错误恢复                                  │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│  Handler (处理层)                            │
│  - 接收 HTTP 请求                            │
│  - 参数验证                                  │
│  - 调用 Service                              │
│  - 返回响应                                  │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│  Service (业务层)                            │
│  - 核心业务逻辑                              │
│  - 业务规则验证                              │
│  - 事务控制                                  │
│  - 调用 Repository                           │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│  Repository (数据层)                         │
│  - 数据库 CRUD                               │
│  - 缓存操作                                  │
│  - 数据查询                                  │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│  Model (模型层)                              │
│  - 数据库表结构                              │
│  - GORM 映射                                 │
└─────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────┐
│              Database                        │
└─────────────────────────────────────────────┘
```

## 请求处理流程

### 完整的请求流程示例

```
1. 用户发送 POST /api/v1/users 创建用户

2. Route 层
   - 匹配路由 POST /api/v1/users -> userHandler.CreateUser

3. Middleware 层
   - Recovery: 捕获 panic
   - Logger: 记录请求日志
   - Auth: 验证用户身份（如果需要）

4. Handler 层 (userHandler.CreateUser)
   - 绑定请求参数到 dto.CreateUserRequest
   - 验证参数是否合法
   - 调用 userService.CreateUser()
   - 将 Model 转换为 DTO
   - 返回 JSON 响应

5. Service 层 (userService.CreateUser)
   - 检查用户名是否已存在（业务规则）
   - 检查邮箱是否已被使用（业务规则）
   - 密码加密
   - 调用 userRepo.Create()
   - 返回创建的用户

6. Repository 层 (userRepo.Create)
   - 执行 GORM 操作
   - db.Create(&user)
   - 返回结果

7. Model 层
   - User 结构体定义
   - GORM 自动处理 CreatedAt, UpdatedAt
   - BeforeCreate 钩子（如果有）

8. 响应返回
   - Repository -> Service -> Handler -> Response
```

## 数据流转

### 请求数据流（向下）

```
HTTP JSON → DTO (Request) → Service → Repository → Model → Database
```

### 响应数据流（向上）

```
Database → Model → Repository → Service → DTO (Response) → HTTP JSON
```

## 各层职责对比

| 层级 | 输入 | 输出 | 主要职责 | 是否包含业务逻辑 |
|-----|------|------|---------|----------------|
| Route | URL | Handler | 路由映射 | ❌ |
| Middleware | Request | Request/Response | 横切关注点 | ❌ |
| Handler | HTTP Request | HTTP Response | 请求处理、参数验证 | ❌ |
| Service | DTO/Params | Model/BusinessError | 业务逻辑 | ✅ |
| Repository | Query Params | Model/error | 数据访问 | ❌ |
| Model | - | - | 数据结构定义 | ❌ |

## 依赖关系

```
Route
  ↓ 依赖
Handler
  ↓ 依赖
Service
  ↓ 依赖
Repository
  ↓ 依赖
Model
```

**原则**：
- 上层可以依赖下层，下层不能依赖上层
- 同层之间尽量不要相互依赖
- 通过接口解耦（特别是 Repository 层）

## DTO vs Model

### Model (数据库模型)
```go
// 完整的数据库表结构
type User struct {
    ID        uint
    Username  string
    Password  string    // 包含敏感信息
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### DTO (数据传输对象)
```go
// 请求 DTO：只包含必需字段
type CreateUserRequest struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// 响应 DTO：排除敏感字段
type UserResponse struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    // 不包含 Password
}
```

## 错误处理

```
Database Error (error)
         ↓
Repository (返回 error)
         ↓
Service (转换为 BusinessError)
         ↓
Handler (返回 HTTP 响应)
```

## 事务管理

事务应该在 **Service 层** 控制：

```go
func (s *UserService) CreateUserWithProfile(req *dto.Request) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 创建用户
        if err := tx.Create(&user).Error; err != nil {
            return err  // 自动回滚
        }

        // 2. 创建用户资料
        if err := tx.Create(&profile).Error; err != nil {
            return err  // 自动回滚
        }

        return nil  // 提交事务
    })
}
```

## 开发新功能的步骤

1. **定义 Model** (`model/xxx.go`)
   - 定义数据库表结构

2. **定义 DTO** (`dto/xxx_request.go`, `dto/xxx_response.go`)
   - 请求和响应的数据结构

3. **实现 Repository** (`repository/xxx_repository.go`)
   - 数据访问层，封装数据库操作

4. **实现 Service** (`service/xxx_service.go`)
   - 业务逻辑层，调用 Repository

5. **实现 Handler** (`handler/xxx_handler.go`)
   - HTTP 处理层，调用 Service

6. **注册路由** (`route/xxx_routes.go`)
   - 定义 URL 路径，绑定 Handler

## 测试策略

```
Unit Test (单元测试)
├── Service 层：Mock Repository 接口
├── Repository 层：使用测试数据库
└── Handler 层：Mock Service

Integration Test (集成测试)
└── 端到端测试：从 HTTP 请求到数据库
```

## 设计原则总结

1. **单一职责**：每一层只做自己该做的事
2. **依赖注入**：通过构造函数注入依赖
3. **接口抽象**：面向接口编程（特别是 Repository 层）
4. **向下依赖**：上层依赖下层，下层不依赖上层
5. **薄 Handler，厚 Service**：Handler 只做参数验证和调用转发，业务逻辑都在 Service

## 常见问题

### Q: 业务逻辑应该放在哪里？
A: **Service 层**。Handler 只做参数验证和调用转发。

### Q: 数据库操作应该放在哪里？
A: **Repository 层**。Service 层通过 Repository 接口操作数据。

### Q: 事务应该在哪里控制？
A: **Service 层**。因为事务往往涉及多个操作，属于业务逻辑范畴。

### Q: DTO 和 Model 的区别？
A:
- **Model**：数据库表的完整映射，包含所有字段
- **DTO**：API 层的数据传输对象，只包含需要的字段，不暴露敏感信息

### Q: 中间件应该放在哪里应用？
A:
- 全局中间件：在 `route.go` 中使用 `router.Use()`
- 路由组中间件：在路由组上使用 `.Use()`
- 单个路由：直接在路由定义中添加
