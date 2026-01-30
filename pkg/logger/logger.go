package logger

import (
	"log/slog"
	"os"
)

// Setup configures the default global logger
func Setup() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Use TextHandler by default for human-readable output in console
	// Could switch to JSONHandler for production
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	slog.SetDefault(logger)
}

// Helper wrappers if needed, or just generally expose slog
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}
