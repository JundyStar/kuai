package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "检查环境配置是否可用",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "配置目录: %s\n", paths.ConfigDir)
			fmt.Fprintf(cmd.OutOrStdout(), "模板目录: %s\n", paths.TemplatesDir)

			if _, err := templateMgr.List(); err != nil {
				return fail("读取模板失败: %v", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✅ Kuai 就绪，可以开始使用模板。")
			return nil
		},
	}
}

