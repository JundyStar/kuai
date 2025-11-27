# Kuai API 使用指南

## JSON 输出功能

现在所有模板管理命令都支持 `--json` 标志，方便前端调用。

## 命令列表

### 1. 获取模板列表

```bash
# 文本输出
kuai template list

# JSON 输出
kuai template list --json
```

**JSON 输出示例**：
```json
[
  {
    "name": "go-service",
    "description": "Go 微服务模板"
  },
  {
    "name": "web-app",
    "description": "Web 应用模板"
  }
]
```

### 2. 获取模板详情

```bash
# 文本输出
kuai template show go-service

# JSON 输出
kuai template show go-service --json
```

### 3. 查看模板目录结构

```bash
# 文本输出（树形结构）
kuai template tree go-service

# JSON 输出
kuai template tree go-service --json

# 限制深度
kuai template tree go-service --max-depth 3
```

**JSON 输出示例**：
```json
{
  "name": "go-service",
  "description": "Go 微服务模板",
  "path": "/Users/username/.kuai/templates/go-service",
  "manifest": {
    "name": "go-service",
    "description": "Go 微服务模板",
    "fields": [
      {
        "name": "Name",
        "prompt": "服务名称",
        "description": "影响 Go module、二进制名等",
        "default": "demo-service",
        "required": true
      },
      {
        "name": "Port",
        "prompt": "监听端口",
        "description": "服务监听端口号",
        "default": "8080",
        "required": false
      }
    ],
    "meta": {
      "version": ""
    }
  }
}
```

**目录结构 JSON 输出示例**：
```json
{
  "templateName": "go-service",
  "path": "/Users/username/.kuai/templates/go-service/template",
  "tree": {
    "name": "template",
    "type": "directory",
    "path": "/Users/username/.kuai/templates/go-service/template",
    "children": [
      {
        "name": "cmd",
        "type": "directory",
        "path": "/Users/username/.kuai/templates/go-service/template/cmd",
        "children": [
          {
            "name": "{{Name}}",
            "type": "directory",
            "path": "/Users/username/.kuai/templates/go-service/template/cmd/{{Name}}",
            "children": [
              {
                "name": "main.go",
                "type": "file",
                "path": "/Users/username/.kuai/templates/go-service/template/cmd/{{Name}}/main.go",
                "size": 1024
              }
            ]
          }
        ]
      },
      {
        "name": "README.md",
        "type": "file",
        "path": "/Users/username/.kuai/templates/go-service/template/README.md",
        "size": 512
      }
    ]
  }
}
```

### 4. 生成项目（使用 JSON 文件）

```bash
# 创建 values.json 文件
cat > values.json << EOF
{
  "Name": "my-service",
  "Port": "3000",
  "RepoBase": "github.com",
  "RepoGroup": "myorg"
}
EOF

# 使用 JSON 文件生成项目
kuai use go-service ./output --values values.json --defaults --force
```

**说明**：
- `--values values.json`：从 JSON 文件加载变量值
- `--defaults`：跳过交互，使用默认值（如果 JSON 中没有提供）
- `--force`：强制覆盖已存在的目录

## 前端调用示例

### Node.js 后端示例

```javascript
const { exec } = require('child_process');
const util = require('util');
const fs = require('fs');
const path = require('path');

const execPromise = util.promisify(exec);

// 1. 获取模板列表
async function getTemplates() {
  try {
    const { stdout } = await execPromise('kuai template list --json');
    return JSON.parse(stdout);
  } catch (error) {
    throw new Error(`获取模板列表失败: ${error.message}`);
  }
}

// 2. 获取模板详情
async function getTemplateInfo(templateName) {
  try {
    const { stdout } = await execPromise(`kuai template show ${templateName} --json`);
    return JSON.parse(stdout);
  } catch (error) {
    throw new Error(`获取模板详情失败: ${error.message}`);
  }
}

// 2.5. 获取模板目录结构
async function getTemplateTree(templateName, maxDepth = 0) {
  try {
    const depthFlag = maxDepth > 0 ? `--max-depth ${maxDepth}` : '';
    const { stdout } = await execPromise(`kuai template tree ${templateName} --json ${depthFlag}`);
    return JSON.parse(stdout);
  } catch (error) {
    throw new Error(`获取模板目录结构失败: ${error.message}`);
  }
}

// 3. 生成项目
async function generateProject(templateName, values, outputDir) {
  try {
    // 创建临时 JSON 文件
    const valuesFile = path.join('/tmp', `values-${Date.now()}.json`);
    fs.writeFileSync(valuesFile, JSON.stringify(values));
    
    // 执行生成命令
    await execPromise(
      `kuai use ${templateName} ${outputDir} --values ${valuesFile} --defaults --force`
    );
    
    // 清理临时文件
    fs.unlinkSync(valuesFile);
    
    return { success: true, outputDir };
  } catch (error) {
    throw new Error(`生成项目失败: ${error.message}`);
  }
}

// Express.js API 示例
const express = require('express');
const app = express();

app.use(express.json());

// GET /api/templates
app.get('/api/templates', async (req, res) => {
  try {
    const templates = await getTemplates();
    res.json(templates);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// GET /api/templates/:name
app.get('/api/templates/:name', async (req, res) => {
  try {
    const info = await getTemplateInfo(req.params.name);
    res.json(info);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// GET /api/templates/:name/tree
app.get('/api/templates/:name/tree', async (req, res) => {
  try {
    const maxDepth = parseInt(req.query.maxDepth) || 0;
    const tree = await getTemplateTree(req.params.name, maxDepth);
    res.json(tree);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// POST /api/generate
app.post('/api/generate', async (req, res) => {
  try {
    const { templateName, values, outputDir } = req.body;
    const result = await generateProject(templateName, values, outputDir);
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.listen(3000, () => {
  console.log('API server running on http://localhost:3000');
});
```

### Python 后端示例

```python
import subprocess
import json
import tempfile
import os

def get_templates():
    """获取模板列表"""
    result = subprocess.run(
        ['kuai', 'template', 'list', '--json'],
        capture_output=True,
        text=True
    )
    if result.returncode != 0:
        raise Exception(f"获取模板列表失败: {result.stderr}")
    return json.loads(result.stdout)

def get_template_info(template_name):
    """获取模板详情"""
    result = subprocess.run(
        ['kuai', 'template', 'show', template_name, '--json'],
        capture_output=True,
        text=True
    )
    if result.returncode != 0:
        raise Exception(f"获取模板详情失败: {result.stderr}")
    return json.loads(result.stdout)

def generate_project(template_name, values, output_dir):
    """生成项目"""
    # 创建临时 JSON 文件
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
        json.dump(values, f)
        values_file = f.name
    
    try:
        # 执行生成命令
        result = subprocess.run(
            ['kuai', 'use', template_name, output_dir,
             '--values', values_file, '--defaults', '--force'],
            capture_output=True,
            text=True
        )
        if result.returncode != 0:
            raise Exception(f"生成项目失败: {result.stderr}")
        return {'success': True, 'outputDir': output_dir}
    finally:
        # 清理临时文件
        os.unlink(values_file)

# Flask API 示例
from flask import Flask, jsonify, request

app = Flask(__name__)

@app.route('/api/templates', methods=['GET'])
def list_templates():
    try:
        templates = get_templates()
        return jsonify(templates)
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/templates/<name>', methods=['GET'])
def show_template(name):
    try:
        info = get_template_info(name)
        return jsonify(info)
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.route('/api/generate', methods=['POST'])
def generate():
    try:
        data = request.json
        result = generate_project(
            data['templateName'],
            data['values'],
            data['outputDir']
        )
        return jsonify(result)
    except Exception as e:
        return jsonify({'error': str(e)}), 500

if __name__ == '__main__':
    app.run(port=3000)
```

### Shell 脚本封装示例

```bash
#!/bin/bash
# kuai-api.sh - 封装 kuai 命令的简单 API

KUAI_CMD="kuai"

case "$1" in
  list)
    $KUAI_CMD template list --json
    ;;
  show)
    if [ -z "$2" ]; then
      echo '{"error": "模板名称不能为空"}' >&2
      exit 1
    fi
    $KUAI_CMD template show "$2" --json
    ;;
  generate)
    TEMPLATE=$2
    OUTPUT=$3
    VALUES=$4
    
    if [ -z "$TEMPLATE" ] || [ -z "$OUTPUT" ] || [ -z "$VALUES" ]; then
      echo '{"error": "参数不完整"}' >&2
      exit 1
    fi
    
    # 创建临时 JSON 文件
    TMPFILE=$(mktemp)
    echo "$VALUES" > "$TMPFILE"
    
    # 执行生成
    $KUAI_CMD use "$TEMPLATE" "$OUTPUT" --values "$TMPFILE" --defaults --force
    
    # 清理
    rm "$TMPFILE"
    ;;
  *)
    echo '{"error": "未知命令"}' >&2
    exit 1
    ;;
esac
```

## 完整工作流程示例

### 前端工作流程

1. **获取模板列表**
   ```javascript
   const templates = await fetch('/api/templates').then(r => r.json());
   ```

2. **用户选择模板，获取详情**
   ```javascript
   const templateInfo = await fetch(`/api/templates/${templateName}`)
     .then(r => r.json());
   ```

3. **用户填写表单，提交生成**
   ```javascript
   const result = await fetch('/api/generate', {
     method: 'POST',
     headers: { 'Content-Type': 'application/json' },
     body: JSON.stringify({
       templateName: 'go-service',
       values: {
         Name: 'my-service',
         Port: '3000'
       },
       outputDir: '/tmp/my-service-output'
     })
   }).then(r => r.json());
   ```

4. **打包并下载**
   ```javascript
   // 后端打包
   exec('cd /tmp/my-service-output && zip -r project.zip .');
   // 返回下载链接
   ```

## 注意事项

1. **路径处理**：确保 `kuai` 命令在 PATH 中，或使用绝对路径
2. **权限问题**：确保运行用户有权限创建输出目录
3. **临时文件**：及时清理临时 JSON 文件
4. **错误处理**：妥善处理命令执行失败的情况
5. **安全性**：验证用户输入，防止命令注入

## 测试命令

```bash
# 测试列表
kuai template list --json

# 测试详情（需要先有模板）
kuai template show <template-name> --json

# 测试生成
echo '{"Name":"test","Port":"8080"}' > /tmp/test-values.json
kuai use <template-name> /tmp/test-output --values /tmp/test-values.json --defaults --force
```

