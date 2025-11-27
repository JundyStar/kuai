# Kuai 运行要求

## ✅ 编译后可直接使用（零依赖）

Go 编译后的二进制文件是**完全自包含的**，所有 Go 依赖库都已打包在二进制文件中。

**核心答案**：**是的，编译后其他人可以直接使用，无需任何额外安装！**

## 📋 运行时要求

### 必需条件（仅 2 项）

#### 1. 文件系统权限
- 需要读写权限来创建模板目录：`~/.kuai/templates/`
- 需要临时目录权限（系统 `/tmp` 目录）

#### 2. 网络端口（仅 Web 界面需要）
- Web 服务器需要监听端口（默认 8080）
- 如果部署到服务器，需要开放相应端口

### ✅ 不需要的条件

- ❌ **不需要 Go 环境**（所有依赖已打包）
- ❌ **不需要 Git**（已移除 Git 功能，只支持 ZIP 上传）
- ❌ **不需要其他外部依赖**

## 🚀 使用场景

### 场景 1：命令行使用
**要求**：仅需二进制文件 + 文件系统权限
```bash
./kuai template add my-template --from /path/to/template
./kuai use my-template ./output
```

### 场景 2：Web 界面
**要求**：二进制文件 + 网络端口
```bash
./kuai web --host 0.0.0.0 --port 8080
```

## 📦 分发方式

### 方式一：提供预编译二进制（推荐）
```bash
# 编译不同平台的版本
GOOS=linux GOARCH=amd64 go build -o kuai-linux-amd64 ./main.go
GOOS=linux GOARCH=arm64 go build -o kuai-linux-arm64 ./main.go
GOOS=darwin GOARCH=amd64 go build -o kuai-darwin-amd64 ./main.go
GOOS=darwin GOARCH=arm64 go build -o kuai-darwin-arm64 ./main.go
GOOS=windows GOARCH=amd64 go build -o kuai-windows-amd64.exe ./main.go
```

### 方式二：提供安装脚本
```bash
#!/bin/bash
# install.sh
curl -L https://github.com/jundy/kuai/releases/latest/download/kuai-linux-amd64 -o /usr/local/bin/kuai
chmod +x /usr/local/bin/kuai
```

## ✅ 总结

**核心答案**：
- ✅ **是的，编译后其他人可以直接使用**
- ✅ **不需要 Go 环境**（所有依赖已打包）
- ✅ **不需要 Git**（只支持 ZIP 文件上传）
- ✅ **不需要任何外部依赖**（完全自包含）

**最简单的部署**：
1. 下载对应平台的二进制文件
2. 赋予执行权限：`chmod +x kuai`
3. 直接运行：`./kuai web`

## 🔍 验证方法

```bash
# 1. 检查二进制是否独立（Linux）
ldd kuai  # 应该显示很少或没有外部依赖

# 2. 测试运行
./kuai --help

# 3. 测试 Web 界面
./kuai web --host 0.0.0.0 --port 8080
# 然后在浏览器访问 http://localhost:8080
```

