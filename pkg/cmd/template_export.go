package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newTemplateExportCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "导出模板为 ZIP 文件",
		Long:  "将指定的模板打包为 ZIP 文件，方便分享和备份",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			
			// 如果没有指定输出路径，使用模板名作为文件名
			if output == "" {
				output = fmt.Sprintf("%s.zip", name)
			}
			
			// 确保输出路径是绝对路径
			if !filepath.IsAbs(output) {
				var err error
				output, err = filepath.Abs(output)
				if err != nil {
					return fmt.Errorf("解析输出路径失败: %w", err)
				}
			}
			
			if err := templateMgr.Export(name, output); err != nil {
				return err
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "✅ 模板 %s 已导出到 %s\n", name, output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "输出 ZIP 文件路径（默认为 <name>.zip）")
	return cmd
}

