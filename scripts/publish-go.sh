#!/bin/bash

# Go 模块发布脚本
set -e

echo "🚀 发布 Go 模块..."

# 检查是否有未提交的更改
if ! git diff --quiet; then
    echo "❌ 有未提交的更改，请先提交代码"
    exit 1
fi

# 获取当前版本
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
echo "当前版本: $CURRENT_VERSION"

# 询问新版本
read -p "输入新版本号 (例如 v1.0.0): " NEW_VERSION

if [ -z "$NEW_VERSION" ]; then
    echo "❌ 版本号不能为空"
    exit 1
fi

# 验证版本号格式
if ! [[ $NEW_VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "❌ 版本号格式不正确，应该是 vX.Y.Z 格式"
    exit 1
fi

# 运行测试
echo "🧪 运行测试..."
go test ./...

# 创建 git tag
echo "🏷️  创建 git tag: $NEW_VERSION"
git tag $NEW_VERSION

# 推送 tag
echo "📤 推送 tag 到远程仓库..."
git push origin $NEW_VERSION

# 创建 GitHub release（如果安装了 gh CLI）
if command -v gh &> /dev/null; then
    echo "📝 创建 GitHub release..."
    gh release create $NEW_VERSION --generate-notes
else
    echo "💡 提示: 安装 GitHub CLI (gh) 可以自动创建 release"
fi

echo "✅ Go 模块发布完成!"
echo "📦 模块地址: github.com/yourusername/keymaster-multisig-pool@$NEW_VERSION" 