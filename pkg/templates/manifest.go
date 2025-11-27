package templates

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var manifestFilenames = []string{"kuai.yaml", "kuai.yml", "kuai.json"}

var skipFiles = map[string]struct{}{
	"kuai.yaml": {},
	"kuai.yml":  {},
	"kuai.json": {},
}

// Manifest 描述模板所需的变量。
type Manifest struct {
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Fields      []Field       `json:"fields" yaml:"fields"`
	Meta        ManifestMeta  `json:"meta" yaml:"meta"`
}

// ManifestMeta 存储额外信息。
type ManifestMeta struct {
	Version string `json:"version" yaml:"version"`
}

// Field 定义模板变量。
type Field struct {
	Name        string `json:"name" yaml:"name"`
	Prompt      string `json:"prompt" yaml:"prompt"`
	Description string `json:"description" yaml:"description"`
	Default     string `json:"default" yaml:"default"`
	Required    bool   `json:"required" yaml:"required"`
}

// LoadManifest 读取模板 Manifest。如果没有找到，会自动扫描模板变量生成默认配置。
func LoadManifest(dir string) (*Manifest, string, error) {
	for _, name := range manifestFilenames {
		full := filepath.Join(dir, name)
		data, err := os.ReadFile(full)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, "", err
		}
		manifest := &Manifest{}
		if strings.HasSuffix(name, ".json") {
			if err := json.Unmarshal(data, manifest); err != nil {
				return nil, "", fmt.Errorf("解析 %s 失败: %w", name, err)
			}
		} else {
			if err := yaml.Unmarshal(data, manifest); err != nil {
				return nil, "", fmt.Errorf("解析 %s 失败: %w", name, err)
			}
		}
		return manifest, full, nil
	}
	// 没有找到 manifest，自动扫描模板变量
	manifest := ScanTemplateVariables(dir)
	return manifest, "", nil
}

// ScanTemplateVariables 扫描模板目录中的所有 {{变量名}}，生成默认 Manifest。
func ScanTemplateVariables(dir string) *Manifest {
	varMap := make(map[string]struct{})
	re := regexp.MustCompile(`\{\{([A-Za-z0-9_]+)\}\}`)

	filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// 跳过 .git 和 manifest 文件
		if entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if _, skip := skipFiles[strings.ToLower(entry.Name())]; skip {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		// 读取文件内容
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		// 提取所有变量名
		matches := re.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			if len(match) > 1 {
				varMap[match[1]] = struct{}{}
			}
		}
		// 路径也可能包含变量
		rel, err := filepath.Rel(dir, path)
		if err == nil {
			pathMatches := re.FindAllStringSubmatch(rel, -1)
			for _, match := range pathMatches {
				if len(match) > 1 {
					varMap[match[1]] = struct{}{}
				}
			}
		}
		return nil
	})

	// 转换为排序后的字段列表
	var names []string
	for name := range varMap {
		// 排除一些常见的系统变量
		if name != "TemplateName" {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	fields := make([]Field, 0, len(names))
	for _, name := range names {
		fields = append(fields, Field{
			Name:        name,
			Prompt:      formatPrompt(name),
			Description: "",
			Default:     "",
			Required:    true,
		})
	}

	return &Manifest{
		Name:        filepath.Base(dir),
		Description: "自动生成的模板配置",
		Fields:      fields,
	}
}

// formatPrompt 将变量名格式化为友好的提示文本。
func formatPrompt(name string) string {
	// 将驼峰命名转换为中文提示
	parts := splitCamelCase(name)
	return strings.Join(parts, " ")
}

// splitCamelCase 将驼峰命名拆分为单词。
func splitCamelCase(s string) []string {
	var parts []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

