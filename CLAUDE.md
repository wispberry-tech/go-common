# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`go-common` is a shared Go utility library (`github.com/wispberry-tech/go-common`) providing HTTP helpers, structured logging, and validation for Go HTTP APIs. It is imported as a dependency by other projects — it is not a standalone application.

Package name is `common`. All source files are in the repo root (no `pkg/` or `internal/` subdirectories).

## Commands

```bash
go build ./...                # Build
go test -v -race ./...        # Run all tests with race detection
go test -v -run TestName ./...  # Run a single test
go vet ./...                  # Vet
```

There is no linter configured locally; CI runs `go vet` only.

## Architecture

Three files, one package:

- **common.go** — HTTP utilities: query parameter parsing (`ParseQueryInt`, `ParseQueryBool`, `ParseQueryTime`, `ParseQueryStringPtr`), JSON response helpers (`WriteJSONResponse`, `WriteJSONError`), and `ReadJSONBody`. All responses use the `ResponseEnvelope` wrapper (`{data, error, meta}`).
- **logging.go** — Wrappers around `charmbracelet/log`. `InitializeLogger()` sets up the global logger with custom lipgloss styles. Log functions (`LogInfo`, `LogError`, etc.) delegate to the global logger. Supports context-based logger propagation via `WithContext`/`FromContext`.
- **validate.go** — Global `Validate` instance (`go-playground/validator/v10`). `FormatValidationErrors` converts validator errors into `ValidationErrorResponse` structs with human-readable messages. Has a hardcoded field display name map in `getFieldDisplayName`.

## Key Dependencies

- `charmbracelet/log` + `charmbracelet/lipgloss` for styled logging
- `go-playground/validator/v10` for struct validation

## CI/Release

- CI runs on Go 1.23.x (matrix) on push/PR to `main`.
- Release workflow auto-bumps patch version and creates a GitHub release on every push to `main`. Can also be triggered manually with a specific version.
- Module requires Go 1.24.0 (`go.mod`), but CI tests on 1.23.x — be aware of potential version mismatch.
