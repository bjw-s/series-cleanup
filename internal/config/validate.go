// Package config implements all configuration aspects of series-cleanup
package config

import (
	"errors"
	"fmt"

	"github.com/bjw-s/series-cleanup/internal/helpers"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"
)

var validate *validator.Validate

// Validate returns if the given configuration is valid and any validation errors
func (c *Config) Validate() []error {
	validate = validator.New()
	validate.RegisterValidation("valid_folder", validateFolder)
	validate.RegisterValidation("valid_loglevel", validateLogLevel)
	err := validate.Struct(c)
	if err != nil {
		var ve validator.ValidationErrors
		var output []error

		if errors.As(err, &ve) {
			for _, fe := range ve {
				output = append(output, errorMessageForFieldError(fe))
			}
		} else {
			output = append(output, err)
		}

		return output
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

func errorMessageForFieldError(fe validator.FieldError) error {
	switch fe.Tag() {
	case "required":
		return fmt.Errorf("%s: This field is required", fe.Namespace())
	case "email":
		return fmt.Errorf("%s: Invalid email address %s", fe.Namespace(), fe.Value())
	}
	return fmt.Errorf("%s: Field validation failed on the '%s' tag", fe.Namespace(), fe.Tag())
}
