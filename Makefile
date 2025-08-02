# Makefile for crypto-info project

# 项目信息
PROJECT_NAME := crypto-info
VERSION := $(shell git describe --tags --always --dirty)
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse HEAD)

# Go相关变量
GO_VERSION := 1.21
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 目录
BIN_DIR := bin
CMD_DIR := cmd
INTERNAL_DIR := internal
API_DIR := api
CONFIGS_DIR := configs
DEPLOYMENTS_DIR := deployments

# Docker相关
DOCKER_IMAGE := $(PROJECT_NAME):$(VERSION)
DOCKER_REGISTRY := your-registry.com

.PHONY: all build clean test lint fmt vet deps docker-build docker-push deploy help

# 默认目标
all: clean fmt vet test build

# 构建
build:
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME) ./$(CMD_DIR)/server

# 构建所有平台
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-linux-amd64 ./$(CMD_DIR)/server
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-darwin-amd64 ./$(CMD_DIR)/server
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME)-windows-amd64.exe ./$(CMD_DIR)/server

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@go clean

# 测试
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 基准测试
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# 代码检查
vet:
	@echo "Running go vet..."
	go vet ./...

# 安装依赖
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# 更新依赖
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# 生成代码
generate:
	@echo "Generating code..."
	go generate ./...

# 构建Docker镜像
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) -f build/Dockerfile .

# 推送Docker镜像
docker-push: docker-build
	@echo "Pushing Docker image..."
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

# 运行
run: build
	@echo "Running $(PROJECT_NAME)..."
	./$(BIN_DIR)/$(PROJECT_NAME)

# 开发模式运行
dev:
	@echo "Running in development mode..."
	go run ./$(CMD_DIR)/server

# 安装工具
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# 生成API文档
swag:
	@echo "Generating API documentation..."
	swag init -g ./$(CMD_DIR)/server/main.go -o ./docs

# 部署到开发环境
deploy-dev:
	@echo "Deploying to development environment..."
	kubectl apply -f $(DEPLOYMENTS_DIR)/dev/

# 部署到生产环境
deploy-prod:
	@echo "Deploying to production environment..."
	kubectl apply -f $(DEPLOYMENTS_DIR)/prod/

# 显示帮助
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  build-all    - Build for all platforms"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  bench        - Run benchmarks"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Install dependencies"
	@echo "  update-deps  - Update dependencies"
	@echo "  generate     - Generate code"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-push  - Push Docker image"
	@echo "  run          - Run the application"
	@echo "  dev          - Run in development mode"
	@echo "  install-tools- Install development tools"
	@echo "  swag         - Generate API documentation"
	@echo "  deploy-dev   - Deploy to development"
	@echo "  deploy-prod  - Deploy to production"
	@echo "  help         - Show this help"