# config 目录

配置管理模块，负责加载和解析应用配置。

## 职责

- 从 YAML 文件加载配置
- 支持环境变量覆盖配置
- 提供类型安全的配置访问接口
- 配置热加载（可选）

## 使用方式

```go
// 在 main.go 中加载配置
config.Load("config.yaml")

// 访问配置
dbHost := config.Conf.Database.Host
serverPort := config.Conf.Server.Port
```

## 配置文件示例 (config.yaml)

```yaml
server:
  host: "localhost"
  port: 8080
  mode: "debug"

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  username: "user"
  password: "password"
  database: "dbname"
  sslmode: false
```

## 环境变量覆盖

支持通过环境变量覆盖配置文件中的值，使用 `APP_` 前缀：

```bash
APP_DATABASE_HOST=prod-db.example.com
APP_SERVER_PORT=9000
```

## 配置结构定义

在 `config.go` 中定义配置结构体：

```go
type AppConfig struct {
    Server   ServerConfig   `koanf:"server"`
    Database DatabaseConfig `koanf:"database"`
    Redis    RedisConfig    `koanf:"redis"`
    Log      LogConfig      `koanf:"log"`
    JWT      JWTConfig      `koanf:"jwt"`
}
```

## 设计原则

1. **单例模式**：使用 `sync.Once` 确保配置只加载一次
2. **类型安全**：使用结构体而非 map 访问配置
3. **环境感知**：支持开发/测试/生产环境的不同配置
4. **敏感信息**：密码等敏感信息应优先从环境变量读取

## 注意事项

- 配置文件不应包含敏感信息（密码、密钥等）
- 生产环境应使用环境变量覆盖敏感配置
- 配置修改后需要重启服务（除非实现了热加载）
