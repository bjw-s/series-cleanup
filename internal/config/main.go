// Package config implements all configuration aspects of series-cleanup
package config

import (
	"os"
	"strings"

	koanf_yaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// Config represents the application configuration structure
type Config struct {
	DryRun         bool             `koanf:"dry_run"`
	FolderRegex    string           `koanf:"folder_regex" validate:"required"`
	FolderSettings []FolderSettings `koanf:"folder_settings" validate:"required,dive"`
	LogFile        string           `koanf:"log_file"`
	LogFormat      string           `koanf:"log_format" validate:"oneof=console json"`
	LogLevel       string           `koanf:"log_level" validate:"valid_loglevel"`
	Rules          globalRules      `koanf:"rules" validate:"required"`
	ScanFolders    []string         `koanf:"scan_folders" validate:"dive,valid_folder"`
	Trakt          traktConfig      `koanf:"trakt" validate:"omitempty"`
	Plex           plexConfig       `koanf:"plex" validate:"omitempty"`
	k              *koanf.Koanf
}

type globalRules struct {
	DeleteWatchedAfterHours int `koanf:"delete_watched_after_hours" validate:"required,min=1"`
	// TODO: implement this after Plex integration
	// DeleteUnwatchedDaysAfterAdded *int
}

type FolderRules struct {
	DeleteWatchedAfterHours int `koanf:"delete_watched_after_hours" validate:"omitempty,gt=0"`
	// TODO: implement this after Plex integration
	// DeleteUnwatchedDaysAfterAdded int   `validate:"omitempty,gt=0"`
	KeepShow    bool  `koanf:"keep_show" validate:"omitempty,excluded_with_all=KeepSeasons"`
	KeepSeasons []int `koanf:"keep_seasons" validate:"omitempty,excluded_with_all=KeepShow,gt=0,dive"`
}

type folderSettingsIdentifiers struct {
	Trakt int    `koanf:"trakt"`
	Slug  string `koanf:"slug"`
	IMDB  string `koanf:"imdb"`
	TVDB  int    `koanf:"tvdb"`
	TMDB  int    `koanf:"tmdb"`
}

// FolderSettings represents settings specific to this folder
type FolderSettings struct {
	Folder      string                    `koanf:"folder" validate:"required"`
	Identifiers folderSettingsIdentifiers `koanf:"identifiers" validate:"omitempty"`
	Rules       FolderRules               `koanf:"rules" validate:"required,dive"`
}

type traktConfig struct {
	Enabled      bool            `koanf:"enabled"`
	CacheDir     string          `koanf:"cache_dir" validate:"required_if=Enabled true,omitempty,valid_folder"`
	ClientID     string          `koanf:"client_id" validate:"required_if=Enabled true,omitempty,gt=0"`
	ClientSecret sensitiveString `koanf:"client_secret" validate:"required_if=Enabled true,omitempty,gt=0"`
	User         string          `koanf:"user" validate:"required_if=Enabled true,omitempty,gt=0"`
}

type plexConfig struct {
	Enabled bool            `koanf:"enabled"`
	Token   sensitiveString `validate:"required_if=Enabled true,omitempty,gt=0"`
}

const (
	DefaultLogFile   = "/dev/stdout"
	DefaultLogFormat = "console"
	DefaultLogLevel  = "info"
)

// LoadConfig instantiates a new Config
func LoadConfig(flags *flag.FlagSet) (*Config, error) {
	const envVarPrefx = "SC_"

	var err error
	var k = koanf.New(".")

	// Defaults
	currentWorkingDir, _ := os.Getwd()
	k.Load(confmap.Provider(map[string]interface{}{
		"dry_run":         false,
		"log_file":        DefaultLogFile,
		"log_format":      DefaultLogFormat,
		"log_level":       DefaultLogLevel,
		"trakt.cache_dir": currentWorkingDir,
	}, "."), nil)

	// Flags
	if err = k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
		return nil, err
	}

	// YAML Config
	yamlConfig := k.String("config")
	if yamlConfig != "" {
		err = k.Load(file.Provider(yamlConfig), koanf_yaml.Parser())
		if err != nil {
			return nil, err
		}
	}

	// Environment variables
	err = k.Load(env.Provider(envVarPrefx, ".", func(s string) string {
		s = strings.TrimPrefix(s, envVarPrefx)
		s = strings.ToLower(s)
		s = strings.Replace(s, "__", ".", -1)
		return s
	}), nil)
	if err != nil {
		return nil, err
	}

	// Load flag overrides again to make sure they override everything
	if err = k.Load(posflag.Provider(flags, ".", k), nil); err != nil {
		return nil, err
	}

	var out Config
	err = k.Unmarshal("", &out)
	if err != nil {
		return nil, err
	}

	out.k = k
	return &out, nil
}
