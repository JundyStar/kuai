package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/jundy/kuai/web"
)

func newWebCmd() *cobra.Command {
	var port int
	var host string

	webCmd := &cobra.Command{
		Use:   "web",
		Short: "启动 Web 界面",
		Long:  "启动一个 Web 服务器，提供图形化界面来管理模板和生成项目",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := web.NewServer(templateMgr, paths)
			addr := fmt.Sprintf("%s:%d", host, port)
			fmt.Fprintf(cmd.OutOrStdout(), "🚀 Kuai Web 界面已启动\n")
			fmt.Fprintf(cmd.OutOrStdout(), "📱 访问地址: http://%s\n", addr)
			fmt.Fprintf(cmd.OutOrStdout(), "按 Ctrl+C 停止服务器\n\n")
			return http.ListenAndServe(addr, server)
		},
	}

	webCmd.Flags().IntVarP(&port, "port", "p", 8080, "服务器端口")
	webCmd.Flags().StringVar(&host, "host", "localhost", "服务器地址")
	return webCmd
}

