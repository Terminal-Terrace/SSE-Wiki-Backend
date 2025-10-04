.PHONY: help run build clean test docs

help:
	@echo "可用命令:"
	@echo "  make install    安装依赖"
	@echo "  make run        运行auth-service服务"
	@echo "  make docs       生成Swagger文档"

install:
	@echo "安装依赖"
	go mod tidy

run:
	@echo "运行auth-service服务"
	go run cmd/server/main.go

docs:
	@echo "生成Swagger文档"
	@which swag > /dev/null || (echo "未安装 swag，正在安装..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@if [ -f ~/go/bin/swag ]; then ~/go/bin/swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal; else swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal; fi
	