package common

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate is the global validator instance with required struct validation enabled.
var Validate = validator.New(validator.WithRequiredStructEnabled())

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents a structured response for validation errors.
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details"`
}

// FormatValidationErrors converts validator.ValidationErrors to user-friendly messages.
// Field display names are derived by splitting camelCase field names into words.
func FormatValidationErrors(err error) ValidationErrorResponse {
	return formatValidationErrors(err, nil)
}

// FormatValidationErrorsFor converts validator.ValidationErrors to user-friendly messages,
// using the "display" struct tag from v for field display names when available.
//
//	type Request struct {
//	    Email string `validate:"required,email" display:"Email Address"`
//	}
func FormatValidationErrorsFor(err error, v any) ValidationErrorResponse {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return formatValidationErrors(err, t)
}

func formatValidationErrors(err error, structType reflect.Type) ValidationErrorResponse {
	var validationErrors []ValidationError

	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrs {
			displayName := resolveDisplayName(fieldError, structType)
			validationError := ValidationError{
				Field: strings.ToLower(fieldError.Field()),
				Value: fmt.Sprintf("%v", fieldError.Value()),
			}

			switch fieldError.Tag() {
			case "required":
				validationError.Message = fmt.Sprintf("%s is required", displayName)
			case "uuid":
				validationError.Message = fmt.Sprintf("%s must be a valid UUID", displayName)
			case "min":
				validationError.Message = fmt.Sprintf("%s must be at least %s characters long", displayName, fieldError.Param())
			case "max":
				validationError.Message = fmt.Sprintf("%s must be no more than %s characters long", displayName, fieldError.Param())
			case "len":
				validationError.Message = fmt.Sprintf("%s must be exactly %s characters long", displayName, fieldError.Param())
			case "email":
				validationError.Message = fmt.Sprintf("%s must be a valid email address", displayName)
			case "url":
				validationError.Message = fmt.Sprintf("%s must be a valid URL", displayName)
			case "numeric":
				validationError.Message = fmt.Sprintf("%s must be numeric", displayName)
			case "alpha":
				validationError.Message = fmt.Sprintf("%s must contain only letters", displayName)
			case "alphanum":
				validationError.Message = fmt.Sprintf("%s must contain only letters and numbers", displayName)
			case "gt":
				validationError.Message = fmt.Sprintf("%s must be greater than %s", displayName, fieldError.Param())
			case "gte":
				validationError.Message = fmt.Sprintf("%s must be greater than or equal to %s", displayName, fieldError.Param())
			case "lt":
				validationError.Message = fmt.Sprintf("%s must be less than %s", displayName, fieldError.Param())
			case "lte":
				validationError.Message = fmt.Sprintf("%s must be less than or equal to %s", displayName, fieldError.Param())
			case "oneof":
				validationError.Message = fmt.Sprintf("%s must be one of: %s", displayName, fieldError.Param())
			case "ip":
				validationError.Message = fmt.Sprintf("%s must be a valid IP address", displayName)
			case "ipv4":
				validationError.Message = fmt.Sprintf("%s must be a valid IPv4 address", displayName)
			case "ipv6":
				validationError.Message = fmt.Sprintf("%s must be a valid IPv6 address", displayName)
			default:
				validationError.Message = fmt.Sprintf("%s is invalid", displayName)
			}

			validationErrors = append(validationErrors, validationError)
		}
	}

	return ValidationErrorResponse{
		Error:   "Validation failed",
		Details: validationErrors,
	}
}

// resolveDisplayName returns a user-friendly name for the field.
// If structType is provided and the field has a "display" tag, that value is used.
// Otherwise falls back to splitting camelCase into words.
func resolveDisplayName(fe validator.FieldError, structType reflect.Type) string {
	if structType != nil {
		if field, ok := structType.FieldByName(fe.StructField()); ok {
			if display := field.Tag.Get("display"); display != "" {
				return display
			}
		}
	}
	return camelCaseToWords(fe.Field())
}

// camelCaseToWords converts a camelCase or PascalCase string to space-separated words.
// Consecutive uppercase letters are kept together as acronyms (e.g., "ClientUUID" → "Client UUID").
func camelCaseToWords(s string) string {
	runes := []rune(s)
	var result strings.Builder
	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			// Insert space before an uppercase letter when:
			// - previous char was lowercase (e.g., "client|U")
			// - previous char was uppercase AND next char is lowercase (e.g., "UUI|D" before lowercase → "UUID |Next")
			if prev >= 'a' && prev <= 'z' {
				result.WriteRune(' ')
			} else if prev >= 'A' && prev <= 'Z' && i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
				result.WriteRune(' ')
			}
		}
		result.WriteRune(r)
	}
	return result.String()
}

func init() {
	// Register a tag name function so validator uses the "json" tag for field names in errors.
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}
