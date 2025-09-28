# Go参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
WIRE=wire

# 项目信息
PROJECT_NAME=go-nexus
BINARY_NAME=go-nexus
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd/server

# 版本信息
VERSION?=v0.1.0
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)

# Docker参数
DOCKER_IMAGE=laolishu/$(PROJECT_NAME)
DOCKER_TAG?=$(VERSION)

# 默认目标
.PHONY: all
all: clean wire-gen build

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - 编译项目"
	@echo "  build-all      - 交叉编译所有平台"
	@echo "  clean          - 清理构建文件"
	@echo "  test           - 运行单元测试"
	@echo "  test-unit      - 运行单元测试（短模式）"
	@echo "  test-integration - 运行集成测试"
	@echo "  test-e2e       - 运行端到端测试"
	@echo "  coverage       - 生成测试覆盖率报告"
	@echo "  benchmark      - 运行性能测试"
	@echo "  race           - 运行竞态检测"
	@echo "  wire-gen       - 生成Wire代码"
	@echo "  lint           - 代码静态检查"
	@echo "  fmt            - 格式化代码"
	@echo "  vet            - 运行go vet"
	@echo "  deps           - 下载依赖"
	@echo "  tidy           - 整理依赖"
	@echo "  docker-build   - 构建Docker镜像"
	@echo "  docker-push    - 推送Docker镜像"
	@echo "  release        - 创建发布版本"
	@echo "  run            - 运行开发服务器"
	@echo "  install-tools  - 安装开发工具"

# 编译
.PHONY: build
build: wire-gen
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_PATH)"

# 交叉编译
.PHONY: build-all
build-all: clean wire-gen
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p bin
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "Cross-compilation complete"

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Wire代码生成
.PHONY: wire-gen
wire-gen:
	@echo "Generating Wire code..."
	@cd $(MAIN_PATH) && $(WIRE)
	@echo "Wire generation complete"

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./test/integration/...

.PHONY: test-e2e
test-e2e:
	@echo "Running e2e tests..."
	$(GOTEST) -v -tags=e2e ./test/e2e/...

# 测试覆盖率
.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 性能测试
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# 竞态检测
.PHONY: race
race:
	@echo "Running race detection..."
	$(GOTEST) -race ./...

# 代码检查
.PHONY: lint
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

# 依赖管理
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Docker
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

.PHONY: docker-push
docker-push: docker-build
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	@echo "Docker image pushed"

# 发布
.PHONY: release
release: clean build-all
	@echo "Creating release $(VERSION)..."
	@mkdir -p release
	@cp bin/* release/
	@cd release && find . -type f -exec sha256sum {} \; > checksums.txt
	@echo "Release $(VERSION) ready in release/ directory"

# 开发运行
.PHONY: run
run: build
	@echo "Starting development server..."
	$(BINARY_PATH) server --config configs/config-dev.yaml

.PHONY: run-prod
run-prod: build
	@echo "Starting production server..."
	$(BINARY_PATH) server --config configs/config-prod.yaml

# 工具安装
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed"

# 数据库迁移
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	$(BINARY_PATH) migrate --config configs/config-dev.yaml --direction up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	$(BINARY_PATH) migrate --config configs/config-dev.yaml --direction down

# API文档生成
.PHONY: docs
docs:
	@echo "Generating API documentation..."
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/server/main.go -o api/docs
	@echo "API documentation generated in api/docs/"

# 代码质量检查
.PHONY: quality
quality: fmt vet lint test-unit
	@echo "Code quality check complete"

# CI流程
.PHONY: ci
ci: deps wire-gen quality test coverage
	@echo "CI pipeline complete"

# 本地开发环境设置
.PHONY: setup
setup: install-tools deps wire-gen
	@echo "Development environment setup complete"
	@echo "Run 'make run' to start the development server"

# 插件相关
.PHONY: build-plugins
build-plugins:
	@echo "Building plugins..."
	@for plugin in plugins/*/; do \
		if [ -d "$$plugin" ]; then \
			echo "Building plugin: $$plugin"; \
			cd "$$plugin" && go build -buildmode=plugin -o "../$$(basename $$plugin).so" . && cd -; \
		fi \
	done
	@echo "Plugins built successfully"

.PHONY: test-plugins
test-plugins:
	@echo "Testing plugins..."
	@for plugin in plugins/*/; do \
		if [ -d "$$plugin" ]; then \
			echo "Testing plugin: $$plugin"; \
			cd "$$plugin" && go test -v . && cd -; \
		fi \
	done

# 性能分析
.PHONY: profile-cpu
profile-cpu:
	@echo "Starting CPU profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

.PHONY: profile-mem
profile-mem:
	@echo "Starting memory profiling..."
	$(GOTEST) -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof