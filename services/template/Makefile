.PHONY: install run build clean test

install:
	@echo "安装依赖"
	go mod tidy

run:
	@echo "运行 template 服务"
	go run cmd/server/main.go

build:
	@echo "构建 template 服务"
	go build -o bin/template cmd/server/main.go

clean:
	@echo "清理构建产物"
	rm -rf bin/

test:
	@echo "运行测试"
	go test -v ./...

test-cover:
	@echo "运行测试并生成覆盖率报告"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
