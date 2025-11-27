package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTemplateRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "删除模板",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := templateMgr.Remove(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "🗑 已删除模板 %s\n", args[0])
			return nil
		},
	}
}

