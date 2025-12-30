.PHONY: help run build clean test test-cover test-integration test-race

help:
	@echo "可用命令:"
	@echo "  make install        安装依赖"
	@echo "  make run            运行sse-wiki服务"
	@echo "  make test           运行单元测试"
	@echo "  make test-cover     运行测试并生成覆盖率报告"
	@echo "  make test-integration 运行集成测试"
	@echo "  make test-race      运行测试并检测数据竞争"

install:
	@echo "安装依赖"
	go mod tidy

run:
	@echo "运行sse-wiki服务"
	go run cmd/server/main.go

test:
	@echo "运行单元测试..."
	go test -v ./...

test-cover:
	@echo "运行测试并生成覆盖率报告..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

test-integration:
	@echo "运行集成测试..."
	go test -v -tags=integration ./...

test-race:
	@echo "运行测试并检测数据竞争..."
	go test -race -v ./...