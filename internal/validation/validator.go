package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Errors returns a map of field names to error messages (lowercase keys)
func Errors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[strings.ToLower(e.Field())] = e.Error()
		}
	}
	return errors
}

// ValidateStruct validates a struct and returns errors if any
func ValidateStruct(s interface{}) map[string]string {
	if err := validate.Struct(s); err != nil {
		return Errors(err)
	}
	return nil
}

// ValidateAndPatchSignals validates and returns signals with errors patched
func ValidateAndPatchSignals(s interface{}) (map[string]any, bool) {
	if errs := ValidateStruct(s); errs != nil {
		return map[string]any{"errors": errs}, false
	}
	return map[string]any{"errors": map[string]string{}}, true
}
