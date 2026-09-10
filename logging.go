package common

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/mattn/go-isatty"
)

// loggerConfig holds configuration for the logger.
type loggerConfig struct {
	level           log.Level
	timeFormat      string
	reportCaller    bool
	reportTimestamp bool
}

// LoggerOption configures the logger when passed to InitializeLogger.
type LoggerOption func(*loggerConfig)

// WithLevel sets the initial log level. Supports: "debug", "info", "warn", "error".
func WithLevel(level string) LoggerOption {
	return func(cfg *loggerConfig) {
		switch level {
		case "debug":
			cfg.level = log.DebugLevel
		case "info":
			cfg.level = log.InfoLevel
		case "warn":
			cfg.level = log.WarnLevel
		case "error":
			cfg.level = log.ErrorLevel
		}
	}
}

// WithTimeFormat sets the time format string for log timestamps.
func WithTimeFormat(format string) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.timeFormat = format
	}
}

// WithCaller enables or disables caller reporting in log output.
func WithCaller(enabled bool) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.reportCaller = enabled
	}
}

// WithTimestamp enables or disables timestamps in log output.
func WithTimestamp(enabled bool) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.reportTimestamp = enabled
	}
}

// InitializeLogger configures the global charmbracelet/log logger
// for beautiful, colorized output with proper formatting.
// This should be called once at application startup.
// Options can be passed to override defaults.
func InitializeLogger(opts ...LoggerOption) {
	cfg := loggerConfig{
		level:           log.InfoLevel,
		timeFormat:      "15:04:05",
		reportCaller:    true,
		reportTimestamp: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{
		TimeFormat:      cfg.timeFormat,
		ReportCaller:    cfg.reportCaller,
		ReportTimestamp: cfg.reportTimestamp,
		Level:           cfg.level,
	})

	logger.SetStyles(getLogStyles())

	switch os.Getenv("LOG_FORMAT") {
	case "json":
		logger.SetFormatter(log.JSONFormatter)
	case "text":
		// Keep the default colored text formatter.
	default:
		// Auto-detect: JSON when stderr is not a terminal (e.g. in
		// containers/CI), colored text otherwise.
		if !isatty.IsTerminal(os.Stderr.Fd()) {
			logger.SetFormatter(log.JSONFormatter)
		}
	}

	log.SetDefault(logger)
}

// SetLogLevel sets the global log level. Supports: debug, info, warn, error.
func SetLogLevel(level string) {
	var logLevel log.Level
	switch level {
	case "debug":
		logLevel = log.DebugLevel
	case "info":
		logLevel = log.InfoLevel
	case "warn":
		logLevel = log.WarnLevel
	case "error":
		logLevel = log.ErrorLevel
	default:
		logLevel = log.InfoLevel
	}

	log.SetLevel(logLevel)
}

// EnableDebugLogging enables debug level logging.
func EnableDebugLogging() {
	log.SetLevel(log.DebugLevel)
}

// DisableDebugLogging disables debug level logging (sets to info).
func DisableDebugLogging() {
	log.SetLevel(log.InfoLevel)
}

// getLogStyles returns custom styles for the logger with beautiful colors
func getLogStyles() *log.Styles {
	return &log.Styles{
		// Style for debug level (cyan)
		Levels: map[log.Level]lipgloss.Style{
			log.DebugLevel: lipgloss.NewStyle().
				SetString("DBG").
				Foreground(lipgloss.Color("36")).
				Bold(true),
			log.InfoLevel: lipgloss.NewStyle().
				SetString("INF").
				Foreground(lipgloss.Color("32")).
				Bold(true),
			log.WarnLevel: lipgloss.NewStyle().
				SetString("WRN").
				Foreground(lipgloss.Color("33")).
				Bold(true),
			log.ErrorLevel: lipgloss.NewStyle().
				SetString("ERR").
				Foreground(lipgloss.Color("31")).
				Bold(true),
		},
		// Style for keys (magenta)
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color("35")),
		// Style for values (white)
		Value: lipgloss.NewStyle().
			Foreground(lipgloss.Color("37")),
		// Style for timestamp (dark gray)
		Timestamp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		// Style for caller (dark gray)
		Caller: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		// Style for separator (dark gray)
		Separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")),
		// Style for message (white)
		Message: lipgloss.NewStyle().
			Foreground(lipgloss.Color("37")),
	}
}

// LogError logs an error message with optional key-value pairs.
func LogError(msg interface{}, keyvals ...interface{}) {
	log.Default().Helper()
	keyvals = append(keyvals, "error", msg)
	log.Error(msg, keyvals...)
}

// LogInfo logs an info message with optional key-value pairs.
func LogInfo(msg interface{}, keyvals ...interface{}) {
	log.Default().Helper()
	log.Info(msg, keyvals...)
}

// LogDebug logs a debug message with optional key-value pairs.
func LogDebug(msg interface{}, keyvals ...interface{}) {
	log.Default().Helper()
	log.Debug(msg, keyvals...)
}

// LogWarn logs a warning message with optional key-value pairs.
func LogWarn(msg interface{}, keyvals ...interface{}) {
	log.Default().Helper()
	log.Warn(msg, keyvals...)
}

// LogErrorf logs a formatted error message.
func LogErrorf(format string, args ...interface{}) {
	log.Default().Helper()
	log.Errorf(format, args...)
}

// LogInfof logs a formatted info message.
func LogInfof(format string, args ...interface{}) {
	log.Default().Helper()
	log.Infof(format, args...)
}

// LogDebugf logs a formatted debug message.
func LogDebugf(format string, args ...interface{}) {
	log.Default().Helper()
	log.Debugf(format, args...)
}

// LogWarnf logs a formatted warning message.
func LogWarnf(format string, args ...interface{}) {
	log.Default().Helper()
	log.Warnf(format, args...)
}

// FromContext returns a logger from the context.
func FromContext(ctx context.Context) *log.Logger {
	return log.FromContext(ctx)
}

// WithContext adds a logger to the context.
func WithContext(ctx context.Context, logger *log.Logger) context.Context {
	return log.WithContext(ctx, logger)
}

// Print logs a message at info level.
func Print(msg interface{}, keyvals ...interface{}) {
	log.Default().Helper()
	log.Print(msg, keyvals...)
}

// Printf logs a formatted message at info level.
func Printf(format string, args ...interface{}) {
	log.Default().Helper()
	log.Printf(format, args...)
}

// TrimSQL collapses runs of whitespace (tabs, newlines, spaces) into a single
// space and trims leading/trailing whitespace. When maxLen is greater than
// zero and the collapsed string is longer, it is truncated rune-safely (a
// UTF-8 codepoint is never split) and a trailing "…" is appended within the
// budget. A maxLen of zero or less collapses only, never truncating.
func TrimSQL(sql string, maxLen int) string {
	collapsed := strings.Join(strings.Fields(sql), " ")
	if maxLen <= 0 {
		return collapsed
	}

	runes := []rune(collapsed)
	if len(runes) <= maxLen {
		return collapsed
	}
	if maxLen == 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}
