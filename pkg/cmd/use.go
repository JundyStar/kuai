package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/jundy/kuai/pkg/templates"
)

func newUseCmd() *cobra.Command {
	var vars []string
	var valuesFile string
	var defaults bool
	var force bool

	useCmd := &cobra.Command{
		Use:   "use <template> <target>",
		Short: "基于模板创建新项目",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, target := args[0], args[1]

			templatePath, err := templateMgr.TemplatePath(name)
			if err != nil {
				return err
			}

			if err := ensureTargetDir(target, force); err != nil {
				return err
			}

			manifest, _, err := templates.LoadManifest(templatePath)
			if err != nil {
				return err
			}

			values, err := templates.CollectValues(templates.ValuesConfig{
				Manifest:   manifest,
				FromFile:   valuesFile,
				RawPairs:   vars,
				UseDefault: defaults,
			})
			if err != nil {
				return err
			}
			values["TemplateName"] = name

			// 如果模板目录里有 template/ 子目录，使用它作为源目录（常见模板仓库结构）
			actualTemplatePath := templatePath
			templateSubdir := filepath.Join(templatePath, "template")
			if info, err := os.Stat(templateSubdir); err == nil && info.IsDir() {
				actualTemplatePath = templateSubdir
			}

			if err := templates.Render(actualTemplatePath, target, values); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "🚀 已在 %s 基于模板 %s 创建项目。\n", target, name)
			return nil
		},
	}

	useCmd.Flags().StringArrayVar(&vars, "var", nil, "以 key=value 设置变量，可多次使用")
	useCmd.Flags().StringVar(&valuesFile, "values", "", "从 JSON/YAML 文件加载变量")
	useCmd.Flags().BoolVar(&defaults, "defaults", false, "跳过交互，直接使用默认值")
	useCmd.Flags().BoolVar(&force, "force", false, "强制覆盖非空目标目录，不询问确认")
	return useCmd
}

func ensureTargetDir(path string, force bool) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fail("目标 %s 已存在且不是目录", path)
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			if force {
				// 强制模式：直接清空目录
				if err := os.RemoveAll(path); err != nil {
					return fmt.Errorf("清空目标目录失败: %w", err)
				}
				return os.MkdirAll(path, 0o755)
			}
			// 交互式确认
			prompt := promptui.Prompt{
				Label:     fmt.Sprintf("目标目录 %s 非空，是否清空并覆盖？(y/N)", path),
				Default:   "N",
				AllowEdit: true,
			}
			result, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("操作已取消")
			}
			if result != "y" && result != "Y" && result != "yes" && result != "Yes" {
				return fmt.Errorf("操作已取消")
			}
			// 用户确认：清空目录
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("清空目标目录失败: %w", err)
			}
			return os.MkdirAll(path, 0o755)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(path, 0o755)
}

