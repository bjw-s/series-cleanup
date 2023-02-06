// Package config implements all configuration aspects of series-cleanup
package config

import (
	"encoding/json"
	"log"
	"path"
	"strings"

	"github.com/knadh/koanf"
	koanf_json "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"

	"github.com/bjw-s/series-cleanup/internal/helpers"
	"github.com/go-playground/validator/v10"
	flag "github.com/spf13/pflag"
)

// AppConfig exposes the collected configuration
var AppConfig Config

type sensitiveString string

func (s sensitiveString) String() string {
	return "[REDACTED]"
}
func (s sensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

type folderSettingsIdentifiers struct {
	Trakt int    `mapstructure:"trakt"`
	Slug  string `mapstructure:"slug"`
	IMDB  string `mapstructure:"imdb"`
	TVDB  int    `mapstructure:"tvdb"`
	TMDB  int    `mapstructure:"tmdb"`
}

// FolderSettings represents settings specific to this folder
type FolderSettings struct {
	Folder      string
	Identifiers folderSettingsIdentifiers
	KeepShow    bool  `validate:"excluded_with_all=KeepSeasons"`
	KeepSeasons []int `validate:"excluded_with_all=KeepShow"`
}

type traktConfig struct {
	Enabled      bool
	CacheFolder  string
	ClientID     string          `validate:"required_if=Enabled true"`
	ClientSecret sensitiveString `validate:"required_if=Enabled true"`
	User         string          `validate:"required_if=Enabled true"`
}

type plexConfig struct {
	Enabled bool
	Token   sensitiveString `validate:"required_if=Enabled true"`
}

// Config represents the application configuration structure
type Config struct {
	DeleteAfterHours int
	DryRun           bool
	FolderRegex      string
	LogLevel         string
	FolderSettings   []FolderSettings `validate:"dive"`
	ScanFolders      []string         `validate:"dive"`
	Sources          struct {
		Trakt traktConfig
		Plex  plexConfig
	} `validate:"dive"`
}

func init() {
	var k = koanf.New(".")

	// Use the POSIX compliant pflag lib instead of Go's flag lib.
	var configFolder = flag.String("configFolder", "/config", "path to store the configuration")
	flag.Parse()

	// Check pre-requisites
	if !helpers.FolderExists(*configFolder) {
		log.Fatalf("Could not find configuration folder: %s", *configFolder)
	}

	// Load default values using the confmap provider.
	// We provide a flat map with the "." delimiter.
	k.Load(confmap.Provider(map[string]interface{}{
		"dryRun":                    false,
		"deleteAfterHours":          24,
		"folderRegex":               "(?P<Show>.*)",
		"loglevel":                  "info",
		"sources.trakt.CacheFolder": configFolder,
	}, "."), nil)

	// Load provided JSON config
	if err := k.Load(file.Provider(path.Join(*configFolder, "settings.json")), koanf_json.Parser()); err != nil {
		log.Fatalf("Error loading file: %v", err)
	}

	// Load environment variables and merge into the loaded config.
	k.Load(env.ProviderWithValue("SC_", ".", func(s string, v string) (string, interface{}) {
		// Strip out the SC_ prefix and lowercase and get the key while also replacing
		// the _ character with . in the key (koanf delimeter).
		key := strings.Replace(strings.ToLower(strings.TrimPrefix(s, "SC_")), "_", ".", -1)

		// If there is a space in the value, split the value into a slice by the space.
		if strings.Contains(v, " ") {
			return key, strings.Split(v, " ")
		}

		// Otherwise, return the plain string.
		return key, v
	}), nil)

	k.Unmarshal("", &AppConfig)

	// Validate the rendered configuration
	validate := validator.New()

	if err := validate.Struct(AppConfig); err != nil {
		log.Fatalf("Configuration validation failed: %v\n", err)
	}
}
