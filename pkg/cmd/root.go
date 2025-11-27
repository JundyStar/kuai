package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jundy/kuai/pkg/config"
	"github.com/jundy/kuai/pkg/templates"
)

var (
	configDir   string
	paths       config.Paths
	templateMgr *templates.Manager
)

// RootCmd is the primary CLI entry point.
var RootCmd = &cobra.Command{
	Use:           "kuai",
	Short:         "Kuai - 快速、稳定的模板管理工具",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		paths, err = config.Resolve(configDir)
		if err != nil {
			return err
		}
		if err := config.Ensure(paths); err != nil {
			return err
		}
		templateMgr = templates.NewManager(paths)
		return nil
	},
}

func init() {
	home, _ := os.UserHomeDir()
	defaultDir := filepath.Join(home, ".kuai")

	RootCmd.PersistentFlags().StringVar(&configDir, "config", defaultDir, "配置目录")

	RootCmd.AddCommand(newUseCmd())
	RootCmd.AddCommand(newTemplateCmd())
	RootCmd.AddCommand(newDoctorCmd())
}

func fail(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

