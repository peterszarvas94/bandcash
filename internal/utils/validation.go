package utils

import (
	"maps"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}
		if name != "" {
			return name
		}
		return field.Name
	})
	return v
}()

// MapErrora returns a map of field names to error messages.
func MapErrora(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = e.Error()
		}
	}
	return errors
}

// Validate validates a struct and returns errors if any
func Validate(s any) map[string]string {
	err := validate.Struct(s)
	if err != nil {
		return MapErrora(err)
	}
	return nil
}

// GetEmptyErrors creates an error map with all fields set to empty string
func GetEmptyErrors(fields []string) map[string]string {
	errs := make(map[string]string, len(fields))
	for _, f := range fields {
		errs[f] = ""
	}
	return errs
}

// WithErrors creates a complete error map with empty fields populated with actual errors
func WithErrors(fields []string, actualErrors map[string]string) map[string]string {
	errs := GetEmptyErrors(fields)
	maps.Copy(errs, actualErrors)
	return errs
}
