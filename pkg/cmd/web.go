package cmd

import (
	"fmt"
	"net"
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
			fmt.Fprintf(cmd.OutOrStdout(), "📱 访问地址: http://%s:%d\n", host, port)
			
			// 如果监听所有接口，显示本机 IP 地址
			if host == "0.0.0.0" || host == "" {
				if ips := getLocalIPs(); len(ips) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "\n💡 可通过以下地址访问:\n")
					for _, ip := range ips {
						fmt.Fprintf(cmd.OutOrStdout(), "   http://%s:%d\n", ip, port)
					}
				}
			}
			
			fmt.Fprintf(cmd.OutOrStdout(), "\n按 Ctrl+C 停止服务器\n\n")
			return http.ListenAndServe(addr, server)
		},
	}

	webCmd.Flags().IntVarP(&port, "port", "p", 8080, "服务器端口")
	webCmd.Flags().StringVar(&host, "host", "0.0.0.0", "服务器地址 (0.0.0.0 表示监听所有网络接口)")
	return webCmd
}

// getLocalIPs 获取本机的非回环 IP 地址列表
func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}

