#!/bin/bash
# 使用 garble 混淆编译的脚本

echo "Building with garble obfuscation..."

# 检查 garble 是否安装
if ! command -v garble &> /dev/null; then
    echo "Error: garble is not installed"
    echo "Install it with: go install mvdan.cc/garble@latest"
    exit 1
fi

# 使用 garble 编译
garble build -o app main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Binary: ./app"
    echo "Note: The binary is obfuscated, making reverse engineering harder."
else
    echo "❌ Build failed"
    exit 1
fi

