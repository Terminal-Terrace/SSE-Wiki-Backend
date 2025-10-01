.PHONY: help run build clean test install

help:
	@echo "用法:"
	@echo "  make run <子包名>    - 运行一个服务"
	@echo "  make build <子包名>  - 构建一个服务"
	@echo "  make clean            - 清理所有构建产物"

install:
	@if [ -d "services/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C services/$(filter-out $@,$(MAKECMDGOALS)) install; \
	elif [ -d "packages/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C packages/$(filter-out $@,$(MAKECMDGOALS)) install; \
	else \
		echo "子包'$(filter-out $@,$(MAKECMDGOALS))'不存在"; \
		exit 1; \
	fi

run:
	@if [ -d "services/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C services/$(filter-out $@,$(MAKECMDGOALS)) run; \
	elif [ -d "packages/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C packages/$(filter-out $@,$(MAKECMDGOALS)) run; \
	else \
		echo "子包'$(filter-out $@,$(MAKECMDGOALS))'不存在"; \
		exit 1; \
	fi

build:
	@if [ -d "services/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C services/$(filter-out $@,$(MAKECMDGOALS)) build; \
	elif [ -d "packages/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C packages/$(filter-out $@,$(MAKECMDGOALS)) build; \
	else \
		echo "子包'$(filter-out $@,$(MAKECMDGOALS))'不存在"; \
		exit 1; \
	fi

clean:
	@find . -name "*.out" -type f -delete
	@find . -name "bin" -type d -exec rm -rf {} +
	@echo "Cleaned all build artifacts"

# This allows passing arguments to make
%:
	@: