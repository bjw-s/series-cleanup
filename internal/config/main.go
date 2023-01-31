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
	"github.com/bjw-s/series-cleanup/internal/mediafile"
	"github.com/go-playground/validator/v10"
	flag "github.com/spf13/pflag"
)

// Config exposes the collected configuration
var Config config

type sensitiveString string

func (s sensitiveString) String() string {
	return "[REDACTED]"
}
func (s sensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal("[REDACTED]")
}

type config struct {
	DeleteAfterHours int    `mapstructure:"deleteAfterHours"`
	DryRun           bool   `mapstructure:"dryRun"`
	FolderRegex      string `mapstructure:"folderRegex"`
	LogLevel         string `mapstructure:"logLevel"`
	Overrides        []struct {
		Folder  string
		Mapping mediafile.ShowMapping
		Skip    bool
	}
	ScanFolders []string `mapstructure:"scanFolders"`
	Trakt       struct {
		CacheFolder  string
		ClientID     string          `mapstructure:"clientId" validate:"required"`
		ClientSecret sensitiveString `mapstructure:"clientSecret" validate:"required"`
		User         string          `mapstructure:"user" validate:"required"`
	}
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
		"dryRun":            false,
		"deleteAfterHours":  24,
		"folderRegex":       "(?P<Show>.*)",
		"loglevel":          "info",
		"trakt.CacheFolder": configFolder,
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

	k.Unmarshal("", &Config)

	// Validate the rendered configuration
	validate := validator.New()

	if err := validate.Struct(&Config); err != nil {
		log.Fatalf("Configuration validation failed: %v\n", err)
	}
}
