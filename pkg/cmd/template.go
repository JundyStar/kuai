package cmd

import "github.com/spf13/cobra"

func newTemplateCmd() *cobra.Command {
	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "管理模板仓库",
	}

	templateCmd.AddCommand(newTemplateAddCmd())
	templateCmd.AddCommand(newTemplateListCmd())
	templateCmd.AddCommand(newTemplateRemoveCmd())
	return templateCmd
}

