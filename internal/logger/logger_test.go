package logger

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		format    string
		shouldErr bool
	}{
		{"debug text", DebugLevel, "text", false},
		{"info json", InfoLevel, "json", false},
		{"warn text", WarnLevel, "text", false},
		{"error text", ErrorLevel, "text", false},
		{"invalid format defaults to text", InfoLevel, "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level, tt.format)
			if (err != nil) != tt.shouldErr {
				t.Errorf("Initialize() error = %v, shouldErr %v", err, tt.shouldErr)
			}
			if defaultLogger == nil {
				t.Error("Initialize() should set defaultLogger")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Create a buffer to capture output
	buf := &bytes.Buffer{}

	// Create a logger that writes to our buffer
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(buf, opts)
	testLogger := &Logger{
		Logger: slog.New(handler),
	}

	// Test different log levels
	testLogger.Debug("debug message", "key", "value")
	testLogger.Info("info message", "key", "value")
	testLogger.Warn("warn message", "key", "value")
	testLogger.Error("error message", "key", "value")

	output := buf.String()
	if len(output) == 0 {
		t.Error("Logger should produce output")
	}
}

func TestGlobalLogger(t *testing.T) {
	// Test that Get() returns a logger
	logger := Get()
	if logger == nil {
		t.Error("Get() should return a non-nil logger")
	}

	// Test convenience functions
	Debug("test debug")
	Info("test info")
	Warn("test warn")
	Error("test error")
}

func TestWithContext(t *testing.T) {
	Initialize(InfoLevel, "text")

	contextLogger := WithContext("request_id", "123", "user", "alice")
	if contextLogger == nil {
		t.Error("WithContext() should return a non-nil logger")
	}

	// Use context logger (should not panic)
	contextLogger.Info("test with context")
}

func TestSetOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	err := SetOutput(buf)
	if err != nil {
		t.Errorf("SetOutput() error = %v", err)
	}

	Info("test message")
	if buf.Len() == 0 {
		t.Error("SetOutput() should direct logs to the new writer")
	}
}
