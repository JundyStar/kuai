package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

// ValuesConfig 定义变量收集的配置。
type ValuesConfig struct {
	Manifest   *Manifest
	FromFile   string
	RawPairs   []string
	UseDefault bool
}

// CollectValues 根据 manifest 加载变量。
func CollectValues(cfg ValuesConfig) (map[string]string, error) {
	values := map[string]string{}

	// 先加载文件
	if cfg.FromFile != "" {
		fileValues, err := loadValuesFile(cfg.FromFile)
		if err != nil {
			return nil, err
		}
		merge(values, fileValues)
	}

	// 解析 CLI 变量
	cliValues, err := parsePairs(cfg.RawPairs)
	if err != nil {
		return nil, err
	}
	merge(values, cliValues)

	// 按 manifest 补齐
	if cfg.Manifest == nil {
		return values, nil
	}

	for _, field := range cfg.Manifest.Fields {
		if _, ok := values[field.Name]; ok {
			continue
		}
		if cfg.UseDefault {
			if field.Default == "" && field.Required {
				return nil, fmt.Errorf("字段 %s 需要提供值", field.Name)
			}
			if field.Default != "" {
				values[field.Name] = field.Default
			}
			continue
		}
		prompt := promptui.Prompt{
			Label:     buildPromptLabel(field),
			Default:   field.Default,
			AllowEdit: true,
		}
		answer, err := prompt.Run()
		if err != nil {
			return nil, fmt.Errorf("读取字段 %s 失败: %w", field.Name, err)
		}
		if answer == "" && field.Required {
			return nil, fmt.Errorf("字段 %s 不能为空", field.Name)
		}
		if answer != "" {
			values[field.Name] = answer
		} else if field.Default != "" {
			values[field.Name] = field.Default
		}
	}

	return values, nil
}

func buildPromptLabel(field Field) string {
	label := field.Name
	if field.Prompt != "" {
		label = field.Prompt
	}
	if field.Description != "" {
		label = fmt.Sprintf("%s (%s)", label, field.Description)
	}
	return label
}

func loadValuesFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取变量文件失败: %w", err)
	}
	result := map[string]string{}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &result); err != nil {
			return nil, fmt.Errorf("解析 YAML 失败: %w", err)
		}
	default:
		if err := json.Unmarshal(content, &result); err != nil {
			return nil, fmt.Errorf("解析 JSON 失败: %w", err)
		}
	}
	return result, nil
}

func parsePairs(items []string) (map[string]string, error) {
	values := map[string]string{}
	for _, item := range items {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("无法解析变量 %q，需要 key=value 形式", item)
		}
		values[parts[0]] = parts[1]
	}
	return values, nil
}

func merge(dst, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}

