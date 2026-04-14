// Package common provides reusable utilities for building HTTP APIs in Go.
// It includes helpers for query parameter parsing, JSON request/response handling,
// structured logging, and request validation.
//
// The package is designed to reduce boilerplate in HTTP handlers and provide
// consistent error handling and logging across services.
package common

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Common error code constants for use with WriteJSONError.
const (
	ErrCodeValidationError = "VALIDATION_ERROR"
	ErrCodeInvalidJSON     = "INVALID_JSON"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeInternalError   = "INTERNAL_ERROR"
)

// ResponseEnvelope wraps API responses in a consistent JSON structure.
// All responses include an optional Data, Error, and Meta field.
type ResponseEnvelope struct {
	Data  any        `json:"data,omitempty"`
	Error *ErrorBody `json:"error,omitempty"`
	Meta  any        `json:"meta,omitempty"`
}

// ErrorBody represents a structured error response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// ParseQueryInt extracts an integer from query parameters.
// Returns defaultValue if the key is missing or invalid.
func ParseQueryInt(r *http.Request, key string, defaultValue int) int {
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// ParseQueryTime extracts a time.Time from query parameters (RFC3339 format).
// Returns time.Time{} if the key is missing or invalid.
func ParseQueryTime(r *http.Request, key string) time.Time {
	value := r.FormValue(key)
	if value == "" {
		return time.Time{}
	}
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsedTime
}

// ParseQueryStringPtr extracts a string pointer from query parameters.
// Returns nil if the key is missing or empty.
func ParseQueryStringPtr(r *http.Request, key string) *string {
	value := r.FormValue(key)
	if value == "" {
		return nil
	}
	return &value
}

// WriteJSONResponse writes a JSON response with the given status code and data.
// The response is wrapped in a ResponseEnvelope.
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	envelope := ResponseEnvelope{Data: data}
	if data == nil {
		envelope.Data = map[string]any{}
	}
	writeEnvelope(w, statusCode, &envelope)
}

// WriteJSONError writes a JSON error response with the given status code,
// error code, message, and optional details.
func WriteJSONError(w http.ResponseWriter, statusCode int, code, message string, details any) {
	envelope := ResponseEnvelope{
		Error: &ErrorBody{Code: code, Message: message, Details: details},
	}
	writeEnvelope(w, statusCode, &envelope)
}

func writeEnvelope(w http.ResponseWriter, statusCode int, envelope *ResponseEnvelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if envelope == nil {
		envelope = &ResponseEnvelope{Data: map[string]any{}}
	}
	if envelope.Data == nil && envelope.Error == nil {
		envelope.Data = map[string]any{}
	}
	if err := json.NewEncoder(w).Encode(envelope); err != nil {
		msg := "Failed to encode response"
		slog.Error(msg, "error", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

// ReadJSONBody decodes the JSON request body into the provided value.
func ReadJSONBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ParseQueryBool extracts a boolean from query parameters.
// Accepts "true", "1", "yes" as true; "false", "0", "no" as false.
// Returns defaultValue if the key is missing or unrecognized.
func ParseQueryBool(r *http.Request, key string, defaultValue bool) bool {
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(value) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return defaultValue
	}
}

// ParseQueryFloat64 extracts a float64 from query parameters.
// Returns defaultValue if the key is missing or invalid.
func ParseQueryFloat64(r *http.Request, key string, defaultValue float64) float64 {
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return f
}

// ParseQueryInt64 extracts an int64 from query parameters.
// Returns defaultValue if the key is missing or invalid.
func ParseQueryInt64(r *http.Request, key string, defaultValue int64) int64 {
	value := r.FormValue(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i
}

// ParseQueryStringSlice extracts a string slice from query parameters.
// Supports repeated keys (e.g., ?tag=a&tag=b). Returns nil if the key is missing.
func ParseQueryStringSlice(r *http.Request, key string) []string {
	if err := r.ParseForm(); err != nil {
		return nil
	}
	values, ok := r.Form[key]
	if !ok {
		return nil
	}
	return values
}

// PaginationParams holds parsed pagination query parameters.
type PaginationParams struct {
	Limit  int
	Offset int
}

// ParsePaginationParams extracts "limit" and "offset" from query parameters.
// Limit is clamped to [1, maxLimit]. Defaults: limit=defaultLimit, offset=0.
func ParsePaginationParams(r *http.Request, defaultLimit, maxLimit int) PaginationParams {
	limit := ParseQueryInt(r, "limit", defaultLimit)
	offset := ParseQueryInt(r, "offset", 0)

	if limit < 1 {
		limit = 1
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}

	return PaginationParams{Limit: limit, Offset: offset}
}
