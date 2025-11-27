package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newTemplateTreeCmd() *cobra.Command {
	var jsonOutput bool
	var maxDepth int

	cmd := &cobra.Command{
		Use:   "tree <name>",
		Short: "显示模板的目录结构",
		Long:  "显示指定模板的目录树结构，方便查看模板包含的文件和目录",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			templatePath, err := templateMgr.TemplatePath(name)
			if err != nil {
				return err
			}

			// 检查是否有 template/ 子目录
			templateSubdir := filepath.Join(templatePath, "template")
			actualPath := templatePath
			if info, err := os.Stat(templateSubdir); err == nil && info.IsDir() {
				actualPath = templateSubdir
			}

			// 构建目录树
			tree, err := buildFileTree(actualPath, maxDepth)
			if err != nil {
				return fmt.Errorf("构建目录树失败: %w", err)
			}

			if jsonOutput {
				// JSON 输出
				output := struct {
					TemplateName string      `json:"templateName"`
					Path         string      `json:"path"`
					Tree         *FileNode   `json:"tree"`
				}{
					TemplateName: name,
					Path:         actualPath,
					Tree:         tree,
				}
				data, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("序列化 JSON 失败: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			// 文本输出
			fmt.Fprintf(cmd.OutOrStdout(), "模板: %s\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "路径: %s\n\n", actualPath)
			fmt.Fprintln(cmd.OutOrStdout(), "目录结构:")
			printTreeToWriter(cmd.OutOrStdout(), tree, "", true)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	cmd.Flags().IntVar(&maxDepth, "max-depth", 0, "最大深度（0 表示不限制）")
	return cmd
}

// FileNode 表示文件树节点
type FileNode struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"` // "file" 或 "directory"
	Path     string      `json:"path,omitempty"`
	Children []*FileNode `json:"children,omitempty"`
	Size     int64       `json:"size,omitempty"`
}

func buildFileTree(root string, maxDepth int) (*FileNode, error) {
	if _, err := os.Stat(root); err != nil {
		return nil, err
	}

	rootNode := &FileNode{
		Name: filepath.Base(root),
		Type: "directory",
		Path: root,
	}

	if err := walkDir(root, rootNode, 0, maxDepth); err != nil {
		return nil, err
	}

	return rootNode, nil
}

func walkDir(dirPath string, parent *FileNode, currentDepth int, maxDepth int) error {
	// 检查深度限制
	if maxDepth > 0 && currentDepth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// 过滤掉 .git 和 manifest 文件
	skipFiles := map[string]struct{}{
		".git":      {},
		"kuai.yaml": {},
		"kuai.yml":  {},
		"kuai.json": {},
	}

	for _, entry := range entries {
		name := entry.Name()
		if _, skip := skipFiles[strings.ToLower(name)]; skip {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		node := &FileNode{
			Name: name,
			Path: fullPath,
		}

		if entry.IsDir() {
			node.Type = "directory"
			node.Children = []*FileNode{}
			parent.Children = append(parent.Children, node)
			// 递归处理子目录
			if err := walkDir(fullPath, node, currentDepth+1, maxDepth); err != nil {
				return err
			}
		} else {
			node.Type = "file"
			if info, err := entry.Info(); err == nil {
				node.Size = info.Size()
			}
			parent.Children = append(parent.Children, node)
		}
	}

	return nil
}

func printTreeToWriter(w io.Writer, node *FileNode, prefix string, isLast bool) {
	// 打印当前节点
	marker := "├── "
	if isLast {
		marker = "└── "
	}
	fmt.Fprintf(w, "%s%s%s", prefix, marker, node.Name)
	if node.Type == "file" && node.Size > 0 {
		fmt.Fprintf(w, " (%d bytes)", node.Size)
	}
	fmt.Fprintln(w)

	// 打印子节点
	if len(node.Children) > 0 {
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}

		for i, child := range node.Children {
			isLastChild := i == len(node.Children)-1
			printTreeToWriter(w, child, childPrefix, isLastChild)
		}
	}
}

