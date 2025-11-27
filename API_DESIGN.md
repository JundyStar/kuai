# Kuai API 设计 - 前后端分离方案

## 方案概述

**核心思路**：Go 项目只提供 CLI 接口，前端通过调用 shell 脚本执行 `kuai` 命令来交互。

## 方案一：Shell 脚本 + JSON 输出（推荐）

### 1. 为 CLI 添加 JSON 输出支持

在 CLI 命令中添加 `--json` 或 `--output json` 标志，输出 JSON 格式数据。

#### 示例：列表模板
```bash
# 普通输出
kuai template list
# NAME    DESCRIPTION
# go-service    Go 微服务模板

# JSON 输出
kuai template list --json
# [{"name":"go-service","description":"Go 微服务模板"}]
```

#### 示例：获取模板详情
```bash
# 新增命令：获取模板 manifest
kuai template show go-service --json
# {
#   "name": "go-service",
#   "description": "Go 微服务模板",
#   "fields": [
#     {"name": "Name", "prompt": "服务名称", "default": "demo-service", "required": true}
#   ]
# }
```

#### 示例：生成项目
```bash
# 使用 JSON 文件作为输入
kuai use go-service ./output --values values.json --defaults
```

### 2. 前端调用方式

#### 方式 A：直接调用 shell 脚本（Node.js/后端）
```javascript
// Node.js 后端 API
const { exec } = require('child_process');
const util = require('util');
const execPromise = util.promisify(exec);

// 获取模板列表
app.get('/api/templates', async (req, res) => {
  try {
    const { stdout } = await execPromise('kuai template list --json');
    const templates = JSON.parse(stdout);
    res.json(templates);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 生成项目
app.post('/api/generate', async (req, res) => {
  const { templateName, values, outputDir } = req.body;
  
  // 创建临时 JSON 文件
  const valuesFile = `/tmp/values-${Date.now()}.json`;
  fs.writeFileSync(valuesFile, JSON.stringify(values));
  
  try {
    await execPromise(`kuai use ${templateName} ${outputDir} --values ${valuesFile} --defaults --force`);
    
    // 打包成 ZIP
    await execPromise(`cd ${outputDir} && zip -r output.zip .`);
    
    // 返回 ZIP 文件
    res.download(`${outputDir}/output.zip`);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});
```

#### 方式 B：通过 Shell 脚本封装（更安全）
```bash
#!/bin/bash
# api.sh - 封装 kuai 命令的 API

case "$1" in
  list)
    kuai template list --json
    ;;
  show)
    kuai template show "$2" --json
    ;;
  generate)
    TEMPLATE=$2
    OUTPUT=$3
    VALUES=$4
    
    # 创建临时 JSON 文件
    TMPFILE=$(mktemp)
    echo "$VALUES" > "$TMPFILE"
    
    kuai use "$TEMPLATE" "$OUTPUT" --values "$TMPFILE" --defaults --force
    
    # 清理
    rm "$TMPFILE"
    ;;
  *)
    echo "Unknown command: $1"
    exit 1
    ;;
esac
```

前端调用：
```javascript
// 通过 HTTP 请求调用后端 API
fetch('/api/templates')
  .then(res => res.json())
  .then(data => console.log(data));
```

### 3. 需要添加的 CLI 功能

#### 3.1 JSON 输出支持
```go
// pkg/cmd/template_list.go
var jsonOutput bool

cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")

if jsonOutput {
    data, _ := json.Marshal(templates)
    fmt.Println(string(data))
} else {
    // 原有的表格输出
}
```

#### 3.2 获取模板详情命令
```go
// pkg/cmd/template_show.go
func newTemplateShowCmd() *cobra.Command {
    var jsonOutput bool
    
    cmd := &cobra.Command{
        Use:   "show <name>",
        Short: "显示模板详情",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            name := args[0]
            templatePath, err := templateMgr.TemplatePath(name)
            if err != nil {
                return err
            }
            
            manifest, _, err := templates.LoadManifest(templatePath)
            if err != nil {
                return err
            }
            
            if jsonOutput {
                data, _ := json.Marshal(manifest)
                fmt.Println(string(data))
            } else {
                // 文本输出
            }
            return nil
        },
    }
    
    cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
    return cmd
}
```

#### 3.3 上传模板（通过文件路径）
```bash
# 前端先上传 ZIP 到服务器，然后调用
kuai template add my-template --from /path/to/extracted --force
```

## 方案二：简单的 HTTP API 服务（备选）

如果不想让前端直接调用 shell，可以创建一个简单的 HTTP API 服务：

```go
// pkg/api/server.go - 简单的 API 服务器
package api

import (
    "encoding/json"
    "net/http"
    "os/exec"
)

func StartAPIServer(port string) {
    http.HandleFunc("/api/templates", handleListTemplates)
    http.HandleFunc("/api/templates/show", handleShowTemplate)
    http.HandleFunc("/api/generate", handleGenerate)
    http.ListenAndServe(":"+port, nil)
}

func handleListTemplates(w http.ResponseWriter, r *http.Request) {
    cmd := exec.Command("kuai", "template", "list", "--json")
    output, _ := cmd.Output()
    w.Header().Set("Content-Type", "application/json")
    w.Write(output)
}
```

## 推荐方案对比

| 方案 | 优点 | 缺点 | 适用场景 |
|------|------|------|----------|
| **方案一：Shell + JSON** | ✅ Go 项目保持简洁<br>✅ 前端完全独立<br>✅ 易于调试 | ⚠️ 需要添加 JSON 输出 | **推荐**：前后端完全分离 |
| **方案二：HTTP API** | ✅ 统一接口<br>✅ 更安全 | ❌ 需要在 Go 项目中维护 API 代码 | 需要统一管理时 |

## 实施步骤

### 步骤 1：添加 JSON 输出支持
1. 为 `template list` 添加 `--json` 标志
2. 新增 `template show <name> --json` 命令
3. 确保 `use` 命令支持 `--values` JSON 文件输入

### 步骤 2：前端开发
1. 前端同事开发独立的 Web 界面
2. 通过 Node.js/Python 等后端调用 `kuai` 命令
3. 或直接通过 shell 脚本封装调用

### 步骤 3：部署
1. Go 项目编译成二进制
2. 前端项目独立部署
3. 前端后端通过 shell 脚本或 HTTP API 通信

## 示例：完整的前后端交互流程

### 1. 前端获取模板列表
```bash
# 后端执行
kuai template list --json
# 返回
[{"name":"go-service","description":"Go 微服务模板"}]
```

### 2. 前端获取模板详情
```bash
# 后端执行
kuai template show go-service --json
# 返回
{
  "name": "go-service",
  "description": "Go 微服务模板",
  "fields": [
    {"name": "Name", "prompt": "服务名称", "default": "demo-service", "required": true},
    {"name": "Port", "prompt": "端口", "default": "8080", "required": false}
  ]
}
```

### 3. 前端提交生成请求
```bash
# 前端提交 JSON
{
  "templateName": "go-service",
  "values": {
    "Name": "my-service",
    "Port": "3000"
  },
  "outputDir": "/tmp/output"
}

# 后端执行
echo '{"Name":"my-service","Port":"3000"}' > /tmp/values.json
kuai use go-service /tmp/output --values /tmp/values.json --defaults --force

# 打包
cd /tmp/output && zip -r project.zip .
```

## 总结

**推荐方案**：方案一（Shell + JSON）
- ✅ Go 项目保持简洁，只提供 CLI
- ✅ 前端完全独立开发
- ✅ 通过 JSON 格式通信
- ✅ 易于测试和调试

**需要做的**：
1. 为 CLI 命令添加 `--json` 输出支持
2. 新增 `template show` 命令获取模板详情
3. 前端通过 shell 脚本或后端 API 调用 `kuai` 命令

