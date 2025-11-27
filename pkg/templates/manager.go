package templates

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/jundy/kuai/pkg/config"
)

// Manager 负责模板的增删查。
type Manager struct {
	paths config.Paths
}

// TemplateInfo 描述一个模板。
type TemplateInfo struct {
	Name        string
	Description string
}

// NewManager 创建模板管理器。
func NewManager(paths config.Paths) *Manager {
	return &Manager{paths: paths}
}

// Add 从本地路径复制模板。
func (m *Manager) Add(name, from string, force bool) error {
	if name == "" {
		return fmt.Errorf("模板名不能为空")
	}
	dst := filepath.Join(m.paths.TemplatesDir, name)
	if _, err := os.Stat(dst); err == nil && !force {
		return fmt.Errorf("模板 %s 已存在，使用 --force 覆盖", name)
	}
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("清理旧模板失败: %w", err)
	}
	if err := copyDir(from, dst); err != nil {
		return err
	}
	return nil
}

// Remove 删除模板。
func (m *Manager) Remove(name string) error {
	if name == "" {
		return fmt.Errorf("模板名不能为空")
	}
	return os.RemoveAll(filepath.Join(m.paths.TemplatesDir, name))
}

// List 返回模板信息。
func (m *Manager) List() ([]TemplateInfo, error) {
	entries, err := os.ReadDir(m.paths.TemplatesDir)
	if err != nil {
		return nil, err
	}
	var infos []TemplateInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		desc := ""
		if manifest, _, _ := LoadManifest(filepath.Join(m.paths.TemplatesDir, name)); manifest != nil {
			desc = manifest.Description
		}
		infos = append(infos, TemplateInfo{
			Name:        name,
			Description: desc,
		})
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos, nil
}

// TemplatePath 返回模板目录。
func (m *Manager) TemplatePath(name string) (string, error) {
	path := filepath.Join(m.paths.TemplatesDir, name)
	if stat, err := os.Stat(path); err != nil || !stat.IsDir() {
		return "", fmt.Errorf("模板 %s 不存在", name)
	}
	return path, nil
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("源路径必须是目录")
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == src {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), info.Mode())
		}
		return copyFile(path, filepath.Join(dst, rel), info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

