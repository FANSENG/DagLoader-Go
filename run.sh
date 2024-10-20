#!/bin/bash

# 设置Go模块
export GO111MODULE=on

# 确保我们在正确的目录
cd "$(dirname "$0")"

# 获取依赖
go mod tidy

# 编译并运行代码
go run .