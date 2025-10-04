package main

import (
	"terminal-terrace/auth-service/internal/route"
	"terminal-terrace/auth-service/config"
	"terminal-terrace/auth-service/internal/database"
)

func main() {
	config.MustLoad("config.yaml")
	database.InitDatabase()  // 添加数据库初始化
	r := route.SetupRouter()

	r.Run(":8081")
}