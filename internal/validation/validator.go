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
