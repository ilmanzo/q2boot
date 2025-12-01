package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger is a wrapper around slog for convenient logging operations
type Logger struct {
	*slog.Logger
}

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

var defaultLogger *Logger

// Initialize sets up the global logger with the specified level and format
func Initialize(level LogLevel, format string) error {
	var logLevel slog.Level
	switch level {
	case DebugLevel:
		logLevel = slog.LevelDebug
	case WarnLevel:
		logLevel = slog.LevelWarn
	case ErrorLevel:
		logLevel = slog.LevelError
	case InfoLevel:
		fallthrough
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stderr, opts)
	case "text":
		fallthrough
	default:
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	defaultLogger = &Logger{
		Logger: slog.New(handler),
	}
	return nil
}

// Get returns the global logger instance
func Get() *Logger {
	if defaultLogger == nil {
		// Initialize with default settings if not already done
		_ = Initialize(InfoLevel, "text")
	}
	return defaultLogger
}

// SetOutput sets the output writer for the logger
func SetOutput(w io.Writer) error {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(w, opts)
	defaultLogger = &Logger{
		Logger: slog.New(handler),
	}
	return nil
}

// Debug logs a debug message with attributes
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info logs an info message with attributes
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn logs a warning message with attributes
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error logs an error message with attributes
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// WithContext returns a new logger with the given context attributes
func (l *Logger) WithContext(args ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// Convenience functions that use the global logger

// Debug logs a debug message
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// WithContext returns a new logger with context
func WithContext(args ...any) *Logger {
	return Get().WithContext(args...)
}
