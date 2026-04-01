package validator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/weprodev/go-pkg/httperr"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps go-playground/validator for use with Echo
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new CustomValidator instance with custom validation rules
func NewValidator() *CustomValidator {
	v := validator.New()

	// Register custom tag name function to use json tags for field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{validator: v}
}

// RegisterValidation allows defining and registering custom validations for specific tags
func (cv *CustomValidator) RegisterValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// Validate validates a struct and returns ValidationErrors wrapped with our custom error format
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var messages []string
			for _, fe := range validationErrors {
				messages = append(messages, FieldErrorToString(fe))
			}
			return httperr.NewServiceError(httperr.StatusUnprocessable, "validation error", messages...)
		}
		return httperr.NewServiceError(httperr.StatusBadRequest, "validation failed")
	}
	return nil
}

// ValidatableDTO interface for DTOs that can validate themselves
type ValidatableDTO interface {
	Validate() error
}

// StrictValidatableDTO interface for DTOs that can validate and unmarshal themselves with strict field checking
type StrictValidatableDTO interface {
	ValidateAndUnmarshal([]byte) error
}

// StrictUnmarshalAndValidate unmarshals JSON data with strict field validation (no unknown fields)
// and then validates the struct using the provided validator function
func StrictUnmarshalAndValidate[T any](data []byte, target *T, validatorFunc func() error) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			return httperr.NewServiceError(httperr.StatusBadRequest, "invalid JSON syntax", fmt.Sprintf("syntax error at offset %d", syntaxErr.Offset))
		}
		if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
			return httperr.NewServiceError(httperr.StatusBadRequest, "invalid field type", fmt.Sprintf("field '%s' expected %s but got %s", typeErr.Field, typeErr.Type, typeErr.Value))
		}
		// Check for unknown fields
		if strings.Contains(err.Error(), "unknown field") {
			return httperr.NewServiceError(httperr.StatusBadRequest, "unknown field in request", err.Error())
		}
		return httperr.NewServiceError(httperr.StatusBadRequest, "invalid JSON format", err.Error())
	}

	// Run validation if provided
	if validatorFunc != nil {
		return validatorFunc()
	}

	return nil
}

// StrictUnmarshalAndValidateWithValidator unmarshals JSON data with strict field validation
// and validates using the common CustomValidator
func StrictUnmarshalAndValidateWithValidator[T any](data []byte, target *T, validator *CustomValidator) error {
	return StrictUnmarshalAndValidate(data, target, func() error {
		return validator.Validate(*target)
	})
}

// StrictUnmarshalValidatable is a convenience function for DTOs that implement ValidatableDTO
func StrictUnmarshalValidatable[T ValidatableDTO](data []byte, target T) error {
	return StrictUnmarshalAndValidate(data, &target, target.Validate)
}
