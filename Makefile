BUILD_DIR=build
BINARY_NAME=bili-parse-api
MAIN_FILE=cmd/server/main.go

.PHONY: build clean run deps test install-deps docker-build docker-run

# 构建应用
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# 清理构建文件
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f data.db

# 运行应用
run:
	@echo "Running $(BINARY_NAME)..."
	@go run $(MAIN_FILE)

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 安装系统依赖 (Ubuntu/Debian)
install-deps:
	@echo "Installing system dependencies..."
	@sudo apt update
	@sudo apt install -y ffmpeg

# 安装系统依赖 (CentOS/RHEL)
install-deps-centos:
	@echo "Installing system dependencies for CentOS/RHEL..."
	@sudo yum install -y epel-release
	@sudo yum install -y ffmpeg

# 构建Docker镜像
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):latest .

# 运行Docker容器
docker-run:
	@echo "Running Docker container..."
	@docker-compose up -d

# 停止Docker容器
docker-stop:
	@echo "Stopping Docker container..."
	@docker-compose down

# 查看Docker日志
docker-logs:
	@docker-compose logs -f

# 开发环境设置
dev-setup: install-deps deps
	@echo "Setting up development environment..."
	@mkdir -p parse_cache logs
	@cp configs/config.example.yaml configs/config.yaml
	@echo "Development environment setup complete!"

# 生产环境构建
prod-build:
	@echo "Building for production..."
	@CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  clean              - Clean build files"
	@echo "  run                - Run the application"
	@echo "  deps               - Download Go dependencies"
	@echo "  test               - Run tests"
	@echo "  install-deps       - Install system dependencies (Ubuntu/Debian)"
	@echo "  install-deps-centos- Install system dependencies (CentOS/RHEL)"
	@echo "  docker-build       - Build Docker image"
	@echo "  docker-run         - Run with Docker Compose"
	@echo "  docker-stop        - Stop Docker containers"
	@echo "  docker-logs        - View Docker logs"
	@echo "  dev-setup          - Setup development environment"
	@echo "  prod-build         - Build for production"
	@echo "  help               - Show this help message"
