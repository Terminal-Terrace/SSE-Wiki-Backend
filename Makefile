.PHONY: help run build clean test install

help:
	@echo "methods:"
	@echo "  make run <sub package name>    - run a services"
	@echo "  make build <sub package name>  - build a services"
	@echo "  make clean            - Clean up intermediate build files"

install:
	@if [ -d "services/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C services/$(filter-out $@,$(MAKECMDGOALS)) install; \
	elif [ -d "packages/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C packages/$(filter-out $@,$(MAKECMDGOALS)) install; \
	else \
		echo "Sub package '$(filter-out $@,$(MAKECMDGOALS))'not exist; \
		exit 1; \
	fi

run:
	@if [ -d "services/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C services/$(filter-out $@,$(MAKECMDGOALS)) run; \
	elif [ -d "packages/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C packages/$(filter-out $@,$(MAKECMDGOALS)) run; \
	else \
		echo "Sub package '$(filter-out $@,$(MAKECMDGOALS))' not exist"; \
		exit 1; \
	fi

build:
	@if [ -d "services/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C services/$(filter-out $@,$(MAKECMDGOALS)) build; \
	elif [ -d "packages/$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		$(MAKE) -C packages/$(filter-out $@,$(MAKECMDGOALS)) build; \
	else \
		echo "Sub package '$(filter-out $@,$(MAKECMDGOALS))' not exist"; \
		exit 1; \
	fi

clean:
	@find . -name "*.out" -type f -delete
	@find . -name "bin" -type d -exec rm -rf {} +
	@echo "Cleaned all build artifacts"

# This allows passing arguments to make
%:
	@: