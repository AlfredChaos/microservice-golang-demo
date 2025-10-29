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
# --dir: 指定要扫描的目录（使用项目根目录，会递归扫描）
# --generalInfo: 指定包含 API 通用信息的文件路径（相对于项目根）
# --output: 输出目录
# --parseDependency: 解析依赖
# --parseInternal: 解析内部包（解析 internal 目录）
swag init \
    --dir ./ \
    --generalInfo ./cmd/api-gateway/main.go \
    --output ./docs \
    --parseDependency \
    --parseInternal 2>&1 | grep -v "warning: failed to get package name" | grep -v "warning: failed to evaluate const"

echo "Swagger documentation generation complete!"
echo "Documentation available at: ./docs/swagger.json"
