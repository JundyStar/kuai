package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths 包含所有需要的目录。
type Paths struct {
	ConfigDir    string
	TemplatesDir string
}

// Resolve 根据用户输入计算目录路径。
func Resolve(custom string) (Paths, error) {
	dir := custom
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, fmt.Errorf("无法获取用户目录: %w", err)
		}
		dir = filepath.Join(home, ".kuai")
	}

	return Paths{
		ConfigDir:    dir,
		TemplatesDir: filepath.Join(dir, "templates"),
	}, nil
}

// Ensure 确保核心目录存在。
func Ensure(p Paths) error {
	if err := os.MkdirAll(p.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	if err := os.MkdirAll(p.TemplatesDir, 0o755); err != nil {
		return fmt.Errorf("创建模板目录失败: %w", err)
	}
	return nil
}

