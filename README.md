# go-common

[![CI](https://github.com/wispberry-tech/go-common/actions/workflows/ci.yml/badge.svg)](https://github.com/wispberry-tech/go-common/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wispberry-tech/go-common)](https://goreportcard.com/report/github.com/wispberry-tech/go-common)
[![Go Reference](https://pkg.go.dev/badge/github.com/wispberry-tech/go-common.svg)](https://pkg.go.dev/github.com/wispberry-tech/go-common)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

A shared Go utility library for building HTTP APIs. Provides query parameter parsing, JSON request/response handling, structured logging, and request validation.

## Installation

```bash
go get github.com/wispberry-tech/go-common
```

```go
import common "github.com/wispberry-tech/go-common"
```

## HTTP Utilities

### Query Parameter Parsing

Parse query parameters with type safety and default values. All parsers return the default when the key is missing or the value is invalid.

```go
// Integers
page := common.ParseQueryInt(r, "page", 1)           // ?page=3 → 3
id := common.ParseQueryInt64(r, "id", 0)              // ?id=9223372036854775807 → 9223372036854775807

// Floats
price := common.ParseQueryFloat64(r, "price", 0.0)    // ?price=19.99 → 19.99

// Booleans — accepts true/1/yes and false/0/no (case-insensitive)
active := common.ParseQueryBool(r, "active", false)    // ?active=yes → true

// Timestamps — expects RFC3339 format
since := common.ParseQueryTime(r, "since")             // ?since=2024-01-15T10:30:00Z → time.Time

// Optional strings — returns *string (nil when missing)
name := common.ParseQueryStringPtr(r, "name")          // ?name=alice → &"alice", missing → nil

// String slices — supports repeated keys
tags := common.ParseQueryStringSlice(r, "tag")         // ?tag=go&tag=rust → ["go", "rust"]
```

### Pagination

Parse `limit` and `offset` query parameters with clamping.

```go
// Reads ?limit=50&offset=10, clamps limit to [1, maxLimit], offset to >= 0
p := common.ParsePaginationParams(r, 20, 100) // defaultLimit=20, maxLimit=100
p.Limit  // 50
p.Offset // 10
```

### JSON Responses

All responses are wrapped in a `ResponseEnvelope` with `data`, `error`, and `meta` fields.

```go
// Success response — {"data": {"id": 1, "name": "alice"}}
common.WriteJSONResponse(w, http.StatusOK, user)

// Error response — {"error": {"code": "NOT_FOUND", "message": "User not found"}}
common.WriteJSONError(w, http.StatusNotFound, common.ErrCodeNotFound, "User not found", nil)

// Error with details — {"error": {"code": "VALIDATION_ERROR", "message": "...", "details": [...]}}
common.WriteJSONError(w, http.StatusBadRequest, common.ErrCodeValidationError, "Invalid input", details)
```

### Error Code Constants

Pre-defined error codes for consistent API responses:

| Constant | Value |
|----------|-------|
| `ErrCodeValidationError` | `"VALIDATION_ERROR"` |
| `ErrCodeInvalidJSON` | `"INVALID_JSON"` |
| `ErrCodeNotFound` | `"NOT_FOUND"` |
| `ErrCodeUnauthorized` | `"UNAUTHORIZED"` |
| `ErrCodeForbidden` | `"FORBIDDEN"` |
| `ErrCodeInternalError` | `"INTERNAL_ERROR"` |

### Reading Request Bodies

```go
var req CreateUserRequest
if err := common.ReadJSONBody(r, &req); err != nil {
    common.WriteJSONError(w, http.StatusBadRequest, common.ErrCodeInvalidJSON, "Invalid JSON body", nil)
    return
}
```

## Logging

Structured, colorized logging built on [charmbracelet/log](https://github.com/charmbracelet/log).

### Initialization

Call `InitializeLogger` once at startup. Options override defaults (info level, timestamps enabled, caller reporting enabled, `15:04:05` time format).

```go
// Use defaults
common.InitializeLogger()

// Or customize with options
common.InitializeLogger(
    common.WithLevel("debug"),
    common.WithTimeFormat("2006-01-02 15:04:05"),
    common.WithCaller(false),
    common.WithTimestamp(true),
)
```

### Log Messages

Structured logging with key-value pairs:

```go
common.LogInfo("User logged in", "user_id", 123, "ip", "10.0.0.1")
common.LogError("Database error", "error", err, "query", "SELECT ...")
common.LogWarn("Rate limit approaching", "requests", 950, "limit", 1000)
common.LogDebug("Cache hit", "key", "user:123")
```

Formatted logging:

```go
common.LogInfof("Processing %d items", count)
common.LogErrorf("Failed to connect to %s: %v", host, err)
common.LogWarnf("Retrying in %s", backoff)
common.LogDebugf("Query took %dms", elapsed)
```

### Change Log Level at Runtime

```go
common.SetLogLevel("debug") // "debug", "info", "warn", "error"

// Convenience toggles
common.EnableDebugLogging()
common.DisableDebugLogging()
```

### Context-Aware Logging

Attach a logger to a context for request-scoped logging:

```go
// Store logger in context
ctx := common.WithContext(r.Context(), logger)

// Retrieve later in the call chain
logger := common.FromContext(ctx)
```

## Validation

Request validation using [go-playground/validator](https://github.com/go-playground/validator), with automatic conversion of validation errors to API-friendly responses.

### Basic Usage

Use the global `Validate` instance to validate structs with `validate` tags:

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8"`
}

if err := common.Validate.Struct(req); err != nil {
    resp := common.FormatValidationErrors(err)
    common.WriteJSONError(w, http.StatusBadRequest, common.ErrCodeValidationError, resp.Error, resp.Details)
    return
}
```

This produces a response like:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {"field": "email", "message": "email must be a valid email address", "value": "notanemail"},
      {"field": "password", "message": "password must be at least 8 characters long", "value": "short"}
    ]
  }
}
```

### Custom Display Names with Struct Tags

Use the `display` struct tag to control how field names appear in error messages. Pass the struct to `FormatValidationErrorsFor` to enable tag lookup:

```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email" display:"Email Address"`
    Age   int    `json:"age" validate:"required,gte=18" display:"Your Age"`
}

if err := common.Validate.Struct(req); err != nil {
    resp := common.FormatValidationErrorsFor(err, req) // pass the struct for display tag lookup
    // → "Email Address must be a valid email address"
    // → "Your Age must be greater than or equal to 18"
}
```

Without `display` tags, field names are derived by splitting camelCase (e.g., `ClientUUID` becomes `Client UUID`).

### Supported Validation Tags

`FormatValidationErrors` generates human-readable messages for these validation tags:

| Tag | Message |
|-----|---------|
| `required` | "{field} is required" |
| `email` | "{field} must be a valid email address" |
| `url` | "{field} must be a valid URL" |
| `uuid` | "{field} must be a valid UUID" |
| `min` | "{field} must be at least {param} characters long" |
| `max` | "{field} must be no more than {param} characters long" |
| `len` | "{field} must be exactly {param} characters long" |
| `gt` | "{field} must be greater than {param}" |
| `gte` | "{field} must be greater than or equal to {param}" |
| `lt` | "{field} must be less than {param}" |
| `lte` | "{field} must be less than or equal to {param}" |
| `oneof` | "{field} must be one of: {param}" |
| `numeric` | "{field} must be numeric" |
| `alpha` | "{field} must contain only letters" |
| `alphanum` | "{field} must contain only letters and numbers" |
| `ip` | "{field} must be a valid IP address" |
| `ipv4` | "{field} must be a valid IPv4 address" |
| `ipv6` | "{field} must be a valid IPv6 address" |
| *(other)* | "{field} is invalid" |

## Dependencies

- [charmbracelet/log](https://github.com/charmbracelet/log) — Structured, colorized logging
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) — Log styling
- [go-playground/validator/v10](https://github.com/go-playground/validator) — Struct validation

## License

GPL-3.0
