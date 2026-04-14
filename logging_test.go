package common

import (
	"context"
	"testing"

	"github.com/charmbracelet/log"
)

func TestInitializeLogger(t *testing.T) {
	// Should not panic with no options.
	InitializeLogger()
}

func TestInitializeLoggerWithOptions(t *testing.T) {
	InitializeLogger(
		WithLevel("debug"),
		WithTimeFormat("2006-01-02"),
		WithCaller(false),
		WithTimestamp(false),
	)

	if log.Default().GetLevel() != log.DebugLevel {
		t.Errorf("level = %v, want DebugLevel", log.Default().GetLevel())
	}

	// Reset to defaults.
	InitializeLogger()
}

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		input string
		want  log.Level
	}{
		{"debug", log.DebugLevel},
		{"info", log.InfoLevel},
		{"warn", log.WarnLevel},
		{"error", log.ErrorLevel},
		{"unknown", log.InfoLevel},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			SetLogLevel(tt.input)
			if log.GetLevel() != tt.want {
				t.Errorf("SetLogLevel(%q): level = %v, want %v", tt.input, log.GetLevel(), tt.want)
			}
		})
	}
	// Reset.
	SetLogLevel("info")
}

func TestEnableDisableDebugLogging(t *testing.T) {
	EnableDebugLogging()
	if log.GetLevel() != log.DebugLevel {
		t.Error("EnableDebugLogging did not set debug level")
	}
	DisableDebugLogging()
	if log.GetLevel() != log.InfoLevel {
		t.Error("DisableDebugLogging did not set info level")
	}
}

func TestContextPropagation(t *testing.T) {
	InitializeLogger()

	logger := log.Default()
	ctx := WithContext(context.Background(), logger)
	got := FromContext(ctx)

	if got == nil {
		t.Fatal("FromContext returned nil")
	}
}
