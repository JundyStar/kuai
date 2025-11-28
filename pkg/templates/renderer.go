package templates

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Render 将模板渲染到目标目录。
// 会遍历源目录中的所有文件，使用 values 中的变量替换模板语法 {{变量名}}。
// 同时支持文件路径和文件内容的模板渲染。
// 安全性：会自动检查渲染后的路径，防止路径遍历攻击。
func Render(srcDir, dstDir string, values map[string]string) error {
	funcs := buildFuncMap(values)
	renderPath := func(rel string) (string, error) {
		tmpl, err := template.New("path").Funcs(funcs).Option("missingkey=error").Parse(rel)
		if err != nil {
			return "", err
		}
		var b bytes.Buffer
		if err := tmpl.Execute(&b, values); err != nil {
			return "", err
		}
		return b.String(), nil
	}

	return filepath.WalkDir(srcDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".git" {
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		targetRel, err := renderPath(rel)
		if err != nil {
			return fmt.Errorf("渲染路径 %s 失败: %w", rel, err)
		}
		// 安全检查：防止路径遍历攻击
		if filepath.IsAbs(targetRel) || strings.Contains(targetRel, "..") {
			return fmt.Errorf("渲染后的路径 %s 包含非法字符，拒绝渲染", targetRel)
		}
		targetPath := filepath.Join(dstDir, targetRel)

		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		if _, skip := skipFiles[strings.ToLower(entry.Name())]; skip {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		tmpl, err := template.New(rel).Funcs(funcs).Option("missingkey=error").Parse(string(data))
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %w", rel, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, values); err != nil {
			return fmt.Errorf("渲染模板 %s 失败: %w", rel, err)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		return os.WriteFile(targetPath, buf.Bytes(), 0o644)
	})
}

// buildFuncMap 将 values 映射转换为 template.FuncMap。
// 这样在模板中可以直接使用 {{变量名}} 而不需要 {{.变量名}}。
func buildFuncMap(values map[string]string) template.FuncMap {
	funcs := template.FuncMap{}
	for k, v := range values {
		val := v // 闭包捕获，确保每个函数返回正确的值
		funcs[k] = func() string {
			return val
		}
	}
	return funcs
}

