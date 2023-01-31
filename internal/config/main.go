package config

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/bjw-s/seriescleanup/internal/mediafile"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
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
	DevelopmentMode  bool   `mapstructure:"devMode"`
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
	// Initialize Viper
	viper.SetConfigName("settings.json")
	viper.SetConfigType("json")
	viper.AddConfigPath("/config")  // /config/settings.json
	viper.AddConfigPath("./config") // ./config/settings.json
	viper.AddConfigPath(".")        // ./settings.json
	viper.SetEnvPrefix("sc")        // SC_CONFIG_ITEM

	viper.SetDefault("devMode", false)
	viper.SetDefault("dryRun", false)
	viper.SetDefault("deleteAfterHours", 24)
	viper.SetDefault("folderRegex", "(?P<Show>.*)")
	viper.SetDefault("logLevel", "info")
	viper.SetDefault("trakt.CacheFolder", "/config")

	// Replace . with _ in Env names
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Read error %v", err)
	}
	if err := viper.Unmarshal(&Config); err != nil {
		log.Fatalf("unable to unmarshall the config %v", err)
	}
	validate := validator.New()
	if err := validate.Struct(&Config); err != nil {
		log.Fatalf("Configuration validation failed: %v\n", err)
	}
}
