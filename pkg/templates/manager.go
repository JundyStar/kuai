package templates

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jundy/kuai/pkg/config"
)

// Manager 负责模板的增删查、验证和导出。
// 所有模板操作都会进行安全性检查，防止路径遍历攻击。
type Manager struct {
	paths config.Paths
}

// TemplateInfo 描述一个模板的基本信息。
type TemplateInfo struct {
	Name        string // 模板名称
	Description string // 模板描述（来自 manifest）
}

// NewManager 创建模板管理器。
func NewManager(paths config.Paths) *Manager {
	return &Manager{paths: paths}
}

// Add 从本地路径复制模板。
// 如果目标模板已存在且 force 为 false，会返回错误。
// 如果 force 为 true，会先备份现有模板（如果存在），然后覆盖。
func (m *Manager) Add(name, from string, force bool) error {
	if name == "" {
		return fmt.Errorf("模板名不能为空")
	}
	// 验证模板名称：防止路径遍历和特殊字符
	if err := validateTemplateName(name); err != nil {
		return err
	}
	dst := filepath.Join(m.paths.TemplatesDir, name)
	
	// 如果模板已存在，处理备份或返回错误
	if _, err := os.Stat(dst); err == nil {
		if !force {
			return fmt.Errorf("模板 %s 已存在，使用 --force 覆盖", name)
		}
		// 备份现有模板
		if err := m.backupTemplate(name); err != nil {
			return fmt.Errorf("备份模板失败: %w", err)
		}
	}
	
	// 清理目标目录
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("清理旧模板失败: %w", err)
	}
	
	// 复制模板
	if err := copyDir(from, dst); err != nil {
		return err
	}
	
	// 验证模板有效性
	if err := m.Validate(name); err != nil {
		// 如果验证失败，尝试恢复备份
		if backupErr := m.restoreTemplate(name); backupErr != nil {
			return fmt.Errorf("模板验证失败: %w，且恢复备份失败: %v", err, backupErr)
		}
		return fmt.Errorf("模板验证失败: %w，已恢复备份", err)
	}
	
	return nil
}

// Remove 删除模板。
func (m *Manager) Remove(name string) error {
	if err := validateTemplateName(name); err != nil {
		return err
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
	// 验证模板名称：防止路径遍历和特殊字符
	if err := validateTemplateName(name); err != nil {
		return "", err
	}
	path := filepath.Join(m.paths.TemplatesDir, name)
	if stat, err := os.Stat(path); err != nil || !stat.IsDir() {
		return "", fmt.Errorf("模板 %s 不存在", name)
	}
	return path, nil
}

// Validate 验证模板是否有效。
// 检查模板目录是否存在，是否包含必要的文件。
func (m *Manager) Validate(name string) error {
	path, err := m.TemplatePath(name)
	if err != nil {
		return err
	}
	
	// 检查目录是否为空
	entries, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("读取模板目录失败: %w", err)
	}
	if len(entries) == 0 {
		return fmt.Errorf("模板目录为空")
	}
	
	// 检查是否有 template/ 子目录或直接包含模板文件
	hasTemplateDir := false
	hasFiles := false
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == "template" {
			hasTemplateDir = true
		}
		if !entry.IsDir() && entry.Name() != "kuai.yaml" && entry.Name() != "kuai.yml" && entry.Name() != "kuai.json" {
			hasFiles = true
		}
	}
	
	if !hasTemplateDir && !hasFiles {
		return fmt.Errorf("模板不包含任何文件")
	}
	
	return nil
}

// Export 将模板导出为 ZIP 文件。
func (m *Manager) Export(name, outputPath string) error {
	path, err := m.TemplatePath(name)
	if err != nil {
		return err
	}
	
	// 创建 ZIP 文件
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}
	defer zipFile.Close()
	
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	// 遍历模板目录并添加到 ZIP
	return filepath.WalkDir(path, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// 跳过 .git 目录
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		
		// 计算相对路径
		relPath, err := filepath.Rel(path, filePath)
		if err != nil {
			return err
		}
		
		// 跳过根目录本身
		if relPath == "." {
			return nil
		}
		
		info, err := entry.Info()
		if err != nil {
			return err
		}
		
		// 创建 ZIP 文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath
		if entry.IsDir() {
			header.Name += "/"
		}
		
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		
		// 如果是文件，写入内容
		if !entry.IsDir() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()
			
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}
		
		return nil
	})
}

// backupTemplate 备份现有模板到备份目录。
func (m *Manager) backupTemplate(name string) error {
	src := filepath.Join(m.paths.TemplatesDir, name)
	backupDir := filepath.Join(m.paths.ConfigDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}
	
	// 使用时间戳作为备份名称
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s-%s", name, timestamp)
	backupPath := filepath.Join(backupDir, backupName)
	
	return copyDir(src, backupPath)
}

// restoreTemplate 从最近的备份恢复模板。
func (m *Manager) restoreTemplate(name string) error {
	backupDir := filepath.Join(m.paths.ConfigDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %w", err)
	}
	
	// 查找匹配的备份（以模板名开头）
	var latestBackup string
	var latestTime time.Time
	prefix := name + "-"
	
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		
		// 尝试解析时间戳
		timestampStr := strings.TrimPrefix(entry.Name(), prefix)
		if t, err := time.Parse("20060102-150405", timestampStr); err == nil {
			if t.After(latestTime) {
				latestTime = t
				latestBackup = entry.Name()
			}
		}
	}
	
	if latestBackup == "" {
		return fmt.Errorf("未找到备份")
	}
	
	backupPath := filepath.Join(backupDir, latestBackup)
	dst := filepath.Join(m.paths.TemplatesDir, name)
	
	// 清理目标目录
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	
	// 恢复备份
	return copyDir(backupPath, dst)
}

// validateTemplateName 验证模板名称是否合法。
func validateTemplateName(name string) error {
	if name == "" {
		return fmt.Errorf("模板名不能为空")
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("模板名包含非法字符（不能包含 ..、/、\\）")
	}
	return nil
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
	// 使用 WalkDir 提升性能（不需要读取文件信息）
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
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
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
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

