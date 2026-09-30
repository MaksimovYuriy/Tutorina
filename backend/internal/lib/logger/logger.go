package logger

import (
	"log/slog"
	"os"
)

func New(environment string) *slog.Logger {
	options := &slog.HandlerOptions{Level: slog.LevelInfo}
	if environment == "dev" {
		options.Level = slog.LevelDebug
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}
	if environment == "test" {
		options.Level = slog.LevelWarn
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}
