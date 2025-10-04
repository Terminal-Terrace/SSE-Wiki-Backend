# cmd 目录

存放可执行文件的入口点。

## 职责

- 初始化应用配置
- 建立数据库连接
- 启动 HTTP 服务器
- 优雅关闭处理

## 目录结构

```
cmd/
└── server/          # HTTP 服务器入口
    └── main.go      # 主函数
```

## 设计原则

1. **保持简洁**：main.go 应该只负责初始化和启动，不包含业务逻辑
2. **依赖注入**：将初始化好的依赖（如数据库连接）传递给其他层
3. **错误处理**：启动阶段的错误应该直接 fatal，确保服务状态正确

## 典型的 main.go 结构

```go
func main() {
    // 1. 加载配置
    config.Load("config.yaml")

    // 2. 初始化基础设施（数据库、Redis、消息队列等）
    db := database.Init()

    // 3. 自动迁移数据库表
    model.InitTable(db)

    // 4. 设置路由（传入依赖）
    router := route.SetupRouter(db)

    // 5. 启动服务器
    router.Run(":8080")
}
```

## 注意事项

- 不要在 cmd 中使用全局变量
- 配置应该先加载，再初始化依赖
- 数据库连接应该作为参数传递，不要在包级别声明
- 考虑添加优雅关闭逻辑（graceful shutdown）
