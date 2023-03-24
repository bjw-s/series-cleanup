// Package config implements all configuration aspects of series-cleanup
package config

import (
	"github.com/bjw-s/series-cleanup/internal/helpers"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"
)

var validate *validator.Validate

// Validate returns if the given configuration is valid and any validation errors
func (c *Config) Validate() error {
	validate = validator.New()
	validate.RegisterValidation("valid_folder", validateFolder)
	validate.RegisterValidation("valid_loglevel", validateLogLevel)
	err := validate.Struct(c)
	if err != nil {
		return err
	}
	return nil
}

func validateLogLevel(fl validator.FieldLevel) bool {
	validLogLevels := []string{}
	for i := zapcore.DebugLevel; i < zapcore.InvalidLevel; i++ {
		validLogLevels = append(validLogLevels, i.String())
	}
	return slices.Contains(validLogLevels, fl.Field().String())
}

func validateFolder(fl validator.FieldLevel) bool {
	return helpers.FolderExists(fl.Field().String())
}
