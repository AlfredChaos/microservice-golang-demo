#!/bin/bash

# 生成 Swagger 文档的脚本

set -e

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)

echo "Generating swagger documentation..."

# 检查 swag 是否安装
if ! command -v swag &> /dev/null; then
    echo "Error: swag is not installed"
    echo "Please run: go install github.com/swaggo/swag/cmd/swag@latest"
    exit 1
fi

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 生成 swagger 文档
# --parseDependency: 解析依赖
# --parseInternal: 解析内部包
# --generalInfo: 指定包含 API 通用信息的文件路径
swag init \
    --dir ./cmd/api-gateway \
    --generalInfo main.go \
    --output ./docs \
    --parseDependency \
    --parseInternal

echo "Swagger documentation generation complete!"
echo "Documentation available at: ./docs/swagger.json"
