# Auth Service

认证服务，提供用户注册、登录、验证码、令牌刷新等功能。

## 启动

```bash
cd services/auth-service
go run cmd/server/main.go
```

服务将在端口 `50051` 上启动。

## 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/login/...

# 查看覆盖率
make test-cover
```

### 测试数据库

测试需要PostgreSQL和Redis，通过Docker Compose启动：

```bash
# 启动测试数据库
docker-compose -f docker-compose.test.yml up -d

# 停止测试数据库
docker-compose -f docker-compose.test.yml down
```

### 覆盖率

```bash
# 生成覆盖率报告
go test -cover ./internal/...

# 生成详细覆盖率报告
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```
