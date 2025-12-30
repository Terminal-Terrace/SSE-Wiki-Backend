# 终端露台 Go 大仓

## TODO:
1. service层重构：
`article/service.go` 中多个方法返回 `map[string]interface{}`，导致 gRPC 层需要大量类型断言和转换代码。


重构方案：
1. 在 `dto/` 下定义响应结构体（如 `ArticleDetailResponse`、`ArticleListResponse`、`ReviewDetailResponse` 等）
2. 修改 service 层方法返回强类型结构体指针
3. gRPC 层直接使用结构体字段，无需类型断言

可以移除大量类型转换代码


## 架构说明

本项目采用纯 gRPC 架构，所有服务通过 gRPC 对外提供接口，由 Node.js BFF 层统一处理 HTTP 请求并转发到 gRPC 服务。

### 服务端口

| 服务 | gRPC 端口 | 说明 |
|------|-----------|------|
| auth-service | 50051 | 认证服务 |
| sse-wiki | 50052 | Wiki 主服务 |
| template | 50053 | 服务模板（参考标准） |

## 项目结构

```
SSE-Wiki-Backend/
├── services/              # 微服务目录
│   ├── auth-service/      # 认证服务 (gRPC: 50051)
│   ├── sse-wiki/          # SSE Wiki 主服务 (gRPC: 50052)
│   └── template/          # gRPC 服务模板（创建新服务时参考）
├── packages/              # 共享包目录
│   ├── database/          # 统一数据库连接管理
│   ├── response/          # 统一响应格式
│   └── email/             # 邮件服务
├── .env.example           # 环境变量模板
├── go.work                # Go Workspace 配置
├── Makefile               # 构建脚本
└── README.md              # 本文件
```

## 文件结构

一个包的结构大概这样:

- config 存放配置
- internal 存放包内部使用的代码, 无法被其它包导入
  - pkg 内部使用的工具
  - ... 其它
- pkg 外部可使用的工具
- cmd 可执行的文件/命令

service最好不要导出东西, 导入package里的就可以了.

## 快速开始

### 环境

待补充各软件版本

### 配置

`.env.example` 为环境变量模板，需要配置拷贝并命名为 `.env`，完成内部相关配置
注意 JWT_SECRET 不要加引号，且要保证和 node 部分一致

### 运行

在子包运行

```sh
make install
make run
```

或者在根目录运行

调用子包的 `make install`

```sh
make install 子包名
```

调用子包的 `make run`

```sh
make run 子包名
```

如果没有make, 也可以跟平时一样运行.

```sh
go mod tidy
go run xxx
```

在window系统中双击run.bat即可运行

## 数据库

项目使用 GORM AutoMigrate 自动同步数据库表结构。

### 新增模型

1. 创建模型文件

```go
// services/auth-service/internal/model/role/role.go
package role

type Role struct {
    ID   int    `gorm:"primaryKey" json:"id"`
    Name string `gorm:"uniqueIndex" json:"name"`
}

func (Role) TableName() string {
    return "auth_roles"
}
```

2. 注册到 model.go

```go
func GetModels() []interface{} {
    return []interface{}{
        &user.User{},
        &role.Role{}, // 新增
    }
}
```

3. 启动服务自动创建表

```bash
go run cmd/server/main.go
```

### 配置

各服务使用独立数据库, 在 config.yaml 中配置数据库名。环境变量覆盖配置文件。

```bash
# .env
DATABASE_HOST=localhost
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=your_password
```

## Proto 同步

使用 `cpb` 工具从 GitHub 同步 proto 文件并生成 Go 代码。

1. 环境

- Node.js 18+ 和 pnpm
- protoc
- protoc-gen-go
- protoc-gen-go-grpc

2. 首次配置

```bash
# 1. 安装 cpb 工具
cd ../SSE-Wiki-Nodejs/packages/pb-compiler
pnpm install && pnpm build && pnpm link --global

# 2. 配置 GitHub 访问
cpb login <your-github-token>
cpb config --owner Terminal-Terrace --repo SSE-WIKI-Proto
```

3. 安装之后，运行一下指令即可分别对两个 service 进行代码同步

```bash
cd services/auth-service
cpb sync --branch <branch-name>
cpb gen-go --go-opt "Mproto/auth_service/auth_service.proto=terminal-terrace/auth-service/protobuf/proto/auth_service"
```

```bash
cd services/sse-wiki
cpb sync --bran <branch-name>
cpb gen-go \
  --go-opt "Mproto/article_service/article_service.proto=terminal-terrace/sse-wiki/protobuf/proto/article_service" \
  --go-opt "Mproto/module_service/module_service.proto=terminal-terrace/sse-wiki/protobuf/proto/module_service" \
  --go-opt "Mproto/review_service/review_service.proto=terminal-terrace/sse-wiki/protobuf/proto/review_service" \
  --go-opt "Mproto/discussion_service/discussion_service.proto=terminal-terrace/sse-wiki/protobuf/proto/discussion_service"
```

## 测试部分

1. Go 测试文件必须以 `_test.go` 结尾，与被测试文件放在同一目录：

```
internal/
├── article/
│   ├── service.go          # 业务逻辑
│   ├── service_test.go     # 对应的测试文件
│   └── repository.go
├── permission/
│   ├── service.go
│   └── service_test.go
└── module/
    ├── service.go
    └── service_test.go
```

2. 函数命名

```go
// 基础测试：Test + 被测函数名
func TestGetRoleLevel(t *testing.T) { ... }

// 场景测试：Test + 功能描述 + 场景
func TestGlobalAdmin_CannotEditBasicInfo(t *testing.T) { ... }

// 属性测试：Test + 功能 + Property
func TestCanDelete_Property(t *testing.T) { ... }
```

3. 代码结构

推荐使用表驱动测试（Table-Driven Tests）：

```go
func TestAddCollaborator_RoleHierarchy(t *testing.T) {
    tests := []struct {
        name         string  // 测试用例名称
        operatorID   uint    // 输入参数
        isAuthor     bool
        operatorRole string
        targetRole   string
        shouldAllow  bool    // 期望结果
    }{
        {
            name:         "Author can add admin",
            operatorID:   1,
            isAuthor:     true,
            operatorRole: "admin",
            targetRole:   "admin",
            shouldAllow:  true,
        },
        {
            name:         "Admin cannot add admin",
            operatorID:   1,
            isAuthor:     false,
            operatorRole: "admin",
            targetRole:   "admin",
            shouldAllow:  false,
        },
        // 更多测试用例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 执行测试逻辑
            result := someFunction(tt.operatorID, tt.isAuthor, tt.operatorRole, tt.targetRole)
            
            if result != tt.shouldAllow {
                t.Errorf("got %v, want %v", result, tt.shouldAllow)
            }
        })
    }
}
```

4. 运行测试

```bash
# 运行单个包的测试
cd services/sse-wiki
go test ./internal/article/...

# 运行多个包的测试
go test ./internal/article/... ./internal/permission/... ./internal/module/...

# 运行特定测试函数
go test ./internal/article/... -run TestGlobalAdmin
```

## 暂时的预期

### auth-service

参见[飞书文档](https://mcn0xmurkm53.feishu.cn/docx/C4z7dMc0co932PxyyXkcPT1Fn5e)

### auth-sdk

预计导出这些东西供外部使用:

- authMiddleware

认证中间件, 所有的服务应该都使用这个中间件. 处理用户鉴权. 顺便将一些用户信息存到上下文里.

已完成

### 多文件存储

改用OSS存储服务，go 后端无须感知文件内容。