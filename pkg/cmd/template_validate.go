package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTemplateValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <name>",
		Short: "验证模板是否有效",
		Long:  "检查模板目录是否存在，是否包含必要的文件",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := templateMgr.Validate(name); err != nil {
				return fmt.Errorf("模板验证失败: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✅ 模板 %s 验证通过\n", name)
			return nil
		},
	}
	return cmd
}

