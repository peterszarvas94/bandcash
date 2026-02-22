package utils

import (
	"context"
	"maps"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
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

// MapErrors returns a map of field names to error messages.
func MapErrors(err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = e.Error()
		}
	}
	return errors
}

// MapErrorsLocalized returns a map of field names to localized error messages.
func MapErrorsLocalized(ctx context.Context, err error) map[string]string {
	errors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = validationMessage(ctx, e)
		}
	}
	return errors
}

// Validate validates a struct and returns errors if any
func Validate(s any) map[string]string {
	err := validate.Struct(s)
	if err != nil {
		return MapErrors(err)
	}
	return nil
}

// ValidateWithLocale validates a struct and returns localized errors if any.
func ValidateWithLocale(ctx context.Context, s any) map[string]string {
	err := validate.Struct(s)
	if err != nil {
		return MapErrorsLocalized(ctx, err)
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

func validationMessage(ctx context.Context, e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return ctxi18n.T(ctx, "validation.required")
	case "min":
		return ctxi18n.T(ctx, "validation.min", e.Param())
	case "max":
		return ctxi18n.T(ctx, "validation.max", e.Param())
	case "gt":
		return ctxi18n.T(ctx, "validation.gt", e.Param())
	case "gte":
		return ctxi18n.T(ctx, "validation.gte", e.Param())
	case "email":
		return ctxi18n.T(ctx, "validation.email")
	default:
		return ctxi18n.T(ctx, "validation.required")
	}
}
