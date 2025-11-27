package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jundy/kuai/pkg/templates"
)

func newTemplateShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "显示模板详情",
		Long:  "显示指定模板的详细信息，包括变量字段和配置",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			templatePath, err := templateMgr.TemplatePath(name)
			if err != nil {
				return err
			}

			// 加载 manifest
			manifest, manifestPath, err := templates.LoadManifest(templatePath)
			if err != nil {
				return fmt.Errorf("加载模板配置失败: %w", err)
			}

			// 如果 manifest 不存在，自动扫描生成
			if manifest == nil {
				manifest = templates.ScanTemplateVariables(templatePath)
				manifest.Name = name
			}

			// 构建输出结构
			output := struct {
				Name        string              `json:"name"`
				Description string              `json:"description"`
				Path        string              `json:"path"`
				Manifest    *templates.Manifest `json:"manifest"`
			}{
				Name:        name,
				Description: manifest.Description,
				Path:        templatePath,
				Manifest:    manifest,
			}

			if jsonOutput {
				// JSON 输出
				data, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return fmt.Errorf("序列化 JSON 失败: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			// 文本输出
			fmt.Fprintf(cmd.OutOrStdout(), "模板名称: %s\n", name)
			fmt.Fprintf(cmd.OutOrStdout(), "描述: %s\n", manifest.Description)
			fmt.Fprintf(cmd.OutOrStdout(), "路径: %s\n", templatePath)
			if manifestPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "配置文件: %s\n", manifestPath)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "配置文件: 自动生成\n")
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\n变量字段:")
			if len(manifest.Fields) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "  无")
			} else {
				for _, field := range manifest.Fields {
					required := ""
					if field.Required {
						required = " (必填)"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  - %s%s\n", field.Name, required)
					if field.Prompt != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    提示: %s\n", field.Prompt)
					}
					if field.Description != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    说明: %s\n", field.Description)
					}
					if field.Default != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "    默认值: %s\n", field.Default)
					}
					fmt.Fprintln(cmd.OutOrStdout())
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

