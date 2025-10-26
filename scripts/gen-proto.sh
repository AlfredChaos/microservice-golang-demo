#!/bin/bash

# 生成 protobuf 代码的脚本

set -e

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
API_DIR="$PROJECT_ROOT/api"

echo "Generating protobuf code..."

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Please install protoc: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Error: protoc-gen-go is not installed"
    echo "Please run: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# 检查 protoc-gen-go-grpc 是否安装
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Error: protoc-gen-go-grpc is not installed"
    echo "Please run: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# 遍历所有 proto 文件并生成代码
find "$API_DIR" -name "*.proto" | while read -r proto_file; do
    echo "Processing: $proto_file"
    
    # 获取 proto 文件所在目录
    proto_dir=$(dirname "$proto_file")
    
    # 生成 Go 代码
    protoc \
        --proto_path="$API_DIR" \
        --go_out="$API_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$API_DIR" \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"
done

echo "Protobuf code generation complete!"
