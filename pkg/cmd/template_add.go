package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTemplateAddCmd() *cobra.Command {
	var from string
	var force bool

	cmd := &cobra.Command{
		Use:   "add <name> --from <path>",
		Short: "从本地目录添加模板",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if from == "" {
				return fail("--from 不能为空")
			}
			name := args[0]
			if err := templateMgr.Add(name, from, force); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "✅ 模板 %s 已添加。\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "模板来源目录")
	cmd.Flags().BoolVar(&force, "force", false, "存在同名模板时覆盖")
	return cmd
}

