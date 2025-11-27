package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newTemplateListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出已安装的模板",
		RunE: func(cmd *cobra.Command, args []string) error {
			templates, err := templateMgr.List()
			if err != nil {
				return err
			}

			if jsonOutput {
				// JSON 输出
				data, err := json.MarshalIndent(templates, "", "  ")
				if err != nil {
					return fmt.Errorf("序列化 JSON 失败: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			// 文本输出
			if len(templates) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "暂无模板，使用 `kuai template add` 添加。")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION")
			for _, tpl := range templates {
				fmt.Fprintf(w, "%s\t%s\n", tpl.Name, tpl.Description)
			}
			return w.Flush()
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "输出 JSON 格式")
	return cmd
}

