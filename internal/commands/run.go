// Package config implements all commands of KoboMail
package commands

import (
	series_cleanup "github.com/bjw-s/series-cleanup/internal/series-cleanup"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run series-cleanup processing",
	Long:  "Run series-cleanup processing.",
	RunE: func(cmd *cobra.Command, args []string) error {
		series_cleanup.AppConfig = conf
		zap.S().Debugw("Running with configuration",
			zap.Any("configuration", conf),
		)
		series_cleanup.Run()
		return nil
	},
}
