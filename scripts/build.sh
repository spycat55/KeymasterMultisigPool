#!/bin/bash

# 构建脚本
set -e

echo "🔨 开始构建项目..."

# 创建构建目录
mkdir -p build dist

# 清理之前的构建
echo "🧹 清理之前的构建..."
rm -rf build/* dist/*

# 构建 TypeScript
echo "📦 构建 TypeScript 项目..."
npm run build

# 构建 Go 项目
echo "🚀 构建 Go 项目..."
echo "构建服务器..."
go build -o build/server ./cmd/server

# 如果有其他 Go 程序，在这里添加
# go build -o build/cli ./cmd/cli

echo "✅ 构建完成!"
echo "📁 TypeScript 输出: dist/"
echo "📁 Go 输出: build/" 