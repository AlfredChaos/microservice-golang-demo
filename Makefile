.PHONY: proto swagger build clean run-gateway run-user run-book run-nice

# 项目配置
PROJECT_NAME=demo
BUILD_DIR=build
API_DIR=api
SCRIPTS_DIR=scripts

# 服务列表
SERVICES=api-gateway user-service book-service nice-service

# 生成 protobuf 代码
proto:
	@echo "Generating protobuf code..."
	@chmod +x $(SCRIPTS_DIR)/gen-proto.sh
	@$(SCRIPTS_DIR)/gen-proto.sh

# 生成 swagger 文档
swagger:
	@echo "Generating swagger documentation..."
	@chmod +x $(SCRIPTS_DIR)/gen-swagger.sh
	@$(SCRIPTS_DIR)/gen-swagger.sh

# 编译所有服务（自动生成 Swagger 文档和 protobuf 代码）
build: swagger proto
	@echo "Building all services..."
	@mkdir -p $(BUILD_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		go build -o $(BUILD_DIR)/$$service ./cmd/$$service; \
	done
	@echo "Build complete!"

# 编译单个服务
build-%:
	@echo "Building $*..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$* ./cmd/$*

# 运行 api-gateway
run-gateway: build-api-gateway
	@echo "Starting api-gateway..."
	@$(BUILD_DIR)/api-gateway

# 运行 user-service
run-user: build-user-service
	@echo "Starting user-service..."
	@$(BUILD_DIR)/user-service

# 运行 book-service
run-book: build-book-service
	@echo "Starting book-service..."
	@$(BUILD_DIR)/book-service

# 运行 nice-service
run-nice: build-nice-service
	@echo "Starting nice-service..."
	@$(BUILD_DIR)/nice-service

# 清理编译产物
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf docs/swagger.json docs/swagger.yaml docs/docs.go
	@echo "Clean complete!"

# 安装开发工具
install-tools:
	@echo "Installing development tools..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Tools installed!"

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies ready!"

# 运行所有服务 (需要在不同终端运行)
help:
	@echo "Available commands:"
	@echo "  make proto          - Generate protobuf code"
	@echo "  make swagger        - Generate swagger documentation"
	@echo "  make build          - Build all services (auto-generate docs & proto)"
	@echo "  make build-<name>   - Build specific service"
	@echo "  make run-gateway    - Run api-gateway"
	@echo "  make run-user       - Run user-service"
	@echo "  make run-book       - Run book-service"
	@echo "  make run-nice       - Run nice-service"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install-tools  - Install development tools"
	@echo "  make deps           - Download dependencies"
