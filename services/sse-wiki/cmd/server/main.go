package main

import (
	"terminal-terrace/sse-wiki/config"
	"terminal-terrace/sse-wiki/internal/database"
	"terminal-terrace/sse-wiki/internal/route"
)

func main() {
	// 1. 加载配置
	config.MustLoad("config.yaml")

	// 2. 初始化数据库
	database.InitDatabase()

	// 3. 设置路由
	r := route.SetupRouter()

	// 4. 启动服务
	r.Run(":8080")
}