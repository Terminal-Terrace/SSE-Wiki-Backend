# Database 包使用说明

## 快速开始

### 1. 在 main.go 中初始化数据库连接

```go
package main

import (
    "log"
    "time"

    "terminal-terrace/database"
    "terminal-terrace/your-service/config"
    "terminal-terrace/your-service/internal/route"
)

func main() {
    // 1. 加载配置
    config.Load("config.yaml")

    // 2. 初始化数据库连接
    dbConfig := &database.PostgresConfig{
        Username:        config.Conf.Database.Username,
        Password:        config.Conf.Database.Password,
        Host:            config.Conf.Database.Host,
        Port:            config.Conf.Database.Port,
        Database:        config.Conf.Database.Database,
        SSLMode:         config.Conf.Database.SSLMode,
        LogLevel:        config.Conf.Log.Level,
        MaxIdleConns:    config.Conf.Database.MaxIdleConns,
        MaxOpenConns:    config.Conf.Database.MaxOpenConns,
        ConnMaxLifetime: time.Duration(config.Conf.Database.MaxLifetime) * time.Second,
    }

    db, err := database.InitPostgres(dbConfig)
    if err != nil {
        log.Fatalf("数据库初始化失败: %v", err)
    }

    // 3. 将 db 传递给路由层
    router := route.SetupRouter(db)

    // 4. 启动服务
    router.Run(":8080")
}
```

### 2. 在 Repository 中使用数据库连接

```go
package repository

import (
    "gorm.io/gorm"
    "terminal-terrace/your-service/internal/model"
)

type UserRepository struct {
    db *gorm.DB
}

// NewUserRepository 创建用户仓库
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// GetByID 根据 ID 查询用户
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.First(&user, id).Error
    return &user, err
}

// Create 创建用户
func (r *UserRepository) Create(user *model.User) error {
    return r.db.Create(user).Error
}
```

## 配置说明

### PostgresConfig 字段

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Username | string | 数据库用户名 | - |
| Password | string | 数据库密码 | - |
| Host | string | 数据库地址 | localhost |
| Port | int | 数据库端口 | 5432 |
| Database | string | 数据库名称 | - |
| SSLMode | bool | 是否启用 SSL | false |
| LogLevel | string | 日志级别 | info |
| MaxIdleConns | int | 最大空闲连接数 | 10 |
| MaxOpenConns | int | 最大打开连接数 | 100 |
| ConnMaxLifetime | time.Duration | 连接最大生命周期 | 1 小时 |

### 日志级别

- `silent`: 不输出任何日志
- `error`: 只输出错误日志
- `warn`: 输出警告和错误日志
- `info`: 输出所有日志（包括 SQL 查询）

## Debug 输出示例

启动时会自动输出配置检查信息：

```
[Database] 配置检查:
  Host: localhost
  Port: 5432
  Username: postgres
  Password: ***
  Database: sse_wiki
  SSLMode: false
  LogLevel: info
[Database] 已连接到 localhost:5432/sse_wiki
```

如果某个配置项未设置，会显示 `未设置`：

```
[Database] 配置检查:
  Host: localhost
  Port: 0
  Username: 未设置
  Password: 未设置
  Database: 未设置
  SSLMode: false
  LogLevel: 未设置
```

## 注意事项

### ✅ 应该做的

1. **只在 main.go 中初始化一次**
2. **将连接传递给 Repository**
3. **在 Repository 中执行具体的 SQL 操作**

### ❌ 不应该做的

1. **不要重复初始化连接**
2. **不要在 Handler/Service 中直接使用 db**
3. **不要在此包中添加业务 SQL**

## 架构流程

```
main.go
  ├─ config.Load()          加载配置
  ├─ database.InitPostgres() 初始化数据库连接
  └─ route.SetupRouter(db)  传递连接给路由
       │
       └─ NewRepository(db)  传递给 Repository
            │
            └─ NewService(repo) 传递给 Service
                 │
                 └─ NewHandler(service) 传递给 Handler
```
