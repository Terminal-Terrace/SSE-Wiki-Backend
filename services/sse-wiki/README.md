# SSE-Wiki Service

SSE-Wiki 主服务，提供文章、模块、讨论等核心功能的 gRPC 接口。

## 架构

本服务采用纯 gRPC 架构，通过 Node.js BFF 层对外提供 HTTP API。

## 启动

```bash
cd services/sse-wiki
go run cmd/server/main.go
```

服务将在 gRPC 端口 `50052` 上启动。

## 文件结构

- `cmd/server` - 服务入口
- `config` - 配置文件和配置加载
- `internal/` - 内部代码
  - `database` - 数据库初始化
  - `grpc` - gRPC 服务实现
  - `model` - 数据库模型
  - `article` - 文章业务逻辑
  - `module` - 模块业务逻辑
  - `discussion` - 讨论业务逻辑
- `protobuf/` - Protocol Buffers 生成代码

## gRPC 服务

- `ArticleService` - 文章管理
- `ModuleService` - 模块管理
- `ReviewService` - 审核管理
- `DiscussionService` - 讨论管理

## 配置

配置文件 `config.yaml`：

```yaml
grpc:
  port: 50052

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  # ...
```
