## Kuai

- **简洁易用**：`kuai template add`/`list`/`remove` + `kuai use` 四个命令涵盖模板管理和实例化。
- **智能扫描**：**无需配置文件**，自动扫描模板中的 `{{变量名}}` 并生成交互式提示。
- **高效稳定**：Go 单一二进制，渲染使用 `text/template`，路径/文件均可使用变量。
- **高度兼容**：支持 macOS、Linux、Windows；可选 manifest 使用 YAML/JSON。

### 安装

#### macOS / Linux

```bash
# 从源码编译安装
git clone https://github.com/JundyStar/kuai.git
cd kuai
go install ./...
```

#### Windows

```powershell
# 从源码编译安装（需要先安装 Go）
git clone https://github.com/JundyStar/kuai.git
cd kuai
go install ./...

# 将 %USERPROFILE%\go\bin 添加到 PATH 环境变量
# 或在 PowerShell 中临时添加：
$env:Path += ";$env:USERPROFILE\go\bin"
```

### 快速开始

#### 命令行方式

```bash
# 添加模板（从本地目录）
kuai template add my-go-service --from /path/to/template
kuai template list
kuai use my-go-service ./demo-service
```

#### Web 界面方式

```bash
# 启动 Web 服务器
kuai web

# 或指定端口
kuai web -p 3000

# 然后在浏览器中访问 http://localhost:8080
```

Web 界面功能：
- 📋 查看所有可用模板
- 📤 上传新模板（ZIP 格式，支持拖拽上传）
- 🎨 可视化填写模板变量
- ⬇️ 一键生成并下载项目
- 🔍 搜索和预览模板

**Windows 路径示例：**
```powershell
kuai template add my-go-service --from C:\path\to\template
kuai use my-go-service .\demo-service
```

### 自动扫描变量

**无需任何配置文件**！Kuai 会自动扫描模板中的所有 `{{变量名}}`，并为每个变量生成友好的交互式提示。

例如，如果模板中有 `{{Name}}`、`{{Port}}` 等变量，运行 `kuai use` 时会自动提示：

```
? Name [default: ]: 
? Port [default: ]: 
```

### 可选：自定义配置

如果想提供更详细的提示、默认值或描述，可以在模板根目录添加 `kuai.yaml`：

```yaml
name: go-service
description: 基本 Go 服务脚手架
fields:
  - name: ServiceName
    prompt: 服务名
    default: demo
    required: true
  - name: Port
    prompt: 监听端口
    default: "8080"
```

