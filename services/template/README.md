# Template Service (gRPC)

这是一个纯 gRPC 服务模板，展示了推荐的项目结构和代码组织方式。

## 目录结构

```
template/
├── cmd/                    # 可执行文件入口
│   └── server/             # gRPC 服务器入口
├── config/                 # 配置管理
├── internal/               # 内部代码（不可被外部导入）
│   ├── grpc/               # gRPC service 实现
│   ├── model/              # 数据模型（GORM Model）
│   └── repository/         # 数据访问层（DAO层）
├── protobuf/               # Protobuf 相关
│   └── proto/              # 生成的 Go 代码
├── go.mod
├── pb.config.json          # protobuf 编译配置
└── Makefile
```

## 架构说明

本项目采用纯 gRPC 架构，HTTP 请求由 Node.js BFF 层处理：

```
请求流程：Node.js BFF → gRPC Client → gRPC Server → Service Logic → Repository → Database
响应流程：Database → Repository → Service Logic → gRPC Response → Node.js BFF → HTTP Response
```

### 各层职责

- **gRPC Service (服务层)**：实现 protobuf 定义的 service 接口，处理业务逻辑
- **Repository (数据层)**：数据库操作、缓存操作
- **Model (模型层)**：数据库表结构定义

## 设计原则

1. **纯 gRPC 通信**：不暴露 HTTP 端口，所有外部请求通过 Node.js BFF 转发
2. **依赖注入**：通过构造函数注入依赖，便于测试和解耦
3. **接口抽象**：Repository 层使用接口定义，方便替换实现
4. **配置集中管理**：使用 config 包统一管理配置，支持环境变量覆盖
5. **错误处理规范**：使用 gRPC status codes 处理错误

## 开发流程

### 1. 定义 Protobuf

在 `SSE-WIKI-Proto` 仓库中定义你的 service：

```protobuf
syntax = "proto3";

package template;

service TemplateService {
  rpc GetExample(GetExampleRequest) returns (GetExampleResponse);
}

message GetExampleRequest {
  int64 id = 1;
}

message GetExampleResponse {
  string name = 1;
}
```

### 2. 生成 Go 代码

配置 `pb.config.json`：

```json
{
  "protoRepoPath": "../../SSE-WIKI-Proto",
  "outputPath": "./protobuf/proto",
  "services": ["template_service"]
}
```

运行生成命令（在 monorepo 根目录）：

```bash
make proto-go SERVICE=template
```

### 3. 实现 gRPC Service

在 `internal/grpc/` 中实现 service：

```go
package grpc

import (
    "context"
    pb "terminal-terrace/template/protobuf/proto/template_service"
)

type TemplateServiceImpl struct {
    pb.UnimplementedTemplateServiceServer
    // 注入 repository
}

func NewTemplateServiceImpl() *TemplateServiceImpl {
    return &TemplateServiceImpl{}
}

func (s *TemplateServiceImpl) GetExample(ctx context.Context, req *pb.GetExampleRequest) (*pb.GetExampleResponse, error) {
    // 实现业务逻辑
    return &pb.GetExampleResponse{Name: "example"}, nil
}
```

### 4. 注册 Service

在 `internal/grpc/server.go` 中注册：

```go
func NewServer(port int, templateService pb.TemplateServiceServer) (*Server, error) {
    // ...
    pb.RegisterTemplateServiceServer(grpcServer, templateService)
    // ...
}
```

### 5. 更新 main.go

```go
templateService := grpcserver.NewTemplateServiceImpl()
server, err := grpcserver.NewServer(grpcPort, templateService)
```

## 配置

配置文件 `config.yaml`：

```yaml
grpc:
  port: 50053

database:
  driver: "postgres"
  host: "localhost"
  # ...
```

## 命令

```bash
# 运行服务
make run

# 构建
make build

# 生成 protobuf
make proto
```

## 端口规划

| 服务 | gRPC 端口 |
|------|-----------|
| auth-service | 50051 |
| sse-wiki | 50052 |
| template | 50053 |
