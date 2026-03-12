# go-common

[![CI](https://github.com/wispberry-tech/go-common/actions/workflows/ci.yml/badge.svg)](https://github.com/wispberry-tech/go-common/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wispberry-tech/go-common)](https://goreportcard.com/report/github.com/wispberry-tech/go-common)
[![Go Reference](https://pkg.go.dev/badge/github.com/wispberry-tech/go-common.svg)](https://pkg.go.dev/github.com/wispberry-tech/go-common)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)


A simple little go library of utils I use in just about every go project from small to large.


## Installation

```bash
go get github.com/wispberry-tech/go-common
```

## Usage

```go
import "github.com/wispberry-tech/go-common/pkg/common"
```

### HTTP Utilities

```go
// Query parameter parsing
page := common.ParseQueryInt(r, "page", 1)
active := common.ParseQueryBool(r, "active", false)
timestamp := common.ParseQueryTime(r, "created_at")
name := common.ParseQueryStringPtr(r, "name") // returns *string

// JSON responses
common.WriteJSONResponse(w, http.StatusOK, data)
common.WriteJSONError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request", nil)

// Request body parsing
var req MyRequest
if err := common.ReadJSONBody(r, &req); err != nil {
    common.WriteJSONError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", nil)
    return
}
```

### Logging

```go
// Initialize at startup
common.InitializeLogger()
common.SetLogLevel("debug") // debug, info, warn, error

// Log messages
common.LogInfo("User logged in", "user_id", 123)
common.LogError("Database error", "error", err)
common.LogWarn("Rate limit approaching", "requests", 950)

// Formatted logging
common.LogInfof("Processing %d items", count)

// Context-aware logging
ctx := common.WithContext(context.Background(), logger)
logger := common.FromContext(ctx)
```

### Validation

```go
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8"`
}

if err := common.Validate.Struct(req); err != nil {
    response := common.FormatValidationErrors(err)
    common.WriteJSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", response.Error, response.Details)
    return
}
```

## Available Utilities

| Utility | Description |
|---------|-------------|
| HTTP Helpers | Query parsing, JSON request/response handling |
| Logging | Structured, colorized logging with charmbracelet/log |
| Validation | Request validation with go-playground/validator |

## Dependencies

- [github.com/charmbracelet/log](https://github.com/charmbracelet/log) - Beautiful logging
- [github.com/go-playground/validator/v10](https://github.com/go-playground/validator) - Struct validation

## License

GPL-3.0
