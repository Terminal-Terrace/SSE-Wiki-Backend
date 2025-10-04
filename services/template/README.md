# Template Service

这是一个标准的服务模板，展示了推荐的项目结构和代码组织方式。

## 目录结构

```
template/
├── cmd/              # 可执行文件入口
│   └── server/       # HTTP 服务器入口
├── config/           # 配置管理
├── internal/         # 内部代码（不可被外部导入）
│   ├── dto/          # 数据传输对象（请求/响应结构）
│   ├── handler/      # HTTP 处理器（Controller层）
│   ├── middleware/   # 中间件
│   ├── model/        # 数据模型（GORM Model）
│   ├── repository/   # 数据访问层（DAO层）
│   ├── route/        # 路由定义
│   └── service/      # 业务逻辑层
├── pkg/              # 可被外部导入的公共工具
├── go.mod
└── Makefile
```

## 架构说明

本项目采用经典的分层架构：

```
请求流程：Route → Middleware → Handler → Service → Repository → Database
响应流程：Database → Repository → Service → Handler → DTO → Response
```

### 各层职责

- **Route (路由层)**：定义 HTTP 路由、绑定处理器、配置中间件
- **Middleware (中间件层)**：处理认证、日志、CORS 等横切关注点
- **Handler (处理层)**：处理 HTTP 请求、参数验证、调用 Service、返回响应
- **Service (业务层)**：核心业务逻辑、事务控制
- **Repository (数据层)**：数据库操作、缓存操作
- **Model (模型层)**：数据库表结构定义
- **DTO (传输层)**：定义 API 的请求和响应结构

## 设计原则

1. **依赖注入**：通过构造函数注入依赖，便于测试和解耦
2. **接口抽象**：Repository 层使用接口定义，方便替换实现
3. **统一响应格式**：使用 `terminal-terrace/response` 包统一响应格式
4. **配置集中管理**：使用 config 包统一管理配置，支持环境变量覆盖
5. **错误处理规范**：使用 `response.BusinessError` 处理业务错误

## 使用建议

### 路由组织

- 按业务模块分文件（如 `user_routes.go`, `article_routes.go`）
- 使用路由组进行版本控制（如 `/api/v1`）
- 中间件按需应用到路由组

### 数据库连接

- 使用 `terminal-terrace/database` 包初始化数据库连接
- 在 main.go 中初始化后传递给各层
- Repository 层接收 `*gorm.DB` 作为依赖

### 错误处理

- Service 层返回 `(result, *response.BusinessError)`
- Handler 层统一处理错误并返回
- 使用预定义的错误码（在 `terminal-terrace/response` 中）

## 开发流程

1. 在 `model/` 定义数据模型
2. 在 `dto/` 定义请求/响应结构
3. 在 `repository/` 实现数据访问
4. 在 `service/` 实现业务逻辑
5. 在 `handler/` 实现 HTTP 处理
6. 在 `route/` 注册路由

## 命令

```bash
# 安装依赖
make install

# 运行服务
make run

# 构建
make build
```
