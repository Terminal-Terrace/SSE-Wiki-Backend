.PHONY: help run build clean test docs sync gen

help:
	@echo "可用命令:"
	@echo "  make install    安装依赖"
	@echo "  make run        运行auth-service服务"
	@echo "  make sync       从GitHub同步proto文件（根据 pb.config.json）"
	@echo "  make gen        生成proto的Go代码（生成 .pb.go 和 _grpc.pb.go）"
	@echo ""
	@echo "工作流程："
	@echo "  1. 编辑 pb.config.json 配置 GitHub 仓库信息"
	@echo "  2. make sync    # 下载最新的 proto 文件"
	@echo "  3. make gen     # 生成 Go 代码"
	@echo "  4. make run     # 运行服务"

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

sync:
	@echo "从GitHub同步proto文件"
	@which jq > /dev/null || (echo "错误: 未安装 jq，请先安装: sudo apt install jq 或 brew install jq" && exit 1)
	@if [ ! -f pb.config.json ]; then echo "错误: pb.config.json 不存在" && exit 1; fi
	@echo "读取配置文件..."
	@OWNER=$$(jq -r '.github.owner' pb.config.json); \
	REPO=$$(jq -r '.github.repo' pb.config.json); \
	BRANCH=$$(jq -r '.github.branch' pb.config.json); \
	if [ "$$OWNER" = "YOUR_GITHUB_OWNER" ] || [ "$$OWNER" = "null" ]; then \
		echo "错误: 请先在 pb.config.json 中配置 GitHub 仓库信息"; \
		echo "示例: { \"github\": { \"owner\": \"your-org\", \"repo\": \"SSE-Wiki-proto\", \"branch\": \"main\" } }"; \
		exit 1; \
	fi; \
	echo "GitHub 仓库: $$OWNER/$$REPO (branch: $$BRANCH)"; \
	if [ -n "$$GITHUB_TOKEN" ]; then \
		echo "检测到 GITHUB_TOKEN，将使用认证访问（适用于私有仓库）"; \
		AUTH_HEADER="-H \"Authorization: token $$GITHUB_TOKEN\""; \
	else \
		echo "未检测到 GITHUB_TOKEN，将使用无认证访问（仅适用于公开仓库）"; \
		echo "提示: 如需访问私有仓库，请设置环境变量 GITHUB_TOKEN"; \
		AUTH_HEADER=""; \
	fi; \
	mkdir -p protobuf; \
	jq -r '.services[]' pb.config.json | while read service; do \
		echo "同步服务: $$service"; \
		SERVICE_DIR="protobuf/$$service"; \
		mkdir -p "$$SERVICE_DIR"; \
		PROTO_URL="https://raw.githubusercontent.com/$$OWNER/$$REPO/$$BRANCH/$$service/$$service.proto"; \
		echo "  下载: $$PROTO_URL"; \
		if [ -n "$$GITHUB_TOKEN" ]; then \
			if curl -fsSL -H "Authorization: token $$GITHUB_TOKEN" "$$PROTO_URL" -o "$$SERVICE_DIR/$$service.proto"; then \
				echo "  ✓ $$service.proto 同步成功"; \
			else \
				echo "  ✗ 下载失败，请检查 URL 或 GITHUB_TOKEN 是否正确"; \
				exit 1; \
			fi; \
		else \
			if curl -fsSL "$$PROTO_URL" -o "$$SERVICE_DIR/$$service.proto"; then \
				echo "  ✓ $$service.proto 同步成功"; \
			else \
				echo "  ✗ 下载失败，请检查 URL 是否正确，或设置 GITHUB_TOKEN（如为私有仓库）"; \
				exit 1; \
			fi; \
		fi; \
	done; \
	echo "所有 proto 文件同步完成！"

gen:
	@echo "生成proto的Go代码"
	@which protoc > /dev/null || (echo "错误: 未安装 protoc，请先安装 Protocol Buffers 编译器" && exit 1)
	@which protoc-gen-go > /dev/null || (echo "未安装 protoc-gen-go，正在安装..." && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
	@which protoc-gen-go-grpc > /dev/null || (echo "未安装 protoc-gen-go-grpc，正在安装..." && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest)
	@mkdir -p pkg/pb
	@if [ ! -d protobuf ]; then echo "错误: protobuf 目录不存在，请先运行 make sync" && exit 1; fi
	@echo "生成 gRPC 代码..."
	@find protobuf -name "*.proto" | while read proto; do \
		echo "  处理: $$proto"; \
		protoc --go_out=pkg/pb --go_opt=paths=source_relative \
			--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
			-I protobuf \
			$$proto; \
	done
	@echo "代码生成完成，输出目录: pkg/pb"
