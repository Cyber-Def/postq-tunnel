package logger

import (
	"log/slog"
	"os"
)

// InitLogger initializes the global slog instance with a JSON handler.
// Also configures the default log level and attaches a timestamp to every log.
func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// Fatal is a helper to log an error and exit the application, simulating log.Fatalf.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
