// Package config implements all commands of KoboMail
package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	conf = &config.Config{}

	rootCmd = &cobra.Command{
		Use:   "series-cleanup",
		Short: "series-cleanup cleans up media files for tv shows watched on Trakt.tv",
		Long: `series-cleanup cleans up media files for tv shows watched on Trakt.tv.
More information available at the Github Repo (https://github.com/bjw-s/series-cleanup)`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	cobra.OnFinalize(finalizeLogger)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringP("config", "c", "", "settings.yaml file for configuring series-cleanup")
	rootCmd.PersistentFlags().String("log_file", "", "Log file location")
	rootCmd.PersistentFlags().StringP("log_level", "l", "", "Log level (debug, info, warn, error, fatal, panic)")
	rootCmd.PersistentFlags().String("log_format", "", "Log format (console, json)")
	rootCmd.PersistentFlags().String("trakt.cache_dir", "", "Folder to store the cache for Trakt.tv queries. Defaults to current working dir")
	rootCmd.PersistentFlags().Bool("dry_run", false, "Run in dry-run mode, don't actually delete any files")
}

func initConfig() {
	var err error
	conf, err = config.LoadConfig(rootCmd.PersistentFlags())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := conf.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initLogger() {
	atom := zap.NewAtomicLevel()

	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	if conf.LogFormat == "json" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var core zapcore.Core
	if conf.LogFile != "/dev/stdout" {
		logFile, _ := os.OpenFile(conf.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.Lock(logFile), atom),
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atom),
		)
	} else {
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), atom),
		)
	}
	logger := zap.New(core)

	// Create a logger with a default level first to ensure config failures are loggable.
	atom.SetLevel(zapcore.InfoLevel)
	zap.ReplaceGlobals(logger)

	lvl, err := zapcore.ParseLevel(conf.LogLevel)
	if err != nil {
		zap.S().Errorf("Invalid log level %s, using default level: info", conf.LogLevel)
		lvl = zapcore.InfoLevel
	}
	atom.SetLevel(lvl)

	zap.S().Debug("Logger initialized")
}

func finalizeLogger() {
	// Flushes buffered log messages
	zap.S().Sync()
}
