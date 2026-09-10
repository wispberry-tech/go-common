package common

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// captureDefaultOutput points the default logger at a buffer and restores a
// fresh default logger (writing to os.Stderr with default settings) when the
// test finishes.
func captureDefaultOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	InitializeLogger()
	buf := new(bytes.Buffer)
	log.Default().SetOutput(buf)
	t.Cleanup(func() { InitializeLogger() })
	return buf
}

func TestCallerReporting(t *testing.T) {
	buf := captureDefaultOutput(t)

	LogInfo("hello")

	out := buf.String()
	if !strings.Contains(out, "logging_test.go") {
		t.Errorf("caller output = %q, want it to reference the test file", out)
	}
	if strings.Contains(out, "logging.go:") {
		t.Errorf("caller output = %q, want it to skip the wrapper frame in logging.go", out)
	}
}

func TestLogFormatJSON(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	buf := captureDefaultOutput(t)

	LogInfo("hello")

	if out := buf.String(); !strings.HasPrefix(out, "{") {
		t.Errorf("output = %q, want JSON object", out)
	}
}

func TestTrimSQL(t *testing.T) {
	tests := []struct {
		name   string
		sql    string
		maxLen int
		want   string
	}{
		{"collapse whitespace", "SELECT *\nFROM users\tWHERE  id = 1", 0, "SELECT * FROM users WHERE id = 1"},
		{"collapse and trim edges", "  SELECT  1  \n\t", 0, "SELECT 1"},
		{"no truncation when fits", "SELECT id FROM users", 100, "SELECT id FROM users"},
		{"truncate at boundary", "abcdefghij", 6, "abcde…"},
		{"exact fit stays whole", "abcdefghij", 10, "abcdefghij"},
		{"maxLen one keeps ellipsis only", "abcdefghij", 1, "…"},
		{"negative maxLen collapses only", "SELECT *\nFROM users", -1, "SELECT * FROM users"},
		{"multibyte runes not split", "日本語テスト", 3, "日本…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimSQL(tt.sql, tt.maxLen); got != tt.want {
				t.Errorf("TrimSQL(%q, %d) = %q, want %q", tt.sql, tt.maxLen, got, tt.want)
			}
		})
	}
}

// decodeLogLine parses the single JSON log line written to the buffer.
func decodeLogLine(t *testing.T, out string) map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("log output = %q, want exactly one line", out)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("log line %q is not valid JSON: %v", lines[0], err)
	}
	return got
}

func TestRequestLogger(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")

	t.Run("honors incoming X-Request-ID", func(t *testing.T) {
		buf := captureDefaultOutput(t)

		handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := RequestID(r.Context()); got != "client-42" {
				t.Errorf("RequestID(ctx) = %q, want %q", got, "client-42")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("hello"))
		}))
		req := httptest.NewRequest(http.MethodGet, "/users?page=2", nil)
		req.Header.Set("X-Request-ID", "client-42")
		handler.ServeHTTP(httptest.NewRecorder(), req)

		got := decodeLogLine(t, buf.String())
		if got["req_id"] != "client-42" {
			t.Errorf(`log req_id = %v, want "client-42"`, got["req_id"])
		}
		if got["method"] != http.MethodGet {
			t.Errorf("log method = %v, want GET", got["method"])
		}
		if got["path"] != "/users" {
			t.Errorf("log path = %v, want /users", got["path"])
		}
		if got["status"] != float64(http.StatusCreated) {
			t.Errorf("log status = %v, want 201", got["status"])
		}
		if got["bytes"] != float64(len("hello")) {
			t.Errorf("log bytes = %v, want 5", got["bytes"])
		}
		if _, ok := got["dur_ms"].(float64); !ok {
			t.Errorf("log dur_ms = %v, want a duration in ms", got["dur_ms"])
		}
	})

	t.Run("generates request ID when header missing", func(t *testing.T) {
		buf := captureDefaultOutput(t)

		var ctxID string
		handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxID = RequestID(r.Context())
		}))
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))

		if len(ctxID) != 16 {
			t.Errorf("generated req_id = %q, want 16 hex chars", ctxID)
		}
		got := decodeLogLine(t, buf.String())
		if got["req_id"] != ctxID {
			t.Errorf("log req_id = %v, context req_id = %q, want equal", got["req_id"], ctxID)
		}
		if got["status"] != float64(http.StatusOK) {
			t.Errorf("log status = %v, want default 200", got["status"])
		}
		if got["bytes"] != float64(0) {
			t.Errorf("log bytes = %v, want 0", got["bytes"])
		}
	})
}

func TestRequestLoggerFlushPassthrough(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	buf := captureDefaultOutput(t)

	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("chunk"))
		w.(http.Flusher).Flush()
	}))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stream", nil))

	if !rec.Flushed {
		t.Error("Flush did not reach the underlying ResponseWriter")
	}
	if _, ok := decodeLogLine(t, buf.String())["bytes"]; !ok {
		t.Error("log line missing bytes field")
	}
}
