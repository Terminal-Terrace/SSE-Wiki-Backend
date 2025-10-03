.PHONY: help run build clean test

install:
	@echo "安装依赖"
	go mod tidy

run:
	@echo "运行sse-wiki服务"
	go run cmd/server/main.go