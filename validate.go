package common

import (
	"fmt"
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
// Returns a ValidationErrorResponse with structured error details suitable for API responses.
func FormatValidationErrors(err error) ValidationErrorResponse {
	var validationErrors []ValidationError

	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validatorErrs {
			validationError := ValidationError{
				Field: strings.ToLower(fieldError.Field()),
				Value: fmt.Sprintf("%v", fieldError.Value()),
			}

			switch fieldError.Tag() {
			case "required":
				validationError.Message = fmt.Sprintf("%s is required", getFieldDisplayName(fieldError.Field()))
			case "uuid":
				validationError.Message = fmt.Sprintf("%s must be a valid UUID", getFieldDisplayName(fieldError.Field()))
			case "min":
				validationError.Message = fmt.Sprintf("%s must be at least %s characters long", getFieldDisplayName(fieldError.Field()), fieldError.Param())
			case "max":
				validationError.Message = fmt.Sprintf("%s must be no more than %s characters long", getFieldDisplayName(fieldError.Field()), fieldError.Param())
			case "email":
				validationError.Message = fmt.Sprintf("%s must be a valid email address", getFieldDisplayName(fieldError.Field()))
			default:
				validationError.Message = fmt.Sprintf("%s is invalid", getFieldDisplayName(fieldError.Field()))
			}

			validationErrors = append(validationErrors, validationError)
		}
	}

	return ValidationErrorResponse{
		Error:   "Validation failed",
		Details: validationErrors,
	}
}

// getFieldDisplayName converts field names to user-friendly display names.
func getFieldDisplayName(field string) string {

	// Example mapping for specific fields, can be expanded as needed
	fieldNames := map[string]string{
		"ClientUUID": "Client",
		"Title":      "Title",
		"Content":    "Content",
		"Position":   "Position",
		"Props":      "Properties",
	}

	if displayName, exists := fieldNames[field]; exists {
		return displayName
	}

	// Convert camelCase to readable format as fallback
	result := ""
	for i, char := range field {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result += " "
		}
		result += string(char)
	}
	return result
}
